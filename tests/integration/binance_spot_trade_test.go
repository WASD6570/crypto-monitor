package integration

import (
	"strconv"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceSpotTradeSupervisorHandoffHappyPath(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept trade frame: %v", err)
	}
	parsed, err := venuebinance.ParseTradeFrame(frame)
	if err != nil {
		t.Fatalf("parse trade frame: %v", err)
	}
	actual, err := service.NormalizeTrade(normalizer.TradeInput{
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
		t.Fatalf("normalize trade: %v", err)
	}
	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
}

func TestIngestionBinanceSpotTradeTimestampDegraded(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/edge-native-timestamp-degraded-trade-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept degraded trade frame: %v", err)
	}
	parsed, err := venuebinance.ParseTradeFrame(frame)
	if err != nil {
		t.Fatalf("parse degraded trade frame: %v", err)
	}
	actual, err := service.NormalizeTrade(normalizer.TradeInput{
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
		t.Fatalf("normalize degraded trade: %v", err)
	}
	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	if actual.TimestampFallbackReason != ingestion.TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, ingestion.TimestampReasonExchangeSkewExceeded)
	}
}

func TestIngestionBinanceSpotTradeDuplicateIdentityStable(t *testing.T) {
	writer := ingestion.NewInMemoryRawEventWriter()
	service, err := normalizer.NewService(
		ingestion.StrictTimestampPolicy(),
		normalizer.WithRawEventWriter(writer, ingestion.RawWriteOptions{
			NormalizerService: "services/normalizer",
			BuildVersion:      "test-build",
		}),
	)
	if err != nil {
		t.Fatalf("new normalizer service: %v", err)
	}
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])
	supervisor := connectedSpotSupervisor(t, recvTime.Add(-180*time.Millisecond))

	var sourceRecordID string
	for i := 0; i < 2; i++ {
		frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], recvTime)
		if err != nil {
			t.Fatalf("accept duplicate trade frame %d: %v", i, err)
		}
		parsed, err := venuebinance.ParseTradeFrame(frame)
		if err != nil {
			t.Fatalf("parse duplicate trade frame %d: %v", i, err)
		}
		actual, err := service.NormalizeTrade(normalizer.TradeInput{
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
				SessionRef:    supervisor.State().SessionRef,
			},
		})
		if err != nil {
			t.Fatalf("normalize duplicate trade %d: %v", i, err)
		}
		if i == 0 {
			sourceRecordID = actual.SourceRecordID
			continue
		}
		if actual.SourceRecordID != sourceRecordID {
			t.Fatalf("sourceRecordId = %q, want %q", actual.SourceRecordID, sourceRecordID)
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
}

func TestIngestionBinanceSpotTradeFlowFeatureInputs(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-trade-flow-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))
	processor, err := features.NewSpotTradeFlowProcessor(features.DefaultSpotTradeFlowConfig())
	if err != nil {
		t.Fatalf("new trade-flow processor: %v", err)
	}

	for index, raw := range fixture.RawMessages {
		frame, err := supervisor.AcceptDataFrame(raw, mustRecvTime(t, raw))
		if err != nil {
			t.Fatalf("accept trade-flow frame %d: %v", index, err)
		}
		parsed, err := venuebinance.ParseTradeFrame(frame)
		if err != nil {
			t.Fatalf("parse trade-flow frame %d: %v", index, err)
		}
		metadata := ingestion.TradeMetadata{Symbol: fixture.Symbol, SourceSymbol: parsed.SourceSymbol, QuoteCurrency: fixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: features.SpotTradeFlowMarketType}
		canonical, err := service.NormalizeTrade(normalizer.TradeInput{Metadata: metadata, Message: parsed.Message})
		if err != nil {
			t.Fatalf("normalize trade-flow frame %d: %v", index, err)
		}
		result, err := processor.Observe(mustSpotTradeFlowObservation(t, canonical, parsed.Message, ingestion.FeedHealthHealthy, nil))
		if err != nil {
			t.Fatalf("observe trade-flow frame %d: %v", index, err)
		}
		if index < 2 && !result.Accepted {
			t.Fatalf("trade-flow frame %d rejected: %+v", index, result)
		}
		if index == 2 && (!result.Duplicate || result.Accepted) {
			t.Fatalf("duplicate trade-flow frame result = %+v", result)
		}
	}

	bucket := findIntegrationTradeFlowBucket(processor.Snapshot(fixture.Symbol), fixture.Symbol, features.BucketFamily30s, "2026-03-06T12:00:30Z")
	if bucket == nil {
		t.Fatal("expected BTC 30s trade-flow bucket")
	}
	if bucket.TradeCount != 2 || bucket.DuplicateCount != 1 || bucket.BuyTradeCount != 1 || bucket.SellTradeCount != 1 {
		t.Fatalf("trade-flow counts = %+v", *bucket)
	}
	if bucket.BuyNotional != 32000 || bucket.SellNotional != 16025 || bucket.NetAggressorNotional != 15975 || bucket.VWAP != 64033.333333 {
		t.Fatalf("trade-flow metrics = %+v", *bucket)
	}
}

func TestIngestionBinanceSpotTradeFlowTimestampDegradedFeatureInput(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/edge-native-timestamp-degraded-trade-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))
	processor, err := features.NewSpotTradeFlowProcessor(features.DefaultSpotTradeFlowConfig())
	if err != nil {
		t.Fatalf("new trade-flow processor: %v", err)
	}

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept degraded trade-flow frame: %v", err)
	}
	parsed, err := venuebinance.ParseTradeFrame(frame)
	if err != nil {
		t.Fatalf("parse degraded trade-flow frame: %v", err)
	}
	canonical, err := service.NormalizeTrade(normalizer.TradeInput{
		Metadata: ingestion.TradeMetadata{Symbol: fixture.Symbol, SourceSymbol: parsed.SourceSymbol, QuoteCurrency: fixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: features.SpotTradeFlowMarketType},
		Message:  parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize degraded trade-flow frame: %v", err)
	}
	_, err = processor.Observe(mustSpotTradeFlowObservation(t, canonical, parsed.Message, ingestion.FeedHealthDegraded, []ingestion.DegradationReason{ingestion.ReasonReconnectLoop}))
	if err != nil {
		t.Fatalf("observe degraded trade-flow frame: %v", err)
	}
	bucket := findIntegrationTradeFlowBucket(processor.Snapshot(fixture.Symbol), fixture.Symbol, features.BucketFamily30s, "2026-03-06T12:02:30Z")
	if bucket == nil {
		t.Fatal("expected degraded ETH 30s trade-flow bucket")
	}
	if bucket.BucketSource != features.BucketSourceRecvTs || bucket.TimestampFallbackCount != 1 {
		t.Fatalf("timestamp fields = %+v, want recv fallback count", *bucket)
	}
	if bucket.FeedHealthState != ingestion.FeedHealthDegraded || len(bucket.FeedHealthReasons) != 1 || bucket.FeedHealthReasons[0] != ingestion.ReasonReconnectLoop {
		t.Fatalf("feed health fields = %+v", *bucket)
	}
}

func connectedSpotSupervisor(t *testing.T, base time.Time) *venuebinance.SpotWebsocketSupervisor {
	t.Helper()
	runtimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	supervisor, err := venuebinance.NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}
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
	return supervisor
}

func mustSpotTradeFlowObservation(t *testing.T, canonical ingestion.CanonicalTradeEvent, message ingestion.TradeMessage, feedHealth ingestion.FeedHealthState, reasons []ingestion.DegradationReason) features.SpotTradeFlowObservation {
	t.Helper()
	price, err := strconv.ParseFloat(message.Price, 64)
	if err != nil {
		t.Fatalf("parse trade price: %v", err)
	}
	size, err := strconv.ParseFloat(message.Size, 64)
	if err != nil {
		t.Fatalf("parse trade size: %v", err)
	}
	recvTs := mustParseTime(t, canonical.RecvTs)
	var exchangeTs time.Time
	if canonical.ExchangeTs != "" {
		parsed, err := time.Parse(time.RFC3339Nano, canonical.ExchangeTs)
		if err == nil {
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
		FeedHealthState:   feedHealth,
		FeedHealthReasons: append([]ingestion.DegradationReason(nil), reasons...),
	}
}

func findIntegrationTradeFlowBucket(buckets []features.SpotTradeFlowBucket, symbol string, family features.BucketFamily, end string) *features.SpotTradeFlowBucket {
	for index := range buckets {
		if buckets[index].Symbol == symbol && buckets[index].Family == family && buckets[index].BucketEnd == end {
			return &buckets[index]
		}
	}
	return nil
}
