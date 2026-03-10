package venuebinance

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const (
	spotRolloverHeadroom = 5 * time.Minute
	spotConnectionMaxAge = 24 * time.Hour

	spotFeedHealthRecordPrefix = "runtime:binance-spot-ws:"
	spotSessionRefPrefix       = "binance-spot-session-"
)

type SpotReconnectCause string

const (
	SpotReconnectCauseTransport SpotReconnectCause = "transport-close"
	SpotReconnectCauseRollover  SpotReconnectCause = "rollover"
)

type SpotSubscribeCommand struct {
	ID     int64
	Method string
	Params []string
}

type SpotRawFrame struct {
	RecvTime time.Time
	Payload  []byte
}

type SpotFeedHealthInput struct {
	Metadata ingestion.FeedHealthMetadata
	Message  ingestion.FeedHealthMessage
}

type SpotReconnectPlan struct {
	Cause      SpotReconnectCause
	Attempt    int
	Delay      time.Duration
	RetryAt    time.Time
	SessionRef string
}

type SpotWebsocketSupervisorState struct {
	DesiredSubscriptions  []string
	ActiveSubscriptions   []string
	PendingSubscribeIDs   []int64
	ConnectionState       ingestion.ConnectionState
	ConnectionOpenedAt    time.Time
	LastFrameAt           time.Time
	LastPongAt            time.Time
	LastSubscribeAckAt    time.Time
	NextReconnectAt       time.Time
	NextRolloverAt        time.Time
	LastReconnectCause    SpotReconnectCause
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	RecentConnectAttempts []time.Time
	SessionRef            string
}

type SpotWebsocketSupervisor struct {
	runtime     *Runtime
	spotRuntime *Runtime
	symbols     []spotSymbolBinding
	state       spotWebsocketSupervisorState
	nextCommand int64
	sessionSeq  int
	connected   bool
}

type spotSymbolBinding struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
}

type spotPendingCommand struct {
	command  SpotSubscribeCommand
	issuedAt time.Time
}

type spotWebsocketSupervisorState struct {
	DesiredSubscriptions  []string
	ActiveSubscriptions   []string
	PendingCommands       map[int64]spotPendingCommand
	ConnectionState       ingestion.ConnectionState
	ConnectionOpenedAt    time.Time
	LastFrameAt           time.Time
	LastPongAt            time.Time
	LastSubscribeAckAt    time.Time
	NextReconnectAt       time.Time
	NextRolloverAt        time.Time
	LastReconnectCause    SpotReconnectCause
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	RecentConnectAttempts []time.Time
	SessionRef            string
	SessionGeneration     int
	AwaitingConnect       bool
}

func NewSpotWebsocketSupervisor(runtime *Runtime) (*SpotWebsocketSupervisor, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}

	spotRuntimeConfig, err := spotSupervisorRuntimeConfig(runtime.config)
	if err != nil {
		return nil, err
	}
	spotRuntime, err := NewRuntime(spotRuntimeConfig)
	if err != nil {
		return nil, err
	}
	bindings, err := spotSymbolBindings(runtime.config.Symbols)
	if err != nil {
		return nil, err
	}
	desiredSubscriptions, err := spotDesiredSubscriptions(spotRuntimeConfig.Symbols, spotRuntimeConfig.Adapter.Streams)
	if err != nil {
		return nil, err
	}

	return &SpotWebsocketSupervisor{
		runtime:     runtime,
		spotRuntime: spotRuntime,
		symbols:     bindings,
		state: spotWebsocketSupervisorState{
			DesiredSubscriptions: append([]string(nil), desiredSubscriptions...),
			PendingCommands:      map[int64]spotPendingCommand{},
			ConnectionState:      ingestion.ConnectionDisconnected,
		},
	}, nil
}

func (s *SpotWebsocketSupervisor) State() SpotWebsocketSupervisorState {
	if s == nil {
		return SpotWebsocketSupervisorState{}
	}
	pendingIDs := make([]int64, 0, len(s.state.PendingCommands))
	for id := range s.state.PendingCommands {
		pendingIDs = append(pendingIDs, id)
	}
	slices.Sort(pendingIDs)
	connectAttempts := append([]time.Time(nil), s.state.RecentConnectAttempts...)
	return SpotWebsocketSupervisorState{
		DesiredSubscriptions:  append([]string(nil), s.state.DesiredSubscriptions...),
		ActiveSubscriptions:   append([]string(nil), s.state.ActiveSubscriptions...),
		PendingSubscribeIDs:   pendingIDs,
		ConnectionState:       s.state.ConnectionState,
		ConnectionOpenedAt:    s.state.ConnectionOpenedAt,
		LastFrameAt:           s.state.LastFrameAt,
		LastPongAt:            s.state.LastPongAt,
		LastSubscribeAckAt:    s.state.LastSubscribeAckAt,
		NextReconnectAt:       s.state.NextReconnectAt,
		NextRolloverAt:        s.state.NextRolloverAt,
		LastReconnectCause:    s.state.LastReconnectCause,
		LocalClockOffset:      s.state.LocalClockOffset,
		ConsecutiveReconnects: s.state.ConsecutiveReconnects,
		RecentConnectAttempts: connectAttempts,
		SessionRef:            s.state.SessionRef,
	}
}

func (s *SpotWebsocketSupervisor) StartConnect(now time.Time) error {
	if s == nil {
		return fmt.Errorf("spot websocket supervisor is required")
	}
	if now.IsZero() {
		return fmt.Errorf("connect time is required")
	}
	if s.connected || s.state.AwaitingConnect {
		return fmt.Errorf("spot websocket session is already active")
	}
	if !s.state.NextReconnectAt.IsZero() && now.Before(s.state.NextReconnectAt) {
		return fmt.Errorf("spot websocket reconnect backoff remains until %s", s.state.NextReconnectAt.UTC().Format(time.RFC3339Nano))
	}
	if err := s.pruneConnectAttempts(now); err != nil {
		return err
	}
	if len(s.state.RecentConnectAttempts) >= s.runtime.config.ConnectsPerMinuteLimit {
		return fmt.Errorf("spot websocket connect rate limit reached")
	}

	s.state.RecentConnectAttempts = append(s.state.RecentConnectAttempts, now)
	s.state.ConnectionState = ingestion.ConnectionConnecting
	s.state.AwaitingConnect = true
	s.sessionSeq++
	s.state.SessionGeneration = s.sessionSeq
	s.state.SessionRef = fmt.Sprintf("%s%d", spotSessionRefPrefix, s.sessionSeq)
	return nil
}

func (s *SpotWebsocketSupervisor) CompleteConnect(now time.Time) (*SpotSubscribeCommand, error) {
	if s == nil {
		return nil, fmt.Errorf("spot websocket supervisor is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("connect completion time is required")
	}
	if !s.state.AwaitingConnect {
		return nil, fmt.Errorf("spot websocket connect has not started")
	}

	s.connected = true
	s.state.AwaitingConnect = false
	s.state.ConnectionState = ingestion.ConnectionConnected
	s.state.ConnectionOpenedAt = now
	s.state.LastFrameAt = now
	s.state.NextReconnectAt = time.Time{}
	s.state.NextRolloverAt = now.Add(spotConnectionMaxAge - spotRolloverHeadroom)
	s.state.PendingCommands = map[int64]spotPendingCommand{}
	s.state.ActiveSubscriptions = nil

	command := s.newSubscribeCommand(now)
	if !s.runtime.config.ResubscribeOnReconnect && s.state.ConsecutiveReconnects > 0 {
		return nil, nil
	}
	return &command, nil
}

func (s *SpotWebsocketSupervisor) AckSubscribe(now time.Time, commandID int64) error {
	if s == nil {
		return fmt.Errorf("spot websocket supervisor is required")
	}
	if now.IsZero() {
		return fmt.Errorf("subscribe ack time is required")
	}
	pending, ok := s.state.PendingCommands[commandID]
	if !ok {
		return fmt.Errorf("unknown spot subscribe command id %d", commandID)
	}

	delete(s.state.PendingCommands, commandID)
	s.state.ActiveSubscriptions = append([]string(nil), pending.command.Params...)
	s.state.LastFrameAt = now
	s.state.LastSubscribeAckAt = now
	s.state.ConsecutiveReconnects = 0
	return nil
}

func (s *SpotWebsocketSupervisor) HandlePing(now time.Time) (bool, error) {
	if err := s.recordFrame(now); err != nil {
		return false, err
	}
	s.state.LastPongAt = now
	return true, nil
}

func (s *SpotWebsocketSupervisor) HandlePong(now time.Time) error {
	if err := s.recordFrame(now); err != nil {
		return err
	}
	s.state.LastPongAt = now
	return nil
}

func (s *SpotWebsocketSupervisor) AcceptDataFrame(raw []byte, now time.Time) (SpotRawFrame, error) {
	if len(raw) == 0 {
		return SpotRawFrame{}, fmt.Errorf("raw payload is required")
	}
	if err := s.recordFrame(now); err != nil {
		return SpotRawFrame{}, err
	}
	return SpotRawFrame{RecvTime: now, Payload: append([]byte(nil), raw...)}, nil
}

func (s *SpotWebsocketSupervisor) UpdateClockOffset(offset time.Duration) {
	if s == nil {
		return
	}
	s.state.LocalClockOffset = offset
}

func (s *SpotWebsocketSupervisor) ShouldRollover(now time.Time) bool {
	if s == nil || now.IsZero() {
		return false
	}
	return s.connected && !s.state.NextRolloverAt.IsZero() && !now.Before(s.state.NextRolloverAt)
}

func (s *SpotWebsocketSupervisor) TriggerRollover(now time.Time) (SpotReconnectPlan, error) {
	if s == nil {
		return SpotReconnectPlan{}, fmt.Errorf("spot websocket supervisor is required")
	}
	if !s.ShouldRollover(now) {
		return SpotReconnectPlan{}, fmt.Errorf("spot websocket rollover is not due")
	}
	return s.beginReconnect(now, SpotReconnectCauseRollover)
}

func (s *SpotWebsocketSupervisor) HandleDisconnect(now time.Time, cause SpotReconnectCause) (SpotReconnectPlan, error) {
	if s == nil {
		return SpotReconnectPlan{}, fmt.Errorf("spot websocket supervisor is required")
	}
	if cause == "" {
		return SpotReconnectPlan{}, fmt.Errorf("disconnect cause is required")
	}
	return s.beginReconnect(now, cause)
}

func (s *SpotWebsocketSupervisor) HealthStatus(now time.Time) (ingestion.FeedHealthStatus, error) {
	if s == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("spot websocket supervisor is required")
	}
	return s.spotRuntime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:       s.state.ConnectionState,
		Now:                   now,
		LastMessageAt:         s.state.LastFrameAt,
		LocalClockOffset:      s.state.LocalClockOffset,
		ConsecutiveReconnects: s.state.ConsecutiveReconnects,
	})
}

func (s *SpotWebsocketSupervisor) FeedHealthInputs(now time.Time) ([]SpotFeedHealthInput, error) {
	if s == nil {
		return nil, fmt.Errorf("spot websocket supervisor is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("feed health time is required")
	}
	status, err := s.HealthStatus(now)
	if err != nil {
		return nil, err
	}
	inputs := make([]SpotFeedHealthInput, 0, len(s.symbols))
	for _, binding := range s.symbols {
		inputs = append(inputs, SpotFeedHealthInput{
			Metadata: ingestion.FeedHealthMetadata{
				Symbol:        binding.Symbol,
				SourceSymbol:  binding.SourceSymbol,
				QuoteCurrency: binding.QuoteCurrency,
				Venue:         ingestion.VenueBinance,
				MarketType:    "spot",
			},
			Message: ingestion.FeedHealthMessage{
				ExchangeTs:     now.UTC().Format(time.RFC3339Nano),
				RecvTs:         now.UTC().Format(time.RFC3339Nano),
				SourceRecordID: spotFeedHealthRecordPrefix + binding.SourceSymbol,
				Status:         status,
			},
		})
	}
	return inputs, nil
}

func (s *SpotWebsocketSupervisor) recordFrame(now time.Time) error {
	if s == nil {
		return fmt.Errorf("spot websocket supervisor is required")
	}
	if now.IsZero() {
		return fmt.Errorf("frame time is required")
	}
	if !s.connected {
		return fmt.Errorf("spot websocket session is not connected")
	}
	s.state.LastFrameAt = now
	return nil
}

func (s *SpotWebsocketSupervisor) beginReconnect(now time.Time, cause SpotReconnectCause) (SpotReconnectPlan, error) {
	if now.IsZero() {
		return SpotReconnectPlan{}, fmt.Errorf("reconnect time is required")
	}
	if !s.connected && !s.state.AwaitingConnect {
		return SpotReconnectPlan{}, fmt.Errorf("spot websocket session is not active")
	}
	s.connected = false
	s.state.AwaitingConnect = false
	s.state.ConnectionState = ingestion.ConnectionReconnecting
	s.state.ConsecutiveReconnects++
	s.state.LastReconnectCause = cause
	s.state.LastFrameAt = now
	s.state.NextRolloverAt = time.Time{}
	s.state.ActiveSubscriptions = nil
	s.state.PendingCommands = map[int64]spotPendingCommand{}
	delay, err := s.runtime.ReconnectDelay(s.state.ConsecutiveReconnects)
	if err != nil {
		return SpotReconnectPlan{}, err
	}
	s.state.NextReconnectAt = now.Add(delay)
	plan := SpotReconnectPlan{
		Cause:      cause,
		Attempt:    s.state.ConsecutiveReconnects,
		Delay:      delay,
		RetryAt:    s.state.NextReconnectAt,
		SessionRef: s.state.SessionRef,
	}
	return plan, nil
}

func (s *SpotWebsocketSupervisor) newSubscribeCommand(now time.Time) SpotSubscribeCommand {
	s.nextCommand++
	command := SpotSubscribeCommand{
		ID:     s.nextCommand,
		Method: "SUBSCRIBE",
		Params: append([]string(nil), s.state.DesiredSubscriptions...),
	}
	s.state.PendingCommands[command.ID] = spotPendingCommand{command: command, issuedAt: now}
	return command
}

func (s *SpotWebsocketSupervisor) pruneConnectAttempts(now time.Time) error {
	if now.IsZero() {
		return fmt.Errorf("current time is required")
	}
	windowStart := now.Add(-time.Minute)
	pruned := s.state.RecentConnectAttempts[:0]
	for _, attempt := range s.state.RecentConnectAttempts {
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
	s.state.RecentConnectAttempts = append([]time.Time(nil), pruned...)
	return nil
}

func spotSupervisorRuntimeConfig(config ingestion.VenueRuntimeConfig) (ingestion.VenueRuntimeConfig, error) {
	filteredStreams := make([]ingestion.StreamDefinition, 0, len(config.Adapter.Streams))
	for _, stream := range config.Adapter.Streams {
		if stream.MarketType != "spot" {
			continue
		}
		if stream.Kind != ingestion.StreamTrades && stream.Kind != ingestion.StreamTopOfBook {
			continue
		}
		filteredStreams = append(filteredStreams, ingestion.StreamDefinition{
			Kind:             stream.Kind,
			MarketType:       stream.MarketType,
			SnapshotRequired: false,
		})
	}
	if len(filteredStreams) == 0 {
		return ingestion.VenueRuntimeConfig{}, fmt.Errorf("binance spot supervisor requires spot trade or top-of-book streams")
	}
	filtered := config
	filtered.Adapter = config.Adapter
	filtered.Adapter.Streams = filteredStreams
	filtered.SnapshotRefreshRequired = false
	filtered.SnapshotRefreshInterval = 0
	return filtered, nil
}

func spotDesiredSubscriptions(symbols []string, streams []ingestion.StreamDefinition) ([]string, error) {
	bindings, err := spotSymbolBindings(symbols)
	if err != nil {
		return nil, err
	}
	var subscriptions []string
	for _, stream := range streams {
		for _, binding := range bindings {
			name, err := spotStreamName(binding.SourceSymbol, stream.Kind)
			if err != nil {
				return nil, err
			}
			subscriptions = append(subscriptions, name)
		}
	}
	return subscriptions, nil
}

func spotSymbolBindings(symbols []string) ([]spotSymbolBinding, error) {
	bindings := make([]spotSymbolBinding, 0, len(symbols))
	for _, symbol := range symbols {
		binding, err := newSpotSymbolBinding(symbol)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

func newSpotSymbolBinding(symbol string) (spotSymbolBinding, error) {
	if !strings.HasSuffix(symbol, "-USD") {
		return spotSymbolBinding{}, fmt.Errorf("unsupported spot supervisor symbol %q", symbol)
	}
	base := strings.TrimSuffix(symbol, "-USD")
	if base == "" {
		return spotSymbolBinding{}, fmt.Errorf("unsupported spot supervisor symbol %q", symbol)
	}
	return spotSymbolBinding{
		Symbol:        symbol,
		SourceSymbol:  base + "USDT",
		QuoteCurrency: "USDT",
	}, nil
}

func spotStreamName(sourceSymbol string, kind ingestion.StreamKind) (string, error) {
	name := strings.ToLower(sourceSymbol)
	switch kind {
	case ingestion.StreamTrades:
		return name + "@trade", nil
	case ingestion.StreamTopOfBook:
		return name + "@bookTicker", nil
	default:
		return "", fmt.Errorf("unsupported spot stream kind %q", kind)
	}
}
