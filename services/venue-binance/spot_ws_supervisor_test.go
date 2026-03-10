package venuebinance

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
)

func TestSpotSubscriptionStateDerivesDesiredSubscriptionsFromConfigOrder(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}

	state := supervisor.State()
	want := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"btcusdt@bookTicker",
		"ethusdt@bookTicker",
	}
	if !reflect.DeepEqual(state.DesiredSubscriptions, want) {
		t.Fatalf("desired subscriptions = %v, want %v", state.DesiredSubscriptions, want)
	}
	if len(state.ActiveSubscriptions) != 0 {
		t.Fatalf("active subscriptions = %v, want none", state.ActiveSubscriptions)
	}
	if state.ConnectionState != ingestion.ConnectionDisconnected {
		t.Fatalf("connection state = %q, want %q", state.ConnectionState, ingestion.ConnectionDisconnected)
	}
}

func TestSpotSubscriptionStateRejectsUnsupportedSymbols(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.Symbols = []string{"BTC-USDT"}
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if _, err := NewSpotWebsocketSupervisor(runtime); err == nil {
		t.Fatal("expected unsupported symbol mapping to fail")
	}
}

func TestSpotWebsocketSupervisorTracksConnectSubscribeReconnectAndRollover(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := supervisor.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := supervisor.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if command == nil {
		t.Fatal("expected subscribe command")
	}
	if command.Method != "SUBSCRIBE" {
		t.Fatalf("method = %q, want %q", command.Method, "SUBSCRIBE")
	}
	if len(command.Params) != 4 {
		t.Fatalf("subscribe params = %v, want 4 entries", command.Params)
	}
	if err := supervisor.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	pong, err := supervisor.HandlePing(base.Add(300 * time.Millisecond))
	if err != nil {
		t.Fatalf("handle ping: %v", err)
	}
	if !pong {
		t.Fatal("expected ping handling to request pong")
	}
	frame, err := supervisor.AcceptDataFrame([]byte(`{"stream":"btcusdt@trade"}`), base.Add(400*time.Millisecond))
	if err != nil {
		t.Fatalf("accept data frame: %v", err)
	}
	if string(frame.Payload) != `{"stream":"btcusdt@trade"}` {
		t.Fatalf("payload = %q", string(frame.Payload))
	}

	state := supervisor.State()
	if state.ConnectionState != ingestion.ConnectionConnected {
		t.Fatalf("connection state = %q, want %q", state.ConnectionState, ingestion.ConnectionConnected)
	}
	if len(state.ActiveSubscriptions) != 4 {
		t.Fatalf("active subscriptions = %v, want 4 entries", state.ActiveSubscriptions)
	}
	if len(state.PendingSubscribeIDs) != 0 {
		t.Fatalf("pending subscribe ids = %v, want none", state.PendingSubscribeIDs)
	}
	if state.LastPongAt.IsZero() {
		t.Fatal("expected last pong time to be recorded")
	}
	if state.SessionRef == "" {
		t.Fatal("expected session ref to be populated")
	}

	plan, err := supervisor.HandleDisconnect(base.Add(time.Second), SpotReconnectCauseTransport)
	if err != nil {
		t.Fatalf("handle disconnect: %v", err)
	}
	if plan.Delay != 500*time.Millisecond {
		t.Fatalf("reconnect delay = %s, want %s", plan.Delay, 500*time.Millisecond)
	}
	if plan.Cause != SpotReconnectCauseTransport {
		t.Fatalf("reconnect cause = %q, want %q", plan.Cause, SpotReconnectCauseTransport)
	}
	state = supervisor.State()
	if state.ConnectionState != ingestion.ConnectionReconnecting {
		t.Fatalf("connection state = %q, want %q", state.ConnectionState, ingestion.ConnectionReconnecting)
	}
	if len(state.ActiveSubscriptions) != 0 {
		t.Fatalf("active subscriptions = %v, want none", state.ActiveSubscriptions)
	}

	reconnectAt := base.Add(time.Second + plan.Delay)
	if err := supervisor.StartConnect(reconnectAt); err != nil {
		t.Fatalf("restart connect: %v", err)
	}
	command, err = supervisor.CompleteConnect(reconnectAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete reconnect: %v", err)
	}
	if err := supervisor.AckSubscribe(reconnectAt.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack reconnect subscribe: %v", err)
	}
	state = supervisor.State()
	if state.ConsecutiveReconnects != 0 {
		t.Fatalf("consecutive reconnects = %d, want 0 after subscribe ack", state.ConsecutiveReconnects)
	}

	rolloverAt := state.NextRolloverAt
	if rolloverAt.IsZero() {
		t.Fatal("expected rollover deadline")
	}
	if !supervisor.ShouldRollover(rolloverAt) {
		t.Fatal("expected rollover to become due at configured deadline")
	}
	plan, err = supervisor.TriggerRollover(rolloverAt)
	if err != nil {
		t.Fatalf("trigger rollover: %v", err)
	}
	if plan.Cause != SpotReconnectCauseRollover {
		t.Fatalf("reconnect cause = %q, want %q", plan.Cause, SpotReconnectCauseRollover)
	}
	if plan.Delay != 500*time.Millisecond {
		t.Fatalf("rollover delay = %s, want %s", plan.Delay, 500*time.Millisecond)
	}
}

func TestSpotWebsocketSupervisorEnforcesReconnectBackoffAndConnectRateLimits(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.ConnectsPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := supervisor.StartConnect(base); err != nil {
		t.Fatalf("start first connect: %v", err)
	}
	if _, err := supervisor.CompleteConnect(base.Add(10 * time.Millisecond)); err != nil {
		t.Fatalf("complete first connect: %v", err)
	}
	plan, err := supervisor.HandleDisconnect(base.Add(20*time.Millisecond), SpotReconnectCauseTransport)
	if err != nil {
		t.Fatalf("disconnect first session: %v", err)
	}
	if err := supervisor.StartConnect(base.Add(200 * time.Millisecond)); err == nil {
		t.Fatal("expected connect before reconnect delay to fail")
	}
	if err := supervisor.StartConnect(plan.RetryAt); err != nil {
		t.Fatalf("start second connect: %v", err)
	}
	if _, err := supervisor.CompleteConnect(plan.RetryAt.Add(10 * time.Millisecond)); err != nil {
		t.Fatalf("complete second connect: %v", err)
	}
	if _, err := supervisor.HandleDisconnect(plan.RetryAt.Add(20*time.Millisecond), SpotReconnectCauseTransport); err != nil {
		t.Fatalf("disconnect second session: %v", err)
	}
	if err := supervisor.StartConnect(base.Add(40 * time.Second)); err == nil {
		t.Fatal("expected connect rate limit to fail inside one-minute window")
	}
	if err := supervisor.StartConnect(base.Add(61 * time.Second)); err != nil {
		t.Fatalf("start third connect after rate-limit window: %v", err)
	}
}

func TestSpotFeedHealthUsesSpotOnlyRuntimeAndRecoversAfterFreshTraffic(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := supervisor.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := supervisor.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := supervisor.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	status, err := supervisor.HealthStatus(base.Add(500 * time.Millisecond))
	if err != nil {
		t.Fatalf("healthy status: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	if status.SnapshotFreshness != ingestion.FreshnessNotApplicable {
		t.Fatalf("snapshot freshness = %q, want %q", status.SnapshotFreshness, ingestion.FreshnessNotApplicable)
	}

	status, err = supervisor.HealthStatus(base.Add(16 * time.Second))
	if err != nil {
		t.Fatalf("stale status: %v", err)
	}
	if status.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthStale)
	}
	if !hasReason(status.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonMessageStale)
	}

	var state SpotWebsocketSupervisorState
	for range 4 {
		if _, err := supervisor.HandleDisconnect(base.Add(16*time.Second), SpotReconnectCauseTransport); err != nil {
			t.Fatalf("handle disconnect: %v", err)
		}
		state = supervisor.State()
		if err := supervisor.StartConnect(state.NextReconnectAt); err != nil {
			t.Fatalf("restart connect: %v", err)
		}
		command, err = supervisor.CompleteConnect(state.NextReconnectAt.Add(50 * time.Millisecond))
		if err != nil {
			t.Fatalf("complete reconnect: %v", err)
		}
	}
	state = supervisor.State()
	status, err = supervisor.HealthStatus(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("reconnect-loop status: %v", err)
	}
	if status.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthDegraded)
	}
	if !hasReason(status.Reasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonReconnectLoop)
	}

	state = supervisor.State()
	if command == nil {
		t.Fatal("expected latest subscribe command")
	}
	if err := supervisor.AckSubscribe(state.LastFrameAt.Add(50*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack latest subscribe: %v", err)
	}
	if _, err := supervisor.AcceptDataFrame([]byte(`{"stream":"ethusdt@bookTicker"}`), state.LastFrameAt.Add(100*time.Millisecond)); err != nil {
		t.Fatalf("accept recovery frame: %v", err)
	}
	state = supervisor.State()
	status, err = supervisor.HealthStatus(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("recovered status: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
}

func TestSpotFeedHealthNormalizationSeamPreservesProvenance(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := supervisor.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := supervisor.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := supervisor.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}

	inputs, err := supervisor.FeedHealthInputs(base.Add(300 * time.Millisecond))
	if err != nil {
		t.Fatalf("feed health inputs: %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("feed health inputs = %d, want 2", len(inputs))
	}

	service, err := normalizer.NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("new normalizer service: %v", err)
	}
	actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: inputs[0].Metadata, Message: inputs[0].Message})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	if actual.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q, want %q", actual.Symbol, "BTC-USD")
	}
	if actual.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", actual.SourceSymbol, "BTCUSDT")
	}
	if actual.FeedHealthState != ingestion.FeedHealthHealthy {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthHealthy)
	}
	if actual.SourceRecordID != spotFeedHealthRecordPrefix+"BTCUSDT" {
		t.Fatalf("source record id = %q, want %q", actual.SourceRecordID, spotFeedHealthRecordPrefix+"BTCUSDT")
	}
}
