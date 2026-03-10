package venuebinance

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestUSDMSubscriptionShapeDerivesDesiredSubscriptionsFromConfigOrder(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	usdm, err := NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	state := usdm.State()
	want := []string{
		"btcusdt@markPrice@1s",
		"ethusdt@markPrice@1s",
		"btcusdt@markPrice@1s",
		"ethusdt@markPrice@1s",
		"btcusdt@forceOrder",
		"ethusdt@forceOrder",
	}
	if !reflect.DeepEqual(state.DesiredSubscriptions, want) {
		t.Fatalf("desired subscriptions = %v, want %v", state.DesiredSubscriptions, want)
	}
}

func TestUSDMSubscriptionShapeRejectsUnsupportedSymbols(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.Symbols = []string{"BTC-USDT"}
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	if _, err := NewUSDMRuntime(runtime); err == nil {
		t.Fatal("expected unsupported symbol mapping to fail")
	}
}

func TestUSDMWebsocketRuntimeTracksConnectResubscribeReconnectAndRollover(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if command == nil || command.Method != "SUBSCRIBE" {
		t.Fatalf("subscribe command = %+v", command)
	}
	if len(command.Params) != 6 {
		t.Fatalf("subscribe params = %v, want 6 entries", command.Params)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept mark price frame: %v", err)
	}
	if _, err := usdm.AcceptForceOrderFrame([]byte(`{"stream":"btcusdt@forceOrder"}`), base.Add(400*time.Millisecond)); err != nil {
		t.Fatalf("accept force order frame: %v", err)
	}

	state := usdm.State()
	if state.LastMarkPriceAt.IsZero() {
		t.Fatal("expected last mark price time")
	}
	if len(state.ActiveSubscriptions) != 6 {
		t.Fatalf("active subscriptions = %v, want 6 entries", state.ActiveSubscriptions)
	}

	plan, err := usdm.HandleDisconnect(base.Add(time.Second), USDMReconnectCauseTransport)
	if err != nil {
		t.Fatalf("handle disconnect: %v", err)
	}
	if plan.Delay != 500*time.Millisecond {
		t.Fatalf("reconnect delay = %s, want %s", plan.Delay, 500*time.Millisecond)
	}
	if err := usdm.StartConnect(plan.RetryAt); err != nil {
		t.Fatalf("restart connect: %v", err)
	}
	command, err = usdm.CompleteConnect(plan.RetryAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete reconnect: %v", err)
	}
	if err := usdm.AckSubscribe(plan.RetryAt.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack reconnect subscribe: %v", err)
	}
	state = usdm.State()
	if state.ConsecutiveReconnects != 0 {
		t.Fatalf("consecutive reconnects = %d, want 0", state.ConsecutiveReconnects)
	}

	rolloverAt := state.NextRolloverAt
	if !usdm.ShouldRollover(rolloverAt) {
		t.Fatal("expected rollover to become due")
	}
	rolloverPlan, err := usdm.TriggerRollover(rolloverAt)
	if err != nil {
		t.Fatalf("trigger rollover: %v", err)
	}
	if rolloverPlan.Cause != USDMReconnectCauseRollover {
		t.Fatalf("rollover cause = %q, want %q", rolloverPlan.Cause, USDMReconnectCauseRollover)
	}
}

func TestUSDMWebsocketRuntimeUsesMarkPriceForFreshnessAndNotForceOrderSparsity(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept initial mark price frame: %v", err)
	}
	if _, err := usdm.AcceptForceOrderFrame([]byte(`{"stream":"btcusdt@forceOrder"}`), base.Add(500*time.Millisecond)); err != nil {
		t.Fatalf("accept force order frame: %v", err)
	}
	status, err := usdm.HealthStatus(base.Add(16 * time.Second))
	if err != nil {
		t.Fatalf("stale status: %v", err)
	}
	if status.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthStale)
	}
	if !hasReason(status.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonMessageStale)
	}

	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(16*time.Second)); err != nil {
		t.Fatalf("accept mark price frame: %v", err)
	}
	status, err = usdm.HealthStatus(base.Add(16*time.Second + 500*time.Millisecond))
	if err != nil {
		t.Fatalf("healthy status: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	if status.SnapshotFreshness != ingestion.FreshnessNotApplicable {
		t.Fatalf("snapshot freshness = %q, want %q", status.SnapshotFreshness, ingestion.FreshnessNotApplicable)
	}
}

func TestUSDMWebsocketRuntimeTracksReconnectLoopAndFeedHealthInputs(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept mark price frame: %v", err)
	}

	for range 4 {
		plan, err := usdm.HandleDisconnect(base.Add(time.Second), USDMReconnectCauseTransport)
		if err != nil {
			t.Fatalf("handle disconnect: %v", err)
		}
		if err := usdm.StartConnect(plan.RetryAt); err != nil {
			t.Fatalf("restart connect: %v", err)
		}
		command, err = usdm.CompleteConnect(plan.RetryAt.Add(50 * time.Millisecond))
		if err != nil {
			t.Fatalf("complete reconnect: %v", err)
		}
	}
	state := usdm.State()
	status, err := usdm.HealthStatus(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("degraded status: %v", err)
	}
	if status.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthDegraded)
	}
	if !hasReason(status.Reasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("reasons = %v, want %q", status.Reasons, ingestion.ReasonReconnectLoop)
	}
	inputs, err := usdm.FeedHealthInputs(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("feed health inputs: %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("feed health inputs = %d, want 2", len(inputs))
	}
	if inputs[0].Metadata.MarketType != "perpetual" {
		t.Fatalf("market type = %q, want %q", inputs[0].Metadata.MarketType, "perpetual")
	}
	if inputs[0].Message.SourceRecordID == "" {
		t.Fatal("expected source record id")
	}
}
