package slowcontext

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type AvailabilityState string

const (
	AvailabilityAvailable   AvailabilityState = "available"
	AvailabilityUnavailable AvailabilityState = "unavailable"
)

type FreshnessState string

const (
	FreshnessFresh       FreshnessState = "fresh"
	FreshnessDelayed     FreshnessState = "delayed"
	FreshnessStale       FreshnessState = "stale"
	FreshnessUnavailable FreshnessState = "unavailable"
)

type AcceptedRecord struct {
	SourceFamily    SourceFamily
	MetricFamily    MetricFamily
	SourceKey       string
	Asset           string
	AsOfTs          time.Time
	PublishedTs     time.Time
	IngestTs        time.Time
	AcceptedAt      time.Time
	DedupeKey       string
	Revision        string
	ExpectedCadence string
	Value           CandidateValue
}

type ThresholdBasis struct {
	ExpectedCadence string    `json:"expectedCadence"`
	DelayedAfterTs  time.Time `json:"delayedAfterTs"`
	StaleAfterTs    time.Time `json:"staleAfterTs"`
	AgeReference    string    `json:"ageReference"`
}

type ContextResponse struct {
	SourceFamily    SourceFamily      `json:"sourceFamily,omitempty"`
	MetricFamily    MetricFamily      `json:"metricFamily"`
	Asset           string            `json:"asset"`
	Availability    AvailabilityState `json:"availability"`
	Freshness       FreshnessState    `json:"freshness"`
	ExpectedCadence string            `json:"expectedCadence,omitempty"`
	AsOfTs          time.Time         `json:"asOfTs,omitempty"`
	PublishedTs     time.Time         `json:"publishedTs,omitempty"`
	IngestTs        time.Time         `json:"ingestTs,omitempty"`
	Age             time.Duration     `json:"age"`
	Revision        string            `json:"revision,omitempty"`
	Value           *CandidateValue   `json:"value,omitempty"`
	PreviousValue   *CandidateValue   `json:"previousValue,omitempty"`
	ThresholdBasis  *ThresholdBasis   `json:"thresholdBasis,omitempty"`
	MessageKey      string            `json:"messageKey"`
	Message         string            `json:"message"`
	Error           string            `json:"error,omitempty"`
}

type AssetContextResponse struct {
	Asset     string            `json:"asset"`
	Contexts  []ContextResponse `json:"contexts"`
	QueriedAt time.Time         `json:"queriedAt"`
}

type AssetQuery struct {
	Asset          string
	MetricFamilies []MetricFamily
	Now            time.Time
}

type Store interface {
	Append(record AcceptedRecord) error
	List(asset string, metricFamilies []MetricFamily) ([]AcceptedRecord, error)
}

type InMemoryStore struct {
	records []AcceptedRecord
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (s *Service) RecordAccepted(result PollResult, expectedCadence string, acceptedAt time.Time) (AcceptedRecord, error) {
	if s == nil {
		return AcceptedRecord{}, fmt.Errorf("slow context service is required")
	}
	if err := result.Validate(); err != nil {
		return AcceptedRecord{}, err
	}
	if result.Status == PublicationStatusNotYet {
		return AcceptedRecord{}, fmt.Errorf("not-yet-published results cannot be accepted")
	}
	if acceptedAt.IsZero() {
		return AcceptedRecord{}, fmt.Errorf("accepted-at time is required")
	}
	if expectedCadence == "" {
		expectedCadence = defaultExpectedCadence(result.MetricFamily)
	}
	record := AcceptedRecord{
		SourceFamily:    result.SourceFamily,
		MetricFamily:    result.MetricFamily,
		SourceKey:       result.SourceKey,
		Asset:           result.Asset,
		AsOfTs:          result.AsOfTs,
		PublishedTs:     result.PublishedTs,
		IngestTs:        result.IngestTs,
		AcceptedAt:      acceptedAt.UTC(),
		DedupeKey:       result.DedupeKey,
		Revision:        result.Revision,
		ExpectedCadence: expectedCadence,
		Value:           *cloneCandidateValue(result.Value),
	}
	if err := record.Validate(); err != nil {
		return AcceptedRecord{}, err
	}
	if err := s.store.Append(record); err != nil {
		return AcceptedRecord{}, err
	}
	return record, nil
}

func (s *Service) QueryAsset(query AssetQuery) (AssetContextResponse, error) {
	if s == nil {
		return AssetContextResponse{}, fmt.Errorf("slow context service is required")
	}
	if query.Asset == "" {
		return AssetContextResponse{}, fmt.Errorf("asset is required")
	}
	if query.Now.IsZero() {
		return AssetContextResponse{}, fmt.Errorf("query time is required")
	}
	metricFamilies := query.MetricFamilies
	if len(metricFamilies) == 0 {
		metricFamilies = defaultMetricFamiliesForAsset(query.Asset)
	}
	records, err := s.store.List(query.Asset, metricFamilies)
	if err != nil {
		return AssetContextResponse{}, err
	}
	recordsByMetric := make(map[MetricFamily][]AcceptedRecord, len(metricFamilies))
	for _, record := range records {
		recordsByMetric[record.MetricFamily] = append(recordsByMetric[record.MetricFamily], record)
	}
	contexts := make([]ContextResponse, 0, len(metricFamilies))
	for _, metricFamily := range metricFamilies {
		relevant := recordsByMetric[metricFamily]
		if len(relevant) == 0 {
			contexts = append(contexts, unavailableContextResponse(query.Asset, metricFamily, ""))
			continue
		}
		sort.Slice(relevant, func(i, j int) bool {
			return relevant[i].AcceptedAt.Before(relevant[j].AcceptedAt)
		})
		latest := relevant[len(relevant)-1]
		var previous *AcceptedRecord
		if len(relevant) > 1 {
			candidate := relevant[len(relevant)-2]
			previous = &candidate
		}
		contexts = append(contexts, buildContextResponse(latest, previous, query.Now))
	}
	return AssetContextResponse{Asset: query.Asset, Contexts: contexts, QueriedAt: query.Now.UTC()}, nil
}

func NewUnavailableAssetContextResponse(asset string, metricFamilies []MetricFamily, queriedAt time.Time, cause string) AssetContextResponse {
	if queriedAt.IsZero() {
		queriedAt = time.Now().UTC()
	}
	if len(metricFamilies) == 0 {
		metricFamilies = defaultMetricFamiliesForAsset(asset)
	}
	contexts := make([]ContextResponse, 0, len(metricFamilies))
	for _, metricFamily := range metricFamilies {
		contexts = append(contexts, unavailableContextResponse(asset, metricFamily, cause))
	}
	return AssetContextResponse{Asset: asset, Contexts: contexts, QueriedAt: queriedAt.UTC()}
}

func (r AssetContextResponse) Context(metricFamily MetricFamily) (ContextResponse, bool) {
	for _, context := range r.Contexts {
		if context.MetricFamily == metricFamily {
			return context, true
		}
	}
	return ContextResponse{}, false
}

func (r AcceptedRecord) Validate() error {
	if err := validateSourceFamily(r.SourceFamily); err != nil {
		return err
	}
	if err := validateMetricFamily(r.SourceFamily, r.MetricFamily); err != nil {
		return err
	}
	if r.SourceKey == "" {
		return fmt.Errorf("source key is required")
	}
	if r.Asset == "" {
		return fmt.Errorf("asset is required")
	}
	if r.AsOfTs.IsZero() {
		return fmt.Errorf("as-of timestamp is required")
	}
	if r.IngestTs.IsZero() {
		return fmt.Errorf("ingest timestamp is required")
	}
	if r.AcceptedAt.IsZero() {
		return fmt.Errorf("accepted-at timestamp is required")
	}
	if r.DedupeKey == "" {
		return fmt.Errorf("dedupe key is required")
	}
	if r.ExpectedCadence == "" {
		return fmt.Errorf("expected cadence is required")
	}
	if r.Value.Amount == "" || r.Value.Unit == "" {
		return fmt.Errorf("accepted record value is required")
	}
	return nil
}

func (s *InMemoryStore) Append(record AcceptedRecord) error {
	if s == nil {
		return fmt.Errorf("store is required")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	s.records = append(s.records, record)
	return nil
}

func (s *InMemoryStore) List(asset string, metricFamilies []MetricFamily) ([]AcceptedRecord, error) {
	if s == nil {
		return nil, fmt.Errorf("store is required")
	}
	metricSet := make(map[MetricFamily]struct{}, len(metricFamilies))
	for _, metricFamily := range metricFamilies {
		metricSet[metricFamily] = struct{}{}
	}
	records := make([]AcceptedRecord, 0, len(s.records))
	for _, record := range s.records {
		if record.Asset != asset {
			continue
		}
		if len(metricSet) > 0 {
			if _, ok := metricSet[record.MetricFamily]; !ok {
				continue
			}
		}
		records = append(records, record)
	}
	return records, nil
}

func buildContextResponse(latest AcceptedRecord, previous *AcceptedRecord, now time.Time) ContextResponse {
	freshness, basis := classifyFreshness(latest, now)
	messageKey, message := freshnessMessage(latest.MetricFamily, freshness)
	response := ContextResponse{
		SourceFamily:    latest.SourceFamily,
		MetricFamily:    latest.MetricFamily,
		Asset:           latest.Asset,
		Availability:    AvailabilityAvailable,
		Freshness:       freshness,
		ExpectedCadence: latest.ExpectedCadence,
		AsOfTs:          latest.AsOfTs,
		PublishedTs:     latest.PublishedTs,
		IngestTs:        latest.IngestTs,
		Age:             now.UTC().Sub(latest.AsOfTs),
		Revision:        latest.Revision,
		Value:           candidateValuePointer(latest.Value),
		ThresholdBasis:  &basis,
		MessageKey:      messageKey,
		Message:         message,
	}
	if previous != nil {
		response.PreviousValue = candidateValuePointer(previous.Value)
	}
	return response
}

func unavailableContextResponse(asset string, metricFamily MetricFamily, cause string) ContextResponse {
	messageKey, message := freshnessMessage(metricFamily, FreshnessUnavailable)
	response := ContextResponse{
		MetricFamily: metricFamily,
		Asset:        asset,
		Availability: AvailabilityUnavailable,
		Freshness:    FreshnessUnavailable,
		MessageKey:   messageKey,
		Message:      message,
	}
	if cause != "" {
		response.Error = cause
	}
	return response
}

func classifyFreshness(record AcceptedRecord, now time.Time) (FreshnessState, ThresholdBasis) {
	delayedAfter := nextPublishWindowEnd(record.SourceFamily, record.AsOfTs)
	staleAfter := delayedAfter.Add(staleDuration(record.SourceFamily))
	basis := ThresholdBasis{
		ExpectedCadence: record.ExpectedCadence,
		DelayedAfterTs:  delayedAfter,
		StaleAfterTs:    staleAfter,
		AgeReference:    "as_of",
	}
	if !now.UTC().Before(staleAfter) {
		return FreshnessStale, basis
	}
	if !now.UTC().Before(delayedAfter) {
		return FreshnessDelayed, basis
	}
	return FreshnessFresh, basis
}

func defaultMetricFamiliesForAsset(asset string) []MetricFamily {
	asset = strings.ToUpper(asset)
	metricFamilies := []MetricFamily{MetricFamilyCMEVolume, MetricFamilyCMEOpenInterest}
	if asset == "BTC" {
		metricFamilies = append(metricFamilies, MetricFamilyETFDailyFlow)
	}
	return metricFamilies
}

func defaultExpectedCadence(metricFamily MetricFamily) string {
	switch metricFamily {
	case MetricFamilyETFDailyFlow:
		return "daily"
	default:
		return "session"
	}
}

func nextPublishWindowEnd(sourceFamily SourceFamily, asOf time.Time) time.Time {
	config, err := DefaultScheduleConfig(sourceFamily)
	if err != nil {
		return asOf.UTC()
	}
	nextDay := time.Date(asOf.UTC().Year(), asOf.UTC().Month(), asOf.UTC().Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return nextDay.Add(time.Duration(config.PublishWindowEndMinute) * time.Minute)
}

func staleDuration(sourceFamily SourceFamily) time.Duration {
	switch sourceFamily {
	case SourceFamilyETF:
		return 48 * time.Hour
	default:
		return 36 * time.Hour
	}
}

func freshnessMessage(metricFamily MetricFamily, freshness FreshnessState) (string, string) {
	metricLabel := metricFamilyLabel(metricFamily)
	keyBase := strings.ToLower(strings.ReplaceAll(metricLabel, " ", "_"))
	switch freshness {
	case FreshnessFresh:
		return keyBase + "_fresh", metricLabel + " is current"
	case FreshnessDelayed:
		return keyBase + "_delayed", metricLabel + " is delayed"
	case FreshnessStale:
		return keyBase + "_stale", metricLabel + " is stale"
	default:
		return keyBase + "_unavailable", metricLabel + " is unavailable"
	}
}

func metricFamilyLabel(metricFamily MetricFamily) string {
	switch metricFamily {
	case MetricFamilyCMEVolume:
		return "CME volume"
	case MetricFamilyCMEOpenInterest:
		return "CME open interest"
	case MetricFamilyETFDailyFlow:
		return "ETF daily flow"
	default:
		return string(metricFamily)
	}
}

func candidateValuePointer(value CandidateValue) *CandidateValue {
	clone := value
	return &clone
}
