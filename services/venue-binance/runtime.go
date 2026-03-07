package venuebinance

import (
	"fmt"
	"sort"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type Runtime struct {
	config ingestion.VenueRuntimeConfig
}

type SnapshotRecoveryStatus struct {
	Ready             bool
	RemainingCooldown time.Duration
}

type SnapshotRecoveryRateLimitStatus struct {
	Allowed            bool
	AttemptsInWindow   int
	RemainingAllowance int
	RetryAfter         time.Duration
}

type ReconnectLoopStatus struct {
	LoopDetected          bool
	ConsecutiveReconnects int
	Threshold             int
	RemainingBeforeLoop   int
}

type ResyncLoopStatus struct {
	LoopDetected        bool
	ResyncCount         int
	Threshold           int
	RemainingBeforeLoop int
}

type AdapterHealthInput struct {
	ConnectionState          ingestion.ConnectionState
	Now                      time.Time
	LastMessageAt            time.Time
	LastSnapshotAt           time.Time
	SequenceGapDetected      bool
	LocalClockOffset         time.Duration
	LastSnapshotRecoveryAt   time.Time
	RecentSnapshotRecoveries []time.Time
	ConsecutiveReconnects    int
	ResyncCount              int
}

type AdapterHealthSnapshot struct {
	ConnectionState           ingestion.ConnectionState
	SequenceGapDetected       bool
	Staleness                 ingestion.FeedHealthStatus
	ClockOffset               ClockOffsetStatus
	ReconnectLoop             ReconnectLoopStatus
	ResyncLoop                ResyncLoopStatus
	SnapshotRecoveryCooldown  SnapshotRecoveryStatus
	SnapshotRecoveryRateLimit SnapshotRecoveryRateLimitStatus
}

type ClockOffsetStatus struct {
	State             ingestion.ClockState
	Offset            time.Duration
	WarningThreshold  time.Duration
	DegradedThreshold time.Duration
}

type AdapterLoopState struct {
	ConnectionState          ingestion.ConnectionState
	LastMessageAt            time.Time
	LastSnapshotAt           time.Time
	SequenceGapDetected      bool
	LocalClockOffset         time.Duration
	LastSnapshotRecoveryAt   time.Time
	RecentSnapshotRecoveries []time.Time
	ConsecutiveReconnects    int
	ResyncCount              int
}

type AdapterLoop struct {
	runtime *Runtime
	state   AdapterLoopState
}

func NewAdapterLoop(runtime *Runtime) (*AdapterLoop, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}

	return &AdapterLoop{
		runtime: runtime,
		state: AdapterLoopState{
			ConnectionState: ingestion.ConnectionDisconnected,
		},
	}, nil
}

func (l *AdapterLoop) State() *AdapterLoopState {
	if l == nil {
		return nil
	}
	return &l.state
}

func (l *AdapterLoop) Decision(now time.Time) (ingestion.FeedHealthStatus, error) {
	if l == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("adapter loop is required")
	}
	if err := l.state.PruneSnapshotRecoveryHistory(now); err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	return l.runtime.EvaluateLoopState(l.state, now)
}

func (s *AdapterLoopState) SetConnectionState(state ingestion.ConnectionState) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if err := validateConnectionState(state); err != nil {
		return err
	}
	s.ConnectionState = state
	return nil
}

func (s *AdapterLoopState) RecordMessage(at time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if at.IsZero() {
		return fmt.Errorf("message time is required")
	}
	s.LastMessageAt = at
	return nil
}

func (s *AdapterLoopState) RecordSnapshot(at time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if at.IsZero() {
		return fmt.Errorf("snapshot time is required")
	}
	s.LastSnapshotAt = at
	return nil
}

func (s *AdapterLoopState) RecordSnapshotRecovery(at time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if at.IsZero() {
		return fmt.Errorf("snapshot recovery time is required")
	}
	s.LastSnapshotRecoveryAt = at
	s.RecentSnapshotRecoveries = append(s.RecentSnapshotRecoveries, at)
	return nil
}

func (s *AdapterLoopState) MarkSequenceGap() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.SequenceGapDetected = true
	return nil
}

func (s *AdapterLoopState) ClearSequenceGap() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.SequenceGapDetected = false
	return nil
}

func (s *AdapterLoopState) IncrementReconnectCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ConsecutiveReconnects++
	return nil
}

func (s *AdapterLoopState) ResetReconnectCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ConsecutiveReconnects = 0
	return nil
}

func (s *AdapterLoopState) IncrementResyncCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ResyncCount++
	return nil
}

func (s *AdapterLoopState) ResetResyncCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ResyncCount = 0
	return nil
}

func (s *AdapterLoopState) PruneSnapshotRecoveryHistory(now time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if now.IsZero() {
		return fmt.Errorf("current time is required")
	}

	windowStart := now.Add(-time.Minute)
	pruned := make([]time.Time, 0, len(s.RecentSnapshotRecoveries))
	for _, attempt := range s.RecentSnapshotRecoveries {
		if attempt.IsZero() {
			return fmt.Errorf("snapshot recovery attempt time is required")
		}
		if attempt.After(now) {
			return fmt.Errorf("snapshot recovery attempt cannot be in the future")
		}
		if attempt.Before(windowStart) {
			continue
		}
		pruned = append(pruned, attempt)
	}

	if len(pruned) == 0 {
		s.RecentSnapshotRecoveries = nil
		return nil
	}
	s.RecentSnapshotRecoveries = pruned
	return nil
}

func (s AdapterLoopState) HealthInput(now time.Time) AdapterHealthInput {
	return AdapterHealthInput{
		ConnectionState:          s.ConnectionState,
		Now:                      now,
		LastMessageAt:            s.LastMessageAt,
		LastSnapshotAt:           s.LastSnapshotAt,
		SequenceGapDetected:      s.SequenceGapDetected,
		LocalClockOffset:         s.LocalClockOffset,
		LastSnapshotRecoveryAt:   s.LastSnapshotRecoveryAt,
		RecentSnapshotRecoveries: append([]time.Time(nil), s.RecentSnapshotRecoveries...),
		ConsecutiveReconnects:    s.ConsecutiveReconnects,
		ResyncCount:              s.ResyncCount,
	}
}

func NewRuntime(config ingestion.VenueRuntimeConfig) (*Runtime, error) {
	if config.Venue != ingestion.VenueBinance {
		return nil, fmt.Errorf("binance runtime requires %q venue config", ingestion.VenueBinance)
	}
	if err := config.Adapter.Validate(); err != nil {
		return nil, err
	}
	if config.ServicePath == "" {
		return nil, fmt.Errorf("service path is required")
	}

	return &Runtime{config: config}, nil
}

func (r *Runtime) ReconnectDelay(attempt int) (time.Duration, error) {
	if r == nil {
		return 0, fmt.Errorf("runtime is required")
	}
	return ingestion.ReconnectDelay(attempt, r.config.Adapter.ReconnectBackoffMin, r.config.Adapter.ReconnectBackoffMax)
}

func (r *Runtime) SnapshotRecoveryStatus(now, lastAttempt time.Time) (SnapshotRecoveryStatus, error) {
	if r == nil {
		return SnapshotRecoveryStatus{}, fmt.Errorf("runtime is required")
	}
	if now.IsZero() {
		return SnapshotRecoveryStatus{}, fmt.Errorf("current time is required")
	}
	if r.config.SnapshotCooldown < 0 {
		return SnapshotRecoveryStatus{}, fmt.Errorf("snapshot cooldown must be non-negative")
	}
	if lastAttempt.IsZero() {
		return SnapshotRecoveryStatus{Ready: true}, nil
	}
	if lastAttempt.After(now) {
		return SnapshotRecoveryStatus{}, fmt.Errorf("last snapshot attempt cannot be in the future")
	}

	remaining := r.config.SnapshotCooldown - now.Sub(lastAttempt)
	if remaining <= 0 {
		return SnapshotRecoveryStatus{Ready: true}, nil
	}

	return SnapshotRecoveryStatus{
		Ready:             false,
		RemainingCooldown: remaining,
	}, nil
}

func (r *Runtime) SnapshotRecoveryRateLimitStatus(now time.Time, recentAttempts []time.Time) (SnapshotRecoveryRateLimitStatus, error) {
	if r == nil {
		return SnapshotRecoveryRateLimitStatus{}, fmt.Errorf("runtime is required")
	}
	if now.IsZero() {
		return SnapshotRecoveryRateLimitStatus{}, fmt.Errorf("current time is required")
	}
	if r.config.SnapshotRecoveryPerMinuteLimit <= 0 {
		return SnapshotRecoveryRateLimitStatus{}, fmt.Errorf("snapshot recovery per-minute limit must be positive")
	}

	windowStart := now.Add(-time.Minute)
	attemptsInWindow := make([]time.Time, 0, len(recentAttempts))
	for _, attempt := range recentAttempts {
		if attempt.IsZero() {
			return SnapshotRecoveryRateLimitStatus{}, fmt.Errorf("snapshot recovery attempt time is required")
		}
		if attempt.After(now) {
			return SnapshotRecoveryRateLimitStatus{}, fmt.Errorf("snapshot recovery attempt cannot be in the future")
		}
		if attempt.Before(windowStart) {
			continue
		}
		attemptsInWindow = append(attemptsInWindow, attempt)
	}

	status := SnapshotRecoveryRateLimitStatus{
		AttemptsInWindow:   len(attemptsInWindow),
		RemainingAllowance: r.config.SnapshotRecoveryPerMinuteLimit - len(attemptsInWindow),
	}
	if status.RemainingAllowance > 0 {
		status.Allowed = true
		return status, nil
	}

	oldestAttempt := attemptsInWindow[0]
	for _, attempt := range attemptsInWindow[1:] {
		if attempt.Before(oldestAttempt) {
			oldestAttempt = attempt
		}
	}
	status.RemainingAllowance = 0
	status.RetryAfter = oldestAttempt.Add(time.Minute).Sub(now)
	return status, nil
}

func (r *Runtime) EvaluateStaleness(now, lastMessageAt, lastSnapshotAt time.Time) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("runtime is required")
	}
	if now.IsZero() {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("current time is required")
	}
	if !lastMessageAt.IsZero() && lastMessageAt.After(now) {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("last message time cannot be in the future")
	}
	if !lastSnapshotAt.IsZero() && lastSnapshotAt.After(now) {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("last snapshot time cannot be in the future")
	}

	return ingestion.EvaluateHealth(r.config.Adapter, ingestion.HealthSnapshot{
		ConnectionState: ingestion.ConnectionConnected,
		Now:             now,
		LastMessageAt:   lastMessageAt,
		LastSnapshotAt:  lastSnapshotAt,
	})
}

func (r *Runtime) EvaluateReconnectLoop(consecutiveReconnects int) (ReconnectLoopStatus, error) {
	if r == nil {
		return ReconnectLoopStatus{}, fmt.Errorf("runtime is required")
	}
	if consecutiveReconnects < 0 {
		return ReconnectLoopStatus{}, fmt.Errorf("consecutive reconnects cannot be negative")
	}
	if r.config.Adapter.ReconnectLoopThreshold <= 0 {
		return ReconnectLoopStatus{}, fmt.Errorf("reconnect loop threshold must be positive")
	}

	status := ReconnectLoopStatus{
		LoopDetected:          consecutiveReconnects >= r.config.Adapter.ReconnectLoopThreshold,
		ConsecutiveReconnects: consecutiveReconnects,
		Threshold:             r.config.Adapter.ReconnectLoopThreshold,
		RemainingBeforeLoop:   r.config.Adapter.ReconnectLoopThreshold - consecutiveReconnects,
	}
	if status.RemainingBeforeLoop < 0 {
		status.RemainingBeforeLoop = 0
	}
	return status, nil
}

func (r *Runtime) EvaluateResyncLoop(resyncCount int) (ResyncLoopStatus, error) {
	if r == nil {
		return ResyncLoopStatus{}, fmt.Errorf("runtime is required")
	}
	if resyncCount < 0 {
		return ResyncLoopStatus{}, fmt.Errorf("resync count cannot be negative")
	}
	if r.config.Adapter.ResyncLoopThreshold <= 0 {
		return ResyncLoopStatus{}, fmt.Errorf("resync loop threshold must be positive")
	}

	status := ResyncLoopStatus{
		LoopDetected:        resyncCount >= r.config.Adapter.ResyncLoopThreshold,
		ResyncCount:         resyncCount,
		Threshold:           r.config.Adapter.ResyncLoopThreshold,
		RemainingBeforeLoop: r.config.Adapter.ResyncLoopThreshold - resyncCount,
	}
	if status.RemainingBeforeLoop < 0 {
		status.RemainingBeforeLoop = 0
	}
	return status, nil
}

func (r *Runtime) BuildHealthSnapshot(input AdapterHealthInput) (ingestion.HealthSnapshot, error) {
	if r == nil {
		return ingestion.HealthSnapshot{}, fmt.Errorf("runtime is required")
	}
	if err := validateConnectionState(input.ConnectionState); err != nil {
		return ingestion.HealthSnapshot{}, err
	}

	return ingestion.HealthSnapshot{
		ConnectionState:       input.ConnectionState,
		Now:                   input.Now,
		LastMessageAt:         input.LastMessageAt,
		LastSnapshotAt:        input.LastSnapshotAt,
		SequenceGapDetected:   input.SequenceGapDetected,
		ConsecutiveReconnects: input.ConsecutiveReconnects,
		ResyncCount:           input.ResyncCount,
		LocalClockOffset:      input.LocalClockOffset,
	}, nil
}

func (r *Runtime) AdapterHealthSnapshot(input AdapterHealthInput) (AdapterHealthSnapshot, error) {
	if input.Now.IsZero() {
		return AdapterHealthSnapshot{}, fmt.Errorf("current time is required")
	}
	if !input.LastMessageAt.IsZero() && input.LastMessageAt.After(input.Now) {
		return AdapterHealthSnapshot{}, fmt.Errorf("last message time cannot be in the future")
	}
	if !input.LastSnapshotAt.IsZero() && input.LastSnapshotAt.After(input.Now) {
		return AdapterHealthSnapshot{}, fmt.Errorf("last snapshot time cannot be in the future")
	}
	healthSnapshot, err := r.BuildHealthSnapshot(input)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	clockOffsetStatus, err := r.EvaluateClockOffset(input.LocalClockOffset)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	staleness, err := ingestion.EvaluateHealth(r.config.Adapter, healthSnapshot)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	reconnectLoop, err := r.EvaluateReconnectLoop(input.ConsecutiveReconnects)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	resyncLoop, err := r.EvaluateResyncLoop(input.ResyncCount)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	snapshotRecoveryCooldown, err := r.SnapshotRecoveryStatus(input.Now, input.LastSnapshotRecoveryAt)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}
	snapshotRecoveryRateLimit, err := r.SnapshotRecoveryRateLimitStatus(input.Now, input.RecentSnapshotRecoveries)
	if err != nil {
		return AdapterHealthSnapshot{}, err
	}

	return AdapterHealthSnapshot{
		ConnectionState:           input.ConnectionState,
		SequenceGapDetected:       input.SequenceGapDetected,
		Staleness:                 staleness,
		ReconnectLoop:             reconnectLoop,
		ResyncLoop:                resyncLoop,
		SnapshotRecoveryCooldown:  snapshotRecoveryCooldown,
		SnapshotRecoveryRateLimit: snapshotRecoveryRateLimit,
		ClockOffset:               clockOffsetStatus,
	}, nil
}

func (r *Runtime) EvaluateClockOffset(offset time.Duration) (ClockOffsetStatus, error) {
	if r == nil {
		return ClockOffsetStatus{}, fmt.Errorf("runtime is required")
	}
	if r.config.Adapter.ClockOffsetWarning <= 0 || r.config.Adapter.ClockOffsetDegraded <= 0 {
		return ClockOffsetStatus{}, fmt.Errorf("clock offset thresholds must be positive")
	}
	if r.config.Adapter.ClockOffsetDegraded < r.config.Adapter.ClockOffsetWarning {
		return ClockOffsetStatus{}, fmt.Errorf("clock degraded threshold must be greater than or equal to warning threshold")
	}

	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}

	status := ClockOffsetStatus{
		State:             ingestion.ClockNormal,
		Offset:            offset,
		WarningThreshold:  r.config.Adapter.ClockOffsetWarning,
		DegradedThreshold: r.config.Adapter.ClockOffsetDegraded,
	}
	if absOffset >= r.config.Adapter.ClockOffsetDegraded {
		status.State = ingestion.ClockDegraded
		return status, nil
	}
	if absOffset >= r.config.Adapter.ClockOffsetWarning {
		status.State = ingestion.ClockWarning
	}
	return status, nil
}

func (r *Runtime) EvaluateAdapterDecision(snapshot AdapterHealthSnapshot) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("runtime is required")
	}

	status := snapshot.Staleness
	if err := validateFeedHealthState(status.State); err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	if err := validateConnectionState(snapshot.ConnectionState); err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	if err := validateClockState(snapshot.ClockOffset.State); err != nil {
		return ingestion.FeedHealthStatus{}, err
	}

	status.ConnectionState = snapshot.ConnectionState
	status.SequenceGapDetected = snapshot.SequenceGapDetected
	status.ClockState = snapshot.ClockOffset.State
	status.Reasons = dedupeReasons(status.Reasons)
	if snapshot.ConnectionState != ingestion.ConnectionConnected {
		status.Reasons = append(status.Reasons, ingestion.ReasonConnectionNotReady)
	}
	if snapshot.SequenceGapDetected {
		status.Reasons = append(status.Reasons, ingestion.ReasonSequenceGap)
	}
	if snapshot.ReconnectLoop.LoopDetected {
		status.Reasons = append(status.Reasons, ingestion.ReasonReconnectLoop)
	}
	if snapshot.ResyncLoop.LoopDetected {
		status.Reasons = append(status.Reasons, ingestion.ReasonResyncLoop)
	}
	if snapshot.ClockOffset.State == ingestion.ClockDegraded {
		status.Reasons = append(status.Reasons, ingestion.ReasonClockDegraded)
	}
	status.Reasons = dedupeReasons(status.Reasons)

	if containsReason(status.Reasons, ingestion.ReasonMessageStale) || containsReason(status.Reasons, ingestion.ReasonSnapshotStale) || status.State == ingestion.FeedHealthStale {
		status.State = ingestion.FeedHealthStale
		return status, nil
	}
	if len(status.Reasons) > 0 || status.State == ingestion.FeedHealthDegraded {
		status.State = ingestion.FeedHealthDegraded
		return status, nil
	}
	status.State = ingestion.FeedHealthHealthy
	return status, nil
}

func (r *Runtime) EvaluateAdapterInput(input AdapterHealthInput) (ingestion.FeedHealthStatus, error) {
	snapshot, err := r.AdapterHealthSnapshot(input)
	if err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	return r.EvaluateAdapterDecision(snapshot)
}

func (r *Runtime) EvaluateLoopState(state AdapterLoopState, now time.Time) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("runtime is required")
	}
	return r.EvaluateAdapterInput(state.HealthInput(now))
}

func validateFeedHealthState(state ingestion.FeedHealthState) error {
	switch state {
	case ingestion.FeedHealthHealthy, ingestion.FeedHealthDegraded, ingestion.FeedHealthStale:
		return nil
	default:
		return fmt.Errorf("unsupported feed health state %q", state)
	}
}

func validateClockState(state ingestion.ClockState) error {
	switch state {
	case ingestion.ClockNormal, ingestion.ClockWarning, ingestion.ClockDegraded:
		return nil
	default:
		return fmt.Errorf("unsupported clock state %q", state)
	}
}

func validateConnectionState(state ingestion.ConnectionState) error {
	switch state {
	case ingestion.ConnectionDisconnected, ingestion.ConnectionConnecting, ingestion.ConnectionConnected, ingestion.ConnectionReconnecting, ingestion.ConnectionResyncing:
		return nil
	default:
		return fmt.Errorf("unsupported connection state %q", state)
	}
}

func dedupeReasons(reasons []ingestion.DegradationReason) []ingestion.DegradationReason {
	if len(reasons) == 0 {
		return nil
	}
	seen := make(map[ingestion.DegradationReason]struct{}, len(reasons))
	result := make([]ingestion.DegradationReason, 0, len(reasons))
	for _, reason := range reasons {
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		result = append(result, reason)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func containsReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
