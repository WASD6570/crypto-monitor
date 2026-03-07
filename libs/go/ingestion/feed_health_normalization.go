package ingestion

import (
	"fmt"
	"sort"
)

type FeedHealthMetadata struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	Venue         Venue
	MarketType    string
}

type FeedHealthMessage struct {
	ExchangeTs     string
	RecvTs         string
	SourceRecordID string
	Status         FeedHealthStatus
}

func NormalizeFeedHealthMessage(metadata FeedHealthMetadata, message FeedHealthMessage, policy TimestampPolicy) (CanonicalFeedHealthEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalFeedHealthEvent{}, fmt.Errorf("feed-health metadata is incomplete")
	}
	if message.SourceRecordID == "" {
		return CanonicalFeedHealthEvent{}, fmt.Errorf("source record id is required")
	}

	resolvedTs, err := resolveMessageTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return CanonicalFeedHealthEvent{}, err
	}

	reasons := append([]DegradationReason(nil), message.Status.Reasons...)
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i] < reasons[j]
	})

	return CanonicalFeedHealthEvent{
		SchemaVersion:           "v1",
		EventType:               "feed-health",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolvedTs.Status,
		FeedHealthState:         message.Status.State,
		DegradationReasons:      reasons,
		SourceRecordID:          message.SourceRecordID,
		CanonicalEventTime:      resolvedTs.EventTime,
		TimestampFallbackReason: resolvedTs.FallbackReason,
	}, nil
}
