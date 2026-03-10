package integration

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestBinanceSpotSupervisor(t *testing.T) {
	config := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(config)
	if err != nil {
		t.Fatalf("new binance runtime: %v", err)
	}
	supervisor, err := venuebinance.NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}
	service, err := normalizer.NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("new normalizer service: %v", err)
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
	if _, err := supervisor.AcceptDataFrame([]byte(`{"stream":"btcusdt@trade"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept initial frame: %v", err)
	}

	healthy, err := supervisor.FeedHealthInputs(base.Add(400 * time.Millisecond))
	if err != nil {
		t.Fatalf("healthy inputs: %v", err)
	}
	if len(healthy) != 2 {
		t.Fatalf("healthy input count = %d, want 2", len(healthy))
	}
	btcHealthy, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: healthy[0].Metadata, Message: healthy[0].Message})
	if err != nil {
		t.Fatalf("normalize healthy feed health: %v", err)
	}
	if btcHealthy.FeedHealthState != ingestion.FeedHealthHealthy {
		t.Fatalf("healthy state = %q, want %q", btcHealthy.FeedHealthState, ingestion.FeedHealthHealthy)
	}

	staleInputs, err := supervisor.FeedHealthInputs(base.Add(17 * time.Second))
	if err != nil {
		t.Fatalf("stale inputs: %v", err)
	}
	btcStale, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: staleInputs[0].Metadata, Message: staleInputs[0].Message})
	if err != nil {
		t.Fatalf("normalize stale feed health: %v", err)
	}
	if btcStale.FeedHealthState != ingestion.FeedHealthStale {
		t.Fatalf("stale state = %q, want %q", btcStale.FeedHealthState, ingestion.FeedHealthStale)
	}
	if !containsReason(btcStale.DegradationReasons, ingestion.ReasonMessageStale) {
		t.Fatalf("stale reasons = %v, want %q", btcStale.DegradationReasons, ingestion.ReasonMessageStale)
	}

	for range 4 {
		plan, err := supervisor.HandleDisconnect(base.Add(17*time.Second), venuebinance.SpotReconnectCauseTransport)
		if err != nil {
			t.Fatalf("handle disconnect: %v", err)
		}
		if err := supervisor.StartConnect(plan.RetryAt); err != nil {
			t.Fatalf("restart connect: %v", err)
		}
		command, err = supervisor.CompleteConnect(plan.RetryAt.Add(50 * time.Millisecond))
		if err != nil {
			t.Fatalf("complete reconnect: %v", err)
		}
	}
	state := supervisor.State()
	degradedInputs, err := supervisor.FeedHealthInputs(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("degraded inputs: %v", err)
	}
	btcDegraded, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: degradedInputs[0].Metadata, Message: degradedInputs[0].Message})
	if err != nil {
		t.Fatalf("normalize degraded feed health: %v", err)
	}
	if btcDegraded.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("degraded state = %q, want %q", btcDegraded.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	if !containsReason(btcDegraded.DegradationReasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("degraded reasons = %v, want %q", btcDegraded.DegradationReasons, ingestion.ReasonReconnectLoop)
	}

	state = supervisor.State()
	if err := supervisor.AckSubscribe(state.LastFrameAt.Add(50*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack latest subscribe: %v", err)
	}
	if _, err := supervisor.AcceptDataFrame([]byte(`{"stream":"ethusdt@bookTicker"}`), state.LastFrameAt.Add(100*time.Millisecond)); err != nil {
		t.Fatalf("accept recovery frame: %v", err)
	}
	state = supervisor.State()
	recoveredInputs, err := supervisor.FeedHealthInputs(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("recovered inputs: %v", err)
	}
	btcRecovered, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: recoveredInputs[0].Metadata, Message: recoveredInputs[0].Message})
	if err != nil {
		t.Fatalf("normalize recovered feed health: %v", err)
	}
	if btcRecovered.FeedHealthState != ingestion.FeedHealthHealthy {
		t.Fatalf("recovered state = %q, want %q", btcRecovered.FeedHealthState, ingestion.FeedHealthHealthy)
	}

	rolloverState := supervisor.State()
	plan, err := supervisor.TriggerRollover(rolloverState.NextRolloverAt)
	if err != nil {
		t.Fatalf("trigger rollover: %v", err)
	}
	if plan.Cause != venuebinance.SpotReconnectCauseRollover {
		t.Fatalf("rollover cause = %q, want %q", plan.Cause, venuebinance.SpotReconnectCauseRollover)
	}
	if err := supervisor.StartConnect(plan.RetryAt); err != nil {
		t.Fatalf("restart after rollover: %v", err)
	}
	command, err = supervisor.CompleteConnect(plan.RetryAt.Add(50 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete rollover reconnect: %v", err)
	}
	if err := supervisor.AckSubscribe(plan.RetryAt.Add(100*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack rollover subscribe: %v", err)
	}
	postRollover, err := supervisor.FeedHealthInputs(plan.RetryAt.Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("post-rollover inputs: %v", err)
	}
	btcPostRollover, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: postRollover[0].Metadata, Message: postRollover[0].Message})
	if err != nil {
		t.Fatalf("normalize post-rollover feed health: %v", err)
	}
	if btcPostRollover.FeedHealthState != ingestion.FeedHealthHealthy {
		t.Fatalf("post-rollover state = %q, want %q", btcPostRollover.FeedHealthState, ingestion.FeedHealthHealthy)
	}
}
