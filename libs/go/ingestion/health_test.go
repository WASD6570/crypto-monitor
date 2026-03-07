package ingestion

import (
	"testing"
	"time"
)

func TestAdapterConfigValidateRejectsInvalidThresholds(t *testing.T) {
	config := validConfig()
	config.ReconnectBackoffMax = config.ReconnectBackoffMin - time.Second

	if err := config.Validate(); err == nil {
		t.Fatal("expected validation error for reconnect backoff range")
	}
}

func TestAdapterConfigValidateRejectsMissingStreams(t *testing.T) {
	config := validConfig()
	config.Streams = nil

	if err := config.Validate(); err == nil {
		t.Fatal("expected validation error for missing streams")
	}
}

func TestEvaluateHealthHealthyStream(t *testing.T) {
	now := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	status, err := EvaluateHealth(validConfig(), HealthSnapshot{
		ConnectionState: ConnectionConnected,
		Now:             now,
		LastMessageAt:   now.Add(-2 * time.Second),
		LastSnapshotAt:  now.Add(-5 * time.Second),
	})
	if err != nil {
		t.Fatalf("evaluate health: %v", err)
	}

	if status.State != FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, FeedHealthHealthy)
	}
	if len(status.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", status.Reasons)
	}
	if status.MessageFreshness != FreshnessFresh {
		t.Fatalf("message freshness = %q, want %q", status.MessageFreshness, FreshnessFresh)
	}
	if status.SnapshotFreshness != FreshnessFresh {
		t.Fatalf("snapshot freshness = %q, want %q", status.SnapshotFreshness, FreshnessFresh)
	}
}

func TestEvaluateHealthMarksStaleWhenMessagesStop(t *testing.T) {
	now := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	status, err := EvaluateHealth(validConfig(), HealthSnapshot{
		ConnectionState: ConnectionConnected,
		Now:             now,
		LastMessageAt:   now.Add(-16 * time.Second),
		LastSnapshotAt:  now.Add(-5 * time.Second),
	})
	if err != nil {
		t.Fatalf("evaluate health: %v", err)
	}

	if status.State != FeedHealthStale {
		t.Fatalf("state = %q, want %q", status.State, FeedHealthStale)
	}
	if !containsReason(status.Reasons, ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ReasonMessageStale)
	}
	if status.MessageFreshness != FreshnessStale {
		t.Fatalf("message freshness = %q, want %q", status.MessageFreshness, FreshnessStale)
	}
}

func TestEvaluateHealthMarksDegradedForReconnectGapAndClock(t *testing.T) {
	now := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	status, err := EvaluateHealth(validConfig(), HealthSnapshot{
		ConnectionState:       ConnectionReconnecting,
		Now:                   now,
		LastMessageAt:         now.Add(-3 * time.Second),
		LastSnapshotAt:        now.Add(-5 * time.Second),
		SequenceGapDetected:   true,
		ConsecutiveReconnects: 3,
		ResyncCount:           2,
		LocalClockOffset:      300 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("evaluate health: %v", err)
	}

	if status.State != FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", status.State, FeedHealthDegraded)
	}
	wantReasons := []DegradationReason{
		ReasonClockDegraded,
		ReasonConnectionNotReady,
		ReasonReconnectLoop,
		ReasonResyncLoop,
		ReasonSequenceGap,
	}
	for _, reason := range wantReasons {
		if !containsReason(status.Reasons, reason) {
			t.Fatalf("reasons = %v, want %q", status.Reasons, reason)
		}
	}
	if status.ClockState != ClockDegraded {
		t.Fatalf("clock state = %q, want %q", status.ClockState, ClockDegraded)
	}
	if status.ConnectionState != ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", status.ConnectionState, ConnectionReconnecting)
	}
}

func validConfig() AdapterConfig {
	return AdapterConfig{
		Venue: VenueKraken,
		Streams: []StreamDefinition{
			{Kind: StreamTrades, MarketType: "spot"},
			{Kind: StreamOrderBook, MarketType: "spot", SnapshotRequired: true},
		},
		MessageStaleAfter:      15 * time.Second,
		SnapshotStaleAfter:     30 * time.Second,
		ReconnectBackoffMin:    time.Second,
		ReconnectBackoffMax:    10 * time.Second,
		ReconnectLoopThreshold: 3,
		ResyncLoopThreshold:    2,
		ClockOffsetWarning:     100 * time.Millisecond,
		ClockOffsetDegraded:    250 * time.Millisecond,
	}
}
