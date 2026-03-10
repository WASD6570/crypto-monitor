package ingestion

import (
	"fmt"
	"sort"
	"time"
)

type Venue string

const (
	VenueBinance  Venue = "BINANCE"
	VenueBybit    Venue = "BYBIT"
	VenueCoinbase Venue = "COINBASE"
	VenueKraken   Venue = "KRAKEN"
)

type StreamKind string

const (
	StreamTrades       StreamKind = "trades"
	StreamTopOfBook    StreamKind = "top-of-book"
	StreamOrderBook    StreamKind = "order-book"
	StreamFundingRate  StreamKind = "funding-rate"
	StreamOpenInterest StreamKind = "open-interest"
	StreamMarkIndex    StreamKind = "mark-index"
	StreamLiquidation  StreamKind = "liquidation"
)

type ConnectionState string

const (
	ConnectionDisconnected ConnectionState = "DISCONNECTED"
	ConnectionConnecting   ConnectionState = "CONNECTING"
	ConnectionConnected    ConnectionState = "CONNECTED"
	ConnectionReconnecting ConnectionState = "RECONNECTING"
	ConnectionResyncing    ConnectionState = "RESYNCING"
)

type FeedHealthState string

const (
	FeedHealthHealthy  FeedHealthState = "HEALTHY"
	FeedHealthDegraded FeedHealthState = "DEGRADED"
	FeedHealthStale    FeedHealthState = "STALE"
)

type FreshnessState string

const (
	FreshnessUnknown       FreshnessState = "UNKNOWN"
	FreshnessFresh         FreshnessState = "FRESH"
	FreshnessStale         FreshnessState = "STALE"
	FreshnessNotApplicable FreshnessState = "NOT_APPLICABLE"
)

type ClockState string

const (
	ClockNormal   ClockState = "NORMAL"
	ClockWarning  ClockState = "WARNING"
	ClockDegraded ClockState = "DEGRADED"
)

type DegradationReason string

const (
	ReasonConnectionNotReady DegradationReason = "connection-not-ready"
	ReasonMessageStale       DegradationReason = "message-stale"
	ReasonSnapshotStale      DegradationReason = "snapshot-stale"
	ReasonSequenceGap        DegradationReason = "sequence-gap"
	ReasonReconnectLoop      DegradationReason = "reconnect-loop"
	ReasonResyncLoop         DegradationReason = "resync-loop"
	ReasonRateLimit          DegradationReason = "rate-limit"
	ReasonClockDegraded      DegradationReason = "clock-degraded"
)

type StreamDefinition struct {
	Kind             StreamKind `json:"kind"`
	MarketType       string     `json:"marketType"`
	SnapshotRequired bool       `json:"snapshotRequired"`
}

type AdapterConfig struct {
	Venue                  Venue
	Streams                []StreamDefinition
	MessageStaleAfter      time.Duration
	SnapshotStaleAfter     time.Duration
	ReconnectBackoffMin    time.Duration
	ReconnectBackoffMax    time.Duration
	ReconnectLoopThreshold int
	ResyncLoopThreshold    int
	ClockOffsetWarning     time.Duration
	ClockOffsetDegraded    time.Duration
}

func (c AdapterConfig) Validate() error {
	if c.Venue == "" {
		return fmt.Errorf("venue is required")
	}
	if len(c.Streams) == 0 {
		return fmt.Errorf("at least one stream definition is required")
	}
	if c.MessageStaleAfter <= 0 {
		return fmt.Errorf("message stale threshold must be positive")
	}
	if c.SnapshotStaleAfter <= 0 {
		return fmt.Errorf("snapshot stale threshold must be positive")
	}
	if c.ReconnectBackoffMin <= 0 || c.ReconnectBackoffMax <= 0 {
		return fmt.Errorf("reconnect backoff thresholds must be positive")
	}
	if c.ReconnectBackoffMax < c.ReconnectBackoffMin {
		return fmt.Errorf("reconnect backoff max must be greater than or equal to min")
	}
	if c.ReconnectLoopThreshold <= 0 {
		return fmt.Errorf("reconnect loop threshold must be positive")
	}
	if c.ResyncLoopThreshold <= 0 {
		return fmt.Errorf("resync loop threshold must be positive")
	}
	if c.ClockOffsetWarning <= 0 || c.ClockOffsetDegraded <= 0 {
		return fmt.Errorf("clock offset thresholds must be positive")
	}
	if c.ClockOffsetDegraded < c.ClockOffsetWarning {
		return fmt.Errorf("clock degraded threshold must be greater than or equal to warning threshold")
	}

	seen := map[StreamKind]struct{}{}
	for _, stream := range c.Streams {
		if stream.Kind == "" {
			return fmt.Errorf("stream kind is required")
		}
		if stream.MarketType == "" {
			return fmt.Errorf("market type is required for stream %q", stream.Kind)
		}
		key := stream.Kind
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate stream definition for %q", stream.Kind)
		}
		seen[key] = struct{}{}
	}

	return nil
}

type HealthSnapshot struct {
	ConnectionState       ConnectionState
	Now                   time.Time
	LastMessageAt         time.Time
	LastSnapshotAt        time.Time
	SequenceGapDetected   bool
	ConsecutiveReconnects int
	ResyncCount           int
	LocalClockOffset      time.Duration
}

type FeedHealthStatus struct {
	State                 FeedHealthState
	ConnectionState       ConnectionState
	MessageFreshness      FreshnessState
	SnapshotFreshness     FreshnessState
	SequenceGapDetected   bool
	ConsecutiveReconnects int
	ResyncCount           int
	ClockState            ClockState
	Reasons               []DegradationReason
}

func EvaluateHealth(config AdapterConfig, snapshot HealthSnapshot) (FeedHealthStatus, error) {
	if err := config.Validate(); err != nil {
		return FeedHealthStatus{}, err
	}
	if snapshot.Now.IsZero() {
		return FeedHealthStatus{}, fmt.Errorf("health snapshot current time is required")
	}

	status := FeedHealthStatus{
		State:                 FeedHealthHealthy,
		ConnectionState:       snapshot.ConnectionState,
		MessageFreshness:      freshnessState(snapshot.Now, snapshot.LastMessageAt, config.MessageStaleAfter),
		SnapshotFreshness:     snapshotFreshness(config, snapshot),
		SequenceGapDetected:   snapshot.SequenceGapDetected,
		ConsecutiveReconnects: snapshot.ConsecutiveReconnects,
		ResyncCount:           snapshot.ResyncCount,
		ClockState:            clockState(snapshot.LocalClockOffset, config.ClockOffsetWarning, config.ClockOffsetDegraded),
	}

	if snapshot.ConnectionState != ConnectionConnected {
		status.Reasons = append(status.Reasons, ReasonConnectionNotReady)
	}
	if status.MessageFreshness == FreshnessStale {
		status.Reasons = append(status.Reasons, ReasonMessageStale)
	}
	if status.SnapshotFreshness == FreshnessStale {
		status.Reasons = append(status.Reasons, ReasonSnapshotStale)
	}
	if snapshot.SequenceGapDetected {
		status.Reasons = append(status.Reasons, ReasonSequenceGap)
	}
	if snapshot.ConsecutiveReconnects >= config.ReconnectLoopThreshold {
		status.Reasons = append(status.Reasons, ReasonReconnectLoop)
	}
	if snapshot.ResyncCount >= config.ResyncLoopThreshold {
		status.Reasons = append(status.Reasons, ReasonResyncLoop)
	}
	if status.ClockState == ClockDegraded {
		status.Reasons = append(status.Reasons, ReasonClockDegraded)
	}

	status.State = deriveState(status)
	sort.Slice(status.Reasons, func(i, j int) bool {
		return status.Reasons[i] < status.Reasons[j]
	})

	return status, nil
}

func snapshotFreshness(config AdapterConfig, snapshot HealthSnapshot) FreshnessState {
	requiresSnapshot := false
	for _, stream := range config.Streams {
		if stream.SnapshotRequired {
			requiresSnapshot = true
			break
		}
	}
	if !requiresSnapshot {
		return FreshnessNotApplicable
	}
	return freshnessState(snapshot.Now, snapshot.LastSnapshotAt, config.SnapshotStaleAfter)
}

func freshnessState(now, lastSeen time.Time, staleAfter time.Duration) FreshnessState {
	if lastSeen.IsZero() {
		return FreshnessUnknown
	}
	if now.Sub(lastSeen) >= staleAfter {
		return FreshnessStale
	}
	return FreshnessFresh
}

func clockState(offset, warningThreshold, degradedThreshold time.Duration) ClockState {
	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}
	if absOffset >= degradedThreshold {
		return ClockDegraded
	}
	if absOffset >= warningThreshold {
		return ClockWarning
	}
	return ClockNormal
}

func deriveState(status FeedHealthStatus) FeedHealthState {
	if containsReason(status.Reasons, ReasonMessageStale) || containsReason(status.Reasons, ReasonSnapshotStale) {
		return FeedHealthStale
	}
	if len(status.Reasons) > 0 {
		return FeedHealthDegraded
	}
	return FeedHealthHealthy
}

func containsReason(reasons []DegradationReason, target DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
