package slowcontext

import (
	"errors"
	"fmt"
	"time"
)

type SourceHealthState string

const (
	SourceHealthHealthy            SourceHealthState = "healthy"
	SourceHealthDelayedPublication SourceHealthState = "delayed_publication"
	SourceHealthSourceUnavailable  SourceHealthState = "source_unavailable"
	SourceHealthParseFailed        SourceHealthState = "parse_failed"
)

type ScheduleConfig struct {
	SourceFamily             SourceFamily
	PollCadenceInWindow      time.Duration
	PollCadenceOutsideWindow time.Duration
	PublishWindowStartMinute int
	PublishWindowEndMinute   int
}

func DefaultScheduleConfig(sourceFamily SourceFamily) (ScheduleConfig, error) {
	if err := validateSourceFamily(sourceFamily); err != nil {
		return ScheduleConfig{}, err
	}
	config := ScheduleConfig{
		SourceFamily:             sourceFamily,
		PollCadenceInWindow:      15 * time.Minute,
		PollCadenceOutsideWindow: time.Hour,
	}
	switch sourceFamily {
	case SourceFamilyCME:
		config.PublishWindowStartMinute = 20 * 60
		config.PublishWindowEndMinute = 21 * 60
	case SourceFamilyETF:
		config.PublishWindowStartMinute = 22 * 60
		config.PublishWindowEndMinute = 23 * 60
	}
	return config, config.Validate()
}

func (c ScheduleConfig) Validate() error {
	if err := validateSourceFamily(c.SourceFamily); err != nil {
		return err
	}
	if c.PollCadenceInWindow <= 0 {
		return fmt.Errorf("poll cadence in window must be positive")
	}
	if c.PollCadenceOutsideWindow <= 0 {
		return fmt.Errorf("poll cadence outside window must be positive")
	}
	if c.PublishWindowStartMinute < 0 || c.PublishWindowStartMinute >= 24*60 {
		return fmt.Errorf("publish window start minute must be between 0 and 1439")
	}
	if c.PublishWindowEndMinute <= 0 || c.PublishWindowEndMinute > 24*60 {
		return fmt.Errorf("publish window end minute must be between 1 and 1440")
	}
	if c.PublishWindowEndMinute <= c.PublishWindowStartMinute {
		return fmt.Errorf("publish window end minute must be greater than start minute")
	}
	return nil
}

func (c ScheduleConfig) PollCadenceAt(now time.Time) (time.Duration, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}
	if now.IsZero() {
		return 0, fmt.Errorf("current time is required")
	}
	if c.IsInPublishWindow(now) {
		return c.PollCadenceInWindow, nil
	}
	return c.PollCadenceOutsideWindow, nil
}

func (c ScheduleConfig) IsInPublishWindow(now time.Time) bool {
	minuteOfDay := now.UTC().Hour()*60 + now.UTC().Minute()
	return minuteOfDay >= c.PublishWindowStartMinute && minuteOfDay < c.PublishWindowEndMinute
}

func (c ScheduleConfig) IsDelayed(now time.Time) bool {
	minuteOfDay := now.UTC().Hour()*60 + now.UTC().Minute()
	return minuteOfDay >= c.PublishWindowEndMinute
}

type PollDecision struct {
	Accepted   bool
	Correction bool
	Duplicate  bool
}

type SourceStatus struct {
	SourceFamily                SourceFamily
	MetricFamily                MetricFamily
	SourceKey                   string
	Asset                       string
	HealthState                 SourceHealthState
	LastSuccessfulPollAt        time.Time
	LastSuccessfulPublicationAt time.Time
	LastNewPublicationAt        time.Time
	LastObservedAsOfTs          time.Time
	LastAcceptedDedupeKey       string
	LastAcceptedRevision        string
	LastError                   string
	ConsecutiveDelayedPolls     int
	ConsecutiveFailures         int
	LastPublicationStatus       PublicationStatus
	LastPublishedValue          *CandidateValue
}

type Tracker struct {
	schedules map[SourceFamily]ScheduleConfig
	statuses  map[string]SourceStatus
}

type SourceIdentity struct {
	SourceFamily SourceFamily
	MetricFamily MetricFamily
	SourceKey    string
	Asset        string
}

func NewTracker(configs ...ScheduleConfig) (*Tracker, error) {
	tracker := &Tracker{
		schedules: make(map[SourceFamily]ScheduleConfig),
		statuses:  make(map[string]SourceStatus),
	}
	if len(configs) == 0 {
		for _, family := range []SourceFamily{SourceFamilyCME, SourceFamilyETF} {
			config, err := DefaultScheduleConfig(family)
			if err != nil {
				return nil, err
			}
			tracker.schedules[family] = config
		}
		return tracker, nil
	}
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return nil, err
		}
		tracker.schedules[config.SourceFamily] = config
	}
	return tracker, nil
}

func (t *Tracker) PollCadence(sourceFamily SourceFamily, now time.Time) (time.Duration, error) {
	config, err := t.scheduleFor(sourceFamily)
	if err != nil {
		return 0, err
	}
	return config.PollCadenceAt(now)
}

func (t *Tracker) ObservePoll(now time.Time, result PollResult) (SourceStatus, PollDecision, error) {
	if t == nil {
		return SourceStatus{}, PollDecision{}, fmt.Errorf("tracker is required")
	}
	if now.IsZero() {
		return SourceStatus{}, PollDecision{}, fmt.Errorf("current time is required")
	}
	if err := result.Validate(); err != nil {
		return SourceStatus{}, PollDecision{}, err
	}
	config, err := t.scheduleFor(result.SourceFamily)
	if err != nil {
		return SourceStatus{}, PollDecision{}, err
	}

	key := sourceStatusKey(result.SourceFamily, result.MetricFamily, result.SourceKey, result.Asset)
	status := t.statuses[key]
	status.SourceFamily = result.SourceFamily
	status.MetricFamily = result.MetricFamily
	status.SourceKey = result.SourceKey
	status.Asset = result.Asset
	status.LastSuccessfulPollAt = now.UTC()
	status.LastObservedAsOfTs = result.AsOfTs
	status.LastPublicationStatus = result.Status
	status.LastError = ""
	status.ConsecutiveFailures = 0

	decision := PollDecision{}
	if result.Status == PublicationStatusNotYet {
		if config.IsDelayed(now) {
			status.HealthState = SourceHealthDelayedPublication
			status.ConsecutiveDelayedPolls++
		} else {
			status.HealthState = SourceHealthHealthy
			status.ConsecutiveDelayedPolls = 0
		}
		t.statuses[key] = status
		return status, decision, nil
	}

	status.HealthState = SourceHealthHealthy
	status.ConsecutiveDelayedPolls = 0
	status.LastSuccessfulPublicationAt = now.UTC()
	previousValue := cloneCandidateValue(status.LastPublishedValue)

	if result.Status == PublicationStatusNew {
		decision.Accepted = true
		status.LastNewPublicationAt = now.UTC()
	} else if status.LastAcceptedDedupeKey == "" {
		decision.Accepted = true
	} else if status.LastAcceptedDedupeKey == result.DedupeKey && status.LastAcceptedRevision == result.Revision && candidateValuesEqual(previousValue, result.Value) {
		decision.Duplicate = true
	} else {
		decision.Accepted = true
		decision.Correction = true
	}

	if decision.Accepted {
		status.LastAcceptedDedupeKey = result.DedupeKey
		status.LastAcceptedRevision = result.Revision
		status.LastPublishedValue = cloneCandidateValue(result.Value)
	}

	t.statuses[key] = status
	return status, decision, nil
}

func (t *Tracker) ObserveFailure(identity SourceIdentity, now time.Time, err error) (SourceStatus, error) {
	if t == nil {
		return SourceStatus{}, fmt.Errorf("tracker is required")
	}
	if now.IsZero() {
		return SourceStatus{}, fmt.Errorf("current time is required")
	}
	if err == nil {
		return SourceStatus{}, fmt.Errorf("failure error is required")
	}
	if err := identity.Validate(); err != nil {
		return SourceStatus{}, err
	}
	key := sourceStatusKey(identity.SourceFamily, identity.MetricFamily, identity.SourceKey, identity.Asset)
	status := t.statuses[key]
	status.SourceFamily = identity.SourceFamily
	status.MetricFamily = identity.MetricFamily
	status.SourceKey = identity.SourceKey
	status.Asset = identity.Asset
	status.ConsecutiveFailures++
	status.LastError = err.Error()
	status.HealthState = SourceHealthSourceUnavailable
	var parseErr *ParseError
	if errors.As(err, &parseErr) {
		status.HealthState = SourceHealthParseFailed
	}
	t.statuses[key] = status
	return status, nil
}

func (i SourceIdentity) Validate() error {
	if err := validateSourceFamily(i.SourceFamily); err != nil {
		return err
	}
	if err := validateMetricFamily(i.SourceFamily, i.MetricFamily); err != nil {
		return err
	}
	if i.SourceKey == "" {
		return fmt.Errorf("source key is required")
	}
	if i.Asset == "" {
		return fmt.Errorf("asset is required")
	}
	return nil
}

func (t *Tracker) scheduleFor(sourceFamily SourceFamily) (ScheduleConfig, error) {
	if t == nil {
		return ScheduleConfig{}, fmt.Errorf("tracker is required")
	}
	if err := validateSourceFamily(sourceFamily); err != nil {
		return ScheduleConfig{}, err
	}
	config, ok := t.schedules[sourceFamily]
	if !ok {
		return ScheduleConfig{}, fmt.Errorf("schedule for source family %q is not configured", sourceFamily)
	}
	return config, nil
}

func sourceStatusKey(sourceFamily SourceFamily, metricFamily MetricFamily, sourceKey, asset string) string {
	return fmt.Sprintf("%s|%s|%s|%s", sourceFamily, metricFamily, sourceKey, asset)
}

func cloneCandidateValue(value *CandidateValue) *CandidateValue {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func candidateValuesEqual(left, right *CandidateValue) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Amount == right.Amount && left.Unit == right.Unit
}
