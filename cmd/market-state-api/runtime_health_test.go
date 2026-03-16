package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestRuntimeHealthSnapshotReturnsStableNotReadyEntries(t *testing.T) {
	_, owner := newRuntimeHealthOwnerForTest(t)
	now := time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)

	snapshot, err := owner.RuntimeHealthSnapshot(context.Background(), now)
	if err != nil {
		t.Fatalf("runtime health snapshot: %v", err)
	}
	if !snapshot.GeneratedAt.Equal(now) {
		t.Fatalf("generated at = %s, want %s", snapshot.GeneratedAt, now)
	}
	if len(snapshot.Symbols) != 2 {
		t.Fatalf("runtime health symbols = %d, want 2", len(snapshot.Symbols))
	}
	if snapshot.Symbols[0].Symbol != "BTC-USD" || snapshot.Symbols[1].Symbol != "ETH-USD" {
		t.Fatalf("runtime health order = [%s %s], want [BTC-USD ETH-USD]", snapshot.Symbols[0].Symbol, snapshot.Symbols[1].Symbol)
	}
	for _, symbol := range snapshot.Symbols {
		if symbol.Readiness != binanceRuntimeHealthNotReady {
			t.Fatalf("%s readiness = %q, want %q", symbol.Symbol, symbol.Readiness, binanceRuntimeHealthNotReady)
		}
		if symbol.FeedHealth.State != ingestion.FeedHealthDegraded {
			t.Fatalf("%s feed health state = %q, want %q", symbol.Symbol, symbol.FeedHealth.State, ingestion.FeedHealthDegraded)
		}
		if !hasReason(symbol.FeedHealth.Reasons, ingestion.ReasonConnectionNotReady) {
			t.Fatalf("%s reasons = %v, want %q", symbol.Symbol, symbol.FeedHealth.Reasons, ingestion.ReasonConnectionNotReady)
		}
		if !symbol.LastAcceptedExchange.IsZero() || !symbol.LastAcceptedRecv.IsZero() {
			t.Fatalf("%s last accepted timestamps should be zero: %+v", symbol.Symbol, symbol)
		}
	}
}

func TestRuntimeHealthSnapshotTracksHealthyDegradedAndRecoveredStatesDeterministically(t *testing.T) {
	runtimeConfig, owner := newRuntimeHealthOwnerForTest(t)
	btc := runtimeHealthStateForSymbol(t, owner, "BTC-USD")
	eth := runtimeHealthStateForSymbol(t, owner, "ETH-USD")

	seedPublishableRuntimeHealthState(t, btc, 1001, 64000.10, 64000.20, time.UnixMilli(1772798400400).UTC(), time.UnixMilli(1772798400400).UTC())
	seedPublishableRuntimeHealthState(t, eth, 2001, 3200.10, 3200.30, time.UnixMilli(1772798400200).UTC(), time.UnixMilli(1772798400200).UTC())

	now := time.UnixMilli(1772798401200).UTC()
	first, err := owner.RuntimeHealthSnapshot(context.Background(), now)
	if err != nil {
		t.Fatalf("first runtime health snapshot: %v", err)
	}
	second, err := owner.RuntimeHealthSnapshot(context.Background(), now)
	if err != nil {
		t.Fatalf("second runtime health snapshot: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated runtime health snapshots differ:\nfirst=%+v\nsecond=%+v", first, second)
	}
	for _, symbol := range first.Symbols {
		if symbol.Readiness != binanceRuntimeHealthReady {
			t.Fatalf("%s readiness = %q, want %q", symbol.Symbol, symbol.Readiness, binanceRuntimeHealthReady)
		}
		if symbol.FeedHealth.State != ingestion.FeedHealthHealthy {
			t.Fatalf("%s feed health state = %q, want %q", symbol.Symbol, symbol.FeedHealth.State, ingestion.FeedHealthHealthy)
		}
		if symbol.DepthStatus.State != venuebinance.SpotDepthRecoverySynchronized {
			t.Fatalf("%s depth status = %q, want %q", symbol.Symbol, symbol.DepthStatus.State, venuebinance.SpotDepthRecoverySynchronized)
		}
	}

	btc.mu.Lock()
	if err := btc.beginReconnectLocked(); err != nil {
		btc.mu.Unlock()
		t.Fatalf("begin reconnect: %v", err)
	}
	btc.connectionState = ingestion.ConnectionReconnecting
	btc.consecutiveReconnects = runtimeConfig.Adapter.ReconnectLoopThreshold
	btc.mu.Unlock()

	degraded, err := owner.RuntimeHealthSnapshot(context.Background(), now.Add(time.Second))
	if err != nil {
		t.Fatalf("degraded runtime health snapshot: %v", err)
	}
	degradedBTC := degraded.Symbols[0]
	if degradedBTC.Readiness != binanceRuntimeHealthReady {
		t.Fatalf("btc readiness during reconnect = %q, want %q", degradedBTC.Readiness, binanceRuntimeHealthReady)
	}
	if degradedBTC.FeedHealth.State != ingestion.FeedHealthDegraded {
		t.Fatalf("btc degraded feed health state = %q, want %q", degradedBTC.FeedHealth.State, ingestion.FeedHealthDegraded)
	}
	if !hasReason(degradedBTC.FeedHealth.Reasons, ingestion.ReasonConnectionNotReady) || !hasReason(degradedBTC.FeedHealth.Reasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("btc degraded reasons = %v, want %q and %q", degradedBTC.FeedHealth.Reasons, ingestion.ReasonConnectionNotReady, ingestion.ReasonReconnectLoop)
	}
	if degradedBTC.DepthStatus.State != venuebinance.SpotDepthRecoveryIdle {
		t.Fatalf("btc depth status during reconnect = %q, want %q", degradedBTC.DepthStatus.State, venuebinance.SpotDepthRecoveryIdle)
	}

	btc.mu.Lock()
	if err := btc.depth.StartSynchronized(runtimeHealthSync(t, btc.binding.SourceSymbol, 1002, time.UnixMilli(1772798401400).UTC(), 64000.10, 64000.20)); err != nil {
		btc.mu.Unlock()
		t.Fatalf("restart synchronized depth: %v", err)
	}
	btc.connectionState = ingestion.ConnectionConnected
	btc.consecutiveReconnects = 0
	btc.markDepthSynchronized()
	btc.mu.Unlock()

	recovered, err := owner.RuntimeHealthSnapshot(context.Background(), now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("recovered runtime health snapshot: %v", err)
	}
	recoveredBTC := recovered.Symbols[0]
	if recoveredBTC.FeedHealth.State != ingestion.FeedHealthHealthy {
		t.Fatalf("btc recovered feed health state = %q, want %q", recoveredBTC.FeedHealth.State, ingestion.FeedHealthHealthy)
	}
	if recoveredBTC.DepthStatus.State != venuebinance.SpotDepthRecoverySynchronized {
		t.Fatalf("btc recovered depth status = %q, want %q", recoveredBTC.DepthStatus.State, venuebinance.SpotDepthRecoverySynchronized)
	}
}

func TestRuntimeHealthSnapshotMarksStaleWithoutDroppingReadiness(t *testing.T) {
	runtimeConfig, owner := newRuntimeHealthOwnerForTest(t)
	btc := runtimeHealthStateForSymbol(t, owner, "BTC-USD")
	recvTs := time.UnixMilli(1772798400400).UTC()
	seedPublishableRuntimeHealthState(t, btc, 1001, 64000.10, 64000.20, recvTs, recvTs)

	staleAfter := runtimeConfig.Adapter.MessageStaleAfter
	if runtimeConfig.Adapter.SnapshotStaleAfter > staleAfter {
		staleAfter = runtimeConfig.Adapter.SnapshotStaleAfter
	}
	entry, err := btc.runtimeHealthSnapshot(recvTs.Add(staleAfter + time.Millisecond))
	if err != nil {
		t.Fatalf("runtime health entry: %v", err)
	}
	if entry.Readiness != binanceRuntimeHealthReady {
		t.Fatalf("readiness = %q, want %q", entry.Readiness, binanceRuntimeHealthReady)
	}
	if entry.FeedHealth.State != ingestion.FeedHealthStale {
		t.Fatalf("feed health state = %q, want %q", entry.FeedHealth.State, ingestion.FeedHealthStale)
	}
	if !hasReason(entry.FeedHealth.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", entry.FeedHealth.Reasons, ingestion.ReasonMessageStale)
	}
	if !hasReason(entry.FeedHealth.Reasons, ingestion.ReasonSnapshotStale) {
		t.Fatalf("reasons = %v, want %q", entry.FeedHealth.Reasons, ingestion.ReasonSnapshotStale)
	}
}

func TestRuntimeHealthSnapshotPreservesRateLimitReasonFromDepthRecovery(t *testing.T) {
	runtimeConfig, _ := testBinanceRuntime(t)
	runtimeConfig.SnapshotRecoveryPerMinuteLimit = 1
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	state, err := newSpotRuntimeState(spotBinding{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT"}, runtime, runtimeHealthStubFetcher{})
	if err != nil {
		t.Fatalf("new spot runtime state: %v", err)
	}
	seedPublishableRuntimeHealthState(t, state, 700, 64020.10, 64020.30, time.UnixMilli(1772798466100).UTC(), time.UnixMilli(1772798466100).UTC())

	state.mu.Lock()
	err = state.depth.MarkSequenceGap(venuebinance.SpotRawFrame{
		RecvTime: time.UnixMilli(1772798521000).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798521000,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.40","0.95"]],"a":[["64020.70","0.80"]]}`),
	})
	state.mu.Unlock()
	if err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}

	entry, err := state.runtimeHealthSnapshot(time.UnixMilli(1772798521000).UTC())
	if err != nil {
		t.Fatalf("runtime health entry: %v", err)
	}
	if entry.DepthStatus.State != venuebinance.SpotDepthRecoveryRateLimitBlocked {
		t.Fatalf("depth status = %q, want %q", entry.DepthStatus.State, venuebinance.SpotDepthRecoveryRateLimitBlocked)
	}
	if !hasReason(entry.FeedHealth.Reasons, ingestion.ReasonRateLimit) {
		t.Fatalf("reasons = %v, want %q", entry.FeedHealth.Reasons, ingestion.ReasonRateLimit)
	}
}

func TestRuntimeHealthSnapshotConcurrentReadsKeepStableSymbolSet(t *testing.T) {
	_, owner := newRuntimeHealthOwnerForTest(t)
	btc := runtimeHealthStateForSymbol(t, owner, "BTC-USD")
	eth := runtimeHealthStateForSymbol(t, owner, "ETH-USD")
	seedPublishableRuntimeHealthState(t, btc, 1001, 64000.10, 64000.20, time.UnixMilli(1772798400400).UTC(), time.UnixMilli(1772798400400).UTC())
	seedPublishableRuntimeHealthState(t, eth, 2001, 3200.10, 3200.30, time.UnixMilli(1772798400200).UTC(), time.UnixMilli(1772798400200).UTC())

	var readers sync.WaitGroup
	errCh := make(chan error, 1)
	for i := 0; i < 8; i++ {
		readers.Add(1)
		go func(offset time.Duration) {
			defer readers.Done()
			for j := 0; j < 50; j++ {
				snapshot, err := owner.RuntimeHealthSnapshot(context.Background(), time.Date(2026, time.March, 15, 12, 10, 0, 0, time.UTC).Add(offset).Add(time.Duration(j)*time.Millisecond))
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				if len(snapshot.Symbols) != 2 || snapshot.Symbols[0].Symbol != "BTC-USD" || snapshot.Symbols[1].Symbol != "ETH-USD" {
					select {
					case errCh <- fmt.Errorf("unexpected snapshot ordering: %+v", snapshot.Symbols):
					default:
					}
					return
				}
			}
		}(time.Duration(i) * time.Second)
	}

	for i := 0; i < 100; i++ {
		btc.mu.Lock()
		if i%2 == 0 {
			btc.connectionState = ingestion.ConnectionConnected
			btc.consecutiveReconnects = 0
		} else {
			btc.connectionState = ingestion.ConnectionReconnecting
			btc.consecutiveReconnects = 1
		}
		btc.localClockOffset = time.Duration(i) * time.Millisecond
		btc.mu.Unlock()
	}

	readers.Wait()
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

type runtimeHealthStubFetcher struct {
	response venuebinance.SpotDepthSnapshotResponse
	err      error
}

func (s runtimeHealthStubFetcher) FetchSpotDepthSnapshot(context.Context, string) (venuebinance.SpotDepthSnapshotResponse, error) {
	if s.err != nil {
		return venuebinance.SpotDepthSnapshotResponse{}, s.err
	}
	return s.response, nil
}

func newRuntimeHealthOwnerForTest(t *testing.T) (ingestion.VenueRuntimeConfig, *binanceSpotRuntimeOwner) {
	t.Helper()
	runtimeConfig, runtime := testBinanceRuntime(t)
	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       http.DefaultClient,
		baseURL:      "http://example.com",
		websocketURL: "ws://example.com",
		now:          func() time.Time { return time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("new runtime health owner: %v", err)
	}
	owner.started = true
	return runtimeConfig, owner
}

func runtimeHealthStateForSymbol(t *testing.T, owner *binanceSpotRuntimeOwner, symbol string) *spotRuntimeState {
	t.Helper()
	state := owner.states[symbol]
	if state == nil {
		t.Fatalf("missing runtime state for %s", symbol)
	}
	return state
}

func seedPublishableRuntimeHealthState(t *testing.T, state *spotRuntimeState, lastUpdateID int64, bid, ask float64, exchangeTs, recvTs time.Time) {
	t.Helper()
	state.mu.Lock()
	defer state.mu.Unlock()
	if err := state.depth.StartSynchronized(runtimeHealthSync(t, state.binding.SourceSymbol, lastUpdateID, recvTs, bid, ask)); err != nil {
		t.Fatalf("start synchronized depth: %v", err)
	}
	state.lastObservation = marketstateapi.SpotCurrentStateObservation{
		Symbol:        state.binding.Symbol,
		SourceSymbol:  state.binding.SourceSymbol,
		QuoteCurrency: state.binding.QuoteCurrency,
		BestBidPrice:  bid,
		BestAskPrice:  ask,
		ExchangeTs:    exchangeTs,
		RecvTs:        recvTs,
	}
	state.haveObservation = true
	state.publishable = true
	state.awaitingDepthSync = false
	state.havePendingObservation = false
	state.lastAcceptedExchange = exchangeTs.UTC()
	state.lastAcceptedRecv = recvTs.UTC()
	state.connectionState = ingestion.ConnectionConnected
	state.localClockOffset = 0
	state.consecutiveReconnects = 0
}

func runtimeHealthSync(t *testing.T, sourceSymbol string, lastUpdateID int64, recvTs time.Time, bid, ask float64) venuebinance.SpotDepthBootstrapSync {
	t.Helper()
	snapshot, err := venuebinance.ParseOrderBookSnapshot([]byte(depthSnapshotPayload(lastUpdateID, sourceSymbol, fmt.Sprintf("%.2f", bid), fmt.Sprintf("%.2f", ask))), recvTs)
	if err != nil {
		t.Fatalf("parse order-book snapshot: %v", err)
	}
	return venuebinance.SpotDepthBootstrapSync{
		SourceSymbol: sourceSymbol,
		Snapshot:     snapshot,
	}
}

func hasReason(reasons []ingestion.DegradationReason, want ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}
