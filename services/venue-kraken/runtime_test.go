package venuekraken

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestRuntimeReconnectDelayUsesKrakenConfigBounds(t *testing.T) {
	config := loadKrakenRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	delay, err := runtime.ReconnectDelay(3)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}
	if delay != 2*time.Second {
		t.Fatalf("delay = %s, want %s", delay, 2*time.Second)
	}
}

func TestRuntimeEvaluateLoopStateReturnsHealthyDecision(t *testing.T) {
	config := loadKrakenRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   now.Add(-5 * time.Second),
		LastSnapshotAt:  now.Add(-10 * time.Second),
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if decision.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthHealthy)
	}
}

func TestRuntimeEvaluateLoopStateReturnsDegradedDecision(t *testing.T) {
	config := loadKrakenRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState:       ingestion.ConnectionReconnecting,
		LastMessageAt:         now.Add(-5 * time.Second),
		LastSnapshotAt:        now.Add(-10 * time.Second),
		SequenceGapDetected:   true,
		LocalClockOffset:      250 * time.Millisecond,
		ConsecutiveReconnects: 3,
		ResyncCount:           2,
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if decision.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthDegraded)
	}
	for _, reason := range []ingestion.DegradationReason{
		ingestion.ReasonClockDegraded,
		ingestion.ReasonConnectionNotReady,
		ingestion.ReasonReconnectLoop,
		ingestion.ReasonResyncLoop,
		ingestion.ReasonSequenceGap,
	} {
		if !hasReason(decision.Reasons, reason) {
			t.Fatalf("reasons = %v, want %q", decision.Reasons, reason)
		}
	}
}

func TestRuntimeEvaluateLoopStateReturnsStaleDecision(t *testing.T) {
	config := loadKrakenRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   now.Add(-13 * time.Second),
		LastSnapshotAt:  now.Add(-26 * time.Second),
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if decision.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthStale)
	}
	for _, reason := range []ingestion.DegradationReason{ingestion.ReasonMessageStale, ingestion.ReasonSnapshotStale} {
		if !hasReason(decision.Reasons, reason) {
			t.Fatalf("reasons = %v, want %q", decision.Reasons, reason)
		}
	}
}

func TestNewRuntimeRejectsWrongVenue(t *testing.T) {
	config := loadKrakenRuntimeConfig(t)
	config.Venue = ingestion.VenueBybit
	if _, err := NewRuntime(config); err == nil {
		t.Fatal("expected non-kraken runtime config to fail")
	}
}

func hasReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
