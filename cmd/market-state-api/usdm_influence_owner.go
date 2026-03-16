package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	"github.com/gorilla/websocket"
)

type binanceUSDMInfluenceOwnerOptions struct {
	client           *http.Client
	baseURL          string
	websocketURL     string
	heartbeatTimeout time.Duration
	now              func() time.Time
}

type binanceUSDMInfluenceOwner struct {
	mu               sync.RWMutex
	stateMu          sync.Mutex
	started          bool
	stopped          bool
	runtime          *venuebinance.USDMRuntime
	inputOwner       *venuebinance.USDMInfluenceInputOwner
	poller           *venuebinance.USDMOpenInterestPoller
	normalizer       *normalizer.Service
	client           *http.Client
	baseURL          string
	websocketURL     string
	heartbeatTimeout time.Duration
	now              func() time.Time
	bindings         map[string]usdmBinding
	cancel           context.CancelFunc
	done             chan struct{}
	conn             *websocket.Conn
	lastErr          error
}

type usdmBinding struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
}

type combinedCurrentStateReader struct {
	spot *binanceSpotRuntimeOwner
	usdm *binanceUSDMInfluenceOwner
}

type usdmEventEnvelope struct {
	EventType    string `json:"e"`
	SourceSymbol string `json:"s"`
	Order        struct {
		SourceSymbol string `json:"s"`
	} `json:"o"`
}

func (r *combinedCurrentStateReader) Snapshot(ctx context.Context, now time.Time) (marketstateapi.SpotCurrentStateSnapshot, error) {
	if r == nil || r.spot == nil {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("spot runtime owner is required")
	}
	return r.spot.Snapshot(ctx, now)
}

func (r *combinedCurrentStateReader) SnapshotUSDMInfluenceInput(ctx context.Context, now time.Time) (features.USDMInfluenceEvaluatorInput, error) {
	if r == nil || r.usdm == nil {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("usdm influence owner is required")
	}
	return r.usdm.SnapshotUSDMInfluenceInput(ctx, now)
}

func newBinanceUSDMInfluenceOwner(runtimeConfig ingestion.VenueRuntimeConfig, runtime *venuebinance.Runtime, options binanceUSDMInfluenceOwnerOptions) (*binanceUSDMInfluenceOwner, error) {
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
	usdmRuntime, err := venuebinance.NewUSDMRuntime(runtime)
	if err != nil {
		return nil, err
	}
	inputOwner, err := venuebinance.NewUSDMInfluenceInputOwner(runtime)
	if err != nil {
		return nil, err
	}
	poller, err := venuebinance.NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		return nil, err
	}
	normalizerService, err := normalizer.NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		return nil, err
	}
	bindings, err := usdmBindings(runtimeConfig.Symbols)
	if err != nil {
		return nil, err
	}
	bySource := make(map[string]usdmBinding, len(bindings))
	for _, binding := range bindings {
		bySource[binding.SourceSymbol] = binding
	}
	return &binanceUSDMInfluenceOwner{
		runtime:          usdmRuntime,
		inputOwner:       inputOwner,
		poller:           poller,
		normalizer:       normalizerService,
		client:           options.client,
		baseURL:          strings.TrimRight(options.baseURL, "/"),
		websocketURL:     options.websocketURL,
		heartbeatTimeout: options.heartbeatTimeout,
		now:              options.now,
		bindings:         bySource,
	}, nil
}

func (o *binanceUSDMInfluenceOwner) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.stopped {
		return fmt.Errorf("binance usdm influence owner is stopped")
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

func (o *binanceUSDMInfluenceOwner) Stop(ctx context.Context) error {
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

func (o *binanceUSDMInfluenceOwner) SnapshotUSDMInfluenceInput(ctx context.Context, now time.Time) (features.USDMInfluenceEvaluatorInput, error) {
	if ctx == nil {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return features.USDMInfluenceEvaluatorInput{}, err
	}
	if now.IsZero() {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("current time is required")
	}
	o.mu.RLock()
	started := o.started
	o.mu.RUnlock()
	if !started {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("binance usdm influence owner is not started")
	}
	if err := o.refreshFeedHealth(now.UTC()); err != nil {
		return features.USDMInfluenceEvaluatorInput{}, err
	}
	return o.inputOwner.Snapshot(now.UTC())
}

func (o *binanceUSDMInfluenceOwner) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	pollDone := make(chan struct{})
	go o.pollLoop(ctx, pollDone)
	defer func() { <-pollDone }()
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
		o.conn = nil
		o.mu.Unlock()
		wait := time.Second
		o.stateMu.Lock()
		plan, planErr := o.runtime.HandleDisconnect(o.now().UTC(), venuebinance.USDMReconnectCauseTransport)
		o.stateMu.Unlock()
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

func (o *binanceUSDMInfluenceOwner) pollLoop(ctx context.Context, done chan struct{}) {
	defer close(done)
	_ = o.refreshOpenInterest(ctx, o.now().UTC())
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = o.refreshOpenInterest(ctx, o.now().UTC())
		}
	}
}

func (o *binanceUSDMInfluenceOwner) runSession(ctx context.Context) error {
	now := o.now().UTC()
	o.stateMu.Lock()
	if err := o.runtime.StartConnect(now); err != nil {
		o.stateMu.Unlock()
		return err
	}
	o.stateMu.Unlock()
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
		o.stateMu.Lock()
		defer o.stateMu.Unlock()
		if _, err := o.runtime.HandlePing(o.now().UTC()); err != nil {
			return err
		}
		if o.heartbeatTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
		}
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})
	conn.SetPongHandler(func(string) error {
		o.stateMu.Lock()
		defer o.stateMu.Unlock()
		if err := o.runtime.HandlePong(o.now().UTC()); err != nil {
			return err
		}
		if o.heartbeatTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(o.heartbeatTimeout))
		}
		return nil
	})
	o.stateMu.Lock()
	command, err := o.runtime.CompleteConnect(now)
	o.stateMu.Unlock()
	if err != nil {
		return err
	}
	if command == nil {
		return fmt.Errorf("usdm runtime owner requires subscribe command")
	}
	o.mu.Lock()
	o.conn = conn
	o.mu.Unlock()
	if err := conn.WriteJSON(binanceSubscribeCommand{Method: command.Method, Params: append([]string(nil), command.Params...), ID: command.ID}); err != nil {
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
		if err := o.handlePayload(payload, o.now().UTC()); err != nil {
			return err
		}
	}
}

func (o *binanceUSDMInfluenceOwner) handlePayload(payload []byte, now time.Time) error {
	if ackID, ok := subscribeAckID(payload); ok {
		o.stateMu.Lock()
		defer o.stateMu.Unlock()
		return o.runtime.AckSubscribe(now, ackID)
	}
	var envelope usdmEventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return fmt.Errorf("decode usdm event envelope: %w", err)
	}
	sourceSymbol := envelope.SourceSymbol
	if sourceSymbol == "" {
		sourceSymbol = envelope.Order.SourceSymbol
	}
	binding, ok := o.bindings[sourceSymbol]
	if !ok {
		return nil
	}
	metadata := derivativesMetadata(binding)
	switch envelope.EventType {
	case "markPriceUpdate":
		o.stateMu.Lock()
		frame, err := o.runtime.AcceptMarkPriceFrame(payload, now)
		o.stateMu.Unlock()
		if err != nil {
			return err
		}
		parsed, err := venuebinance.ParseUSDMMarkPrice(frame.Payload, frame.RecvTime)
		if err != nil {
			return err
		}
		funding, err := o.normalizer.NormalizeFunding(normalizer.FundingInput{Metadata: metadata, Message: parsed.Funding})
		if err != nil {
			return err
		}
		if err := o.inputOwner.AcceptFunding(funding); err != nil {
			return err
		}
		markIndex, err := o.normalizer.NormalizeMarkIndex(normalizer.MarkIndexInput{Metadata: metadata, Message: parsed.MarkIndex})
		if err != nil {
			return err
		}
		return o.inputOwner.AcceptMarkIndex(markIndex)
	case "forceOrder":
		o.stateMu.Lock()
		frame, err := o.runtime.AcceptForceOrderFrame(payload, now)
		o.stateMu.Unlock()
		if err != nil {
			return err
		}
		parsed, err := venuebinance.ParseUSDMForceOrder(frame.Payload, frame.RecvTime)
		if err != nil {
			return err
		}
		liquidation, err := o.normalizer.NormalizeLiquidation(normalizer.LiquidationInput{Metadata: metadata, Message: parsed.Message})
		if err != nil {
			return err
		}
		return o.inputOwner.AcceptLiquidation(liquidation)
	default:
		return nil
	}
}

func (o *binanceUSDMInfluenceOwner) pingLoop(ctx context.Context, conn *websocket.Conn, done <-chan struct{}) {
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
			now := o.now().UTC()
			o.stateMu.Lock()
			shouldRollover := o.runtime.ShouldRollover(now)
			var rolloverErr error
			if shouldRollover {
				_, rolloverErr = o.runtime.TriggerRollover(now)
			}
			o.stateMu.Unlock()
			if shouldRollover && rolloverErr == nil {
					_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "rollover"), time.Now().Add(time.Second))
					_ = conn.Close()
					return
			}
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
				return
			}
		}
	}
}

func (o *binanceUSDMInfluenceOwner) refreshOpenInterest(ctx context.Context, now time.Time) error {
	o.stateMu.Lock()
	plans, err := o.poller.DuePolls(now)
	o.stateMu.Unlock()
	if err != nil {
		return err
	}
	for _, plan := range plans {
		if !plan.Allowed {
			continue
		}
		o.stateMu.Lock()
		request, err := o.poller.BeginPoll(plan.Request.SourceSymbol, now)
		o.stateMu.Unlock()
		if err != nil {
			o.recordError(err)
			continue
		}
		payload, err := o.fetchOpenInterest(ctx, request.Path)
		if err != nil {
			o.recordError(err)
			continue
		}
		o.stateMu.Lock()
		parsed, err := o.poller.AcceptSnapshot(payload, now)
		o.stateMu.Unlock()
		if err != nil {
			o.recordError(err)
			continue
		}
		binding, ok := o.bindings[parsed.SourceSymbol]
		if !ok {
			return fmt.Errorf("unsupported usdm open interest source symbol %q", parsed.SourceSymbol)
		}
		event, err := o.normalizer.NormalizeOpenInterest(normalizer.OpenInterestInput{Metadata: derivativesMetadata(binding), Message: parsed.Message})
		if err != nil {
			o.recordError(err)
			continue
		}
		if err := o.inputOwner.AcceptOpenInterest(event); err != nil {
			return err
		}
	}
	return nil
}

func (o *binanceUSDMInfluenceOwner) refreshFeedHealth(now time.Time) error {
	o.stateMu.Lock()
	websocketInputs, err := o.runtime.FeedHealthInputs(now)
	o.stateMu.Unlock()
	if err != nil {
		return err
	}
	for _, input := range websocketInputs {
		event, err := o.normalizer.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
		if err != nil {
			return err
		}
		if err := o.inputOwner.ObserveWebsocketFeedHealth(event); err != nil {
			return err
		}
	}
	o.stateMu.Lock()
	openInterestInputs, err := o.poller.FeedHealthInputs(now)
	o.stateMu.Unlock()
	if err != nil {
		return err
	}
	for _, input := range openInterestInputs {
		event, err := o.normalizer.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
		if err != nil {
			return err
		}
		if err := o.inputOwner.ObserveOpenInterestFeedHealth(event); err != nil {
			return err
		}
	}
	return nil
}

func (o *binanceUSDMInfluenceOwner) fetchOpenInterest(ctx context.Context, path string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create open interest request: %w", err)
	}
	response, err := o.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request open interest: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open interest status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read open interest response: %w", err)
	}
	return body, nil
}

func derivativesMetadata(binding usdmBinding) ingestion.DerivativesMetadata {
	return ingestion.DerivativesMetadata{
		Symbol:        binding.Symbol,
		SourceSymbol:  binding.SourceSymbol,
		QuoteCurrency: binding.QuoteCurrency,
		Venue:         ingestion.VenueBinance,
		MarketType:    "perpetual",
	}
}

func usdmBindings(symbols []string) ([]usdmBinding, error) {
	bindings := make([]usdmBinding, 0, len(symbols))
	for _, symbol := range symbols {
		if !strings.HasSuffix(symbol, "-USD") {
			return nil, fmt.Errorf("unsupported usdm symbol %q", symbol)
		}
		base := strings.TrimSuffix(symbol, "-USD")
		if base == "" {
			return nil, fmt.Errorf("unsupported usdm symbol %q", symbol)
		}
		bindings = append(bindings, usdmBinding{Symbol: symbol, SourceSymbol: base + "USDT", QuoteCurrency: "USDT"})
	}
	return bindings, nil
}

func (o *binanceUSDMInfluenceOwner) recordError(err error) {
	if err == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.lastErr = err
}
