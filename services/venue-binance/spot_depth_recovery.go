package venuebinance

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const spotDepthFeedHealthRecordPrefix = "runtime:binance-spot-depth:"

type SpotDepthRecoveryState string

const (
	SpotDepthRecoveryIdle             SpotDepthRecoveryState = "idle"
	SpotDepthRecoverySynchronized     SpotDepthRecoveryState = "synchronized"
	SpotDepthRecoveryResyncing        SpotDepthRecoveryState = "resyncing"
	SpotDepthRecoveryCooldownBlocked  SpotDepthRecoveryState = "cooldown-blocked"
	SpotDepthRecoveryRateLimitBlocked SpotDepthRecoveryState = "rate-limit-blocked"
	SpotDepthRecoveryBootstrapFailed  SpotDepthRecoveryState = "bootstrap-failed"
)

type SpotDepthRecoveryTrigger string

const (
	SpotDepthRecoveryTriggerNone            SpotDepthRecoveryTrigger = ""
	SpotDepthRecoveryTriggerSequenceGap     SpotDepthRecoveryTrigger = "sequence-gap"
	SpotDepthRecoveryTriggerBootstrapFail   SpotDepthRecoveryTrigger = "bootstrap-failed"
	SpotDepthRecoveryTriggerSnapshotRefresh SpotDepthRecoveryTrigger = "snapshot-refresh"
)

type SpotDepthRecoveryStatus struct {
	State                 SpotDepthRecoveryState
	Trigger               SpotDepthRecoveryTrigger
	SourceSymbol          string
	LastAcceptedSequence  int64
	BufferedDeltaCount    int
	LastMessageAt         time.Time
	LastSnapshotAt        time.Time
	LastRecoveryAttemptAt time.Time
	RemainingCooldown     time.Duration
	RetryAfter            time.Duration
	ResyncCount           int
	SequenceGapDetected   bool
	RefreshDue            bool
	Synchronized          bool
}

type SpotDepthFeedHealthOptions struct {
	Symbol                string
	QuoteCurrency         string
	Now                   time.Time
	ConnectionState       ingestion.ConnectionState
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
}

type SpotDepthRecoveryOwner struct {
	runtime                *Runtime
	fetcher                SpotDepthSnapshotFetcher
	state                  SpotDepthRecoveryState
	trigger                SpotDepthRecoveryTrigger
	sourceSymbol           string
	lastAcceptedSequence   int64
	lastMessageAt          time.Time
	lastSnapshotAt         time.Time
	lastRecoveryAttemptAt  time.Time
	recentRecoveryAttempts []time.Time
	resyncCount            int
	sequenceGapDetected    bool
	refreshDue             bool
	buffered               []ParsedOrderBook
	remainingCooldown      time.Duration
	retryAfter             time.Duration
}

func NewSpotDepthRecoveryOwner(runtime *Runtime, fetcher SpotDepthSnapshotFetcher) (*SpotDepthRecoveryOwner, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}
	if fetcher == nil {
		return nil, fmt.Errorf("snapshot fetcher is required")
	}
	if !hasSpotOrderBookStream(runtime.config.Adapter.Streams) {
		return nil, fmt.Errorf("binance spot depth recovery requires spot order-book stream")
	}
	return &SpotDepthRecoveryOwner{runtime: runtime, fetcher: fetcher, state: SpotDepthRecoveryIdle}, nil
}

func (o *SpotDepthRecoveryOwner) Status() SpotDepthRecoveryStatus {
	if o == nil {
		return SpotDepthRecoveryStatus{}
	}
	return SpotDepthRecoveryStatus{
		State:                 o.state,
		Trigger:               o.trigger,
		SourceSymbol:          o.sourceSymbol,
		LastAcceptedSequence:  o.lastAcceptedSequence,
		BufferedDeltaCount:    len(o.buffered),
		LastMessageAt:         o.lastMessageAt,
		LastSnapshotAt:        o.lastSnapshotAt,
		LastRecoveryAttemptAt: o.lastRecoveryAttemptAt,
		RemainingCooldown:     o.remainingCooldown,
		RetryAfter:            o.retryAfter,
		ResyncCount:           o.resyncCount,
		SequenceGapDetected:   o.sequenceGapDetected,
		RefreshDue:            o.refreshDue,
		Synchronized:          o.state == SpotDepthRecoverySynchronized,
	}
}

func (o *SpotDepthRecoveryOwner) StatusAt(now time.Time) (SpotDepthRecoveryStatus, error) {
	if o == nil {
		return SpotDepthRecoveryStatus{}, fmt.Errorf("spot depth recovery owner is required")
	}
	if now.IsZero() {
		return SpotDepthRecoveryStatus{}, fmt.Errorf("current time is required")
	}
	status := o.Status()
	if status.State == SpotDepthRecoveryIdle {
		return status, nil
	}
	if status.State == SpotDepthRecoverySynchronized {
		refreshStatus, err := o.runtime.SnapshotRefreshStatus(now, status.LastSnapshotAt)
		if err != nil {
			return SpotDepthRecoveryStatus{}, err
		}
		status.RefreshDue = refreshStatus.Due
		return status, nil
	}

	cooldownStatus, err := o.runtime.SnapshotRecoveryStatus(now, o.lastRecoveryAttemptAt)
	if err != nil {
		return SpotDepthRecoveryStatus{}, err
	}
	rateLimitStatus, err := o.runtime.SnapshotRecoveryRateLimitStatus(now, o.recentRecoveryAttempts)
	if err != nil {
		return SpotDepthRecoveryStatus{}, err
	}

	status.RemainingCooldown = 0
	status.RetryAfter = 0
	status.Synchronized = false
	if !cooldownStatus.Ready {
		status.State = SpotDepthRecoveryCooldownBlocked
		status.RemainingCooldown = cooldownStatus.RemainingCooldown
		return status, nil
	}
	if !rateLimitStatus.Allowed {
		status.State = SpotDepthRecoveryRateLimitBlocked
		status.RetryAfter = rateLimitStatus.RetryAfter
		return status, nil
	}
	if status.Trigger == SpotDepthRecoveryTriggerBootstrapFail {
		status.State = SpotDepthRecoveryBootstrapFailed
		return status, nil
	}
	status.State = SpotDepthRecoveryResyncing
	return status, nil
}

func (o *SpotDepthRecoveryOwner) StartSynchronized(sync SpotDepthBootstrapSync) error {
	if o == nil {
		return fmt.Errorf("spot depth recovery owner is required")
	}
	if sync.SourceSymbol == "" {
		return fmt.Errorf("source symbol is required")
	}
	snapshotRecv, err := parsedOrderBookRecvTime(sync.Snapshot)
	if err != nil {
		return err
	}
	lastSequence := sync.Snapshot.FinalSequence
	lastMessageAt := snapshotRecv
	for _, delta := range sync.Deltas {
		if delta.SourceSymbol != sync.SourceSymbol {
			return fmt.Errorf("aligned delta source symbol = %q, want %q", delta.SourceSymbol, sync.SourceSymbol)
		}
		recvTime, err := parsedOrderBookRecvTime(delta)
		if err != nil {
			return err
		}
		if recvTime.After(lastMessageAt) {
			lastMessageAt = recvTime
		}
		if delta.FinalSequence > lastSequence {
			lastSequence = delta.FinalSequence
		}
	}

	o.state = SpotDepthRecoverySynchronized
	o.trigger = SpotDepthRecoveryTriggerNone
	o.sourceSymbol = sync.SourceSymbol
	o.lastAcceptedSequence = lastSequence
	o.lastMessageAt = lastMessageAt
	o.lastSnapshotAt = snapshotRecv
	o.lastRecoveryAttemptAt = snapshotRecv
	o.recentRecoveryAttempts = []time.Time{snapshotRecv}
	o.resyncCount = 0
	o.sequenceGapDetected = false
	o.refreshDue = false
	o.buffered = nil
	o.remainingCooldown = 0
	o.retryAfter = 0
	return nil
}

func (o *SpotDepthRecoveryOwner) MarkBootstrapFailure(sourceSymbol string) error {
	if o == nil {
		return fmt.Errorf("spot depth recovery owner is required")
	}
	if sourceSymbol == "" {
		return fmt.Errorf("source symbol is required")
	}
	if o.sourceSymbol == "" {
		o.sourceSymbol = sourceSymbol
	} else if o.sourceSymbol != sourceSymbol {
		return fmt.Errorf("bootstrap failure source symbol = %q, want %q", sourceSymbol, o.sourceSymbol)
	}
	o.state = SpotDepthRecoveryBootstrapFailed
	o.trigger = SpotDepthRecoveryTriggerBootstrapFail
	o.sequenceGapDetected = true
	o.refreshDue = false
	o.buffered = nil
	return nil
}

func (o *SpotDepthRecoveryOwner) MarkSequenceGap(frame SpotRawFrame) error {
	if o == nil {
		return fmt.Errorf("spot depth recovery owner is required")
	}
	parsed, err := ParseOrderBookFrame(frame)
	if err != nil {
		return err
	}
	if err := o.ensureSourceSymbol(parsed.SourceSymbol); err != nil {
		return err
	}
	o.lastMessageAt = frame.RecvTime
	o.sequenceGapDetected = true
	o.refreshDue = false
	o.trigger = SpotDepthRecoveryTriggerSequenceGap
	o.buffered = append(o.buffered, parsed)
	return o.updateRecoveryState(frame.RecvTime)
}

func (o *SpotDepthRecoveryOwner) AcceptSynchronizedDelta(frame SpotRawFrame) error {
	if o == nil {
		return fmt.Errorf("spot depth recovery owner is required")
	}
	if o.state != SpotDepthRecoverySynchronized {
		return fmt.Errorf("spot depth recovery owner is not synchronized")
	}
	parsed, err := ParseOrderBookFrame(frame)
	if err != nil {
		return err
	}
	if err := o.ensureSourceSymbol(parsed.SourceSymbol); err != nil {
		return err
	}
	o.lastMessageAt = frame.RecvTime
	if parsed.FinalSequence <= o.lastAcceptedSequence {
		return nil
	}
	expected := o.lastAcceptedSequence + 1
	if parsed.FirstSequence <= expected && parsed.FinalSequence >= expected {
		o.lastAcceptedSequence = parsed.FinalSequence
		return nil
	}
	o.sequenceGapDetected = true
	o.refreshDue = false
	o.trigger = SpotDepthRecoveryTriggerSequenceGap
	o.buffered = append(o.buffered, parsed)
	return o.updateRecoveryState(frame.RecvTime)
}

func (o *SpotDepthRecoveryOwner) BufferRecoveryDelta(frame SpotRawFrame) error {
	if o == nil {
		return fmt.Errorf("spot depth recovery owner is required")
	}
	parsed, err := ParseOrderBookFrame(frame)
	if err != nil {
		return err
	}
	if err := o.ensureSourceSymbol(parsed.SourceSymbol); err != nil {
		return err
	}
	if parsed.FinalSequence <= o.lastAcceptedSequence {
		o.lastMessageAt = frame.RecvTime
		return nil
	}
	o.lastMessageAt = frame.RecvTime
	o.buffered = append(o.buffered, parsed)
	return nil
}

func (o *SpotDepthRecoveryOwner) SnapshotRefreshStatus(now time.Time) (SnapshotRefreshStatus, error) {
	if o == nil {
		return SnapshotRefreshStatus{}, fmt.Errorf("spot depth recovery owner is required")
	}
	status, err := o.runtime.SnapshotRefreshStatus(now, o.lastSnapshotAt)
	if err != nil {
		return SnapshotRefreshStatus{}, err
	}
	o.refreshDue = status.Due
	return status, nil
}

func (o *SpotDepthRecoveryOwner) Refresh(ctx context.Context, now time.Time) (SpotDepthBootstrapSync, error) {
	if o == nil {
		return SpotDepthBootstrapSync{}, fmt.Errorf("spot depth recovery owner is required")
	}
	status, err := o.runtime.SnapshotRefreshStatus(now, o.lastSnapshotAt)
	if err != nil {
		return SpotDepthBootstrapSync{}, err
	}
	if !status.Required || !status.Due {
		return SpotDepthBootstrapSync{}, fmt.Errorf("snapshot refresh is not due")
	}
	o.refreshDue = true
	o.trigger = SpotDepthRecoveryTriggerSnapshotRefresh
	return o.attemptRecovery(ctx, now, false)
}

func (o *SpotDepthRecoveryOwner) Recover(ctx context.Context, now time.Time) (SpotDepthBootstrapSync, error) {
	if o == nil {
		return SpotDepthBootstrapSync{}, fmt.Errorf("spot depth recovery owner is required")
	}
	if o.trigger == SpotDepthRecoveryTriggerNone {
		return SpotDepthBootstrapSync{}, fmt.Errorf("recovery trigger is required")
	}
	return o.attemptRecovery(ctx, now, true)
}

func (o *SpotDepthRecoveryOwner) HealthStatus(now time.Time, connectionState ingestion.ConnectionState, localClockOffset time.Duration, consecutiveReconnects int) (ingestion.FeedHealthStatus, error) {
	if o == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("spot depth recovery owner is required")
	}
	if now.IsZero() {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("current time is required")
	}
	recoveryStatus, err := o.StatusAt(now)
	if err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	healthConnectionState := connectionState
	if recoveryStatus.State != SpotDepthRecoverySynchronized && recoveryStatus.State != SpotDepthRecoveryIdle {
		healthConnectionState = ingestion.ConnectionResyncing
	}
	status, err := o.runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:          healthConnectionState,
		Now:                      now,
		LastMessageAt:            recoveryStatus.LastMessageAt,
		LastSnapshotAt:           recoveryStatus.LastSnapshotAt,
		SequenceGapDetected:      recoveryStatus.SequenceGapDetected,
		LocalClockOffset:         localClockOffset,
		LastSnapshotRecoveryAt:   recoveryStatus.LastRecoveryAttemptAt,
		RecentSnapshotRecoveries: append([]time.Time(nil), o.recentRecoveryAttempts...),
		ConsecutiveReconnects:    consecutiveReconnects,
		ResyncCount:              recoveryStatus.ResyncCount,
	})
	if err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	if recoveryStatus.State == SpotDepthRecoveryRateLimitBlocked {
		status.Reasons = append(status.Reasons, ingestion.ReasonRateLimit)
	}
	status.Reasons = dedupeDepthRecoveryReasons(status.Reasons)
	if containsDepthRecoveryReason(status.Reasons, ingestion.ReasonMessageStale) || containsDepthRecoveryReason(status.Reasons, ingestion.ReasonSnapshotStale) {
		status.State = ingestion.FeedHealthStale
	} else if len(status.Reasons) > 0 {
		status.State = ingestion.FeedHealthDegraded
	} else {
		status.State = ingestion.FeedHealthHealthy
	}
	return status, nil
}

func (o *SpotDepthRecoveryOwner) FeedHealthInput(options SpotDepthFeedHealthOptions) (SpotFeedHealthInput, error) {
	if o == nil {
		return SpotFeedHealthInput{}, fmt.Errorf("spot depth recovery owner is required")
	}
	if options.Symbol == "" || options.QuoteCurrency == "" {
		return SpotFeedHealthInput{}, fmt.Errorf("feed health symbol and quote currency are required")
	}
	if options.Now.IsZero() {
		return SpotFeedHealthInput{}, fmt.Errorf("feed health time is required")
	}
	if o.sourceSymbol == "" {
		return SpotFeedHealthInput{}, fmt.Errorf("source symbol is required")
	}
	status, err := o.HealthStatus(options.Now, options.ConnectionState, options.LocalClockOffset, options.ConsecutiveReconnects)
	if err != nil {
		return SpotFeedHealthInput{}, err
	}
	timestamp := options.Now.UTC().Format(time.RFC3339Nano)
	return SpotFeedHealthInput{
		Metadata: ingestion.FeedHealthMetadata{
			Symbol:        options.Symbol,
			SourceSymbol:  o.sourceSymbol,
			QuoteCurrency: options.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message: ingestion.FeedHealthMessage{
			ExchangeTs:     timestamp,
			RecvTs:         timestamp,
			SourceRecordID: spotDepthFeedHealthRecordPrefix + o.sourceSymbol,
			Status:         status,
		},
	}, nil
}

func (o *SpotDepthRecoveryOwner) attemptRecovery(ctx context.Context, now time.Time, requireBridge bool) (SpotDepthBootstrapSync, error) {
	if ctx == nil {
		return SpotDepthBootstrapSync{}, fmt.Errorf("context is required")
	}
	if now.IsZero() {
		return SpotDepthBootstrapSync{}, fmt.Errorf("current time is required")
	}
	if o.sourceSymbol == "" {
		return SpotDepthBootstrapSync{}, fmt.Errorf("source symbol is required")
	}
	if requireBridge && len(o.buffered) == 0 {
		return SpotDepthBootstrapSync{}, fmt.Errorf("buffered depth delta is required before recovery")
	}
	if err := o.updateRecoveryState(now); err != nil {
		return SpotDepthBootstrapSync{}, err
	}
	if o.state == SpotDepthRecoveryCooldownBlocked {
		return SpotDepthBootstrapSync{}, fmt.Errorf("snapshot recovery cooldown remains for %s", o.remainingCooldown)
	}
	if o.state == SpotDepthRecoveryRateLimitBlocked {
		return SpotDepthBootstrapSync{}, fmt.Errorf("snapshot recovery rate limit blocks retry for %s", o.retryAfter)
	}

	o.lastRecoveryAttemptAt = now
	o.recentRecoveryAttempts = append(o.recentRecoveryAttempts, now)
	o.resyncCount++
	o.remainingCooldown = 0
	o.retryAfter = 0
	if requireBridge {
		o.state = SpotDepthRecoveryResyncing
	}

	snapshotResponse, err := o.fetcher.FetchSpotDepthSnapshot(ctx, o.sourceSymbol)
	if err != nil {
		if o.trigger == SpotDepthRecoveryTriggerBootstrapFail {
			o.state = SpotDepthRecoveryBootstrapFailed
		}
		return SpotDepthBootstrapSync{}, fmt.Errorf("fetch spot depth snapshot: %w", err)
	}
	snapshot, err := ParseOrderBookSnapshotWithSourceSymbol(snapshotResponse.Payload, o.sourceSymbol, snapshotResponse.RecvTime)
	if err != nil {
		if o.trigger == SpotDepthRecoveryTriggerBootstrapFail {
			o.state = SpotDepthRecoveryBootstrapFailed
		}
		return SpotDepthBootstrapSync{}, fmt.Errorf("parse spot depth snapshot: %w", err)
	}

	aligned := []ParsedOrderBook(nil)
	if len(o.buffered) > 0 {
		aligned, err = alignBufferedDeltas(snapshot.FinalSequence, o.buffered)
		if err != nil {
			if o.trigger == SpotDepthRecoveryTriggerBootstrapFail {
				o.state = SpotDepthRecoveryBootstrapFailed
			} else {
				o.state = SpotDepthRecoveryResyncing
			}
			return SpotDepthBootstrapSync{}, err
		}
	}

	sync := SpotDepthBootstrapSync{SourceSymbol: o.sourceSymbol, Snapshot: snapshot, Deltas: aligned}
	if err := o.StartSynchronized(sync); err != nil {
		return SpotDepthBootstrapSync{}, err
	}
	return sync, nil
}

func (o *SpotDepthRecoveryOwner) updateRecoveryState(now time.Time) error {
	cooldownStatus, err := o.runtime.SnapshotRecoveryStatus(now, o.lastRecoveryAttemptAt)
	if err != nil {
		return err
	}
	rateLimitStatus, err := o.runtime.SnapshotRecoveryRateLimitStatus(now, o.recentRecoveryAttempts)
	if err != nil {
		return err
	}
	o.remainingCooldown = 0
	o.retryAfter = 0
	if !cooldownStatus.Ready {
		o.state = SpotDepthRecoveryCooldownBlocked
		o.remainingCooldown = cooldownStatus.RemainingCooldown
		return nil
	}
	if !rateLimitStatus.Allowed {
		o.state = SpotDepthRecoveryRateLimitBlocked
		o.retryAfter = rateLimitStatus.RetryAfter
		return nil
	}
	if o.trigger == SpotDepthRecoveryTriggerBootstrapFail {
		o.state = SpotDepthRecoveryBootstrapFailed
		return nil
	}
	o.state = SpotDepthRecoveryResyncing
	return nil
}

func (o *SpotDepthRecoveryOwner) ensureSourceSymbol(sourceSymbol string) error {
	if sourceSymbol == "" {
		return fmt.Errorf("source symbol is required")
	}
	if o.sourceSymbol == "" {
		o.sourceSymbol = sourceSymbol
		return nil
	}
	if o.sourceSymbol != sourceSymbol {
		return fmt.Errorf("depth recovery source symbol = %q, want %q", sourceSymbol, o.sourceSymbol)
	}
	return nil
}

func parsedOrderBookRecvTime(parsed ParsedOrderBook) (time.Time, error) {
	if parsed.Message.RecvTs == "" {
		return time.Time{}, fmt.Errorf("order-book recv timestamp is required")
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, parsed.Message.RecvTs)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse order-book recv timestamp: %w", err)
	}
	return parsedTime, nil
}

func dedupeDepthRecoveryReasons(reasons []ingestion.DegradationReason) []ingestion.DegradationReason {
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
	slices.Sort(result)
	return result
}

func containsDepthRecoveryReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
