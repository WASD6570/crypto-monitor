package slowcontext

import (
	"fmt"
	"time"
)

type SourceFamily string

const (
	SourceFamilyCME SourceFamily = "CME"
	SourceFamilyETF SourceFamily = "ETF"
)

type MetricFamily string

const (
	MetricFamilyCMEVolume       MetricFamily = "cme_volume"
	MetricFamilyCMEOpenInterest MetricFamily = "cme_open_interest"
	MetricFamilyETFDailyFlow    MetricFamily = "etf_daily_flow"
)

type PublicationStatus string

const (
	PublicationStatusNew      PublicationStatus = "published_new_value"
	PublicationStatusSameAsOf PublicationStatus = "published_same_value_or_same_asof"
	PublicationStatusNotYet   PublicationStatus = "not_yet_published"
)

type CandidateValue struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

type PollResult struct {
	SourceFamily SourceFamily      `json:"sourceFamily"`
	MetricFamily MetricFamily      `json:"metricFamily"`
	Status       PublicationStatus `json:"status"`
	SourceKey    string            `json:"sourceKey"`
	Asset        string            `json:"asset"`
	AsOfTs       time.Time         `json:"asOfTs"`
	PublishedTs  time.Time         `json:"publishedTs,omitempty"`
	IngestTs     time.Time         `json:"ingestTs"`
	DedupeKey    string            `json:"dedupeKey"`
	Revision     string            `json:"revision,omitempty"`
	Value        *CandidateValue   `json:"value,omitempty"`
}

func (r PollResult) Validate() error {
	if err := validateSourceFamily(r.SourceFamily); err != nil {
		return err
	}
	if err := validateMetricFamily(r.SourceFamily, r.MetricFamily); err != nil {
		return err
	}
	if err := validatePublicationStatus(r.Status); err != nil {
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
	if r.DedupeKey == "" {
		return fmt.Errorf("dedupe key is required")
	}
	if r.Status == PublicationStatusNotYet {
		if r.Value != nil {
			return fmt.Errorf("not-yet-published results cannot include a value")
		}
		return nil
	}
	if r.Value == nil {
		return fmt.Errorf("published results require a value")
	}
	if r.Value.Amount == "" {
		return fmt.Errorf("value amount is required")
	}
	if r.Value.Unit == "" {
		return fmt.Errorf("value unit is required")
	}
	return nil
}

type ParseError struct {
	SourceFamily SourceFamily
	Err          error
}

func (e *ParseError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s parse failed: %v", e.SourceFamily, e.Err)
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func BuildDedupeKey(sourceFamily SourceFamily, metricFamily MetricFamily, asset, sourceKey string, asOf time.Time) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", sourceFamily, metricFamily, asset, sourceKey, asOf.UTC().Format(time.RFC3339))
}

func validateSourceFamily(sourceFamily SourceFamily) error {
	switch sourceFamily {
	case SourceFamilyCME, SourceFamilyETF:
		return nil
	default:
		return fmt.Errorf("unsupported source family %q", sourceFamily)
	}
}

func validateMetricFamily(sourceFamily SourceFamily, metricFamily MetricFamily) error {
	switch sourceFamily {
	case SourceFamilyCME:
		switch metricFamily {
		case MetricFamilyCMEVolume, MetricFamilyCMEOpenInterest:
			return nil
		}
	case SourceFamilyETF:
		switch metricFamily {
		case MetricFamilyETFDailyFlow:
			return nil
		}
	}
	return fmt.Errorf("unsupported metric family %q for source family %q", metricFamily, sourceFamily)
}

func validatePublicationStatus(status PublicationStatus) error {
	switch status {
	case PublicationStatusNew, PublicationStatusSameAsOf, PublicationStatusNotYet:
		return nil
	default:
		return fmt.Errorf("unsupported publication status %q", status)
	}
}
