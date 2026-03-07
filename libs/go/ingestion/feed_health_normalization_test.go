package ingestion

import (
	"testing"
	"time"
)

func TestNormalizeFeedHealthMessagePreservesStateAndReasons(t *testing.T) {
	actual, err := NormalizeFeedHealthMessage(FeedHealthMetadata{
		Symbol:        "ETH-USD",
		SourceSymbol:  "ETH/USD",
		QuoteCurrency: "USD",
		Venue:         VenueKraken,
		MarketType:    "spot",
	}, FeedHealthMessage{
		ExchangeTs:     "2026-03-06T12:02:00Z",
		RecvTs:         "2026-03-06T12:02:00.12Z",
		SourceRecordID: "runtime:kraken-gap",
		Status: FeedHealthStatus{
			State:   FeedHealthDegraded,
			Reasons: []DegradationReason{ReasonResyncLoop, ReasonSequenceGap},
		},
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize feed-health: %v", err)
	}

	if actual.FeedHealthState != FeedHealthDegraded {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, FeedHealthDegraded)
	}
	if len(actual.DegradationReasons) != 2 {
		t.Fatalf("degradation reasons = %v, want 2 reasons", actual.DegradationReasons)
	}
	if actual.DegradationReasons[0] != ReasonResyncLoop || actual.DegradationReasons[1] != ReasonSequenceGap {
		t.Fatalf("degradation reasons = %v, want sorted reasons", actual.DegradationReasons)
	}
	if !actual.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 2, 0, 0, time.UTC)) {
		t.Fatalf("canonical event time = %s, want exchange time", actual.CanonicalEventTime)
	}
}
