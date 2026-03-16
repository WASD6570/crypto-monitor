package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	"github.com/gorilla/websocket"
)

var testHTTPClient = &http.Client{Timeout: time.Second}

func TestBinanceSpotRuntimeOwnerStartupReturnsNoObservationsUntilPublishable(t *testing.T) {
	runtimeConfig, runtime := testBinanceRuntime(t)
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new runtime owner: %v", err)
	}
	if err := owner.Start(context.Background()); err != nil {
		t.Fatalf("start runtime owner: %v", err)
	}
	defer func() {
		if err := owner.Stop(context.Background()); err != nil {
			t.Fatalf("stop runtime owner: %v", err)
		}
	}()

	snapshot, err := owner.Snapshot(context.Background(), time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Observations) != 0 {
		t.Fatalf("observations = %d, want 0", len(snapshot.Observations))
	}
	if requests.Load() != 0 {
		t.Fatalf("depth snapshot requests = %d, want 0", requests.Load())
	}
}

func TestBinanceSpotRuntimeOwnerPublishesDeterministicallyAndPreservesDegradedState(t *testing.T) {
	runtimeConfig, runtime := testBinanceRuntime(t)
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Query().Get("symbol") {
		case "BTCUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
		case "ETHUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(2001, "ETHUSDT", "3200.10", "3200.30")))
		default:
			http.Error(w, "unexpected symbol", http.StatusBadRequest)
		}
	}))
	defer restServer.Close()
	disconnect := make(chan struct{})
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		writeTextFrame(t, conn, `{"u":2002,"E":1772798400200,"s":"ETHUSDT","b":"3200.10","B":"1.0","a":"3200.30","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400200,"s":"ETHUSDT","U":2002,"u":2002,"b":[["3200.10","1.0"]],"a":[["3200.30","1.0"]]}`)
		writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
		<-disconnect
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
	})

	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new runtime owner: %v", err)
	}
	if err := owner.Start(context.Background()); err != nil {
		t.Fatalf("start runtime owner: %v", err)
	}
	defer func() {
		if err := owner.Stop(context.Background()); err != nil {
			t.Fatalf("stop runtime owner: %v", err)
		}
	}()

	var snapshot marketstateapi.SpotCurrentStateSnapshot
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		current, err := owner.Snapshot(context.Background(), time.Now().UTC())
		if err == nil && len(current.Observations) == 2 {
			snapshot = current
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(snapshot.Observations) != 2 {
		latest, err := owner.Snapshot(context.Background(), time.Now().UTC())
		t.Fatalf("expected two observations, got %d, requests=%d, err=%v, latest=%+v, supervisor=%+v, lastErr=%v", len(latest.Observations), requests.Load(), err, latest, owner.supervisor.State(), owner.lastError())
	}
	if snapshot.Observations[0].Symbol != "BTC-USD" || snapshot.Observations[1].Symbol != "ETH-USD" {
		t.Fatalf("observation order = [%s %s], want [BTC-USD ETH-USD]", snapshot.Observations[0].Symbol, snapshot.Observations[1].Symbol)
	}
	if requests.Load() != 2 {
		t.Fatalf("depth snapshot requests = %d, want 2", requests.Load())
	}

	close(disconnect)
	wsServer.CloseClientConnections()
	wsServer.Close()

	var degraded marketstateapi.SpotCurrentStateSnapshot
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		current, err := owner.Snapshot(context.Background(), time.Now().UTC())
		if err == nil && len(current.Observations) == 2 && current.Observations[0].FeedHealth.ConnectionState != ingestion.ConnectionConnected {
			degraded = current
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(degraded.Observations) != 2 {
		latest, err := owner.Snapshot(context.Background(), time.Now().UTC())
		t.Fatalf("expected degraded snapshot, err=%v, latest=%+v, supervisor=%+v, lastErr=%v", err, latest, owner.supervisor.State(), owner.lastError())
	}
	btc := degraded.Observations[0]
	if btc.BestBidPrice != 64000.10 || btc.BestAskPrice != 64000.20 {
		t.Fatalf("btc observation changed during degradation: %+v", btc)
	}
	if btc.FeedHealth.State == "" {
		t.Fatalf("expected degraded feed health state")
	}
	if btc.DepthStatus.State != venuebinance.SpotDepthRecoveryIdle {
		t.Fatalf("expected reconnect path to reset depth posture until fresh sync: %+v", btc.DepthStatus)
	}
}

func TestBinanceSpotRuntimeOwnerPreservesLastObservationUntilReconnectDepthSync(t *testing.T) {
	runtimeConfig, runtime := testBinanceRuntime(t)
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("symbol") {
		case "BTCUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
		default:
			http.Error(w, "unexpected symbol", http.StatusBadRequest)
		}
	}))
	defer restServer.Close()
	firstDisconnect := make(chan struct{})
	secondSession := make(chan struct{}, 1)
	var sessions atomic.Int32
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		session := sessions.Add(1)
		switch session {
		case 1:
			writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
			writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
			<-firstDisconnect
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
		case 2:
			select {
			case secondSession <- struct{}{}:
			default:
			}
			writeTextFrame(t, conn, `{"u":1003,"E":1772798400500,"s":"BTCUSDT","b":"65000.10","B":"1.0","a":"65000.20","A":"1.0"}`)
			blockUntilSocketClosed(conn)
		default:
			blockUntilSocketClosed(conn)
		}
	})
	defer wsServer.Close()

	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new runtime owner: %v", err)
	}
	if err := owner.Start(context.Background()); err != nil {
		t.Fatalf("start runtime owner: %v", err)
	}
	defer func() {
		if err := owner.Stop(context.Background()); err != nil {
			t.Fatalf("stop runtime owner: %v", err)
		}
	}()

	var initial marketstateapi.SpotCurrentStateSnapshot
	waitFor(t, 2*time.Second, func() bool {
		current, err := owner.Snapshot(context.Background(), time.Now().UTC())
		if err != nil || len(current.Observations) != 1 {
			return false
		}
		if current.Observations[0].BestBidPrice != 64000.10 {
			return false
		}
		initial = current
		return true
	})
	close(firstDisconnect)
	select {
	case <-secondSession:
	case <-time.After(2 * time.Second):
		t.Fatal("second websocket session did not start")
	}

	var reconnectSnapshot marketstateapi.SpotCurrentStateSnapshot
	waitFor(t, 2*time.Second, func() bool {
		current, err := owner.Snapshot(context.Background(), time.Now().UTC())
		if err != nil || len(current.Observations) != 1 {
			return false
		}
		if current.Observations[0].BestBidPrice != initial.Observations[0].BestBidPrice {
			return false
		}
		reconnectSnapshot = current
		return current.Observations[0].FeedHealth.ConnectionState == ingestion.ConnectionConnected
	})
	if reconnectSnapshot.Observations[0].BestBidPrice != 64000.10 || reconnectSnapshot.Observations[0].BestAskPrice != 64000.20 {
		t.Fatalf("reconnect snapshot overwrote last accepted observation: %+v", reconnectSnapshot.Observations[0])
	}
}

func TestBinanceSpotRuntimeOwnerRefreshesDueSnapshots(t *testing.T) {
	runtimeConfig, _ := testBinanceRuntime(t)
	runtimeConfig.HeartbeatTimeout = 20 * time.Millisecond
	runtimeConfig.SnapshotRefreshRequired = true
	runtimeConfig.SnapshotRefreshInterval = 20 * time.Millisecond
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new runtime owner: %v", err)
	}
	if err := owner.Start(context.Background()); err != nil {
		t.Fatalf("start runtime owner: %v", err)
	}
	defer func() {
		if err := owner.Stop(context.Background()); err != nil {
			t.Fatalf("stop runtime owner: %v", err)
		}
	}()

	waitFor(t, 2*time.Second, func() bool {
		current, err := owner.Snapshot(context.Background(), time.Now().UTC())
		return err == nil && len(current.Observations) == 1 && requests.Load() >= 1
	})
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if requests.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if requests.Load() < 2 {
		t.Fatalf("expected refresh snapshot requests, got %d, supervisor=%+v, lastErr=%v", requests.Load(), owner.supervisor.State(), owner.lastError())
	}
}

func TestBinanceSpotRuntimeOwnerTriggersRolloverReconnect(t *testing.T) {
	runtimeConfig, _ := testBinanceRuntime(t)
	runtimeConfig.HeartbeatTimeout = 20 * time.Millisecond
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	var sessions atomic.Int32
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		sessions.Add(1)
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	base := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	startedAt := time.Now()
	var nowMu sync.Mutex
	extraOffset := time.Duration(0)
	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
		now: func() time.Time {
			nowMu.Lock()
			defer nowMu.Unlock()
			return base.Add(time.Since(startedAt) + extraOffset)
		},
	})
	if err != nil {
		t.Fatalf("new runtime owner: %v", err)
	}
	if err := owner.Start(context.Background()); err != nil {
		t.Fatalf("start runtime owner: %v", err)
	}
	defer func() {
		if err := owner.Stop(context.Background()); err != nil {
			t.Fatalf("stop runtime owner: %v", err)
		}
	}()

	waitFor(t, 2*time.Second, func() bool { return sessions.Load() >= 1 })
	nowMu.Lock()
	extraOffset = 25 * time.Hour
	nowMu.Unlock()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sessions.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if sessions.Load() < 2 {
		t.Fatalf("expected rollover reconnect, sessions=%d, supervisor=%+v, lastErr=%v", sessions.Load(), owner.supervisor.State(), owner.lastError())
	}
}

func TestBinanceSpotRuntimeOwnerDeterministicRepeatedInput(t *testing.T) {
	run := func() marketstateapi.SpotCurrentStateSnapshot {
		runtimeConfig, runtime := testBinanceRuntime(t)
		runtimeConfig.HeartbeatTimeout = 0
		var requests atomic.Int32
		restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
		}))
		defer restServer.Close()
		wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
			writeSubscribeAck(t, conn, command.ID)
			writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
			writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
			blockUntilSocketClosed(conn)
		})
		defer wsServer.Close()

		base := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
		sequence := []time.Time{
			base,
			base.Add(100 * time.Millisecond),
			base.Add(200 * time.Millisecond),
			base.Add(300 * time.Millisecond),
			base.Add(400 * time.Millisecond),
			base.Add(500 * time.Millisecond),
		}
		var index atomic.Int64
		nowFunc := func() time.Time {
			current := index.Add(1) - 1
			if int(current) >= len(sequence) {
				return sequence[len(sequence)-1]
			}
			return sequence[current]
		}

		owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
			client:       restServer.Client(),
			baseURL:      restServer.URL,
			websocketURL: websocketURLForServer(wsServer),
			now:          nowFunc,
		})
		if err != nil {
			t.Fatalf("new runtime owner: %v", err)
		}
		if err := owner.Start(context.Background()); err != nil {
			t.Fatalf("start runtime owner: %v", err)
		}
		defer func() {
			if err := owner.Stop(context.Background()); err != nil {
				t.Fatalf("stop runtime owner: %v", err)
			}
		}()

		var snapshot marketstateapi.SpotCurrentStateSnapshot
		waitFor(t, 2*time.Second, func() bool {
			current, err := owner.Snapshot(context.Background(), sequence[len(sequence)-1])
			if err != nil || len(current.Observations) != 1 {
				return false
			}
			snapshot = current
			return requests.Load() == 1
		})
		return snapshot
	}

	first := run()
	second := run()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("deterministic owner snapshots differ\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func TestNewProviderWithOptionsUsesRuntimeOwnerStartupState(t *testing.T) {
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	provider, err := newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC) },
		client:           restServer.Client(),
		configPath:       testConfigPath(),
		binanceURL:       restServer.URL,
		websocketURL:     websocketURLForServer(wsServer),
		binanceUSDMURL:   restServer.URL,
		usdmWebsocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	if closer, ok := provider.(interface{ Close(context.Context) error }); ok {
		defer func() {
			if err := closer.Close(context.Background()); err != nil {
				t.Fatalf("close provider: %v", err)
			}
		}()
	}

	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q, want BTC-USD", response.Symbol)
	}
	if response.SchemaVersion == "" {
		t.Fatalf("expected schema version in response")
	}
	if response.Composite.World.CompositePrice != nil {
		t.Fatalf("expected no world composite price during runtime warm-up")
	}
	if !response.Composite.World.Unavailable {
		t.Fatalf("expected world composite to remain unavailable during runtime warm-up: %+v", response.Composite.World)
	}
	if !response.Composite.USA.Unavailable {
		t.Fatalf("expected usa composite to remain unavailable: %+v", response.Composite.USA)
	}
	if requests.Load() > 2 {
		t.Fatalf("request count = %d, want <= 2", requests.Load())
	}

	_, err = provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("second current symbol state: %v", err)
	}
	if requests.Load() > 2 {
		t.Fatalf("request count after second read = %d, want <= 2", requests.Load())
	}
}

func TestNewProviderWithOptionsPassesClockIntoRuntimeOwner(t *testing.T) {
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	fixedNow := time.Date(2026, time.March, 10, 12, 34, 56, 0, time.UTC)
	provider, err := newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return fixedNow },
		client:           restServer.Client(),
		configPath:       testConfigPath(),
		binanceURL:       restServer.URL,
		websocketURL:     websocketURLForServer(wsServer),
		binanceUSDMURL:   restServer.URL,
		usdmWebsocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	wrapped, ok := provider.(*providerWithRuntime)
	if !ok {
		t.Fatalf("provider type = %T, want *providerWithRuntime", provider)
	}
	defer func() {
		if err := wrapped.Close(context.Background()); err != nil {
			t.Fatalf("close provider: %v", err)
		}
	}()
	if got := wrapped.owner.now(); !got.Equal(fixedNow) {
		t.Fatalf("runtime owner clock = %s, want %s", got, fixedNow)
	}
}

func TestNewProviderWithOptionsServesCurrentStateRoutes(t *testing.T) {
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Query().Get("symbol") {
		case "BTCUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
		case "ETHUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(2001, "ETHUSDT", "3200.10", "3200.30")))
		default:
			http.Error(w, "unexpected symbol", http.StatusBadRequest)
		}
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		writeTextFrame(t, conn, `{"u":2002,"E":1772798400200,"s":"ETHUSDT","b":"3200.10","B":"1.0","a":"3200.30","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400200,"s":"ETHUSDT","U":2002,"u":2002,"b":[["3200.10","1.0"]],"a":[["3200.30","1.0"]]}`)
		writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	provider, err := newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return time.Date(2026, time.March, 10, 12, 0, 1, 0, time.UTC) },
		client:           restServer.Client(),
		configPath:       testConfigPath(),
		binanceURL:       restServer.URL,
		websocketURL:     websocketURLForServer(wsServer),
		binanceUSDMURL:   restServer.URL,
		usdmWebsocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	closer, ok := provider.(interface{ Close(context.Context) error })
	if !ok {
		t.Fatalf("provider type = %T, want closer", provider)
	}
	defer func() {
		if err := closer.Close(context.Background()); err != nil {
			t.Fatalf("close provider: %v", err)
		}
	}()

	handler, err := marketstateapi.NewHandler(provider)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	defer server.Close()

	var symbolResponse marketstateapi.SymbolStateResponse
	waitFor(t, 2*time.Second, func() bool {
		current, status, err := tryGetJSON[marketstateapi.SymbolStateResponse](server.URL + "/api/market-state/BTC-USD")
		if err != nil || status != http.StatusOK {
			return false
		}
		symbolResponse = current
		return !current.Composite.World.Unavailable && current.Symbol == "BTC-USD"
	})
	if symbolResponse.SchemaVersion != features.MarketStateCurrentResponseSchema {
		t.Fatalf("schema version = %q", symbolResponse.SchemaVersion)
	}
	if symbolResponse.SlowContext.Asset != "BTC" {
		t.Fatalf("slow context asset = %q", symbolResponse.SlowContext.Asset)
	}

	globalResponse, _ := mustGetJSON[features.MarketStateCurrentGlobalResponse](t, server.URL+"/api/market-state/global", http.StatusOK)
	if globalResponse.SchemaVersion != features.MarketStateCurrentGlobalSchema {
		t.Fatalf("global schema version = %q", globalResponse.SchemaVersion)
	}
	if len(globalResponse.Symbols) != 2 {
		t.Fatalf("global symbol count = %d, want 2", len(globalResponse.Symbols))
	}
	if globalResponse.Symbols[0].Symbol != "BTC-USD" || globalResponse.Symbols[1].Symbol != "ETH-USD" {
		t.Fatalf("global symbols = %+v, want [BTC-USD ETH-USD]", globalResponse.Symbols)
	}

	unsupported, _ := mustGetJSON[apiErrorPayload](t, server.URL+"/api/market-state/SOL-USD", http.StatusNotFound)
	if unsupported.Error != "unsupported symbol: SOL-USD" {
		t.Fatalf("unsupported symbol error = %q, want %q", unsupported.Error, "unsupported symbol: SOL-USD")
	}
	if requests.Load() != 4 {
		t.Fatalf("runtime bootstrap requests = %d, want 4", requests.Load())
	}
	if _, ok := provider.(*providerWithRuntime); !ok {
		t.Fatalf("provider type = %T, want *providerWithRuntime", provider)
	}
}

func TestNewProviderWithOptionsServesRuntimeStatusDuringWarmup(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	provider, err := newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return fixedNow },
		client:           restServer.Client(),
		configPath:       testConfigPath(),
		binanceURL:       restServer.URL,
		websocketURL:     websocketURLForServer(wsServer),
		binanceUSDMURL:   restServer.URL,
		usdmWebsocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	closer, ok := provider.(interface{ Close(context.Context) error })
	if !ok {
		t.Fatalf("provider type = %T, want closer", provider)
	}
	defer func() {
		if err := closer.Close(context.Background()); err != nil {
			t.Fatalf("close provider: %v", err)
		}
	}()

	handler, err := marketstateapi.NewHandler(provider)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	defer server.Close()

	runtimeStatus, _ := mustGetJSON[marketstateapi.RuntimeStatusResponse](t, server.URL+"/api/runtime-status", http.StatusOK)
	if !runtimeStatus.GeneratedAt.Equal(fixedNow) {
		t.Fatalf("generated at = %s, want %s", runtimeStatus.GeneratedAt, fixedNow)
	}
	if len(runtimeStatus.Symbols) != 2 {
		t.Fatalf("runtime status symbols = %d, want 2", len(runtimeStatus.Symbols))
	}
	if runtimeStatus.Symbols[0].Symbol != "BTC-USD" || runtimeStatus.Symbols[1].Symbol != "ETH-USD" {
		t.Fatalf("runtime status order = [%s %s], want [BTC-USD ETH-USD]", runtimeStatus.Symbols[0].Symbol, runtimeStatus.Symbols[1].Symbol)
	}
	for _, symbol := range runtimeStatus.Symbols {
		if symbol.Readiness != marketstateapi.RuntimeStatusNotReady {
			t.Fatalf("%s readiness = %q, want %q", symbol.Symbol, symbol.Readiness, marketstateapi.RuntimeStatusNotReady)
		}
		if symbol.FeedHealth.State == "" {
			t.Fatalf("%s feed health state should stay explicit", symbol.Symbol)
		}
		if symbol.LastAcceptedExchange != nil || symbol.LastAcceptedRecv != nil {
			t.Fatalf("%s accepted timestamps should be nil: %+v", symbol.Symbol, symbol)
		}
	}

	healthPayload, _ := mustGetJSON[map[string]any](t, server.URL+"/healthz", http.StatusOK)
	if len(healthPayload) != 1 || healthPayload["status"] != "ok" {
		t.Fatalf("health payload = %+v, want only status=ok", healthPayload)
	}
	if requests.Load() > 2 {
		t.Fatalf("runtime warmup requests = %d, want <= 2", requests.Load())
	}
}

func TestNewProviderWithOptionsReflectsDegradedAPIState(t *testing.T) {
	var requests atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Query().Get("symbol") {
		case "BTCUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(1001, "BTCUSDT", "64000.10", "64000.20")))
		case "ETHUSDT":
			_, _ = w.Write([]byte(depthSnapshotPayload(2001, "ETHUSDT", "3200.10", "3200.30")))
		default:
			http.Error(w, "unexpected symbol", http.StatusBadRequest)
		}
	}))
	defer restServer.Close()
	disconnect := make(chan struct{})
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		writeTextFrame(t, conn, `{"u":2002,"E":1772798400200,"s":"ETHUSDT","b":"3200.10","B":"1.0","a":"3200.30","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400200,"s":"ETHUSDT","U":2002,"u":2002,"b":[["3200.10","1.0"]],"a":[["3200.30","1.0"]]}`)
		writeTextFrame(t, conn, `{"u":1002,"E":1772798400400,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400400,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
		<-disconnect
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
	})
	provider, err := newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return time.Now().UTC() },
		client:           restServer.Client(),
		configPath:       testConfigPath(),
		binanceURL:       restServer.URL,
		websocketURL:     websocketURLForServer(wsServer),
		binanceUSDMURL:   restServer.URL,
		usdmWebsocketURL: websocketURLForServer(wsServer),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	closer, ok := provider.(interface{ Close(context.Context) error })
	if !ok {
		t.Fatalf("provider type = %T, want closer", provider)
	}
	defer func() {
		if err := closer.Close(context.Background()); err != nil {
			t.Fatalf("close provider: %v", err)
		}
	}()
	handler, err := marketstateapi.NewHandler(provider)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	defer server.Close()

	waitFor(t, 2*time.Second, func() bool {
		current, status, err := tryGetJSON[marketstateapi.SymbolStateResponse](server.URL + "/api/market-state/BTC-USD")
		return err == nil && status == http.StatusOK && !current.Composite.World.Unavailable
	})

	close(disconnect)
	wsServer.CloseClientConnections()
	wsServer.Close()

	var degraded marketstateapi.SymbolStateResponse
	waitFor(t, 2*time.Second, func() bool {
		current, status, err := tryGetJSON[marketstateapi.SymbolStateResponse](server.URL + "/api/market-state/BTC-USD")
		if err != nil || status != http.StatusOK {
			return false
		}
		degraded = current
		return !current.Composite.World.Unavailable && current.Composite.World.Degraded
	})
	if degraded.Composite.Availability == features.CurrentStateAvailabilityUnavailable {
		t.Fatalf("expected degraded symbol to remain readable: %+v", degraded.Composite)
	}
	if requests.Load() != 4 {
		t.Fatalf("runtime bootstrap requests = %d, want 4", requests.Load())
	}
	if degraded.Regime.Symbol.State == "" {
		t.Fatalf("expected symbol regime state during degraded response: %+v", degraded.Regime.Symbol)
	}

	var runtimeStatus marketstateapi.RuntimeStatusResponse
	waitFor(t, 2*time.Second, func() bool {
		current, status, err := tryGetJSON[marketstateapi.RuntimeStatusResponse](server.URL + "/api/runtime-status")
		if err != nil || status != http.StatusOK || len(current.Symbols) != 2 {
			return false
		}
		runtimeStatus = current
		return current.Symbols[0].Symbol == "BTC-USD" && current.Symbols[0].FeedHealth.State == ingestion.FeedHealthDegraded && hasReason(current.Symbols[0].FeedHealth.Reasons, ingestion.ReasonConnectionNotReady)
	})
	if runtimeStatus.Symbols[0].Readiness != marketstateapi.RuntimeStatusReady {
		t.Fatalf("btc readiness during degradation = %q, want %q", runtimeStatus.Symbols[0].Readiness, marketstateapi.RuntimeStatusReady)
	}
	if runtimeStatus.Symbols[0].ConnectionState == ingestion.ConnectionConnected {
		t.Fatalf("btc connection state should reflect degradation: %+v", runtimeStatus.Symbols[0])
	}

	healthPayload, _ := mustGetJSON[map[string]any](t, server.URL+"/healthz", http.StatusOK)
	if len(healthPayload) != 1 || healthPayload["status"] != "ok" {
		t.Fatalf("health payload during degradation = %+v, want only status=ok", healthPayload)
	}
	global, _ := mustGetJSON[features.MarketStateCurrentGlobalResponse](t, server.URL+"/api/market-state/global", http.StatusOK)
	if len(global.Symbols) != 2 {
		t.Fatalf("global symbol count during degradation = %d, want 2", len(global.Symbols))
	}
	if global.Global.State == "" {
		t.Fatalf("expected global state during degraded response: %+v", global.Global)
	}
	for _, symbol := range global.Symbols {
		if symbol.Symbol == "" || symbol.EffectiveState == "" {
			t.Fatalf("expected machine-readable global symbol summary during degradation: %+v", symbol)
		}
	}
}

func TestNewProviderWithOptionsRejectsMissingWebsocketURL(t *testing.T) {
	_, err := newProviderWithOptions(providerOptions{
		clock:      func() time.Time { return time.Now().UTC() },
		client:     http.DefaultClient,
		configPath: testConfigPath(),
		binanceURL: defaultBinanceURL,
	})
	if err == nil {
		t.Fatalf("expected websocket url error")
	}
}

func TestNewProviderWithOptionsLoadsBinanceEnvironmentProfiles(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/openInterest" {
			http.Error(w, "unexpected request", http.StatusInternalServerError)
			return
		}
		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			http.Error(w, "missing symbol", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(fmt.Sprintf(`{"symbol":"%s","openInterest":"10659.509","time":1772798404000}`, symbol)))
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	for _, environment := range []string{"local", "dev", "prod"} {
		environment := environment
		t.Run(environment, func(t *testing.T) {
			provider, err := newProviderWithOptions(providerOptions{
				clock:            func() time.Time { return fixedNow },
				client:           restServer.Client(),
				configPath:       testEnvironmentConfigPath(environment),
				binanceURL:       restServer.URL,
				websocketURL:     websocketURLForServer(wsServer),
				binanceUSDMURL:   restServer.URL,
				usdmWebsocketURL: websocketURLForServer(wsServer),
			})
			if err != nil {
				t.Fatalf("new provider: %v", err)
			}
			wrapped, ok := provider.(*providerWithRuntime)
			if !ok {
				t.Fatalf("provider type = %T, want *providerWithRuntime", provider)
			}
			defer func() {
				if err := wrapped.Close(context.Background()); err != nil {
					t.Fatalf("close provider: %v", err)
				}
			}()

			status, err := wrapped.RuntimeStatus(context.Background())
			if err != nil {
				t.Fatalf("runtime status: %v", err)
			}
			if !status.GeneratedAt.Equal(fixedNow) {
				t.Fatalf("generated at = %s, want %s", status.GeneratedAt, fixedNow)
			}
			if len(status.Symbols) != 2 {
				t.Fatalf("runtime status symbols = %d, want 2", len(status.Symbols))
			}
			if status.Symbols[0].Symbol != "BTC-USD" || status.Symbols[1].Symbol != "ETH-USD" {
				t.Fatalf("runtime status order = [%s %s], want [BTC-USD ETH-USD]", status.Symbols[0].Symbol, status.Symbols[1].Symbol)
			}
			for _, symbol := range status.Symbols {
				if symbol.Readiness != marketstateapi.RuntimeStatusNotReady {
					t.Fatalf("%s readiness = %q, want %q", symbol.Symbol, symbol.Readiness, marketstateapi.RuntimeStatusNotReady)
				}
			}
		})
	}
}

func TestNewProviderWithOptionsRejectsSpotOverridesWithoutUSDMOverrides(t *testing.T) {
	_, err := newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       http.DefaultClient,
		configPath:   testConfigPath(),
		binanceURL:   "https://example.test",
		websocketURL: "wss://example.test/ws",
	})
	if err == nil {
		t.Fatal("expected usdm override error")
	}
	if !strings.Contains(err.Error(), "binance usdm base url is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestNewProviderWithOptionsRejectsMissingConfig(t *testing.T) {
	_, err := newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       http.DefaultClient,
		configPath:   filepath.Join(t.TempDir(), "missing.json"),
		binanceURL:   defaultBinanceURL,
		websocketURL: defaultBinanceWebsocketURL,
	})
	if err == nil {
		t.Fatalf("expected config error")
	}
}

func TestNewProviderWithOptionsRejectsRuntimeStatusSymbolOverrides(t *testing.T) {
	contents, err := os.ReadFile(testConfigPath())
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	rewritten := strings.Replace(string(contents), `"ETH-USD"`, `"SOL-USD"`, 1)
	configPath := filepath.Join(t.TempDir(), "ingestion.v1.json")
	if err := os.WriteFile(configPath, []byte(rewritten), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err = newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       http.DefaultClient,
		configPath:   configPath,
		binanceURL:   defaultBinanceURL,
		websocketURL: defaultBinanceWebsocketURL,
	})
	if err == nil {
		t.Fatal("expected runtime-status symbol validation error")
	}
	if !strings.Contains(err.Error(), "binance runtime symbols must stay") {
		t.Fatalf("error = %v", err)
	}
}

func TestServerAddressUsesDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	if got := serverAddress(); got != ":8080" {
		t.Fatalf("server address = %q, want :8080", got)
	}
}

func TestServerAddressUsesPortEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	if got := serverAddress(); got != ":9090" {
		t.Fatalf("server address = %q, want :9090", got)
	}
}

func testConfigPath() string {
	return testEnvironmentConfigPath("local")
}

func testEnvironmentConfigPath(environment string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "configs", environment, "ingestion.v1.json"))
}

func testBinanceRuntime(t *testing.T) (ingestion.VenueRuntimeConfig, *venuebinance.Runtime) {
	t.Helper()
	envConfig, err := ingestion.LoadEnvironmentConfig(testConfigPath())
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}
	runtimeConfig, err := envConfig.RuntimeConfigFor(ingestion.VenueBinance)
	if err != nil {
		t.Fatalf("load binance runtime config: %v", err)
	}
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	return runtimeConfig, runtime
}

func newSpotWebsocketTestServer(t *testing.T, script func(*websocket.Conn, binanceSubscribeCommand)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()
		var command binanceSubscribeCommand
		if err := conn.ReadJSON(&command); err != nil {
			return
		}
		script(conn, command)
	}))
	t.Cleanup(server.Close)
	return server
}

func websocketURLForServer(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func writeSubscribeAck(t *testing.T, conn *websocket.Conn, id int64) {
	t.Helper()
	if err := conn.WriteJSON(map[string]any{"result": nil, "id": id}); err != nil {
		t.Errorf("write subscribe ack: %v", err)
	}
}

func writeTextFrame(t *testing.T, conn *websocket.Conn, payload string) {
	t.Helper()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		t.Errorf("write text frame: %v", err)
	}
}

func blockUntilSocketClosed(conn *websocket.Conn) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func waitFor(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not satisfied before timeout")
}

func depthSnapshotPayload(lastUpdateID int64, sourceSymbol string, bid string, ask string) string {
	return fmt.Sprintf(`{"lastUpdateId":%d,"symbol":"%s","bids":[["%s","1.0"]],"asks":[["%s","1.0"]]}`,
		lastUpdateID,
		sourceSymbol,
		bid,
		ask,
	)
}

type apiErrorPayload struct {
	Error string `json:"error"`
}

func mustGetJSON[T any](t *testing.T, url string, wantStatus int) (T, int) {
	t.Helper()
	payload, status, err := tryGetJSON[T](url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	if status != wantStatus {
		t.Fatalf("status for %s = %d, want %d", url, status, wantStatus)
	}
	return payload, status
}

func tryGetJSON[T any](url string) (T, int, error) {
	var payload T
	response, err := testHTTPClient.Get(url)
	if err != nil {
		return payload, 0, err
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return payload, response.StatusCode, err
	}
	return payload, response.StatusCode, nil
}
