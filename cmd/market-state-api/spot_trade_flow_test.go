package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	"github.com/gorilla/websocket"
)

func TestBinanceSpotRuntimeOwnerRecordsTradeFlowBuckets(t *testing.T) {
	runtimeConfig, runtime := testBinanceRuntime(t)
	runtimeConfig.HeartbeatTimeout = 0
	base := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
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
		writeTextFrame(t, conn, spotTradeFlowTradeFrame("ETHUSDT", 2001, "3500.00", "2.0", base.Add(350*time.Millisecond), true))
		writeTextFrame(t, conn, spotTradeFlowTradeFrame("BTCUSDT", 1002, "64100.00", "0.25", base.Add(450*time.Millisecond), true))
		writeTextFrame(t, conn, spotTradeFlowTradeFrame("BTCUSDT", 1001, "64000.00", "0.5", base.Add(550*time.Millisecond), false))
		writeTextFrame(t, conn, spotTradeFlowTradeFrame("BTCUSDT", 1001, "64000.00", "0.5", base.Add(550*time.Millisecond), false))
		writeTextFrame(t, conn, `{"u":2002,"E":1772798400600,"s":"ETHUSDT","b":"3200.10","B":"1.0","a":"3200.30","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400600,"s":"ETHUSDT","U":2002,"u":2002,"b":[["3200.10","1.0"]],"a":[["3200.30","1.0"]]}`)
		writeTextFrame(t, conn, `{"u":1002,"E":1772798400700,"s":"BTCUSDT","b":"64000.10","B":"1.0","a":"64000.20","A":"1.0"}`)
		writeTextFrame(t, conn, `{"e":"depthUpdate","E":1772798400700,"s":"BTCUSDT","U":1002,"u":1002,"b":[["64000.10","1.0"]],"a":[["64000.20","1.0"]]}`)
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	now := monotonicTestClock(base, 100*time.Millisecond)
	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
		now:          now,
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
		current, err := owner.Snapshot(context.Background(), base.Add(5*time.Second))
		if err != nil || len(current.Observations) != 2 || len(current.TradeFlow) != 6 {
			return false
		}
		snapshot = current
		return requests.Load() == 2
	})
	if snapshot.TradeFlow[0].Symbol != "BTC-USD" || snapshot.TradeFlow[0].Family != features.BucketFamily30s || snapshot.TradeFlow[3].Symbol != "ETH-USD" {
		t.Fatalf("trade-flow order = %+v, want BTC families before ETH families", snapshot.TradeFlow)
	}
	btc := findRuntimeTradeFlowBucket(snapshot.TradeFlow, "BTC-USD", features.BucketFamily30s, "2026-03-06T12:00:30Z")
	if btc == nil {
		t.Fatal("expected BTC 30s trade-flow bucket")
	}
	if btc.TradeCount != 2 || btc.DuplicateCount != 1 || btc.BuyTradeCount != 1 || btc.SellTradeCount != 1 {
		t.Fatalf("BTC trade-flow counts = %+v", *btc)
	}
	if btc.BuyNotional != 32000 || btc.SellNotional != 16025 || btc.TotalNotional != 48025 || btc.VWAP != 64033.333333 {
		t.Fatalf("BTC trade-flow metrics = %+v", *btc)
	}
	again, err := owner.Snapshot(context.Background(), base.Add(5*time.Second))
	if err != nil {
		t.Fatalf("second snapshot: %v", err)
	}
	if !reflect.DeepEqual(snapshot.TradeFlow, again.TradeFlow) {
		t.Fatalf("trade-flow snapshots differ:\nfirst=%+v\nsecond=%+v", snapshot.TradeFlow, again.TradeFlow)
	}
}

func TestBinanceSpotRuntimeOwnerTradeFlowTimestampDegradedDoesNotPublishCurrentState(t *testing.T) {
	runtimeConfig, runtime := testBinanceRuntime(t)
	runtimeConfig.HeartbeatTimeout = 0
	base := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unexpected depth request", http.StatusInternalServerError)
	}))
	defer restServer.Close()
	wsServer := newSpotWebsocketTestServer(t, func(conn *websocket.Conn, command binanceSubscribeCommand) {
		writeSubscribeAck(t, conn, command.ID)
		writeTextFrame(t, conn, spotTradeFlowTradeFrame("ETHUSDT", 2001, "3500.00", "2.0", base.Add(-10*time.Second), false))
		blockUntilSocketClosed(conn)
	})
	defer wsServer.Close()

	owner, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       restServer.Client(),
		baseURL:      restServer.URL,
		websocketURL: websocketURLForServer(wsServer),
		now:          monotonicTestClock(base, 100*time.Millisecond),
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
		current, err := owner.Snapshot(context.Background(), base.Add(time.Second))
		if err != nil || len(current.TradeFlow) != 3 {
			return false
		}
		snapshot = current
		return true
	})
	if len(snapshot.Observations) != 0 {
		t.Fatalf("observations = %d, want 0 so trade flow cannot publish current state by itself", len(snapshot.Observations))
	}
	bucket := findRuntimeTradeFlowBucket(snapshot.TradeFlow, "ETH-USD", features.BucketFamily30s, "2026-03-06T12:00:30Z")
	if bucket == nil {
		t.Fatalf("trade-flow snapshot = %+v, want ETH 30s bucket", snapshot.TradeFlow)
	}
	if bucket.BucketSource != features.BucketSourceRecvTs || bucket.TimestampFallbackCount != 1 {
		t.Fatalf("timestamp fields = %+v, want recv fallback count", *bucket)
	}
}

func TestBinanceSpotRuntimeOwnerTradeFlowRejectsInvalidNumericFields(t *testing.T) {
	_, runtime := testBinanceRuntime(t)
	state, err := newSpotRuntimeState(spotBinding{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT"}, runtime, spotTradeFlowStubFetcher{})
	if err != nil {
		t.Fatalf("new spot runtime state: %v", err)
	}
	err = state.recordTrade(venuebinance.ParsedTrade{
		SourceSymbol: "BTCUSDT",
		Message: ingestion.TradeMessage{
			Type:       "trade",
			TradeID:    "1001",
			Price:      "not-a-number",
			Size:       "0.5",
			Side:       features.SpotTradeFlowSideBuy,
			ExchangeTs: "2026-03-06T12:00:00.100Z",
			RecvTs:     "2026-03-06T12:00:00.200Z",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "parse trade price") {
		t.Fatalf("error = %v, want parse trade price", err)
	}
}

type spotTradeFlowStubFetcher struct{}

func (spotTradeFlowStubFetcher) FetchSpotDepthSnapshot(context.Context, string) (venuebinance.SpotDepthSnapshotResponse, error) {
	return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("unexpected depth snapshot")
}

func spotTradeFlowTradeFrame(sourceSymbol string, tradeID int64, price string, quantity string, eventTime time.Time, buyerMaker bool) string {
	return fmt.Sprintf(`{"e":"trade","E":%d,"s":"%s","t":%d,"p":"%s","q":"%s","T":%d,"m":%t,"M":true}`,
		eventTime.UnixMilli(), sourceSymbol, tradeID, price, quantity, eventTime.UnixMilli(), buyerMaker)
}

func monotonicTestClock(start time.Time, step time.Duration) func() time.Time {
	var index atomic.Int64
	return func() time.Time {
		current := index.Add(1) - 1
		return start.Add(time.Duration(current) * step)
	}
}

func findRuntimeTradeFlowBucket(buckets []features.SpotTradeFlowBucket, symbol string, family features.BucketFamily, end string) *features.SpotTradeFlowBucket {
	for index := range buckets {
		if buckets[index].Symbol == symbol && buckets[index].Family == family && buckets[index].BucketEnd == end {
			return &buckets[index]
		}
	}
	return nil
}
