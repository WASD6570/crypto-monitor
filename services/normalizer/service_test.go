package normalizer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	venuecoinbase "github.com/crypto-market-copilot/alerts/services/venue-coinbase"
	venuekraken "github.com/crypto-market-copilot/alerts/services/venue-kraken"
)

type fixtureEnvelope struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
}

type fixtureMessage struct {
	RecvTs string `json:"recvTs"`
}

func TestServiceNormalizeTradePreservesCanonicalOutput(t *testing.T) {
	service := newService(t)
	fixture := loadFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])

	parsed, err := venuecoinbase.ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	actual, err := service.NormalizeTrade(TradeInput{
		Metadata: ingestion.TradeMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueCoinbase,
			MarketType:    "spot",
		},
		Message: parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize trade: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
}

func TestServiceNormalizeBinanceTradePreservesCanonicalOutput(t *testing.T) {
	service := newService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-trade-usdt.fixture.v1.json")
	recvTime := time.UnixMilli(1772798400180).UTC()

	parsed, err := venuebinance.ParseTradeEvent([]byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.10","q":"0.015","T":1772798400100,"m":false,"M":true}`), recvTime)
	if err != nil {
		t.Fatalf("parse binance trade event: %v", err)
	}

	actual, err := service.NormalizeTrade(TradeInput{
		Metadata: ingestion.TradeMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message: parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize binance trade: %v", err)
	}
	if actual.EventType != "market-trade" || actual.Symbol != fixture.Symbol || actual.SourceSymbol != "BTCUSDT" {
		t.Fatalf("unexpected binance trade event: %+v", actual)
	}
	if actual.QuoteCurrency != fixture.QuoteCurrency || actual.Venue != ingestion.VenueBinance || actual.MarketType != "spot" {
		t.Fatalf("unexpected binance trade provenance: %+v", actual)
	}
	if actual.ExchangeTs != "2026-03-06T12:00:00.100Z" {
		t.Fatalf("exchangeTs = %q, want %q", actual.ExchangeTs, "2026-03-06T12:00:00.100Z")
	}
	if actual.RecvTs != "2026-03-06T12:00:00.18Z" {
		t.Fatalf("recvTs = %q, want %q", actual.RecvTs, "2026-03-06T12:00:00.18Z")
	}
	if actual.SourceRecordID != "trade:1001" || actual.TimestampStatus != ingestion.TimestampStatusNormal {
		t.Fatalf("unexpected binance trade identity/status: %+v", actual)
	}
}

func TestServiceNormalizeBinanceTradeTimestampDegraded(t *testing.T) {
	service := newService(t)
	recvTime := time.UnixMilli(1772798540030).UTC()

	parsed, err := venuebinance.ParseTradeEvent([]byte(`{"e":"trade","E":1772798540030,"s":"ETHUSDT","t":2010,"p":"3501.20","q":"0.80","T":1,"m":true,"M":true}`), recvTime)
	if err != nil {
		t.Fatalf("parse degraded binance trade event: %v", err)
	}

	actual, err := service.NormalizeTrade(TradeInput{
		Metadata: ingestion.TradeMetadata{
			Symbol:        "ETH-USD",
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: "USDT",
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message: parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize degraded binance trade: %v", err)
	}
	if actual.Symbol != "ETH-USD" || actual.SourceSymbol != "ETHUSDT" || actual.EventType != "market-trade" {
		t.Fatalf("unexpected degraded binance trade: %+v", actual)
	}
	if actual.TimestampStatus != ingestion.TimestampStatusDegraded {
		t.Fatalf("timestamp status = %q, want %q", actual.TimestampStatus, ingestion.TimestampStatusDegraded)
	}
	if actual.TimestampFallbackReason != ingestion.TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, ingestion.TimestampReasonExchangeSkewExceeded)
	}
}

func TestServiceNormalizeOrderBookPreservesCanonicalOutput(t *testing.T) {
	service := newService(t)
	fixture := loadFixture(t, "tests/fixtures/events/kraken/ETH-USD/happy-order-book-snapshot-delta-usd.fixture.v1.json")
	sequencer := &ingestion.OrderBookSequencer{}

	snapshotRaw := []byte(`{"channel":"book","type":"snapshot","pair":"ETH/USD","sequence":500,"bids":[["3499.10","1.20"]],"asks":[["3499.40","0.95"]],"exchangeTs":"2026-03-06T12:00:02.000Z"}`)
	deltaRaw := []byte(`{"channel":"book","type":"delta","pair":"ETH/USD","sequence":501,"bids":[["3499.20","1.00"]],"asks":[["3499.50","0.90"]],"exchangeTs":"2026-03-06T12:00:02.100Z"}`)

	snapshot, err := venuekraken.ParseOrderBookEvent(snapshotRaw, "2026-03-06T12:00:02.040Z")
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	delta, err := venuekraken.ParseOrderBookEvent(deltaRaw, "2026-03-06T12:00:02.130Z")
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}

	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueKraken,
		MarketType:    "spot",
	}

	first, err := service.NormalizeOrderBook(OrderBookInput{Metadata: metadata, Message: snapshot.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := service.NormalizeOrderBook(OrderBookInput{Metadata: metadata, Message: delta.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize delta: %v", err)
	}

	if first.OrderBookEvent == nil || second.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events")
	}
	assertCanonicalOrderBookMatchesFixture(t, *first.OrderBookEvent, fixture.ExpectedCanonical[0])
	assertCanonicalOrderBookMatchesFixture(t, *second.OrderBookEvent, fixture.ExpectedCanonical[1])
}

func TestServiceNormalizeBinanceTopOfBookFallsBackToRecvTime(t *testing.T) {
	service := newService(t)
	recvTime := time.UnixMilli(1772798400180).UTC()

	parsed, err := venuebinance.ParseTopOfBookEvent([]byte(`{"u":9001,"s":"BTCUSDT","b":"64000.10","B":"1.25","a":"64000.20","A":"0.98"}`), recvTime)
	if err != nil {
		t.Fatalf("parse binance top-of-book event: %v", err)
	}

	actual, err := service.NormalizeOrderBook(OrderBookInput{
		Metadata: ingestion.BookMetadata{
			Symbol:        "BTC-USD",
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: "USDT",
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message:   parsed.Message,
		Sequencer: &ingestion.OrderBookSequencer{},
	})
	if err != nil {
		t.Fatalf("normalize binance top-of-book: %v", err)
	}
	if actual.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	if actual.SequenceResult.Action != ingestion.SequenceAcceptedTopOfBook {
		t.Fatalf("sequence action = %q, want %q", actual.SequenceResult.Action, ingestion.SequenceAcceptedTopOfBook)
	}
	if actual.OrderBookEvent.BookAction != string(ingestion.BookUpdateTopOfBook) {
		t.Fatalf("book action = %q, want %q", actual.OrderBookEvent.BookAction, ingestion.BookUpdateTopOfBook)
	}
	if actual.OrderBookEvent.Symbol != "BTC-USD" || actual.OrderBookEvent.SourceSymbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol provenance: %+v", actual.OrderBookEvent)
	}
	if actual.OrderBookEvent.SourceRecordID != "ticker:9001" {
		t.Fatalf("sourceRecordId = %q, want %q", actual.OrderBookEvent.SourceRecordID, "ticker:9001")
	}
	if actual.OrderBookEvent.TimestampStatus != ingestion.TimestampStatusDegraded {
		t.Fatalf("timestamp status = %q, want %q", actual.OrderBookEvent.TimestampStatus, ingestion.TimestampStatusDegraded)
	}
	if actual.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", actual.OrderBookEvent.TimestampFallbackReason, ingestion.TimestampReasonExchangeMissingOrInvalid)
	}
	if actual.OrderBookEvent.RecvTs != "2026-03-06T12:00:00.18Z" {
		t.Fatalf("recvTs = %q, want %q", actual.OrderBookEvent.RecvTs, "2026-03-06T12:00:00.18Z")
	}
}

func TestServiceNormalizeOrderBookAcceptsWindowedBinanceDeltaAfterBootstrap(t *testing.T) {
	service := newService(t)
	sequencer := &ingestion.OrderBookSequencer{}
	metadata := ingestion.BookMetadata{
		Symbol:        "BTC-USD",
		SourceSymbol:  "BTCUSDT",
		QuoteCurrency: "USDT",
		Venue:         ingestion.VenueBinance,
		MarketType:    "spot",
	}

	snapshot, err := service.NormalizeOrderBook(OrderBookInput{
		Metadata: metadata,
		Message: ingestion.OrderBookMessage{
			Type:          "snapshot",
			FirstSequence: 700,
			Sequence:      700,
			BestBidPrice:  "64020.00",
			BestAskPrice:  "64020.50",
			RecvTs:        "2026-03-06T12:02:00.25Z",
		},
		Sequencer: sequencer,
	})
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	delta, err := service.NormalizeOrderBook(OrderBookInput{
		Metadata: metadata,
		Message: ingestion.OrderBookMessage{
			Type:          "delta",
			FirstSequence: 700,
			Sequence:      701,
			BestBidPrice:  "64020.10",
			BestAskPrice:  "64020.40",
			ExchangeTs:    "2026-03-06T12:02:00.2Z",
			RecvTs:        "2026-03-06T12:02:00.3Z",
		},
		Sequencer: sequencer,
	})
	if err != nil {
		t.Fatalf("normalize windowed delta: %v", err)
	}

	if snapshot.OrderBookEvent == nil || delta.OrderBookEvent == nil {
		t.Fatal("expected canonical snapshot and delta events")
	}
	if delta.SequenceResult.Action != ingestion.SequenceAcceptedDelta {
		t.Fatalf("sequence action = %q, want %q", delta.SequenceResult.Action, ingestion.SequenceAcceptedDelta)
	}
	if delta.OrderBookEvent.SourceRecordID != "book:701" {
		t.Fatalf("sourceRecordId = %q, want %q", delta.OrderBookEvent.SourceRecordID, "book:701")
	}
	if delta.OrderBookEvent.Symbol != "BTC-USD" || delta.OrderBookEvent.SourceSymbol != "BTCUSDT" {
		t.Fatalf("unexpected provenance: %+v", delta.OrderBookEvent)
	}
	if delta.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", delta.OrderBookEvent.TimestampFallbackReason)
	}
}

func TestServiceNormalizeFeedHealthPreservesDegradationMetadata(t *testing.T) {
	service := newService(t)
	config := loadKrakenRuntimeConfig(t)
	runtime, err := venuekraken.NewRuntime(config)
	if err != nil {
		t.Fatalf("new kraken runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	status, err := runtime.EvaluateLoopState(venuekraken.AdapterLoopState{
		ConnectionState:       ingestion.ConnectionReconnecting,
		LastMessageAt:         now.Add(-5 * time.Second),
		LastSnapshotAt:        now.Add(-10 * time.Second),
		SequenceGapDetected:   true,
		LocalClockOffset:      250 * time.Millisecond,
		ConsecutiveReconnects: 3,
		ResyncCount:           2,
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}

	actual, err := service.NormalizeFeedHealth(FeedHealthInput{
		Metadata: ingestion.FeedHealthMetadata{
			Symbol:        "ETH-USD",
			SourceSymbol:  "ETH/USD",
			QuoteCurrency: "USD",
			Venue:         ingestion.VenueKraken,
			MarketType:    "spot",
		},
		Message: ingestion.FeedHealthMessage{
			ExchangeTs:     "2026-03-06T12:02:05Z",
			RecvTs:         "2026-03-06T12:02:05.02Z",
			SourceRecordID: "runtime:kraken-loop",
			Status:         status,
		},
	})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}

	if actual.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	for _, reason := range []ingestion.DegradationReason{
		ingestion.ReasonClockDegraded,
		ingestion.ReasonConnectionNotReady,
		ingestion.ReasonReconnectLoop,
		ingestion.ReasonResyncLoop,
		ingestion.ReasonSequenceGap,
	} {
		if !containsReason(actual.DegradationReasons, reason) {
			t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, reason)
		}
	}
	if actual.SourceRecordID != "runtime:kraken-loop" {
		t.Fatalf("source record id = %q, want %q", actual.SourceRecordID, "runtime:kraken-loop")
	}
	if actual.ExchangeTs != "2026-03-06T12:02:05Z" {
		t.Fatalf("exchangeTs = %q, want %q", actual.ExchangeTs, "2026-03-06T12:02:05Z")
	}
}

func TestServiceNormalizeFeedHealthPreservesBinanceDepthRecoveryReasons(t *testing.T) {
	service := newService(t)
	config := loadBinanceRuntimeConfig(t)
	runtime, err := venuebinance.NewRuntime(config)
	if err != nil {
		t.Fatalf("new binance runtime: %v", err)
	}
	owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, stubBinanceDepthRecoverySnapshotFetcher{})
	if err != nil {
		t.Fatalf("new depth recovery owner: %v", err)
	}
	snapshot, err := venuebinance.ParseOrderBookSnapshotWithSourceSymbol([]byte(`{"lastUpdateId":700,"bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`), "BTCUSDT", time.UnixMilli(1772798520000).UTC())
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	delta, err := venuebinance.ParseOrderBookDelta([]byte(`{"e":"depthUpdate","E":1772798520100,"s":"BTCUSDT","U":700,"u":701,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`), time.UnixMilli(1772798520100).UTC())
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}
	if err := owner.StartSynchronized(venuebinance.SpotDepthBootstrapSync{SourceSymbol: "BTCUSDT", Snapshot: snapshot, Deltas: []venuebinance.ParsedOrderBook{delta}}); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	if err := owner.MarkSequenceGap(venuebinance.SpotRawFrame{
		RecvTime: time.UnixMilli(1772798520600).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798520600,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.40","0.95"]],"a":[["64020.70","0.80"]]}`),
	}); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{
		Symbol:          "BTC-USD",
		QuoteCurrency:   "USDT",
		Now:             time.UnixMilli(1772798520600).UTC(),
		ConnectionState: ingestion.ConnectionConnected,
	})
	if err != nil {
		t.Fatalf("feed health input: %v", err)
	}
	actual, err := service.NormalizeFeedHealth(FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	if actual.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	if actual.SourceRecordID != "runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("source record id = %q, want %q", actual.SourceRecordID, "runtime:binance-spot-depth:BTCUSDT")
	}
	if !containsReason(actual.DegradationReasons, ingestion.ReasonConnectionNotReady) {
		t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, ingestion.ReasonConnectionNotReady)
	}
	if !containsReason(actual.DegradationReasons, ingestion.ReasonSequenceGap) {
		t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, ingestion.ReasonSequenceGap)
	}
}

func TestServiceNormalizeFunding(t *testing.T) {
	service := newService(t)
	parsed, err := venuebinance.ParseUSDMMarkPrice([]byte(`{"e":"markPriceUpdate","E":1772798405000,"s":"BTCUSDT","p":"11794.15000000","i":"11784.62659091","r":"0.00038167","T":1772812800000}`), time.UnixMilli(1772798405040).UTC())
	if err != nil {
		t.Fatalf("parse mark price: %v", err)
	}
	actual, err := service.NormalizeFunding(FundingInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: parsed.SourceSymbol, QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.Funding,
	})
	if err != nil {
		t.Fatalf("normalize funding: %v", err)
	}
	if actual.EventType != "funding-rate" || actual.SourceSymbol != "BTCUSDT" || actual.MarketType != "perpetual" {
		t.Fatalf("unexpected funding event: %+v", actual)
	}
}

func TestServiceNormalizeMarkIndex(t *testing.T) {
	service := newService(t)
	parsed, err := venuebinance.ParseUSDMMarkPrice([]byte(`{"e":"markPriceUpdate","E":1772798405000,"s":"ETHUSDT","p":"3500.10","i":"3499.80","r":"0.00010000","T":1772812800000}`), time.UnixMilli(1772798405040).UTC())
	if err != nil {
		t.Fatalf("parse mark price: %v", err)
	}
	actual, err := service.NormalizeMarkIndex(MarkIndexInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: "ETH-USD", SourceSymbol: parsed.SourceSymbol, QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.MarkIndex,
	})
	if err != nil {
		t.Fatalf("normalize mark index: %v", err)
	}
	if actual.EventType != "mark-index" || actual.Symbol != "ETH-USD" || actual.SourceSymbol != "ETHUSDT" {
		t.Fatalf("unexpected mark index event: %+v", actual)
	}
}

func TestServiceNormalizeOpenInterest(t *testing.T) {
	service := newService(t)
	parsed, err := venuebinance.ParseUSDMOpenInterest([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), time.UnixMilli(1772798404060).UTC())
	if err != nil {
		t.Fatalf("parse open interest: %v", err)
	}
	actual, err := service.NormalizeOpenInterest(OpenInterestInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: parsed.SourceSymbol, QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize open interest: %v", err)
	}
	if actual.EventType != "open-interest-snapshot" || actual.OpenInterest != "10659.509" {
		t.Fatalf("unexpected open interest event: %+v", actual)
	}
}

func TestServiceNormalizeLiquidation(t *testing.T) {
	service := newService(t)
	parsed, err := venuebinance.ParseUSDMForceOrder([]byte(`{"e":"forceOrder","E":1772798406000,"o":{"s":"BTCUSDT","S":"SELL","q":"0.014","p":"9910","T":1772798406000}}`), time.UnixMilli(1772798406020).UTC())
	if err != nil {
		t.Fatalf("parse force order: %v", err)
	}
	actual, err := service.NormalizeLiquidation(LiquidationInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: parsed.SourceSymbol, QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize liquidation: %v", err)
	}
	if actual.EventType != "liquidation-print" || actual.Side != "sell" {
		t.Fatalf("unexpected liquidation event: %+v", actual)
	}
}

func TestNormalizerRawWriteBoundary(t *testing.T) {
	writer := ingestion.NewInMemoryRawEventWriter()
	service, err := NewService(
		ingestion.StrictTimestampPolicy(),
		WithRawEventWriter(writer, ingestion.RawWriteOptions{
			NormalizerService: "services/normalizer",
			BuildVersion:      "test-build",
		}),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	fixture := loadFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])
	parsed, err := venuecoinbase.ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	_, err = service.NormalizeTrade(TradeInput{
		Metadata: ingestion.TradeMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueCoinbase,
			MarketType:    "spot",
		},
		Message: parsed.Message,
		Raw: ingestion.RawWriteContext{
			ConnectionRef: "coinbase-ws-1",
			SessionRef:    "session-1",
		},
	})
	if err != nil {
		t.Fatalf("normalize trade with raw writer: %v", err)
	}

	entries := writer.Entries()
	if len(entries) != 1 {
		t.Fatalf("raw entry count = %d, want 1", len(entries))
	}
	if entries[0].NormalizerService != "services/normalizer" {
		t.Fatalf("normalizer service = %q, want %q", entries[0].NormalizerService, "services/normalizer")
	}
	if entries[0].BuildVersion != "test-build" {
		t.Fatalf("build version = %q, want %q", entries[0].BuildVersion, "test-build")
	}
	if entries[0].ConnectionRef != "coinbase-ws-1" || entries[0].SessionRef != "session-1" {
		t.Fatalf("ingest provenance = (%q, %q), want (%q, %q)", entries[0].ConnectionRef, entries[0].SessionRef, "coinbase-ws-1", "session-1")
	}
}

func TestNormalizerRawWriteBoundaryPreservesBinanceTradeDuplicateIdentity(t *testing.T) {
	writer := ingestion.NewInMemoryRawEventWriter()
	service, err := NewService(
		ingestion.StrictTimestampPolicy(),
		WithRawEventWriter(writer, ingestion.RawWriteOptions{
			NormalizerService: "services/normalizer",
			BuildVersion:      "test-build",
		}),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-trade-usdt.fixture.v1.json")
	recvTime := time.UnixMilli(1772798400180).UTC()
	parsed, err := venuebinance.ParseTradeEvent([]byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.10","q":"0.015","T":1772798400100,"m":false,"M":true}`), recvTime)
	if err != nil {
		t.Fatalf("parse binance trade event: %v", err)
	}

	var firstID string
	for i := 0; i < 2; i++ {
		actual, err := service.NormalizeTrade(TradeInput{
			Metadata: ingestion.TradeMetadata{
				Symbol:        fixture.Symbol,
				SourceSymbol:  parsed.SourceSymbol,
				QuoteCurrency: fixture.QuoteCurrency,
				Venue:         ingestion.VenueBinance,
				MarketType:    "spot",
			},
			Message: parsed.Message,
			Raw: ingestion.RawWriteContext{
				ConnectionRef: "binance-spot-ws",
				SessionRef:    "binance-spot-session-1",
			},
		})
		if err != nil {
			t.Fatalf("normalize binance trade %d: %v", i, err)
		}
		if i == 0 {
			firstID = actual.SourceRecordID
			continue
		}
		if actual.SourceRecordID != firstID {
			t.Fatalf("sourceRecordId = %q, want %q", actual.SourceRecordID, firstID)
		}
	}

	entries := writer.Entries()
	if len(entries) != 2 {
		t.Fatalf("raw entry count = %d, want 2", len(entries))
	}
	if entries[0].StreamFamily != string(ingestion.StreamTrades) {
		t.Fatalf("stream family = %q, want %q", entries[0].StreamFamily, ingestion.StreamTrades)
	}
	if entries[1].DuplicateAudit.Duplicate != true || entries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("duplicate audit = %+v, want duplicate occurrence 2", entries[1].DuplicateAudit)
	}
	if entries[0].ConnectionRef != "binance-spot-ws" || entries[0].SessionRef != "binance-spot-session-1" {
		t.Fatalf("ingest provenance = (%q, %q), want (%q, %q)", entries[0].ConnectionRef, entries[0].SessionRef, "binance-spot-ws", "binance-spot-session-1")
	}
}

func TestNormalizerRawWriteBoundaryPreservesBinanceTopOfBookIdentity(t *testing.T) {
	writer := ingestion.NewInMemoryRawEventWriter()
	service, err := NewService(
		ingestion.StrictTimestampPolicy(),
		WithRawEventWriter(writer, ingestion.RawWriteOptions{
			NormalizerService: "services/normalizer",
			BuildVersion:      "test-build",
		}),
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	recvTime := time.UnixMilli(1772798400180).UTC()
	parsed, err := venuebinance.ParseTopOfBookEvent([]byte(`{"u":9001,"s":"BTCUSDT","b":"64000.10","B":"1.25","a":"64000.20","A":"0.98"}`), recvTime)
	if err != nil {
		t.Fatalf("parse binance top-of-book event: %v", err)
	}

	var firstID string
	for i := 0; i < 2; i++ {
		actual, err := service.NormalizeOrderBook(OrderBookInput{
			Metadata: ingestion.BookMetadata{
				Symbol:        "BTC-USD",
				SourceSymbol:  parsed.SourceSymbol,
				QuoteCurrency: "USDT",
				Venue:         ingestion.VenueBinance,
				MarketType:    "spot",
			},
			Message:   parsed.Message,
			Sequencer: &ingestion.OrderBookSequencer{},
			Raw: ingestion.RawWriteContext{
				ConnectionRef: "binance-spot-ws",
				SessionRef:    "binance-spot-session-1",
			},
		})
		if err != nil {
			t.Fatalf("normalize binance top-of-book %d: %v", i, err)
		}
		if actual.OrderBookEvent == nil {
			t.Fatalf("expected top-of-book event for iteration %d", i)
		}
		if i == 0 {
			firstID = actual.OrderBookEvent.SourceRecordID
			continue
		}
		if actual.OrderBookEvent.SourceRecordID != firstID {
			t.Fatalf("sourceRecordId = %q, want %q", actual.OrderBookEvent.SourceRecordID, firstID)
		}
	}

	entries := writer.Entries()
	if len(entries) != 2 {
		t.Fatalf("raw entry count = %d, want 2", len(entries))
	}
	if entries[0].StreamFamily != string(ingestion.StreamTopOfBook) {
		t.Fatalf("stream family = %q, want %q", entries[0].StreamFamily, ingestion.StreamTopOfBook)
	}
	if entries[0].VenueSequence != 9001 {
		t.Fatalf("venue sequence = %d, want %d", entries[0].VenueSequence, 9001)
	}
	if entries[1].DuplicateAudit.Duplicate != true || entries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("duplicate audit = %+v, want duplicate occurrence 2", entries[1].DuplicateAudit)
	}
	if entries[0].ConnectionRef != "binance-spot-ws" || entries[0].SessionRef != "binance-spot-session-1" {
		t.Fatalf("ingest provenance = (%q, %q), want (%q, %q)", entries[0].ConnectionRef, entries[0].SessionRef, "binance-spot-ws", "binance-spot-session-1")
	}
}

func newService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

func loadFixture(t *testing.T, relativePath string) fixtureEnvelope {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}
	var fixture fixtureEnvelope
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	return fixture
}

func mustRecvTime(t *testing.T, raw json.RawMessage) time.Time {
	t.Helper()
	var message fixtureMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("decode fixture raw message: %v", err)
	}
	recvTime, err := time.Parse(time.RFC3339Nano, message.RecvTs)
	if err != nil {
		t.Fatalf("parse recv timestamp: %v", err)
	}
	return recvTime
}

func loadKrakenRuntimeConfig(t *testing.T) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}
	runtimeConfig, err := config.RuntimeConfigFor(ingestion.VenueKraken)
	if err != nil {
		t.Fatalf("load kraken runtime config: %v", err)
	}
	return runtimeConfig
}

func loadBinanceRuntimeConfig(t *testing.T) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}
	runtimeConfig, err := config.RuntimeConfigFor(ingestion.VenueBinance)
	if err != nil {
		t.Fatalf("load binance runtime config: %v", err)
	}
	return runtimeConfig
}

type stubBinanceDepthRecoverySnapshotFetcher struct{}

func (stubBinanceDepthRecoverySnapshotFetcher) FetchSpotDepthSnapshot(context.Context, string) (venuebinance.SpotDepthSnapshotResponse, error) {
	return venuebinance.SpotDepthSnapshotResponse{}, nil
}

func assertCanonicalTradeMatchesFixture(t *testing.T, actual ingestion.CanonicalTradeEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalTradeEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected trade event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical trade mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalOrderBookMatchesFixture(t *testing.T, actual ingestion.CanonicalOrderBookEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalOrderBookEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected order-book event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical order-book mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func containsReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}
