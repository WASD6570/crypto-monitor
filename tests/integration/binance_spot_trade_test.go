package integration

import (
	"testing"
	"time"

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
