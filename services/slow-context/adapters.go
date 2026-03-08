package slowcontext

import (
	"encoding/json"
	"fmt"
	"time"
)

type fixtureEnvelope struct {
	SourceKey    string            `json:"sourceKey"`
	MetricFamily MetricFamily      `json:"metricFamily"`
	PublishState PublicationStatus `json:"publishState"`
	Asset        string            `json:"asset"`
	AsOf         string            `json:"asOf"`
	PublishedAt  string            `json:"publishedAt,omitempty"`
	Revision     string            `json:"revision,omitempty"`
	Value        *CandidateValue   `json:"value,omitempty"`
}

type cmeAdapter struct{}

func NewCMEAdapter() Adapter {
	return cmeAdapter{}
}

func (c cmeAdapter) SourceFamily() SourceFamily {
	return SourceFamilyCME
}

func (c cmeAdapter) ParsePoll(raw []byte, ingestTs string) (PollResult, error) {
	return parseFixturePoll(SourceFamilyCME, raw, ingestTs)
}

type etfAdapter struct{}

func NewETFAdapter() Adapter {
	return etfAdapter{}
}

func (e etfAdapter) SourceFamily() SourceFamily {
	return SourceFamilyETF
}

func (e etfAdapter) ParsePoll(raw []byte, ingestTs string) (PollResult, error) {
	return parseFixturePoll(SourceFamilyETF, raw, ingestTs)
}

func parseFixturePoll(sourceFamily SourceFamily, raw []byte, ingestTs string) (PollResult, error) {
	var envelope fixtureEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return PollResult{}, &ParseError{SourceFamily: sourceFamily, Err: err}
	}

	ingestTime, err := time.Parse(time.RFC3339, ingestTs)
	if err != nil {
		return PollResult{}, &ParseError{SourceFamily: sourceFamily, Err: fmt.Errorf("parse ingest timestamp: %w", err)}
	}
	asOfTime, err := time.Parse(time.RFC3339, envelope.AsOf)
	if err != nil {
		return PollResult{}, &ParseError{SourceFamily: sourceFamily, Err: fmt.Errorf("parse as-of timestamp: %w", err)}
	}

	result := PollResult{
		SourceFamily: sourceFamily,
		MetricFamily: envelope.MetricFamily,
		Status:       envelope.PublishState,
		SourceKey:    envelope.SourceKey,
		Asset:        envelope.Asset,
		AsOfTs:       asOfTime.UTC(),
		IngestTs:     ingestTime.UTC(),
		Revision:     envelope.Revision,
		Value:        envelope.Value,
	}
	result.DedupeKey = BuildDedupeKey(result.SourceFamily, result.MetricFamily, result.Asset, result.SourceKey, result.AsOfTs)

	if envelope.PublishedAt != "" {
		publishedTime, err := time.Parse(time.RFC3339, envelope.PublishedAt)
		if err != nil {
			return PollResult{}, &ParseError{SourceFamily: sourceFamily, Err: fmt.Errorf("parse published timestamp: %w", err)}
		}
		result.PublishedTs = publishedTime.UTC()
	}

	if err := result.Validate(); err != nil {
		return PollResult{}, &ParseError{SourceFamily: sourceFamily, Err: err}
	}

	return result, nil
}
