package venuebinance

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const (
	usdmFeedHealthRecordPrefix = "runtime:binance-usdm-ws:"
	usdmSessionRefPrefix       = "binance-usdm-session-"
)

type USDMReconnectCause string

const (
	USDMReconnectCauseTransport USDMReconnectCause = "transport-close"
	USDMReconnectCauseRollover  USDMReconnectCause = "rollover"
)

type USDMSubscribeCommand struct {
	ID     int64
	Method string
	Params []string
}

type USDMRawFrame struct {
	RecvTime time.Time
	Payload  []byte
}

type USDMFeedHealthInput struct {
	Metadata ingestion.FeedHealthMetadata
	Message  ingestion.FeedHealthMessage
}

type USDMReconnectPlan struct {
	Cause      USDMReconnectCause
	Attempt    int
	Delay      time.Duration
	RetryAt    time.Time
	SessionRef string
}

type USDMRuntimeState struct {
	DesiredSubscriptions  []string
	ActiveSubscriptions   []string
	PendingSubscribeIDs   []int64
	ConnectionState       ingestion.ConnectionState
	ConnectionOpenedAt    time.Time
	LastFrameAt           time.Time
	LastPongAt            time.Time
	LastMarkPriceAt       time.Time
	LastSubscribeAckAt    time.Time
	NextReconnectAt       time.Time
	NextRolloverAt        time.Time
	LastReconnectCause    USDMReconnectCause
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	RecentConnectAttempts []time.Time
	SessionRef            string
}

type USDMRuntime struct {
	runtime     *Runtime
	usdmRuntime *Runtime
	symbols     []usdmSymbolBinding
	state       usdmRuntimeState
	nextCommand int64
	sessionSeq  int
	connected   bool
}

type usdmSymbolBinding struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
}

type usdmPendingCommand struct {
	command  USDMSubscribeCommand
	issuedAt time.Time
}

type usdmRuntimeState struct {
	DesiredSubscriptions  []string
	ActiveSubscriptions   []string
	PendingCommands       map[int64]usdmPendingCommand
	ConnectionState       ingestion.ConnectionState
	ConnectionOpenedAt    time.Time
	LastFrameAt           time.Time
	LastPongAt            time.Time
	LastMarkPriceAt       time.Time
	LastSubscribeAckAt    time.Time
	NextReconnectAt       time.Time
	NextRolloverAt        time.Time
	LastReconnectCause    USDMReconnectCause
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	RecentConnectAttempts []time.Time
	SessionRef            string
	SessionGeneration     int
	AwaitingConnect       bool
}

func NewUSDMRuntime(runtime *Runtime) (*USDMRuntime, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}

	filteredConfig, err := usdmRuntimeConfig(runtime.config)
	if err != nil {
		return nil, err
	}
	filteredRuntime, err := NewRuntime(filteredConfig)
	if err != nil {
		return nil, err
	}
	bindings, err := usdmSymbolBindings(runtime.config.Symbols)
	if err != nil {
		return nil, err
	}
	desiredSubscriptions, err := usdmDesiredSubscriptions(filteredConfig.Symbols, filteredConfig.Adapter.Streams)
	if err != nil {
		return nil, err
	}

	return &USDMRuntime{
		runtime:     runtime,
		usdmRuntime: filteredRuntime,
		symbols:     bindings,
		state: usdmRuntimeState{
			DesiredSubscriptions: append([]string(nil), desiredSubscriptions...),
			PendingCommands:      map[int64]usdmPendingCommand{},
			ConnectionState:      ingestion.ConnectionDisconnected,
		},
	}, nil
}

func (r *USDMRuntime) State() USDMRuntimeState {
	if r == nil {
		return USDMRuntimeState{}
	}
	pendingIDs := make([]int64, 0, len(r.state.PendingCommands))
	for id := range r.state.PendingCommands {
		pendingIDs = append(pendingIDs, id)
	}
	slices.Sort(pendingIDs)
	return USDMRuntimeState{
		DesiredSubscriptions:  append([]string(nil), r.state.DesiredSubscriptions...),
		ActiveSubscriptions:   append([]string(nil), r.state.ActiveSubscriptions...),
		PendingSubscribeIDs:   pendingIDs,
		ConnectionState:       r.state.ConnectionState,
		ConnectionOpenedAt:    r.state.ConnectionOpenedAt,
		LastFrameAt:           r.state.LastFrameAt,
		LastPongAt:            r.state.LastPongAt,
		LastMarkPriceAt:       r.state.LastMarkPriceAt,
		LastSubscribeAckAt:    r.state.LastSubscribeAckAt,
		NextReconnectAt:       r.state.NextReconnectAt,
		NextRolloverAt:        r.state.NextRolloverAt,
		LastReconnectCause:    r.state.LastReconnectCause,
		LocalClockOffset:      r.state.LocalClockOffset,
		ConsecutiveReconnects: r.state.ConsecutiveReconnects,
		RecentConnectAttempts: append([]time.Time(nil), r.state.RecentConnectAttempts...),
		SessionRef:            r.state.SessionRef,
	}
}

func (r *USDMRuntime) StartConnect(now time.Time) error {
	if r == nil {
		return fmt.Errorf("usdm runtime is required")
	}
	if now.IsZero() {
		return fmt.Errorf("connect time is required")
	}
	if r.connected || r.state.AwaitingConnect {
		return fmt.Errorf("usdm websocket session is already active")
	}
	if !r.state.NextReconnectAt.IsZero() && now.Before(r.state.NextReconnectAt) {
		return fmt.Errorf("usdm websocket reconnect backoff remains until %s", r.state.NextReconnectAt.UTC().Format(time.RFC3339Nano))
	}
	if err := r.pruneConnectAttempts(now); err != nil {
		return err
	}
	if len(r.state.RecentConnectAttempts) >= r.runtime.config.ConnectsPerMinuteLimit {
		return fmt.Errorf("usdm websocket connect rate limit reached")
	}

	r.state.RecentConnectAttempts = append(r.state.RecentConnectAttempts, now)
	r.state.ConnectionState = ingestion.ConnectionConnecting
	r.state.AwaitingConnect = true
	r.sessionSeq++
	r.state.SessionGeneration = r.sessionSeq
	r.state.SessionRef = fmt.Sprintf("%s%d", usdmSessionRefPrefix, r.sessionSeq)
	return nil
}

func (r *USDMRuntime) CompleteConnect(now time.Time) (*USDMSubscribeCommand, error) {
	if r == nil {
		return nil, fmt.Errorf("usdm runtime is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("connect completion time is required")
	}
	if !r.state.AwaitingConnect {
		return nil, fmt.Errorf("usdm websocket connect has not started")
	}

	r.connected = true
	r.state.AwaitingConnect = false
	r.state.ConnectionState = ingestion.ConnectionConnected
	r.state.ConnectionOpenedAt = now
	r.state.LastFrameAt = now
	r.state.NextReconnectAt = time.Time{}
	r.state.NextRolloverAt = now.Add(spotConnectionMaxAge - spotRolloverHeadroom)
	r.state.PendingCommands = map[int64]usdmPendingCommand{}
	r.state.ActiveSubscriptions = nil

	command := r.newSubscribeCommand(now)
	if !r.runtime.config.ResubscribeOnReconnect && r.state.ConsecutiveReconnects > 0 {
		return nil, nil
	}
	return &command, nil
}

func (r *USDMRuntime) AckSubscribe(now time.Time, commandID int64) error {
	if r == nil {
		return fmt.Errorf("usdm runtime is required")
	}
	if now.IsZero() {
		return fmt.Errorf("subscribe ack time is required")
	}
	pending, ok := r.state.PendingCommands[commandID]
	if !ok {
		return fmt.Errorf("unknown usdm subscribe command id %d", commandID)
	}
	delete(r.state.PendingCommands, commandID)
	r.state.ActiveSubscriptions = append([]string(nil), pending.command.Params...)
	r.state.LastFrameAt = now
	r.state.LastSubscribeAckAt = now
	r.state.ConsecutiveReconnects = 0
	return nil
}

func (r *USDMRuntime) HandlePing(now time.Time) (bool, error) {
	if err := r.recordFrame(now); err != nil {
		return false, err
	}
	r.state.LastPongAt = now
	return true, nil
}

func (r *USDMRuntime) HandlePong(now time.Time) error {
	if err := r.recordFrame(now); err != nil {
		return err
	}
	r.state.LastPongAt = now
	return nil
}

func (r *USDMRuntime) AcceptMarkPriceFrame(raw []byte, now time.Time) (USDMRawFrame, error) {
	if err := r.recordFrame(now); err != nil {
		return USDMRawFrame{}, err
	}
	r.state.LastMarkPriceAt = now
	return USDMRawFrame{RecvTime: now, Payload: append([]byte(nil), raw...)}, nil
}

func (r *USDMRuntime) AcceptForceOrderFrame(raw []byte, now time.Time) (USDMRawFrame, error) {
	if err := r.recordFrame(now); err != nil {
		return USDMRawFrame{}, err
	}
	return USDMRawFrame{RecvTime: now, Payload: append([]byte(nil), raw...)}, nil
}

func (r *USDMRuntime) UpdateClockOffset(offset time.Duration) {
	if r == nil {
		return
	}
	r.state.LocalClockOffset = offset
}

func (r *USDMRuntime) ShouldRollover(now time.Time) bool {
	if r == nil || now.IsZero() {
		return false
	}
	return r.connected && !r.state.NextRolloverAt.IsZero() && !now.Before(r.state.NextRolloverAt)
}

func (r *USDMRuntime) TriggerRollover(now time.Time) (USDMReconnectPlan, error) {
	if r == nil {
		return USDMReconnectPlan{}, fmt.Errorf("usdm runtime is required")
	}
	if !r.ShouldRollover(now) {
		return USDMReconnectPlan{}, fmt.Errorf("usdm websocket rollover is not due")
	}
	return r.beginReconnect(now, USDMReconnectCauseRollover)
}

func (r *USDMRuntime) HandleDisconnect(now time.Time, cause USDMReconnectCause) (USDMReconnectPlan, error) {
	if r == nil {
		return USDMReconnectPlan{}, fmt.Errorf("usdm runtime is required")
	}
	if cause == "" {
		return USDMReconnectPlan{}, fmt.Errorf("disconnect cause is required")
	}
	return r.beginReconnect(now, cause)
}

func (r *USDMRuntime) HealthStatus(now time.Time) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("usdm runtime is required")
	}
	return r.usdmRuntime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:       r.state.ConnectionState,
		Now:                   now,
		LastMessageAt:         r.state.LastMarkPriceAt,
		LocalClockOffset:      r.state.LocalClockOffset,
		ConsecutiveReconnects: r.state.ConsecutiveReconnects,
	})
}

func (r *USDMRuntime) FeedHealthInputs(now time.Time) ([]USDMFeedHealthInput, error) {
	if r == nil {
		return nil, fmt.Errorf("usdm runtime is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("feed health time is required")
	}
	status, err := r.HealthStatus(now)
	if err != nil {
		return nil, err
	}
	inputs := make([]USDMFeedHealthInput, 0, len(r.symbols))
	for _, binding := range r.symbols {
		inputs = append(inputs, USDMFeedHealthInput{
			Metadata: ingestion.FeedHealthMetadata{
				Symbol:        binding.Symbol,
				SourceSymbol:  binding.SourceSymbol,
				QuoteCurrency: binding.QuoteCurrency,
				Venue:         ingestion.VenueBinance,
				MarketType:    "perpetual",
			},
			Message: ingestion.FeedHealthMessage{
				ExchangeTs:     now.UTC().Format(time.RFC3339Nano),
				RecvTs:         now.UTC().Format(time.RFC3339Nano),
				SourceRecordID: usdmFeedHealthRecordPrefix + binding.SourceSymbol,
				Status:         status,
			},
		})
	}
	return inputs, nil
}

func (r *USDMRuntime) recordFrame(now time.Time) error {
	if r == nil {
		return fmt.Errorf("usdm runtime is required")
	}
	if now.IsZero() {
		return fmt.Errorf("frame time is required")
	}
	if !r.connected {
		return fmt.Errorf("usdm websocket session is not connected")
	}
	r.state.LastFrameAt = now
	return nil
}

func (r *USDMRuntime) beginReconnect(now time.Time, cause USDMReconnectCause) (USDMReconnectPlan, error) {
	if now.IsZero() {
		return USDMReconnectPlan{}, fmt.Errorf("reconnect time is required")
	}
	if !r.connected && !r.state.AwaitingConnect {
		return USDMReconnectPlan{}, fmt.Errorf("usdm websocket session is not active")
	}
	r.connected = false
	r.state.AwaitingConnect = false
	r.state.ConnectionState = ingestion.ConnectionReconnecting
	r.state.ConsecutiveReconnects++
	r.state.LastReconnectCause = cause
	r.state.LastFrameAt = now
	r.state.NextRolloverAt = time.Time{}
	r.state.ActiveSubscriptions = nil
	r.state.PendingCommands = map[int64]usdmPendingCommand{}
	delay, err := r.runtime.ReconnectDelay(r.state.ConsecutiveReconnects)
	if err != nil {
		return USDMReconnectPlan{}, err
	}
	r.state.NextReconnectAt = now.Add(delay)
	return USDMReconnectPlan{
		Cause:      cause,
		Attempt:    r.state.ConsecutiveReconnects,
		Delay:      delay,
		RetryAt:    r.state.NextReconnectAt,
		SessionRef: r.state.SessionRef,
	}, nil
}

func (r *USDMRuntime) newSubscribeCommand(now time.Time) USDMSubscribeCommand {
	r.nextCommand++
	command := USDMSubscribeCommand{
		ID:     r.nextCommand,
		Method: "SUBSCRIBE",
		Params: append([]string(nil), r.state.DesiredSubscriptions...),
	}
	r.state.PendingCommands[command.ID] = usdmPendingCommand{command: command, issuedAt: now}
	return command
}

func (r *USDMRuntime) pruneConnectAttempts(now time.Time) error {
	if now.IsZero() {
		return fmt.Errorf("current time is required")
	}
	windowStart := now.Add(-time.Minute)
	pruned := r.state.RecentConnectAttempts[:0]
	for _, attempt := range r.state.RecentConnectAttempts {
		if attempt.IsZero() {
			return fmt.Errorf("connect attempt time is required")
		}
		if attempt.After(now) {
			return fmt.Errorf("connect attempt cannot be in the future")
		}
		if attempt.Before(windowStart) {
			continue
		}
		pruned = append(pruned, attempt)
	}
	r.state.RecentConnectAttempts = append([]time.Time(nil), pruned...)
	return nil
}

func usdmRuntimeConfig(config ingestion.VenueRuntimeConfig) (ingestion.VenueRuntimeConfig, error) {
	filteredStreams := make([]ingestion.StreamDefinition, 0, len(config.Adapter.Streams))
	for _, stream := range config.Adapter.Streams {
		if stream.MarketType != "perpetual" {
			continue
		}
		if stream.Kind != ingestion.StreamFundingRate && stream.Kind != ingestion.StreamMarkIndex && stream.Kind != ingestion.StreamLiquidation {
			continue
		}
		filteredStreams = append(filteredStreams, ingestion.StreamDefinition{Kind: stream.Kind, MarketType: stream.MarketType, SnapshotRequired: false})
	}
	if len(filteredStreams) == 0 {
		return ingestion.VenueRuntimeConfig{}, fmt.Errorf("binance usdm runtime requires perpetual funding, mark-index, or liquidation streams")
	}
	filtered := config
	filtered.Adapter = config.Adapter
	filtered.Adapter.Streams = filteredStreams
	filtered.SnapshotRefreshRequired = false
	filtered.SnapshotRefreshInterval = 0
	return filtered, nil
}

func usdmDesiredSubscriptions(symbols []string, streams []ingestion.StreamDefinition) ([]string, error) {
	bindings, err := usdmSymbolBindings(symbols)
	if err != nil {
		return nil, err
	}
	var subscriptions []string
	for _, stream := range streams {
		for _, binding := range bindings {
			name, err := usdmStreamName(binding.SourceSymbol, stream.Kind)
			if err != nil {
				return nil, err
			}
			subscriptions = append(subscriptions, name)
		}
	}
	return subscriptions, nil
}

func usdmSymbolBindings(symbols []string) ([]usdmSymbolBinding, error) {
	bindings := make([]usdmSymbolBinding, 0, len(symbols))
	for _, symbol := range symbols {
		binding, err := newUSDMSymbolBinding(symbol)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

func newUSDMSymbolBinding(symbol string) (usdmSymbolBinding, error) {
	if !strings.HasSuffix(symbol, "-USD") {
		return usdmSymbolBinding{}, fmt.Errorf("unsupported usdm runtime symbol %q", symbol)
	}
	base := strings.TrimSuffix(symbol, "-USD")
	if base == "" {
		return usdmSymbolBinding{}, fmt.Errorf("unsupported usdm runtime symbol %q", symbol)
	}
	return usdmSymbolBinding{Symbol: symbol, SourceSymbol: base + "USDT", QuoteCurrency: "USDT"}, nil
}

func usdmStreamName(sourceSymbol string, kind ingestion.StreamKind) (string, error) {
	name := strings.ToLower(sourceSymbol)
	switch kind {
	case ingestion.StreamFundingRate, ingestion.StreamMarkIndex:
		return name + "@markPrice@1s", nil
	case ingestion.StreamLiquidation:
		return name + "@forceOrder", nil
	default:
		return "", fmt.Errorf("unsupported usdm stream kind %q", kind)
	}
}
