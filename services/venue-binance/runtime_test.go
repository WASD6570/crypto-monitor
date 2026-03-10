package venuebinance

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestRuntimeReconnectDelayUsesBinanceConfigBounds(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
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

func TestRuntimeReconnectDelayClampsAtConfiguredMaximum(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	delay, err := runtime.ReconnectDelay(8)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}

	if delay != 5*time.Second {
		t.Fatalf("delay = %s, want %s", delay, 5*time.Second)
	}
}

func TestRuntimeReconnectDelayRejectsInvalidAttempt(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.ReconnectDelay(0); err == nil {
		t.Fatal("expected invalid reconnect attempt to fail")
	}
}

func TestRuntimeSnapshotRecoveryStatusAllowsFirstAttempt(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.SnapshotRecoveryStatus(time.UnixMilli(1772798461000).UTC(), time.Time{})
	if err != nil {
		t.Fatalf("snapshot recovery status: %v", err)
	}

	if !status.Ready {
		t.Fatal("expected first snapshot recovery attempt to be ready")
	}
	if status.RemainingCooldown != 0 {
		t.Fatalf("remaining cooldown = %s, want %s", status.RemainingCooldown, 0*time.Second)
	}
}

func TestRuntimeSnapshotRecoveryStatusReportsRemainingCooldown(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798461500).UTC()
	lastAttempt := time.UnixMilli(1772798461000).UTC()
	status, err := runtime.SnapshotRecoveryStatus(now, lastAttempt)
	if err != nil {
		t.Fatalf("snapshot recovery status: %v", err)
	}

	if status.Ready {
		t.Fatal("expected snapshot recovery cooldown to still be active")
	}
	if status.RemainingCooldown != 500*time.Millisecond {
		t.Fatalf("remaining cooldown = %s, want %s", status.RemainingCooldown, 500*time.Millisecond)
	}
}

func TestRuntimeSnapshotRecoveryStatusAllowsRetryAfterCooldown(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798462000).UTC()
	lastAttempt := time.UnixMilli(1772798461000).UTC()
	status, err := runtime.SnapshotRecoveryStatus(now, lastAttempt)
	if err != nil {
		t.Fatalf("snapshot recovery status: %v", err)
	}

	if !status.Ready {
		t.Fatal("expected snapshot recovery to be ready after cooldown")
	}
	if status.RemainingCooldown != 0 {
		t.Fatalf("remaining cooldown = %s, want %s", status.RemainingCooldown, 0*time.Second)
	}
}

func TestRuntimeSnapshotRecoveryStatusRejectsFutureAttemptTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798461000).UTC()
	lastAttempt := time.UnixMilli(1772798461500).UTC()
	if _, err := runtime.SnapshotRecoveryStatus(now, lastAttempt); err == nil {
		t.Fatal("expected future snapshot attempt time to fail")
	}
}

func TestRuntimeSnapshotRecoveryStatusRejectsMissingCurrentTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.SnapshotRecoveryStatus(time.Time{}, time.UnixMilli(1772798461000).UTC()); err == nil {
		t.Fatal("expected missing current time to fail")
	}
}

func TestRuntimeSnapshotRecoveryRateLimitStatusAllowsAttemptBelowLimit(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798465000).UTC()
	status, err := runtime.SnapshotRecoveryRateLimitStatus(now, []time.Time{now.Add(-30 * time.Second)})
	if err != nil {
		t.Fatalf("snapshot recovery rate-limit status: %v", err)
	}

	if !status.Allowed {
		t.Fatal("expected snapshot recovery attempt below limit to be allowed")
	}
	if status.AttemptsInWindow != 1 {
		t.Fatalf("attempts in window = %d, want %d", status.AttemptsInWindow, 1)
	}
	if status.RemainingAllowance != 1 {
		t.Fatalf("remaining allowance = %d, want %d", status.RemainingAllowance, 1)
	}
	if status.RetryAfter != 0 {
		t.Fatalf("retry after = %s, want %s", status.RetryAfter, 0*time.Second)
	}
}

func TestRuntimeSnapshotRecoveryRateLimitStatusBlocksAttemptAtLimit(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798465000).UTC()
	attempts := []time.Time{
		now.Add(-50 * time.Second),
		now.Add(-10 * time.Second),
	}
	status, err := runtime.SnapshotRecoveryRateLimitStatus(now, attempts)
	if err != nil {
		t.Fatalf("snapshot recovery rate-limit status: %v", err)
	}

	if status.Allowed {
		t.Fatal("expected snapshot recovery attempt at limit to be blocked")
	}
	if status.AttemptsInWindow != 2 {
		t.Fatalf("attempts in window = %d, want %d", status.AttemptsInWindow, 2)
	}
	if status.RemainingAllowance != 0 {
		t.Fatalf("remaining allowance = %d, want %d", status.RemainingAllowance, 0)
	}
	if status.RetryAfter != 10*time.Second {
		t.Fatalf("retry after = %s, want %s", status.RetryAfter, 10*time.Second)
	}
}

func TestRuntimeSnapshotRecoveryRateLimitStatusIgnoresAttemptsOutsideWindow(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 1
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798465000).UTC()
	status, err := runtime.SnapshotRecoveryRateLimitStatus(now, []time.Time{now.Add(-61 * time.Second)})
	if err != nil {
		t.Fatalf("snapshot recovery rate-limit status: %v", err)
	}

	if !status.Allowed {
		t.Fatal("expected attempts outside the one-minute window to be ignored")
	}
	if status.AttemptsInWindow != 0 {
		t.Fatalf("attempts in window = %d, want %d", status.AttemptsInWindow, 0)
	}
}

func TestRuntimeSnapshotRecoveryRateLimitStatusRejectsFutureAttemptTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798465000).UTC()
	if _, err := runtime.SnapshotRecoveryRateLimitStatus(now, []time.Time{now.Add(time.Second)}); err == nil {
		t.Fatal("expected future snapshot recovery attempt to fail")
	}
}

func TestRuntimeSnapshotRecoveryRateLimitStatusRejectsMissingCurrentTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.SnapshotRecoveryRateLimitStatus(time.Time{}, nil); err == nil {
		t.Fatal("expected missing current time to fail")
	}
}

func TestRuntimeSnapshotRefreshStatusReportsDueAfterInterval(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.SnapshotRefreshStatus(time.UnixMilli(1772798820000).UTC(), time.UnixMilli(1772798520000).UTC())
	if err != nil {
		t.Fatalf("snapshot refresh status: %v", err)
	}
	if !status.Required {
		t.Fatal("expected snapshot refresh to be required")
	}
	if !status.Due {
		t.Fatal("expected snapshot refresh to be due after interval")
	}
	if status.Remaining != 0 {
		t.Fatalf("remaining = %s, want %s", status.Remaining, 0*time.Second)
	}
}

func TestRuntimeSnapshotRefreshStatusReportsRemainingBeforeDue(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.SnapshotRefreshStatus(time.UnixMilli(1772798700000).UTC(), time.UnixMilli(1772798520000).UTC())
	if err != nil {
		t.Fatalf("snapshot refresh status: %v", err)
	}
	if status.Due {
		t.Fatal("did not expect snapshot refresh to be due early")
	}
	if status.Remaining != 2*time.Minute {
		t.Fatalf("remaining = %s, want %s", status.Remaining, 2*time.Minute)
	}
}

func TestRuntimeSnapshotRefreshStatusRejectsFutureSnapshotTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	if _, err := runtime.SnapshotRefreshStatus(now, now.Add(time.Second)); err == nil {
		t.Fatal("expected future snapshot time to fail")
	}
}

func TestRuntimeEvaluateStalenessReturnsHealthyForFreshMessageAndSnapshot(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	status, err := runtime.EvaluateStaleness(now, now.Add(-5*time.Second), now.Add(-10*time.Second))
	if err != nil {
		t.Fatalf("evaluate staleness: %v", err)
	}

	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	if status.MessageFreshness != ingestion.FreshnessFresh {
		t.Fatalf("message freshness = %q, want %q", status.MessageFreshness, ingestion.FreshnessFresh)
	}
	if status.SnapshotFreshness != ingestion.FreshnessFresh {
		t.Fatalf("snapshot freshness = %q, want %q", status.SnapshotFreshness, ingestion.FreshnessFresh)
	}
	if len(status.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", status.Reasons)
	}
}

func TestRuntimeEvaluateStalenessReturnsStaleWhenMessagesStop(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	status, err := runtime.EvaluateStaleness(now, now.Add(-16*time.Second), now.Add(-10*time.Second))
	if err != nil {
		t.Fatalf("evaluate staleness: %v", err)
	}

	if status.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthStale)
	}
	if status.MessageFreshness != ingestion.FreshnessStale {
		t.Fatalf("message freshness = %q, want %q", status.MessageFreshness, ingestion.FreshnessStale)
	}
	if !hasReason(status.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonMessageStale)
	}
}

func TestRuntimeEvaluateStalenessReturnsStaleWhenSnapshotStops(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	status, err := runtime.EvaluateStaleness(now, now.Add(-5*time.Second), now.Add(-31*time.Second))
	if err != nil {
		t.Fatalf("evaluate staleness: %v", err)
	}

	if status.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthStale)
	}
	if status.SnapshotFreshness != ingestion.FreshnessStale {
		t.Fatalf("snapshot freshness = %q, want %q", status.SnapshotFreshness, ingestion.FreshnessStale)
	}
	if !hasReason(status.Reasons, ingestion.ReasonSnapshotStale) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonSnapshotStale)
	}
}

func TestRuntimeEvaluateStalenessRejectsFutureLastMessageTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	if _, err := runtime.EvaluateStaleness(now, now.Add(time.Second), now); err == nil {
		t.Fatal("expected future last message time to fail")
	}
}

func TestRuntimeEvaluateStalenessRejectsFutureLastSnapshotTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798520000).UTC()
	if _, err := runtime.EvaluateStaleness(now, now, now.Add(time.Second)); err == nil {
		t.Fatal("expected future last snapshot time to fail")
	}
}

func TestRuntimeEvaluateStalenessRejectsMissingCurrentTime(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.EvaluateStaleness(time.Time{}, time.UnixMilli(1772798520000).UTC(), time.UnixMilli(1772798520000).UTC()); err == nil {
		t.Fatal("expected missing current time to fail")
	}
}

func TestRuntimeEvaluateReconnectLoopReturnsNormalBelowThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateReconnectLoop(2)
	if err != nil {
		t.Fatalf("evaluate reconnect loop: %v", err)
	}

	if status.LoopDetected {
		t.Fatal("expected reconnect count below threshold to stay normal")
	}
	if status.Threshold != 4 {
		t.Fatalf("threshold = %d, want %d", status.Threshold, 4)
	}
	if status.RemainingBeforeLoop != 2 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 2)
	}
}

func TestRuntimeEvaluateReconnectLoopDetectsLoopAtThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateReconnectLoop(4)
	if err != nil {
		t.Fatalf("evaluate reconnect loop: %v", err)
	}

	if !status.LoopDetected {
		t.Fatal("expected reconnect count at threshold to detect loop")
	}
	if status.RemainingBeforeLoop != 0 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 0)
	}
}

func TestRuntimeEvaluateReconnectLoopClampsRemainingBelowZero(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateReconnectLoop(6)
	if err != nil {
		t.Fatalf("evaluate reconnect loop: %v", err)
	}

	if !status.LoopDetected {
		t.Fatal("expected reconnect count above threshold to detect loop")
	}
	if status.RemainingBeforeLoop != 0 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 0)
	}
}

func TestRuntimeEvaluateReconnectLoopRejectsNegativeCount(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.EvaluateReconnectLoop(-1); err == nil {
		t.Fatal("expected negative reconnect count to fail")
	}
}

func TestRuntimeEvaluateResyncLoopReturnsNormalBelowThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateResyncLoop(2)
	if err != nil {
		t.Fatalf("evaluate resync loop: %v", err)
	}

	if status.LoopDetected {
		t.Fatal("expected resync count below threshold to stay normal")
	}
	if status.Threshold != 3 {
		t.Fatalf("threshold = %d, want %d", status.Threshold, 3)
	}
	if status.RemainingBeforeLoop != 1 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 1)
	}
}

func TestRuntimeEvaluateResyncLoopDetectsLoopAtThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateResyncLoop(3)
	if err != nil {
		t.Fatalf("evaluate resync loop: %v", err)
	}

	if !status.LoopDetected {
		t.Fatal("expected resync count at threshold to detect loop")
	}
	if status.RemainingBeforeLoop != 0 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 0)
	}
}

func TestRuntimeEvaluateResyncLoopClampsRemainingBelowZero(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateResyncLoop(5)
	if err != nil {
		t.Fatalf("evaluate resync loop: %v", err)
	}

	if !status.LoopDetected {
		t.Fatal("expected resync count above threshold to detect loop")
	}
	if status.RemainingBeforeLoop != 0 {
		t.Fatalf("remaining before loop = %d, want %d", status.RemainingBeforeLoop, 0)
	}
}

func TestRuntimeEvaluateResyncLoopRejectsNegativeCount(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := runtime.EvaluateResyncLoop(-1); err == nil {
		t.Fatal("expected negative resync count to fail")
	}
}

func TestRuntimeBuildHealthSnapshotPreservesAdapterInputFields(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	healthSnapshot, err := runtime.BuildHealthSnapshot(AdapterHealthInput{
		ConnectionState:       ingestion.ConnectionReconnecting,
		Now:                   now,
		LastMessageAt:         now.Add(-5 * time.Second),
		LastSnapshotAt:        now.Add(-10 * time.Second),
		SequenceGapDetected:   true,
		LocalClockOffset:      150 * time.Millisecond,
		ConsecutiveReconnects: 2,
		ResyncCount:           1,
	})
	if err != nil {
		t.Fatalf("build health snapshot: %v", err)
	}

	if healthSnapshot.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", healthSnapshot.ConnectionState, ingestion.ConnectionReconnecting)
	}
	if !healthSnapshot.Now.Equal(now) {
		t.Fatalf("now = %s, want %s", healthSnapshot.Now, now)
	}
	if !healthSnapshot.LastMessageAt.Equal(now.Add(-5 * time.Second)) {
		t.Fatalf("last message = %s, want %s", healthSnapshot.LastMessageAt, now.Add(-5*time.Second))
	}
	if !healthSnapshot.LastSnapshotAt.Equal(now.Add(-10 * time.Second)) {
		t.Fatalf("last snapshot = %s, want %s", healthSnapshot.LastSnapshotAt, now.Add(-10*time.Second))
	}
	if !healthSnapshot.SequenceGapDetected {
		t.Fatal("expected sequence gap to be preserved")
	}
	if healthSnapshot.LocalClockOffset != 150*time.Millisecond {
		t.Fatalf("clock offset = %s, want %s", healthSnapshot.LocalClockOffset, 150*time.Millisecond)
	}
	if healthSnapshot.ConsecutiveReconnects != 2 {
		t.Fatalf("consecutive reconnects = %d, want %d", healthSnapshot.ConsecutiveReconnects, 2)
	}
	if healthSnapshot.ResyncCount != 1 {
		t.Fatalf("resync count = %d, want %d", healthSnapshot.ResyncCount, 1)
	}
}

func TestRuntimeBuildHealthSnapshotRejectsUnsupportedConnectionState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.BuildHealthSnapshot(AdapterHealthInput{ConnectionState: ingestion.ConnectionState("BROKEN")})
	if err == nil {
		t.Fatal("expected unsupported connection state to fail")
	}
}

func TestAdapterLoopStateHealthInputPreservesStateFields(t *testing.T) {
	now := time.UnixMilli(1772798525000).UTC()
	state := AdapterLoopState{
		ConnectionState:          ingestion.ConnectionReconnecting,
		LastMessageAt:            now.Add(-5 * time.Second),
		LastSnapshotAt:           now.Add(-10 * time.Second),
		SequenceGapDetected:      true,
		LocalClockOffset:         150 * time.Millisecond,
		LastSnapshotRecoveryAt:   now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{now.Add(-20 * time.Second)},
		ConsecutiveReconnects:    2,
		ResyncCount:              1,
	}

	input := state.HealthInput(now)
	if input.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", input.ConnectionState, ingestion.ConnectionReconnecting)
	}
	if !input.Now.Equal(now) {
		t.Fatalf("now = %s, want %s", input.Now, now)
	}
	if !input.LastMessageAt.Equal(state.LastMessageAt) {
		t.Fatalf("last message = %s, want %s", input.LastMessageAt, state.LastMessageAt)
	}
	if !input.LastSnapshotAt.Equal(state.LastSnapshotAt) {
		t.Fatalf("last snapshot = %s, want %s", input.LastSnapshotAt, state.LastSnapshotAt)
	}
	if !input.SequenceGapDetected {
		t.Fatal("expected sequence gap to be preserved")
	}
	if input.LocalClockOffset != 150*time.Millisecond {
		t.Fatalf("clock offset = %s, want %s", input.LocalClockOffset, 150*time.Millisecond)
	}
	if !input.LastSnapshotRecoveryAt.Equal(state.LastSnapshotRecoveryAt) {
		t.Fatalf("last snapshot recovery = %s, want %s", input.LastSnapshotRecoveryAt, state.LastSnapshotRecoveryAt)
	}
	if len(input.RecentSnapshotRecoveries) != 1 {
		t.Fatalf("recent snapshot recoveries = %d, want %d", len(input.RecentSnapshotRecoveries), 1)
	}
	if input.ConsecutiveReconnects != 2 {
		t.Fatalf("consecutive reconnects = %d, want %d", input.ConsecutiveReconnects, 2)
	}
	if input.ResyncCount != 1 {
		t.Fatalf("resync count = %d, want %d", input.ResyncCount, 1)
	}
}

func TestAdapterLoopStateSetConnectionStateUpdatesState(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.SetConnectionState(ingestion.ConnectionReconnecting); err != nil {
		t.Fatalf("set connection state: %v", err)
	}
	if state.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", state.ConnectionState, ingestion.ConnectionReconnecting)
	}
}

func TestAdapterLoopStateSetConnectionStateRejectsUnsupportedState(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.SetConnectionState(ingestion.ConnectionState("BROKEN")); err == nil {
		t.Fatal("expected unsupported connection state to fail")
	}
}

func TestAdapterLoopStateRecordMessageUpdatesTimestamp(t *testing.T) {
	state := &AdapterLoopState{}
	at := time.UnixMilli(1772798525000).UTC()
	if err := state.RecordMessage(at); err != nil {
		t.Fatalf("record message: %v", err)
	}
	if !state.LastMessageAt.Equal(at) {
		t.Fatalf("last message = %s, want %s", state.LastMessageAt, at)
	}
}

func TestAdapterLoopStateRecordMessageRejectsMissingTime(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.RecordMessage(time.Time{}); err == nil {
		t.Fatal("expected missing message time to fail")
	}
}

func TestAdapterLoopStateRecordSnapshotUpdatesTimestamp(t *testing.T) {
	state := &AdapterLoopState{}
	at := time.UnixMilli(1772798525000).UTC()
	if err := state.RecordSnapshot(at); err != nil {
		t.Fatalf("record snapshot: %v", err)
	}
	if !state.LastSnapshotAt.Equal(at) {
		t.Fatalf("last snapshot = %s, want %s", state.LastSnapshotAt, at)
	}
}

func TestAdapterLoopStateRecordSnapshotRejectsMissingTime(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.RecordSnapshot(time.Time{}); err == nil {
		t.Fatal("expected missing snapshot time to fail")
	}
}

func TestAdapterLoopStateRecordSnapshotRecoveryUpdatesHistory(t *testing.T) {
	state := &AdapterLoopState{}
	first := time.UnixMilli(1772798524000).UTC()
	second := time.UnixMilli(1772798525000).UTC()
	if err := state.RecordSnapshotRecovery(first); err != nil {
		t.Fatalf("record first snapshot recovery: %v", err)
	}
	if err := state.RecordSnapshotRecovery(second); err != nil {
		t.Fatalf("record second snapshot recovery: %v", err)
	}
	if !state.LastSnapshotRecoveryAt.Equal(second) {
		t.Fatalf("last snapshot recovery = %s, want %s", state.LastSnapshotRecoveryAt, second)
	}
	if len(state.RecentSnapshotRecoveries) != 2 {
		t.Fatalf("recent snapshot recoveries = %d, want %d", len(state.RecentSnapshotRecoveries), 2)
	}
	if !state.RecentSnapshotRecoveries[0].Equal(first) || !state.RecentSnapshotRecoveries[1].Equal(second) {
		t.Fatalf("snapshot recovery history = %v, want [%s %s]", state.RecentSnapshotRecoveries, first, second)
	}
}

func TestAdapterLoopStateRecordSnapshotRecoveryRejectsMissingTime(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.RecordSnapshotRecovery(time.Time{}); err == nil {
		t.Fatal("expected missing snapshot recovery time to fail")
	}
}

func TestAdapterLoopStateMarkSequenceGapSetsFlag(t *testing.T) {
	state := &AdapterLoopState{}
	if err := state.MarkSequenceGap(); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	if !state.SequenceGapDetected {
		t.Fatal("expected sequence gap to be marked")
	}
}

func TestAdapterLoopStateClearSequenceGapClearsFlag(t *testing.T) {
	state := &AdapterLoopState{SequenceGapDetected: true}
	if err := state.ClearSequenceGap(); err != nil {
		t.Fatalf("clear sequence gap: %v", err)
	}
	if state.SequenceGapDetected {
		t.Fatal("expected sequence gap to be cleared")
	}
}

func TestAdapterLoopStateIncrementReconnectCountUpdatesCount(t *testing.T) {
	state := &AdapterLoopState{ConsecutiveReconnects: 1}
	if err := state.IncrementReconnectCount(); err != nil {
		t.Fatalf("increment reconnect count: %v", err)
	}
	if state.ConsecutiveReconnects != 2 {
		t.Fatalf("consecutive reconnects = %d, want %d", state.ConsecutiveReconnects, 2)
	}
}

func TestAdapterLoopStateResetReconnectCountClearsCount(t *testing.T) {
	state := &AdapterLoopState{ConsecutiveReconnects: 3}
	if err := state.ResetReconnectCount(); err != nil {
		t.Fatalf("reset reconnect count: %v", err)
	}
	if state.ConsecutiveReconnects != 0 {
		t.Fatalf("consecutive reconnects = %d, want %d", state.ConsecutiveReconnects, 0)
	}
}

func TestAdapterLoopStateIncrementResyncCountUpdatesCount(t *testing.T) {
	state := &AdapterLoopState{ResyncCount: 1}
	if err := state.IncrementResyncCount(); err != nil {
		t.Fatalf("increment resync count: %v", err)
	}
	if state.ResyncCount != 2 {
		t.Fatalf("resync count = %d, want %d", state.ResyncCount, 2)
	}
}

func TestAdapterLoopStateResetResyncCountClearsCount(t *testing.T) {
	state := &AdapterLoopState{ResyncCount: 3}
	if err := state.ResetResyncCount(); err != nil {
		t.Fatalf("reset resync count: %v", err)
	}
	if state.ResyncCount != 0 {
		t.Fatalf("resync count = %d, want %d", state.ResyncCount, 0)
	}
}

func TestAdapterLoopStatePruneSnapshotRecoveryHistoryRemovesAttemptsOutsideWindow(t *testing.T) {
	now := time.UnixMilli(1772798525000).UTC()
	state := &AdapterLoopState{
		LastSnapshotRecoveryAt: now.Add(-5 * time.Second),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-61 * time.Second),
			now.Add(-time.Minute),
			now.Add(-5 * time.Second),
		},
	}

	if err := state.PruneSnapshotRecoveryHistory(now); err != nil {
		t.Fatalf("prune snapshot recovery history: %v", err)
	}
	if !state.LastSnapshotRecoveryAt.Equal(now.Add(-5 * time.Second)) {
		t.Fatalf("last snapshot recovery = %s, want %s", state.LastSnapshotRecoveryAt, now.Add(-5*time.Second))
	}
	if len(state.RecentSnapshotRecoveries) != 2 {
		t.Fatalf("recent snapshot recoveries = %d, want %d", len(state.RecentSnapshotRecoveries), 2)
	}
	if !state.RecentSnapshotRecoveries[0].Equal(now.Add(-time.Minute)) {
		t.Fatalf("first recent snapshot recovery = %s, want %s", state.RecentSnapshotRecoveries[0], now.Add(-time.Minute))
	}
	if !state.RecentSnapshotRecoveries[1].Equal(now.Add(-5 * time.Second)) {
		t.Fatalf("second recent snapshot recovery = %s, want %s", state.RecentSnapshotRecoveries[1], now.Add(-5*time.Second))
	}
}

func TestAdapterLoopStatePruneSnapshotRecoveryHistoryRejectsFutureAttempt(t *testing.T) {
	now := time.UnixMilli(1772798525000).UTC()
	state := &AdapterLoopState{RecentSnapshotRecoveries: []time.Time{now.Add(time.Second)}}
	if err := state.PruneSnapshotRecoveryHistory(now); err == nil {
		t.Fatal("expected future snapshot recovery attempt to fail")
	}
}

func TestRuntimeAdapterHealthSnapshotComposesHealthyStatuses(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	snapshot, err := runtime.AdapterHealthSnapshot(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionConnected,
		Now:                    now,
		LastMessageAt:          now.Add(-5 * time.Second),
		LastSnapshotAt:         now.Add(-10 * time.Second),
		SequenceGapDetected:    false,
		LocalClockOffset:       50 * time.Millisecond,
		LastSnapshotRecoveryAt: now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-61 * time.Second),
		},
		ConsecutiveReconnects: 1,
		ResyncCount:           1,
	})
	if err != nil {
		t.Fatalf("adapter health snapshot: %v", err)
	}

	if snapshot.Staleness.State != ingestion.FeedHealthHealthy {
		t.Fatalf("staleness state = %q, want %q", snapshot.Staleness.State, ingestion.FeedHealthHealthy)
	}
	if snapshot.ConnectionState != ingestion.ConnectionConnected {
		t.Fatalf("connection state = %q, want %q", snapshot.ConnectionState, ingestion.ConnectionConnected)
	}
	if snapshot.Staleness.ConnectionState != ingestion.ConnectionConnected {
		t.Fatalf("staleness connection state = %q, want %q", snapshot.Staleness.ConnectionState, ingestion.ConnectionConnected)
	}
	if snapshot.SequenceGapDetected {
		t.Fatal("expected sequence gap to be false")
	}
	if snapshot.Staleness.SequenceGapDetected {
		t.Fatal("expected staleness sequence gap to be false")
	}
	if snapshot.ClockOffset.State != ingestion.ClockNormal {
		t.Fatalf("clock state = %q, want %q", snapshot.ClockOffset.State, ingestion.ClockNormal)
	}
	if snapshot.ReconnectLoop.LoopDetected {
		t.Fatal("expected reconnect loop to be false")
	}
	if snapshot.ResyncLoop.LoopDetected {
		t.Fatal("expected resync loop to be false")
	}
	if snapshot.SnapshotRecoveryCooldown.Ready {
		t.Fatal("expected snapshot cooldown to still be active")
	}
	if !snapshot.SnapshotRecoveryRateLimit.Allowed {
		t.Fatal("expected snapshot recovery rate-limit to allow attempt")
	}
}

func TestRuntimeAdapterHealthSnapshotComposesDegradedStatuses(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	snapshot, err := runtime.AdapterHealthSnapshot(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionReconnecting,
		Now:                    now,
		LastMessageAt:          now.Add(-16 * time.Second),
		LastSnapshotAt:         now.Add(-31 * time.Second),
		SequenceGapDetected:    true,
		LocalClockOffset:       250 * time.Millisecond,
		LastSnapshotRecoveryAt: now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-50 * time.Second),
			now.Add(-10 * time.Second),
		},
		ConsecutiveReconnects: 4,
		ResyncCount:           3,
	})
	if err != nil {
		t.Fatalf("adapter health snapshot: %v", err)
	}

	if snapshot.Staleness.State != ingestion.FeedHealthStale {
		t.Fatalf("staleness state = %q, want %q", snapshot.Staleness.State, ingestion.FeedHealthStale)
	}
	if snapshot.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", snapshot.ConnectionState, ingestion.ConnectionReconnecting)
	}
	if !snapshot.SequenceGapDetected {
		t.Fatal("expected sequence gap to be true")
	}
	if snapshot.ClockOffset.State != ingestion.ClockDegraded {
		t.Fatalf("clock state = %q, want %q", snapshot.ClockOffset.State, ingestion.ClockDegraded)
	}
	if !snapshot.ReconnectLoop.LoopDetected {
		t.Fatal("expected reconnect loop to be detected")
	}
	if !snapshot.ResyncLoop.LoopDetected {
		t.Fatal("expected resync loop to be detected")
	}
	if snapshot.SnapshotRecoveryCooldown.Ready {
		t.Fatal("expected snapshot cooldown to still be active")
	}
	if snapshot.SnapshotRecoveryCooldown.RemainingCooldown != 500*time.Millisecond {
		t.Fatalf("remaining cooldown = %s, want %s", snapshot.SnapshotRecoveryCooldown.RemainingCooldown, 500*time.Millisecond)
	}
	if snapshot.SnapshotRecoveryRateLimit.Allowed {
		t.Fatal("expected snapshot recovery rate-limit to block attempt")
	}
	if snapshot.SnapshotRecoveryRateLimit.RetryAfter != 10*time.Second {
		t.Fatalf("retry after = %s, want %s", snapshot.SnapshotRecoveryRateLimit.RetryAfter, 10*time.Second)
	}
}

func TestRuntimeAdapterHealthSnapshotPropagatesInvalidInput(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.AdapterHealthSnapshot(AdapterHealthInput{
		ConnectionState:       ingestion.ConnectionConnected,
		Now:                   time.UnixMilli(1772798525000).UTC(),
		LastMessageAt:         time.UnixMilli(1772798525000).UTC().Add(time.Second),
		LastSnapshotAt:        time.UnixMilli(1772798525000).UTC(),
		SequenceGapDetected:   false,
		LocalClockOffset:      0,
		ConsecutiveReconnects: 0,
		ResyncCount:           0,
	})
	if err == nil {
		t.Fatal("expected invalid adapter health input to fail")
	}
}

func TestRuntimeAdapterHealthSnapshotRejectsUnsupportedConnectionState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 1
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.AdapterHealthSnapshot(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionState("BROKEN"),
		Now:                    time.UnixMilli(1772798525000).UTC(),
		LastMessageAt:          time.UnixMilli(1772798520000).UTC(),
		LastSnapshotAt:         time.UnixMilli(1772798520000).UTC(),
		SequenceGapDetected:    false,
		LastSnapshotRecoveryAt: time.UnixMilli(1772798520000).UTC(),
	})
	if err == nil {
		t.Fatal("expected unsupported connection state to fail")
	}
}

func TestRuntimeEvaluateClockOffsetReturnsNormalBelowWarningThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateClockOffset(50 * time.Millisecond)
	if err != nil {
		t.Fatalf("evaluate clock offset: %v", err)
	}

	if status.State != ingestion.ClockNormal {
		t.Fatalf("state = %q, want %q", status.State, ingestion.ClockNormal)
	}
}

func TestRuntimeEvaluateClockOffsetReturnsWarningAtThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateClockOffset(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("evaluate clock offset: %v", err)
	}

	if status.State != ingestion.ClockWarning {
		t.Fatalf("state = %q, want %q", status.State, ingestion.ClockWarning)
	}
}

func TestRuntimeEvaluateClockOffsetReturnsDegradedAtThreshold(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateClockOffset(250 * time.Millisecond)
	if err != nil {
		t.Fatalf("evaluate clock offset: %v", err)
	}

	if status.State != ingestion.ClockDegraded {
		t.Fatalf("state = %q, want %q", status.State, ingestion.ClockDegraded)
	}
}

func TestRuntimeEvaluateClockOffsetUsesAbsoluteOffset(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	status, err := runtime.EvaluateClockOffset(-150 * time.Millisecond)
	if err != nil {
		t.Fatalf("evaluate clock offset: %v", err)
	}

	if status.State != ingestion.ClockWarning {
		t.Fatalf("state = %q, want %q", status.State, ingestion.ClockWarning)
	}
}

func TestRuntimeEvaluateClockOffsetRejectsInvalidThresholdState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	runtime.config.Adapter.ClockOffsetWarning = 0

	if _, err := runtime.EvaluateClockOffset(50 * time.Millisecond); err == nil {
		t.Fatal("expected invalid clock threshold state to fail")
	}
}

func TestRuntimeEvaluateAdapterDecisionReturnsHealthyWhenSnapshotIsHealthy(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	decision, err := runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionConnected,
		SequenceGapDetected: false,
		Staleness: ingestion.FeedHealthStatus{
			State:             ingestion.FeedHealthHealthy,
			MessageFreshness:  ingestion.FreshnessFresh,
			SnapshotFreshness: ingestion.FreshnessFresh,
		},
		ClockOffset:   ClockOffsetStatus{State: ingestion.ClockNormal},
		ReconnectLoop: ReconnectLoopStatus{LoopDetected: false},
		ResyncLoop:    ResyncLoopStatus{LoopDetected: false},
	})
	if err != nil {
		t.Fatalf("evaluate adapter decision: %v", err)
	}

	if decision.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthHealthy)
	}
	if decision.ConnectionState != ingestion.ConnectionConnected {
		t.Fatalf("connection state = %q, want %q", decision.ConnectionState, ingestion.ConnectionConnected)
	}
	if decision.SequenceGapDetected {
		t.Fatal("expected sequence gap to be false")
	}
	if decision.ClockState != ingestion.ClockNormal {
		t.Fatalf("clock state = %q, want %q", decision.ClockState, ingestion.ClockNormal)
	}
	if len(decision.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", decision.Reasons)
	}
}

func TestRuntimeEvaluateAdapterDecisionReturnsDegradedForLoopAndClockReasons(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	decision, err := runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionReconnecting,
		SequenceGapDetected: true,
		Staleness: ingestion.FeedHealthStatus{
			State:             ingestion.FeedHealthHealthy,
			MessageFreshness:  ingestion.FreshnessFresh,
			SnapshotFreshness: ingestion.FreshnessFresh,
		},
		ClockOffset:   ClockOffsetStatus{State: ingestion.ClockDegraded},
		ReconnectLoop: ReconnectLoopStatus{LoopDetected: true},
		ResyncLoop:    ResyncLoopStatus{LoopDetected: true},
	})
	if err != nil {
		t.Fatalf("evaluate adapter decision: %v", err)
	}

	if decision.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthDegraded)
	}
	if decision.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", decision.ConnectionState, ingestion.ConnectionReconnecting)
	}
	if !decision.SequenceGapDetected {
		t.Fatal("expected sequence gap to be true")
	}
	if decision.ClockState != ingestion.ClockDegraded {
		t.Fatalf("clock state = %q, want %q", decision.ClockState, ingestion.ClockDegraded)
	}
	wantReasons := []ingestion.DegradationReason{
		ingestion.ReasonClockDegraded,
		ingestion.ReasonConnectionNotReady,
		ingestion.ReasonReconnectLoop,
		ingestion.ReasonResyncLoop,
		ingestion.ReasonSequenceGap,
	}
	for _, reason := range wantReasons {
		if !hasReason(decision.Reasons, reason) {
			t.Fatalf("reasons = %v, want %q", decision.Reasons, reason)
		}
	}
}

func TestRuntimeEvaluateAdapterDecisionReturnsDegradedForSequenceGap(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	decision, err := runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionConnected,
		SequenceGapDetected: true,
		Staleness: ingestion.FeedHealthStatus{
			State:               ingestion.FeedHealthHealthy,
			ConnectionState:     ingestion.ConnectionConnected,
			MessageFreshness:    ingestion.FreshnessFresh,
			SnapshotFreshness:   ingestion.FreshnessFresh,
			SequenceGapDetected: false,
		},
		ClockOffset:   ClockOffsetStatus{State: ingestion.ClockNormal},
		ReconnectLoop: ReconnectLoopStatus{LoopDetected: false},
		ResyncLoop:    ResyncLoopStatus{LoopDetected: false},
	})
	if err != nil {
		t.Fatalf("evaluate adapter decision: %v", err)
	}

	if decision.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthDegraded)
	}
	if !decision.SequenceGapDetected {
		t.Fatal("expected sequence gap to be true")
	}
	if !hasReason(decision.Reasons, ingestion.ReasonSequenceGap) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonSequenceGap)
	}
}

func TestRuntimeEvaluateAdapterDecisionPreservesStalePrecedence(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	decision, err := runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionResyncing,
		SequenceGapDetected: true,
		Staleness: ingestion.FeedHealthStatus{
			State:             ingestion.FeedHealthStale,
			MessageFreshness:  ingestion.FreshnessStale,
			SnapshotFreshness: ingestion.FreshnessFresh,
			Reasons:           []ingestion.DegradationReason{ingestion.ReasonMessageStale},
		},
		ClockOffset:   ClockOffsetStatus{State: ingestion.ClockDegraded},
		ReconnectLoop: ReconnectLoopStatus{LoopDetected: true},
		ResyncLoop:    ResyncLoopStatus{LoopDetected: false},
	})
	if err != nil {
		t.Fatalf("evaluate adapter decision: %v", err)
	}

	if decision.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthStale)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonMessageStale)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonConnectionNotReady) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonConnectionNotReady)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonSequenceGap) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonSequenceGap)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonReconnectLoop)
	}
}

func TestRuntimeEvaluateAdapterDecisionRejectsUnsupportedState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionConnected,
		SequenceGapDetected: false,
		Staleness:           ingestion.FeedHealthStatus{State: ingestion.FeedHealthState("BROKEN")},
		ClockOffset:         ClockOffsetStatus{State: ingestion.ClockNormal},
	})
	if err == nil {
		t.Fatal("expected unsupported feed health state to fail")
	}
}

func TestRuntimeEvaluateAdapterDecisionRejectsUnsupportedClockState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionConnected,
		SequenceGapDetected: false,
		Staleness:           ingestion.FeedHealthStatus{State: ingestion.FeedHealthHealthy},
		ClockOffset:         ClockOffsetStatus{State: ingestion.ClockState("BROKEN")},
	})
	if err == nil {
		t.Fatal("expected unsupported clock state to fail")
	}
}

func TestRuntimeEvaluateAdapterDecisionRejectsUnsupportedConnectionState(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.EvaluateAdapterDecision(AdapterHealthSnapshot{
		ConnectionState:     ingestion.ConnectionState("BROKEN"),
		SequenceGapDetected: false,
		Staleness:           ingestion.FeedHealthStatus{State: ingestion.FeedHealthHealthy},
		ClockOffset:         ClockOffsetStatus{State: ingestion.ClockNormal},
	})
	if err == nil {
		t.Fatal("expected unsupported connection state to fail")
	}
}

func TestRuntimeEvaluateAdapterInputReturnsHealthyDecision(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionConnected,
		Now:                    now,
		LastMessageAt:          now.Add(-5 * time.Second),
		LastSnapshotAt:         now.Add(-10 * time.Second),
		SequenceGapDetected:    false,
		LocalClockOffset:       50 * time.Millisecond,
		LastSnapshotRecoveryAt: now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-61 * time.Second),
		},
		ConsecutiveReconnects: 1,
		ResyncCount:           1,
	})
	if err != nil {
		t.Fatalf("evaluate adapter input: %v", err)
	}

	if decision.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthHealthy)
	}
	if decision.ConnectionState != ingestion.ConnectionConnected {
		t.Fatalf("connection state = %q, want %q", decision.ConnectionState, ingestion.ConnectionConnected)
	}
	if decision.SequenceGapDetected {
		t.Fatal("expected sequence gap to be false")
	}
}

func TestRuntimeEvaluateAdapterInputReturnsDegradedDecision(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionReconnecting,
		Now:                    now,
		LastMessageAt:          now.Add(-5 * time.Second),
		LastSnapshotAt:         now.Add(-10 * time.Second),
		SequenceGapDetected:    true,
		LocalClockOffset:       250 * time.Millisecond,
		LastSnapshotRecoveryAt: now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-50 * time.Second),
			now.Add(-10 * time.Second),
		},
		ConsecutiveReconnects: 4,
		ResyncCount:           3,
	})
	if err != nil {
		t.Fatalf("evaluate adapter input: %v", err)
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

func TestRuntimeEvaluateAdapterInputPreservesStalePrecedence(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:        ingestion.ConnectionResyncing,
		Now:                    now,
		LastMessageAt:          now.Add(-16 * time.Second),
		LastSnapshotAt:         now.Add(-10 * time.Second),
		SequenceGapDetected:    true,
		LocalClockOffset:       250 * time.Millisecond,
		LastSnapshotRecoveryAt: now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{
			now.Add(-61 * time.Second),
		},
		ConsecutiveReconnects: 4,
		ResyncCount:           1,
	})
	if err != nil {
		t.Fatalf("evaluate adapter input: %v", err)
	}

	if decision.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthStale)
	}
	if !hasReason(decision.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", decision.Reasons, ingestion.ReasonMessageStale)
	}
}

func TestRuntimeEvaluateAdapterInputPropagatesInvalidInput(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState: ingestion.ConnectionState("BROKEN"),
	})
	if err == nil {
		t.Fatal("expected invalid adapter input to fail")
	}
}

func TestRuntimeEvaluateLoopStateReturnsHealthyDecision(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState:          ingestion.ConnectionConnected,
		LastMessageAt:            now.Add(-5 * time.Second),
		LastSnapshotAt:           now.Add(-10 * time.Second),
		LocalClockOffset:         50 * time.Millisecond,
		LastSnapshotRecoveryAt:   now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{now.Add(-61 * time.Second)},
		ConsecutiveReconnects:    1,
		ResyncCount:              1,
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}

	if decision.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", decision.State, ingestion.FeedHealthHealthy)
	}
}

func TestRuntimeEvaluateLoopStateTracksDegradedTransition(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	state := AdapterLoopState{
		ConnectionState:          ingestion.ConnectionConnected,
		LastMessageAt:            now.Add(-5 * time.Second),
		LastSnapshotAt:           now.Add(-10 * time.Second),
		LocalClockOffset:         50 * time.Millisecond,
		LastSnapshotRecoveryAt:   now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{now.Add(-61 * time.Second)},
		ConsecutiveReconnects:    1,
		ResyncCount:              1,
	}

	if err := state.SetConnectionState(ingestion.ConnectionReconnecting); err != nil {
		t.Fatalf("set connection state: %v", err)
	}
	if err := state.MarkSequenceGap(); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	state.LocalClockOffset = 250 * time.Millisecond
	for range 3 {
		if err := state.IncrementReconnectCount(); err != nil {
			t.Fatalf("increment reconnect count: %v", err)
		}
	}
	for range 2 {
		if err := state.IncrementResyncCount(); err != nil {
			t.Fatalf("increment resync count: %v", err)
		}
	}

	decision, err := runtime.EvaluateLoopState(state, now)
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

func TestRuntimeEvaluateLoopStateTracksStaleTransition(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	decision, err := runtime.EvaluateLoopState(AdapterLoopState{
		ConnectionState:          ingestion.ConnectionResyncing,
		LastMessageAt:            now.Add(-16 * time.Second),
		LastSnapshotAt:           now.Add(-10 * time.Second),
		SequenceGapDetected:      true,
		LocalClockOffset:         250 * time.Millisecond,
		LastSnapshotRecoveryAt:   now.Add(-500 * time.Millisecond),
		RecentSnapshotRecoveries: []time.Time{now.Add(-61 * time.Second)},
		ConsecutiveReconnects:    4,
		ResyncCount:              1,
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

func TestRuntimeEvaluateLoopStateRecoversAfterClearingGapAndResettingCounters(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	state := &AdapterLoopState{
		ConnectionState:          ingestion.ConnectionConnected,
		LastMessageAt:            now.Add(-5 * time.Second),
		LastSnapshotAt:           now.Add(-10 * time.Second),
		LocalClockOffset:         50 * time.Millisecond,
		LastSnapshotRecoveryAt:   now.Add(-5 * time.Second),
		RecentSnapshotRecoveries: []time.Time{now.Add(-61 * time.Second), now.Add(-5 * time.Second)},
	}

	if err := state.MarkSequenceGap(); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	for range 4 {
		if err := state.IncrementReconnectCount(); err != nil {
			t.Fatalf("increment reconnect count: %v", err)
		}
	}
	for range 3 {
		if err := state.IncrementResyncCount(); err != nil {
			t.Fatalf("increment resync count: %v", err)
		}
	}
	degraded, err := runtime.EvaluateLoopState(*state, now)
	if err != nil {
		t.Fatalf("evaluate degraded loop state: %v", err)
	}
	if degraded.State != ingestion.FeedHealthDegraded {
		t.Fatalf("degraded state = %q, want %q", degraded.State, ingestion.FeedHealthDegraded)
	}

	if err := state.ClearSequenceGap(); err != nil {
		t.Fatalf("clear sequence gap: %v", err)
	}
	if err := state.ResetReconnectCount(); err != nil {
		t.Fatalf("reset reconnect count: %v", err)
	}
	if err := state.ResetResyncCount(); err != nil {
		t.Fatalf("reset resync count: %v", err)
	}
	if err := state.PruneSnapshotRecoveryHistory(now); err != nil {
		t.Fatalf("prune snapshot recovery history: %v", err)
	}

	recovered, err := runtime.EvaluateLoopState(*state, now)
	if err != nil {
		t.Fatalf("evaluate recovered loop state: %v", err)
	}
	if recovered.State != ingestion.FeedHealthHealthy {
		t.Fatalf("recovered state = %q, want %q", recovered.State, ingestion.FeedHealthHealthy)
	}
	if len(recovered.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", recovered.Reasons)
	}
}

func TestRuntimeEvaluateLoopStatePropagatesInvalidInput(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	_, err = runtime.EvaluateLoopState(AdapterLoopState{ConnectionState: ingestion.ConnectionState("BROKEN")}, time.UnixMilli(1772798525000).UTC())
	if err == nil {
		t.Fatal("expected invalid loop state to fail")
	}
}

func TestNewAdapterLoopRejectsMissingRuntime(t *testing.T) {
	if _, err := NewAdapterLoop(nil); err == nil {
		t.Fatal("expected missing runtime to fail")
	}
}

func TestAdapterLoopDecisionMatchesRuntimeHelpers(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	loop, err := NewAdapterLoop(runtime)
	if err != nil {
		t.Fatalf("new adapter loop: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	state := loop.State()
	if state == nil {
		t.Fatal("expected adapter loop state")
	}
	if err := state.SetConnectionState(ingestion.ConnectionReconnecting); err != nil {
		t.Fatalf("set connection state: %v", err)
	}
	if err := state.RecordMessage(now.Add(-5 * time.Second)); err != nil {
		t.Fatalf("record message: %v", err)
	}
	if err := state.RecordSnapshot(now.Add(-10 * time.Second)); err != nil {
		t.Fatalf("record snapshot: %v", err)
	}
	if err := state.RecordSnapshotRecovery(now.Add(-61 * time.Second)); err != nil {
		t.Fatalf("record old snapshot recovery: %v", err)
	}
	if err := state.RecordSnapshotRecovery(now.Add(-5 * time.Second)); err != nil {
		t.Fatalf("record recent snapshot recovery: %v", err)
	}
	if err := state.MarkSequenceGap(); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	for range 4 {
		if err := state.IncrementReconnectCount(); err != nil {
			t.Fatalf("increment reconnect count: %v", err)
		}
	}
	for range 3 {
		if err := state.IncrementResyncCount(); err != nil {
			t.Fatalf("increment resync count: %v", err)
		}
	}
	state.LocalClockOffset = 250 * time.Millisecond

	decision, err := loop.Decision(now)
	if err != nil {
		t.Fatalf("adapter loop decision: %v", err)
	}
	want, err := runtime.EvaluateLoopState(*state, now)
	if err != nil {
		t.Fatalf("runtime evaluate loop state: %v", err)
	}
	if !reflect.DeepEqual(decision, want) {
		t.Fatalf("decision = %#v, want %#v", decision, want)
	}
	if len(state.RecentSnapshotRecoveries) != 1 {
		t.Fatalf("recent snapshot recoveries = %d, want %d", len(state.RecentSnapshotRecoveries), 1)
	}
	if !state.RecentSnapshotRecoveries[0].Equal(now.Add(-5 * time.Second)) {
		t.Fatalf("recent snapshot recovery = %s, want %s", state.RecentSnapshotRecoveries[0], now.Add(-5*time.Second))
	}
}

func TestAdapterLoopDecisionTracksHealthyDegradedStaleAndRecoveryTransitions(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.SnapshotRecoveryPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	loop, err := NewAdapterLoop(runtime)
	if err != nil {
		t.Fatalf("new adapter loop: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	state := loop.State()
	if err := state.SetConnectionState(ingestion.ConnectionConnected); err != nil {
		t.Fatalf("set connection state: %v", err)
	}
	if err := state.RecordMessage(base.Add(-5 * time.Second)); err != nil {
		t.Fatalf("record message: %v", err)
	}
	if err := state.RecordSnapshot(base.Add(-10 * time.Second)); err != nil {
		t.Fatalf("record snapshot: %v", err)
	}

	healthy, err := loop.Decision(base)
	if err != nil {
		t.Fatalf("healthy decision: %v", err)
	}
	if healthy.State != ingestion.FeedHealthHealthy {
		t.Fatalf("healthy state = %q, want %q", healthy.State, ingestion.FeedHealthHealthy)
	}

	if err := state.MarkSequenceGap(); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	degraded, err := loop.Decision(base)
	if err != nil {
		t.Fatalf("degraded decision: %v", err)
	}
	if degraded.State != ingestion.FeedHealthDegraded {
		t.Fatalf("degraded state = %q, want %q", degraded.State, ingestion.FeedHealthDegraded)
	}
	if !hasReason(degraded.Reasons, ingestion.ReasonSequenceGap) {
		t.Fatalf("reasons = %v, want %q", degraded.Reasons, ingestion.ReasonSequenceGap)
	}

	stale, err := loop.Decision(base.Add(11 * time.Second))
	if err != nil {
		t.Fatalf("stale decision: %v", err)
	}
	if stale.State != ingestion.FeedHealthStale {
		t.Fatalf("stale state = %q, want %q", stale.State, ingestion.FeedHealthStale)
	}
	if !hasReason(stale.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", stale.Reasons, ingestion.ReasonMessageStale)
	}

	if err := state.ClearSequenceGap(); err != nil {
		t.Fatalf("clear sequence gap: %v", err)
	}
	if err := state.RecordMessage(base.Add(12 * time.Second)); err != nil {
		t.Fatalf("record fresh message: %v", err)
	}
	if err := state.RecordSnapshot(base.Add(11 * time.Second)); err != nil {
		t.Fatalf("record fresh snapshot: %v", err)
	}
	recovered, err := loop.Decision(base.Add(12 * time.Second))
	if err != nil {
		t.Fatalf("recovered decision: %v", err)
	}
	if recovered.State != ingestion.FeedHealthHealthy {
		t.Fatalf("recovered state = %q, want %q", recovered.State, ingestion.FeedHealthHealthy)
	}
	if len(recovered.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", recovered.Reasons)
	}
}

func TestNewRuntimeRejectsWrongVenue(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.Venue = ingestion.VenueKraken

	if _, err := NewRuntime(config); err == nil {
		t.Fatal("expected non-binance runtime config to fail")
	}
}

func loadBinanceRuntimeConfig(t *testing.T) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}

	runtimeConfig, err := config.RuntimeConfigFor(ingestion.VenueBinance)
	if err != nil {
		t.Fatalf("load binance runtime config: %v", err)
	}
	return runtimeConfig
}

func hasReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
