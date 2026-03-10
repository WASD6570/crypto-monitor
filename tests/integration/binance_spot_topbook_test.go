package integration

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceSpotTopOfBookSupervisorHandoffHappyPath(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-top-book-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept top-of-book frame: %v", err)
	}
	parsed, err := venuebinance.ParseTopOfBookFrame(frame)
	if err != nil {
		t.Fatalf("parse top-of-book frame: %v", err)
	}
	actual, err := service.NormalizeOrderBook(normalizer.OrderBookInput{
		Metadata: ingestion.BookMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message:   parsed.Message,
		Sequencer: &ingestion.OrderBookSequencer{},
	})
	if err != nil {
		t.Fatalf("normalize top-of-book: %v", err)
	}
	if actual.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	assertCanonicalOrderBookMatchesFixture(t, *actual.OrderBookEvent, fixture.ExpectedCanonical[0])
	if actual.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", actual.OrderBookEvent.TimestampFallbackReason, ingestion.TimestampReasonExchangeMissingOrInvalid)
	}
}

func TestIngestionBinanceSpotTopOfBookTimestampDegraded(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/edge-native-timestamp-degraded-top-book-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept degraded top-of-book frame: %v", err)
	}
	parsed, err := venuebinance.ParseTopOfBookFrame(frame)
	if err != nil {
		t.Fatalf("parse degraded top-of-book frame: %v", err)
	}
	actual, err := service.NormalizeOrderBook(normalizer.OrderBookInput{
		Metadata: ingestion.BookMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		},
		Message:   parsed.Message,
		Sequencer: &ingestion.OrderBookSequencer{},
	})
	if err != nil {
		t.Fatalf("normalize degraded top-of-book: %v", err)
	}
	if actual.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	assertCanonicalOrderBookMatchesFixture(t, *actual.OrderBookEvent, fixture.ExpectedCanonical[0])
	if actual.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.OrderBookEvent.TimestampFallbackReason, ingestion.TimestampReasonExchangeSkewExceeded)
	}
}

func TestIngestionBinanceSpotTopOfBookDuplicateIdentityStable(t *testing.T) {
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
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-top-book-usdt.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])
	supervisor := connectedSpotSupervisor(t, recvTime.Add(-180*time.Millisecond))

	var sourceRecordID string
	for i := 0; i < 2; i++ {
		frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], recvTime)
		if err != nil {
			t.Fatalf("accept duplicate top-of-book frame %d: %v", i, err)
		}
		parsed, err := venuebinance.ParseTopOfBookFrame(frame)
		if err != nil {
			t.Fatalf("parse duplicate top-of-book frame %d: %v", i, err)
		}
		actual, err := service.NormalizeOrderBook(normalizer.OrderBookInput{
			Metadata: ingestion.BookMetadata{
				Symbol:        fixture.Symbol,
				SourceSymbol:  parsed.SourceSymbol,
				QuoteCurrency: fixture.QuoteCurrency,
				Venue:         ingestion.VenueBinance,
				MarketType:    "spot",
			},
			Message:   parsed.Message,
			Sequencer: &ingestion.OrderBookSequencer{},
			Raw: ingestion.RawWriteContext{
				ConnectionRef: "binance-spot-ws",
				SessionRef:    supervisor.State().SessionRef,
			},
		})
		if err != nil {
			t.Fatalf("normalize duplicate top-of-book %d: %v", i, err)
		}
		if actual.OrderBookEvent == nil {
			t.Fatalf("expected top-of-book event for iteration %d", i)
		}
		if i == 0 {
			sourceRecordID = actual.OrderBookEvent.SourceRecordID
			continue
		}
		if actual.OrderBookEvent.SourceRecordID != sourceRecordID {
			t.Fatalf("sourceRecordId = %q, want %q", actual.OrderBookEvent.SourceRecordID, sourceRecordID)
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
}
