package features

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestSpotTradeFlowAggregatesAndSuppressesDuplicates(t *testing.T) {
	processor := newSpotTradeFlowTestProcessor(t)
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	observations := []SpotTradeFlowObservation{
		spotTradeFlowTestObservation("ETH-USD", "trade:2001", SpotTradeFlowSideBuy, 3500, 2, start.Add(5*time.Second)),
		spotTradeFlowTestObservation("BTC-USD", "trade:1002", SpotTradeFlowSideSell, 64100, 0.25, start.Add(20*time.Second)),
		spotTradeFlowTestObservation("BTC-USD", "trade:1001", SpotTradeFlowSideBuy, 64000, 0.5, start.Add(10*time.Second)),
	}
	for _, observation := range observations {
		result, err := processor.Observe(observation)
		if err != nil {
			t.Fatalf("observe %s: %v", observation.SourceRecordID, err)
		}
		if !result.Accepted || result.Duplicate {
			t.Fatalf("result for %s = %+v, want accepted non-duplicate", observation.SourceRecordID, result)
		}
	}
	duplicate, err := processor.Observe(spotTradeFlowTestObservation("BTC-USD", "trade:1001", SpotTradeFlowSideBuy, 64000, 0.5, start.Add(10*time.Second)))
	if err != nil {
		t.Fatalf("observe duplicate: %v", err)
	}
	if duplicate.Accepted || !duplicate.Duplicate {
		t.Fatalf("duplicate result = %+v, want rejected duplicate", duplicate)
	}

	snapshot := processor.Snapshot()
	if len(snapshot) != 6 {
		t.Fatalf("snapshot bucket count = %d, want 6", len(snapshot))
	}
	if snapshot[0].Symbol != "BTC-USD" || snapshot[0].Family != BucketFamily30s || snapshot[3].Symbol != "ETH-USD" || snapshot[3].Family != BucketFamily30s {
		t.Fatalf("snapshot order = %+v, want BTC families before ETH families", snapshot)
	}

	btc30s := findSpotTradeFlowBucket(snapshot, "BTC-USD", BucketFamily30s, "2026-03-06T12:00:30Z")
	if btc30s == nil {
		t.Fatal("expected BTC 30s bucket")
	}
	if btc30s.TradeCount != 2 || btc30s.DuplicateCount != 1 || btc30s.BuyTradeCount != 1 || btc30s.SellTradeCount != 1 {
		t.Fatalf("BTC trade counts = %+v", *btc30s)
	}
	if btc30s.BuyNotional != 32000 || btc30s.SellNotional != 16025 || btc30s.NetAggressorNotional != 15975 || btc30s.TotalNotional != 48025 {
		t.Fatalf("BTC notional fields = %+v", *btc30s)
	}
	if btc30s.VWAP != 64033.333333 || btc30s.FirstPrice != 64000 || btc30s.LastPrice != 64100 || btc30s.PriceChangeBps != 15.625 {
		t.Fatalf("BTC price fields = %+v", *btc30s)
	}
	btc2m := findSpotTradeFlowBucket(snapshot, "BTC-USD", BucketFamily2m, "2026-03-06T12:02:00Z")
	if btc2m == nil || btc2m.TradeCount != btc30s.TradeCount || btc2m.DuplicateCount != btc30s.DuplicateCount {
		t.Fatalf("BTC 2m rollup-like bucket = %+v, want matching counts", btc2m)
	}
}

func TestSpotTradeFlowTimestampFallbackAndFeedHealth(t *testing.T) {
	processor := newSpotTradeFlowTestProcessor(t)
	recv := time.Date(2026, 3, 6, 12, 0, 31, 0, time.UTC)
	observation := spotTradeFlowTestObservation("ETH-USD", "trade:2002", SpotTradeFlowSideSell, 3510, 1.5, recv)
	observation.ExchangeTs = time.Time{}
	observation.RecvTs = recv
	observation.TimestampStatus = ingestion.TimestampStatusDegraded
	observation.FeedHealthState = ingestion.FeedHealthDegraded
	observation.FeedHealthReasons = []ingestion.DegradationReason{ingestion.ReasonSequenceGap, ingestion.ReasonReconnectLoop}

	if _, err := processor.Observe(observation); err != nil {
		t.Fatalf("observe degraded trade: %v", err)
	}
	bucket := findSpotTradeFlowBucket(processor.Snapshot("ETH-USD"), "ETH-USD", BucketFamily30s, "2026-03-06T12:01:00Z")
	if bucket == nil {
		t.Fatal("expected degraded ETH 30s bucket")
	}
	if bucket.BucketSource != BucketSourceRecvTs || bucket.TimestampFallbackCount != 1 {
		t.Fatalf("timestamp fields = %+v, want recv fallback count", *bucket)
	}
	if bucket.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("feed health state = %q", bucket.FeedHealthState)
	}
	wantReasons := []ingestion.DegradationReason{ingestion.ReasonReconnectLoop, ingestion.ReasonSequenceGap}
	if !reflect.DeepEqual(bucket.FeedHealthReasons, wantReasons) {
		t.Fatalf("feed health reasons = %v, want %v", bucket.FeedHealthReasons, wantReasons)
	}
}

func TestSpotTradeFlowRejectsInvalidInput(t *testing.T) {
	processor := newSpotTradeFlowTestProcessor(t)
	valid := spotTradeFlowTestObservation("BTC-USD", "trade:1001", SpotTradeFlowSideBuy, 64000, 0.5, time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC))
	tests := []struct {
		name string
		edit func(*SpotTradeFlowObservation)
		want string
	}{
		{name: "unsupported symbol", edit: func(o *SpotTradeFlowObservation) { o.Symbol = "SOL-USD" }, want: "unsupported spot trade-flow symbol"},
		{name: "source mismatch", edit: func(o *SpotTradeFlowObservation) { o.SourceSymbol = "BTCUSD" }, want: "source symbol"},
		{name: "venue", edit: func(o *SpotTradeFlowObservation) { o.Venue = ingestion.VenueCoinbase }, want: "venue"},
		{name: "market type", edit: func(o *SpotTradeFlowObservation) { o.MarketType = "perpetual" }, want: "market type"},
		{name: "empty id", edit: func(o *SpotTradeFlowObservation) { o.SourceRecordID = "" }, want: "source record id"},
		{name: "bad side", edit: func(o *SpotTradeFlowObservation) { o.Side = "hold" }, want: "unsupported spot trade-flow side"},
		{name: "bad price", edit: func(o *SpotTradeFlowObservation) { o.Price = 0 }, want: "price must be positive"},
		{name: "bad size", edit: func(o *SpotTradeFlowObservation) { o.Size = -1 }, want: "size must be positive"},
		{name: "missing recv", edit: func(o *SpotTradeFlowObservation) { o.RecvTs = time.Time{} }, want: "recv timestamp"},
		{name: "missing normal exchange", edit: func(o *SpotTradeFlowObservation) { o.ExchangeTs = time.Time{} }, want: "exchange timestamp"},
		{name: "bad timestamp status", edit: func(o *SpotTradeFlowObservation) { o.TimestampStatus = "unknown" }, want: "timestamp status"},
		{name: "bad feed health", edit: func(o *SpotTradeFlowObservation) { o.FeedHealthState = "UNKNOWN" }, want: "feed health state"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observation := valid
			test.edit(&observation)
			_, err := processor.Observe(observation)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestSpotTradeFlowRepeatedRunsAreStable(t *testing.T) {
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	observations := []SpotTradeFlowObservation{
		spotTradeFlowTestObservation("BTC-USD", "trade:1002", SpotTradeFlowSideSell, 64100, 0.25, start.Add(20*time.Second)),
		spotTradeFlowTestObservation("BTC-USD", "trade:1001", SpotTradeFlowSideBuy, 64000, 0.5, start.Add(10*time.Second)),
		spotTradeFlowTestObservation("ETH-USD", "trade:2001", SpotTradeFlowSideBuy, 3500, 2, start.Add(5*time.Second)),
	}
	first := spotTradeFlowSnapshotFromObservations(t, observations)
	second := spotTradeFlowSnapshotFromObservations(t, observations)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("snapshots differ across repeated runs:\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func TestSpotTradeFlowRejectsAfterWatermarkLateInput(t *testing.T) {
	processor := newSpotTradeFlowTestProcessor(t)
	observation := spotTradeFlowTestObservation("BTC-USD", "trade:1001", SpotTradeFlowSideBuy, 64000, 0.5, time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC))
	observation.ObservedAt = time.Date(2026, 3, 6, 12, 1, 0, 0, time.UTC)
	result, err := processor.Observe(observation)
	if err != nil {
		t.Fatalf("observe late trade: %v", err)
	}
	if result.Accepted || result.Assignment.LateDisposition != LateEventAfterWatermark {
		t.Fatalf("late result = %+v, want rejected after watermark", result)
	}
	if len(processor.Snapshot()) != 0 {
		t.Fatalf("snapshot after rejected late input = %+v, want empty", processor.Snapshot())
	}
}

func newSpotTradeFlowTestProcessor(t *testing.T) *SpotTradeFlowProcessor {
	t.Helper()
	processor, err := NewSpotTradeFlowProcessor(DefaultSpotTradeFlowConfig())
	if err != nil {
		t.Fatalf("new spot trade-flow processor: %v", err)
	}
	return processor
}

func spotTradeFlowTestObservation(symbol, sourceRecordID, side string, price, size float64, exchangeTs time.Time) SpotTradeFlowObservation {
	sourceSymbol := "BTCUSDT"
	if symbol == "ETH-USD" {
		sourceSymbol = "ETHUSDT"
	}
	return SpotTradeFlowObservation{
		Symbol:          symbol,
		Venue:           ingestion.VenueBinance,
		MarketType:      SpotTradeFlowMarketType,
		SourceSymbol:    sourceSymbol,
		SourceRecordID:  sourceRecordID,
		Side:            side,
		Price:           price,
		Size:            size,
		ExchangeTs:      exchangeTs,
		RecvTs:          exchangeTs.Add(100 * time.Millisecond),
		TimestampStatus: ingestion.TimestampStatusNormal,
		FeedHealthState: ingestion.FeedHealthHealthy,
	}
}

func spotTradeFlowSnapshotFromObservations(t *testing.T, observations []SpotTradeFlowObservation) []SpotTradeFlowBucket {
	t.Helper()
	processor := newSpotTradeFlowTestProcessor(t)
	for _, observation := range observations {
		if _, err := processor.Observe(observation); err != nil {
			t.Fatalf("observe %s: %v", observation.SourceRecordID, err)
		}
	}
	return processor.Snapshot()
}

func findSpotTradeFlowBucket(buckets []SpotTradeFlowBucket, symbol string, family BucketFamily, end string) *SpotTradeFlowBucket {
	for index := range buckets {
		if buckets[index].Symbol == symbol && buckets[index].Family == family && buckets[index].BucketEnd == end {
			return &buckets[index]
		}
	}
	return nil
}
