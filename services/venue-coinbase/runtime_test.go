package venuecoinbase

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestRuntimeReconnectDelayUsesCoinbaseConfigBounds(t *testing.T) {
	config := loadCoinbaseRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	delay, err := runtime.ReconnectDelay(4)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}
	if delay != 4*time.Second {
		t.Fatalf("delay = %s, want %s", delay, 4*time.Second)
	}
}

func TestRuntimeEvaluateLoopStateReturnsHealthyDecision(t *testing.T) {
	config := loadCoinbaseRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   now.Add(-5 * time.Second),
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if decision.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthHealthy)
	}
	if decision.SnapshotFreshness != ingestion.FreshnessNotApplicable {
		t.Fatalf("snapshot freshness = %q, want %q", decision.SnapshotFreshness, ingestion.FreshnessNotApplicable)
	}
}

func TestRuntimeEvaluateLoopStateReturnsDegradedDecision(t *testing.T) {
	config := loadCoinbaseRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState:       ingestion.ConnectionReconnecting,
		LastMessageAt:         now.Add(-5 * time.Second),
		LocalClockOffset:      250 * time.Millisecond,
		ConsecutiveReconnects: 4,
		ResyncCount:           3,
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
	} {
		if !hasReason(decision.Reasons, reason) {
			t.Fatalf("reasons = %v, want %q", decision.Reasons, reason)
		}
	}
}

func TestRuntimeEvaluateLoopStateReturnsStaleDecision(t *testing.T) {
	config := loadCoinbaseRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   now.Add(-16 * time.Second),
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if decision.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthStale)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonMessageStale)
	}
}

func TestNewRuntimeRejectsWrongVenue(t *testing.T) {
	config := loadCoinbaseRuntimeConfig(t)
	config.Venue = ingestion.VenueBinance

	if _, err := NewRuntime(config); err == nil {
		t.Fatal("expected non-coinbase runtime config to fail")
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
