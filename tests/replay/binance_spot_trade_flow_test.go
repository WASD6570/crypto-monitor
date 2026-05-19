package replay

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestReplayBinanceSpotTradeFlow(t *testing.T) {
	base := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	inputs := []replaySpotTradeFlowInput{
		{
			Symbol:        "BTC-USD",
			SourceSymbol:  "BTCUSDT",
			QuoteCurrency: "USDT",
			Raw:           []byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.00","q":"0.50","T":1772798400100,"m":false,"M":true}`),
			RecvTs:        base.Add(180 * time.Millisecond),
			Health:        ingestion.FeedHealthHealthy,
		},
		{
			Symbol:        "BTC-USD",
			SourceSymbol:  "BTCUSDT",
			QuoteCurrency: "USDT",
			Raw:           []byte(`{"e":"trade","E":1772798400280,"s":"BTCUSDT","t":1002,"p":"64100.00","q":"0.25","T":1772798400200,"m":true,"M":true}`),
			RecvTs:        base.Add(280 * time.Millisecond),
			Health:        ingestion.FeedHealthHealthy,
		},
		{
			Symbol:        "BTC-USD",
			SourceSymbol:  "BTCUSDT",
			QuoteCurrency: "USDT",
			Raw:           []byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.00","q":"0.50","T":1772798400100,"m":false,"M":true}`),
			RecvTs:        base.Add(180 * time.Millisecond),
			Health:        ingestion.FeedHealthHealthy,
		},
		{
			Symbol:        "ETH-USD",
			SourceSymbol:  "ETHUSDT",
			QuoteCurrency: "USDT",
			Raw:           []byte(`{"e":"trade","E":1772798540030,"s":"ETHUSDT","t":2010,"p":"3501.20","q":"0.80","T":1,"m":true,"M":true}`),
			RecvTs:        base.Add(2*time.Minute + 20*time.Second + 30*time.Millisecond),
			Health:        ingestion.FeedHealthDegraded,
			Reasons:       []ingestion.DegradationReason{ingestion.ReasonReconnectLoop},
		},
	}

	first := replayBinanceSpotTradeFlow(t, inputs)
	second := replayBinanceSpotTradeFlow(t, inputs)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("trade-flow replay output drift:\nfirst=%+v\nsecond=%+v", first, second)
	}
	if len(first) != 6 {
		t.Fatalf("bucket count = %d, want 6", len(first))
	}
	btc := findReplayTradeFlowBucket(first, "BTC-USD", features.BucketFamily30s, "2026-03-06T12:00:30Z")
	if btc == nil {
		t.Fatal("expected BTC 30s bucket")
	}
	if btc.TradeCount != 2 || btc.DuplicateCount != 1 || btc.BuyNotional != 32000 || btc.SellNotional != 16025 || btc.VWAP != 64033.333333 {
		t.Fatalf("BTC bucket = %+v", *btc)
	}
	eth := findReplayTradeFlowBucket(first, "ETH-USD", features.BucketFamily30s, "2026-03-06T12:02:30Z")
	if eth == nil {
		t.Fatal("expected ETH 30s bucket")
	}
	if eth.BucketSource != features.BucketSourceRecvTs || eth.TimestampFallbackCount != 1 || eth.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("ETH degraded bucket = %+v", *eth)
	}
}

type replaySpotTradeFlowInput struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	Raw           []byte
	RecvTs        time.Time
	Health        ingestion.FeedHealthState
	Reasons       []ingestion.DegradationReason
}

func replayBinanceSpotTradeFlow(t *testing.T, inputs []replaySpotTradeFlowInput) []features.SpotTradeFlowBucket {
	t.Helper()
	processor, err := features.NewSpotTradeFlowProcessor(features.DefaultSpotTradeFlowConfig())
	if err != nil {
		t.Fatalf("new trade-flow processor: %v", err)
	}
	for index, input := range inputs {
		parsed, err := venuebinance.ParseTradeEvent(input.Raw, input.RecvTs)
		if err != nil {
			t.Fatalf("parse replay trade %d: %v", index, err)
		}
		canonical, err := ingestion.NormalizeTradeMessage(ingestion.TradeMetadata{
			Symbol:        input.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: input.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    features.SpotTradeFlowMarketType,
		}, parsed.Message, ingestion.StrictTimestampPolicy())
		if err != nil {
			t.Fatalf("normalize replay trade %d: %v", index, err)
		}
		result, err := processor.Observe(mustReplayTradeFlowObservation(t, canonical, parsed.Message, input.Health, input.Reasons))
		if err != nil {
			t.Fatalf("observe replay trade %d: %v", index, err)
		}
		if index == 2 && (!result.Duplicate || result.Accepted) {
			t.Fatalf("duplicate replay trade result = %+v", result)
		}
	}
	return processor.Snapshot()
}

func mustReplayTradeFlowObservation(t *testing.T, canonical ingestion.CanonicalTradeEvent, message ingestion.TradeMessage, health ingestion.FeedHealthState, reasons []ingestion.DegradationReason) features.SpotTradeFlowObservation {
	t.Helper()
	price, err := strconv.ParseFloat(message.Price, 64)
	if err != nil {
		t.Fatalf("parse price: %v", err)
	}
	size, err := strconv.ParseFloat(message.Size, 64)
	if err != nil {
		t.Fatalf("parse size: %v", err)
	}
	recvTs, err := time.Parse(time.RFC3339Nano, canonical.RecvTs)
	if err != nil {
		t.Fatalf("parse recv time: %v", err)
	}
	var exchangeTs time.Time
	if canonical.ExchangeTs != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, canonical.ExchangeTs); err == nil {
			exchangeTs = parsed
		}
	}
	return features.SpotTradeFlowObservation{
		Symbol:            canonical.Symbol,
		Venue:             canonical.Venue,
		MarketType:        canonical.MarketType,
		SourceSymbol:      canonical.SourceSymbol,
		SourceRecordID:    canonical.SourceRecordID,
		Side:              message.Side,
		Price:             price,
		Size:              size,
		ExchangeTs:        exchangeTs,
		RecvTs:            recvTs,
		ObservedAt:        recvTs,
		TimestampStatus:   canonical.TimestampStatus,
		FeedHealthState:   health,
		FeedHealthReasons: append([]ingestion.DegradationReason(nil), reasons...),
	}
}

func findReplayTradeFlowBucket(buckets []features.SpotTradeFlowBucket, symbol string, family features.BucketFamily, end string) *features.SpotTradeFlowBucket {
	for index := range buckets {
		if buckets[index].Symbol == symbol && buckets[index].Family == family && buckets[index].BucketEnd == end {
			return &buckets[index]
		}
	}
	return nil
}
