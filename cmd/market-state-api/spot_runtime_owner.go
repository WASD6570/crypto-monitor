package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	"github.com/gorilla/websocket"
)

const depthStreamSuffix = "@depth@100ms"

type binanceSpotRuntimeOwnerOptions struct {
	client       *http.Client
	baseURL      string
	websocketURL string
	now          func() time.Time
}

type binanceSpotRuntimeOwner struct {
	mu               sync.RWMutex
	started          bool
	stopped          bool
	runtime          *venuebinance.Runtime
	supervisor       *venuebinance.SpotWebsocketSupervisor
	snapshotFetcher  venuebinance.SpotDepthSnapshotFetcher
	heartbeatTimeout time.Duration
	now              func() time.Time
	websocketURL     string
	states           map[string]*spotRuntimeState
	sourceSymbols    map[string]*spotRuntimeState
	order            []string
	cancel           context.CancelFunc
	done             chan struct{}
	conn             *websocket.Conn
	lastErr          error
}

type spotRuntimeState struct {
	mu                     sync.RWMutex
	runtime                *venuebinance.Runtime
	fetcher                venuebinance.SpotDepthSnapshotFetcher
	binding                spotBinding
	bootstrap              *venuebinance.SpotDepthBootstrapOwner
	depth                  *venuebinance.SpotDepthRecoveryOwner
	lastObservation        marketstateapi.SpotCurrentStateObservation
	pendingObservation     marketstateapi.SpotCurrentStateObservation
	haveObservation        bool
	havePendingObservation bool
	publishable            bool
	awaitingDepthSync      bool
	lastAcceptedExchange   time.Time
	lastAcceptedRecv       time.Time
	connectionState        ingestion.ConnectionState
	localClockOffset       time.Duration
	consecutiveReconnects  int
}

type spotBinding struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
}

type binanceSubscribeAck struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
}

type binanceEventHeader struct {
	EventType    string          `json:"e"`
	EventTime    int64           `json:"E"`
	SourceSymbol string          `json:"s"`
	UpdateID     int64           `json:"u"`
	BestBidPrice json.RawMessage `json:"b"`
	BestAskPrice json.RawMessage `json:"a"`
	BestBidSize  json.RawMessage `json:"B"`
	BestAskSize  json.RawMessage `json:"A"`
}

type binanceSubscribeCommand struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}

type binanceSpotDepthSnapshotFetcher struct {
	client  *http.Client
	baseURL string
	now     func() time.Time
}

func newBinanceSpotRuntimeOwner(runtimeConfig ingestion.VenueRuntimeConfig, runtime *venuebinance.Runtime, options binanceSpotRuntimeOwnerOptions) (*binanceSpotRuntimeOwner, error) {
	if runtime == nil {
		return nil, fmt.Errorf("binance runtime is required")
	}
	if options.client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if options.baseURL == "" {
		return nil, fmt.Errorf("binance base url is required")
	}
	if options.websocketURL == "" {
		return nil, fmt.Errorf("binance websocket url is required")
	}
	if options.now == nil {
		options.now = func() time.Time { return time.Now().UTC() }
	}
	supervisor, err := venuebinance.NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		return nil, fmt.Errorf("create spot websocket supervisor: %w", err)
	}
	bindings, err := spotBindings(runtimeConfig.Symbols)
	if err != nil {
		return nil, err
	}
	fetcher := binanceSpotDepthSnapshotFetcher{client: options.client, baseURL: strings.TrimRight(options.baseURL, "/"), now: options.now}
	states := make(map[string]*spotRuntimeState, len(bindings))
	sourceSymbols := make(map[string]*spotRuntimeState, len(bindings))
	order := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		state, err := newSpotRuntimeState(binding, runtime, fetcher)
		if err != nil {
			return nil, fmt.Errorf("create runtime state for %s: %w", binding.Symbol, err)
		}
		states[binding.Symbol] = state
		sourceSymbols[binding.SourceSymbol] = state
		order = append(order, binding.Symbol)
	}
	return &binanceSpotRuntimeOwner{
		runtime:          runtime,
		supervisor:       supervisor,
		snapshotFetcher:  fetcher,
		heartbeatTimeout: runtimeConfig.HeartbeatTimeout,
		now:              options.now,
		websocketURL:     options.websocketURL,
		states:           states,
		sourceSymbols:    sourceSymbols,
		order:            order,
	}, nil
}

func newSpotRuntimeState(binding spotBinding, runtime *venuebinance.Runtime, fetcher venuebinance.SpotDepthSnapshotFetcher) (*spotRuntimeState, error) {
	depth, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, fetcher)
	if err != nil {
		return nil, err
	}
	bootstrap, err := venuebinance.NewSpotDepthBootstrapOwner(runtime, fetcher)
	if err != nil {
		return nil, err
	}
	return &spotRuntimeState{
		runtime:         runtime,
		fetcher:         fetcher,
		binding:         binding,
		bootstrap:       bootstrap,
		depth:           depth,
		connectionState: ingestion.ConnectionDisconnected,
	}, nil
}

func (o *binanceSpotRuntimeOwner) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.stopped {
		return fmt.Errorf("binance spot runtime owner is stopped")
	}
	if o.started {
		return nil
	}
	runCtx, cancel := context.WithCancel(context.Background())
	o.started = true
	o.cancel = cancel
	o.done = make(chan struct{})
	go o.run(runCtx, o.done)
	return nil
}

func (o *binanceSpotRuntimeOwner) Stop(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	cancel := o.cancel
	done := o.done
	conn := o.conn
	o.stopped = true
	o.started = false
	o.cancel = nil
	o.done = nil
	o.conn = nil
	o.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if conn != nil {
		_ = conn.Close()
	}
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (o *binanceSpotRuntimeOwner) Snapshot(ctx context.Context, now time.Time) (marketstateapi.SpotCurrentStateSnapshot, error) {
	if ctx == nil {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return marketstateapi.SpotCurrentStateSnapshot{}, err
	}
	if now.IsZero() {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("current time is required")
	}

	o.mu.RLock()
	defer o.mu.RUnlock()
	if !o.started {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("binance spot runtime owner is not started")
	}

	observations := make([]marketstateapi.SpotCurrentStateObservation, 0, len(o.order))
	for _, symbol := range o.order {
		state := o.states[symbol]
		if state == nil {
			continue
		}
		observation, ok, err := state.snapshot(now)
		if err != nil {
			return marketstateapi.SpotCurrentStateSnapshot{}, err
		}
		if ok {
			observations = append(observations, observation)
		}
	}
	return marketstateapi.SpotCurrentStateSnapshot{Observations: observations}, nil
}

func (o *binanceSpotRuntimeOwner) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	for {
		if ctx.Err() != nil {
			return
		}
		err := o.runSession(ctx)
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			return
		}
		o.mu.Lock()
		o.lastErr = err
		o.mu.Unlock()
		now := o.now()
		wait := time.Second
		o.mu.Lock()
		state := o.supervisor.State()
		o.conn = nil
		if state.ConnectionState == ingestion.ConnectionReconnecting && !state.NextReconnectAt.IsZero() {
			wait = time.Until(state.NextReconnectAt)
			o.mu.Unlock()
			if wait < 0 {
				wait = 0
			}
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			continue
		}
		plan, planErr := o.supervisor.HandleDisconnect(now, venuebinance.SpotReconnectCauseTransport)
		o.applySupervisorStateLocked()
		o.mu.Unlock()
		if planErr == nil {
			wait = time.Until(plan.RetryAt)
			if wait < 0 {
				wait = 0
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (o *binanceSpotRuntimeOwner) lastError() error {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.lastErr
}

func (o *binanceSpotRuntimeOwner) runSession(ctx context.Context) error {
	now := o.now()
	o.mu.Lock()
	if err := o.supervisor.StartConnect(now); err != nil {
		o.mu.Unlock()
		return err
	}
	o.applySupervisorStateLocked()
	o.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, o.websocketURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	if o.heartbeatTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
	}
	pingDone := make(chan struct{})
	defer close(pingDone)
	go o.pingLoop(ctx, conn, pingDone)

	conn.SetPingHandler(func(appData string) error {
		o.mu.Lock()
		defer o.mu.Unlock()
		if _, err := o.supervisor.HandlePing(o.now()); err != nil {
			return err
		}
		o.applySupervisorStateLocked()
		if o.heartbeatTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
		}
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})
	conn.SetPongHandler(func(string) error {
		o.mu.Lock()
		defer o.mu.Unlock()
		if err := o.supervisor.HandlePong(o.now()); err != nil {
			return err
		}
		o.applySupervisorStateLocked()
		if o.heartbeatTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
		}
		return nil
	})

	o.mu.Lock()
	o.conn = conn
	command, err := o.supervisor.CompleteConnect(o.now())
	if err != nil {
		o.conn = nil
		o.mu.Unlock()
		return err
	}
	o.applySupervisorStateLocked()
	o.mu.Unlock()

	if command == nil {
		return fmt.Errorf("spot runtime owner requires subscribe command")
	}
	if err := conn.WriteJSON(binanceSubscribeCommand{
		Method: command.Method,
		Params: append(command.Params, o.depthSubscriptions()...),
		ID:     command.ID,
	}); err != nil {
		return err
	}

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if o.heartbeatTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		if err := o.handlePayload(ctx, payload, o.now()); err != nil {
			return err
		}
	}
}

func (o *binanceSpotRuntimeOwner) handlePayload(ctx context.Context, payload []byte, now time.Time) error {
	if ackID, ok := subscribeAckID(payload); ok {
		o.mu.Lock()
		defer o.mu.Unlock()
		if err := o.supervisor.AckSubscribe(now, ackID); err != nil {
			return err
		}
		o.applySupervisorStateLocked()
		return nil
	}

	o.mu.Lock()
	frame, err := o.supervisor.AcceptDataFrame(payload, now)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	o.applySupervisorStateLocked()
	header, err := parseEventHeader(frame.Payload)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	state, ok := o.sourceSymbols[header.SourceSymbol]
	o.mu.Unlock()
	if !ok {
		return nil
	}
	switch header.EventType {
	case "trade":
		_, err = venuebinance.ParseTradeFrame(frame)
		return err
	case "bookTicker":
		parsed, err := venuebinance.ParseTopOfBookFrame(frame)
		if err != nil {
			return err
		}
		return state.recordTopOfBook(parsed)
	case "depthUpdate":
		return state.handleDepthFrame(ctx, frame)
	default:
		return nil
	}
}

func (o *binanceSpotRuntimeOwner) applySupervisorStateLocked() {
	state := o.supervisor.State()
	for _, symbol := range o.order {
		runtimeState := o.states[symbol]
		if runtimeState == nil {
			continue
		}
		runtimeState.mu.Lock()
		if runtimeState.connectionState == ingestion.ConnectionConnected && state.ConnectionState != ingestion.ConnectionConnected {
			if err := runtimeState.beginReconnectLocked(); err != nil {
				o.lastErr = err
			}
		}
		runtimeState.connectionState = state.ConnectionState
		runtimeState.localClockOffset = state.LocalClockOffset
		runtimeState.consecutiveReconnects = state.ConsecutiveReconnects
		runtimeState.mu.Unlock()
	}
}

func (o *binanceSpotRuntimeOwner) pingLoop(ctx context.Context, conn *websocket.Conn, done <-chan struct{}) {
	if o.heartbeatTimeout <= 0 {
		return
	}
	interval := o.heartbeatTimeout / 2
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			now := o.now()
			if o.triggerRolloverIfDue(conn, now) {
				return
			}
			o.refreshDueStates(ctx, now)
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
				return
			}
		}
	}
}

func (o *binanceSpotRuntimeOwner) triggerRolloverIfDue(conn *websocket.Conn, now time.Time) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.supervisor.ShouldRollover(now) {
		return false
	}
	if _, err := o.supervisor.TriggerRollover(now); err != nil {
		o.lastErr = err
		return false
	}
	o.applySupervisorStateLocked()
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "rollover"), time.Now().Add(time.Second))
	_ = conn.Close()
	return true
}

func (o *binanceSpotRuntimeOwner) refreshDueStates(ctx context.Context, now time.Time) {
	o.mu.RLock()
	states := make([]*spotRuntimeState, 0, len(o.order))
	for _, symbol := range o.order {
		if state := o.states[symbol]; state != nil {
			states = append(states, state)
		}
	}
	o.mu.RUnlock()
	for _, state := range states {
		if err := state.refreshIfDue(ctx, now); err != nil {
			o.mu.Lock()
			o.lastErr = err
			o.mu.Unlock()
		}
	}
}

func (o *binanceSpotRuntimeOwner) depthSubscriptions() []string {
	params := make([]string, 0, len(o.order))
	for _, symbol := range o.order {
		state := o.states[symbol]
		if state == nil {
			continue
		}
		params = append(params, strings.ToLower(state.binding.SourceSymbol)+depthStreamSuffix)
	}
	return params
}

func (s *spotRuntimeState) handleDepthFrame(ctx context.Context, frame venuebinance.SpotRawFrame) error {
	if s == nil {
		return fmt.Errorf("spot runtime state is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.depth.Status()
	switch status.State {
	case venuebinance.SpotDepthRecoveryIdle, venuebinance.SpotDepthRecoveryBootstrapFailed:
		return s.bootstrapDepth(ctx, frame)
	case venuebinance.SpotDepthRecoverySynchronized:
		if err := s.depth.AcceptSynchronizedDelta(frame); err != nil {
			return err
		}
		if s.depth.Status().Synchronized && s.haveObservation {
			s.publishable = true
			return nil
		}
		return s.tryRecoverDepth(ctx, frame.RecvTime)
	default:
		if err := s.depth.BufferRecoveryDelta(frame); err != nil {
			return err
		}
		return s.tryRecoverDepth(ctx, frame.RecvTime)
	}
}

func (s *spotRuntimeState) bootstrapDepth(ctx context.Context, frame venuebinance.SpotRawFrame) error {
	if err := s.bootstrap.BufferDelta(frame); err != nil {
		return err
	}
	sync, err := s.bootstrap.Synchronize(ctx)
	if err != nil {
		if markErr := s.depth.MarkBootstrapFailure(s.binding.SourceSymbol); markErr != nil {
			return markErr
		}
		return s.resetBootstrap()
	}
	if err := s.depth.StartSynchronized(sync); err != nil {
		return err
	}
	s.markDepthSynchronized()
	return s.resetBootstrap()
}

func (s *spotRuntimeState) tryRecoverDepth(ctx context.Context, now time.Time) error {
	sync, err := s.depth.Recover(ctx, now)
	if err != nil {
		return nil
	}
	if err := s.depth.StartSynchronized(sync); err != nil {
		return err
	}
	s.markDepthSynchronized()
	return nil
}

func (s *spotRuntimeState) refreshIfDue(ctx context.Context, now time.Time) error {
	if s == nil {
		return fmt.Errorf("spot runtime state is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.depth.Status()
	if status.State != venuebinance.SpotDepthRecoverySynchronized && status.State != venuebinance.SpotDepthRecoveryCooldownBlocked && status.State != venuebinance.SpotDepthRecoveryRateLimitBlocked {
		return nil
	}
	refreshStatus, err := s.depth.SnapshotRefreshStatus(now)
	if err != nil {
		return err
	}
	if !refreshStatus.Required || !refreshStatus.Due {
		return nil
	}
	sync, err := s.depth.Refresh(ctx, now)
	if err != nil {
		return nil
	}
	if err := s.depth.StartSynchronized(sync); err != nil {
		return err
	}
	s.markDepthSynchronized()
	return nil
}

func (s *spotRuntimeState) recordTopOfBook(parsed venuebinance.ParsedTopOfBook) error {
	if s == nil {
		return fmt.Errorf("spot runtime state is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if parsed.SourceSymbol != s.binding.SourceSymbol {
		return fmt.Errorf("top-of-book source symbol = %q, want %q", parsed.SourceSymbol, s.binding.SourceSymbol)
	}
	bestBidPrice, err := strconv.ParseFloat(parsed.Message.BestBidPrice, 64)
	if err != nil {
		return fmt.Errorf("parse best bid price: %w", err)
	}
	bestAskPrice, err := strconv.ParseFloat(parsed.Message.BestAskPrice, 64)
	if err != nil {
		return fmt.Errorf("parse best ask price: %w", err)
	}
	recvTs, err := time.Parse(time.RFC3339Nano, parsed.Message.RecvTs)
	if err != nil {
		return fmt.Errorf("parse recv timestamp: %w", err)
	}
	var exchangeTs time.Time
	if parsed.Message.ExchangeTs != "" {
		exchangeTs, err = time.Parse(time.RFC3339Nano, parsed.Message.ExchangeTs)
		if err != nil {
			return fmt.Errorf("parse exchange timestamp: %w", err)
		}
	}
	observation := marketstateapi.SpotCurrentStateObservation{
		Symbol:        s.binding.Symbol,
		SourceSymbol:  s.binding.SourceSymbol,
		QuoteCurrency: s.binding.QuoteCurrency,
		BestBidPrice:  bestBidPrice,
		BestAskPrice:  bestAskPrice,
		ExchangeTs:    exchangeTs,
		RecvTs:        recvTs,
	}
	if s.awaitingDepthSync {
		s.pendingObservation = observation
		s.havePendingObservation = true
		return nil
	}
	s.lastObservation = observation
	s.haveObservation = true
	s.lastAcceptedExchange = exchangeTs.UTC()
	s.lastAcceptedRecv = recvTs.UTC()
	if s.depth.Status().Synchronized {
		s.publishable = true
	}
	return nil
}

func (s *spotRuntimeState) snapshot(now time.Time) (marketstateapi.SpotCurrentStateObservation, bool, error) {
	if s == nil {
		return marketstateapi.SpotCurrentStateObservation{}, false, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.haveObservation || !s.publishable {
		return marketstateapi.SpotCurrentStateObservation{}, false, nil
	}
	feedHealth, err := s.depth.HealthStatus(now.UTC(), s.connectionState, s.localClockOffset, s.consecutiveReconnects)
	if err != nil {
		return marketstateapi.SpotCurrentStateObservation{}, false, err
	}
	observation := s.lastObservation
	observation.FeedHealth = feedHealth
	observation.DepthStatus = s.depth.Status()
	return observation, true, nil
}

func (s *spotRuntimeState) resetBootstrap() error {
	bootstrap, err := venuebinance.NewSpotDepthBootstrapOwner(s.runtime, s.fetcher)
	if err != nil {
		return err
	}
	s.bootstrap = bootstrap
	return nil
}

func (s *spotRuntimeState) beginReconnectLocked() error {
	depth, err := venuebinance.NewSpotDepthRecoveryOwner(s.runtime, s.fetcher)
	if err != nil {
		return err
	}
	bootstrap, err := venuebinance.NewSpotDepthBootstrapOwner(s.runtime, s.fetcher)
	if err != nil {
		return err
	}
	s.depth = depth
	s.bootstrap = bootstrap
	s.awaitingDepthSync = s.haveObservation
	s.havePendingObservation = false
	s.publishable = s.haveObservation
	return nil
}

func (s *spotRuntimeState) markDepthSynchronized() {
	if s.awaitingDepthSync {
		s.awaitingDepthSync = false
		if s.havePendingObservation {
			s.lastObservation = s.pendingObservation
			s.haveObservation = true
			s.lastAcceptedExchange = s.pendingObservation.ExchangeTs.UTC()
			s.lastAcceptedRecv = s.pendingObservation.RecvTs.UTC()
			s.havePendingObservation = false
		}
	}
	if s.haveObservation {
		s.publishable = true
	}
}

func spotBindings(symbols []string) ([]spotBinding, error) {
	bindings := make([]spotBinding, 0, len(symbols))
	for _, symbol := range symbols {
		if !strings.HasSuffix(symbol, "-USD") {
			return nil, fmt.Errorf("unsupported spot symbol %q", symbol)
		}
		base := strings.TrimSuffix(symbol, "-USD")
		if base == "" {
			return nil, fmt.Errorf("unsupported spot symbol %q", symbol)
		}
		bindings = append(bindings, spotBinding{
			Symbol:        symbol,
			SourceSymbol:  base + "USDT",
			QuoteCurrency: "USDT",
		})
	}
	return bindings, nil
}

func subscribeAckID(raw []byte) (int64, bool) {
	var ack binanceSubscribeAck
	if err := json.Unmarshal(raw, &ack); err != nil {
		return 0, false
	}
	if ack.ID == 0 || string(ack.Result) != "null" {
		return 0, false
	}
	return ack.ID, true
}

func parseEventHeader(raw []byte) (binanceEventHeader, error) {
	var header binanceEventHeader
	if err := json.Unmarshal(raw, &header); err != nil {
		return binanceEventHeader{}, fmt.Errorf("decode binance event header: %w", err)
	}
	if header.EventType == "" && header.SourceSymbol != "" && header.UpdateID > 0 && len(header.BestBidPrice) > 0 && len(header.BestAskPrice) > 0 && len(header.BestBidSize) > 0 && len(header.BestAskSize) > 0 {
		header.EventType = "bookTicker"
	}
	if header.EventType == "" || header.SourceSymbol == "" {
		return binanceEventHeader{}, fmt.Errorf("binance event type and source symbol are required")
	}
	return header, nil
}

func (f binanceSpotDepthSnapshotFetcher) FetchSpotDepthSnapshot(ctx context.Context, sourceSymbol string) (venuebinance.SpotDepthSnapshotResponse, error) {
	if ctx == nil {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("context is required")
	}
	if sourceSymbol == "" {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("source symbol is required")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, f.baseURL+"/api/v3/depth?symbol="+sourceSymbol+"&limit=5", nil)
	if err != nil {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("create depth request: %w", err)
	}
	response, err := f.client.Do(request)
	if err != nil {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("request depth snapshot: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("depth snapshot status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("read depth snapshot: %w", err)
	}
	return venuebinance.SpotDepthSnapshotResponse{Payload: body, RecvTime: f.now()}, nil
}
