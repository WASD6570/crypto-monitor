package venuebybit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type orderBookFixture struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

func TestParseOrderBookEventFeedsCanonicalNormalizationHappyPath(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/bybit/BTC-USD/happy-order-book-usdt.fixture.v1.json")
	sequencer := &ingestion.OrderBookSequencer{}
	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueBybit,
		MarketType:    "spot",
	}

	first, err := ParseOrderBookEvent(fixture.RawMessages[0], mustFixtureRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	second, err := ParseOrderBookEvent(fixture.RawMessages[1], mustFixtureRecvTime(t, fixture.RawMessages[1]))
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}

	normalizedSnapshot, err := ingestion.NormalizeOrderBookMessage(metadata, first.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	normalizedDelta, err := ingestion.NormalizeOrderBookMessage(metadata, second.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize delta: %v", err)
	}

	if normalizedSnapshot.OrderBookEvent == nil || normalizedDelta.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events for snapshot and delta")
	}
	if normalizedSnapshot.FeedHealthEvent != nil || normalizedDelta.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health events for happy order-book flow")
	}
	assertCanonicalOrderBookMatchesFixture(t, *normalizedSnapshot.OrderBookEvent, fixture.ExpectedCanonical[0])
	assertCanonicalOrderBookMatchesFixture(t, *normalizedDelta.OrderBookEvent, fixture.ExpectedCanonical[1])
}

func TestParseOrderBookEventFeedsDeterministicGapDegradation(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/bybit/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json")
	sequencer := &ingestion.OrderBookSequencer{}
	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueBybit,
		MarketType:    "spot",
	}

	first, err := ParseOrderBookEvent(fixture.RawMessages[0], mustFixtureRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	second, err := ParseOrderBookEvent(fixture.RawMessages[1], mustFixtureRecvTime(t, fixture.RawMessages[1]))
	if err != nil {
		t.Fatalf("parse gap delta: %v", err)
	}

	normalizedSnapshot, err := ingestion.NormalizeOrderBookMessage(metadata, first.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	normalizedGap, err := ingestion.NormalizeOrderBookMessage(metadata, second.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize gap delta: %v", err)
	}

	if normalizedSnapshot.OrderBookEvent == nil {
		t.Fatal("expected snapshot event before gap")
	}
	if normalizedGap.OrderBookEvent != nil {
		t.Fatal("did not expect order-book event after sequence gap")
	}
	if normalizedGap.FeedHealthEvent == nil {
		t.Fatal("expected feed-health event after sequence gap")
	}
	assertCanonicalFeedHealthMatchesFixture(t, *normalizedGap.FeedHealthEvent, fixture.ExpectedCanonical[0])
	if normalizedGap.SequenceResult.Action != ingestion.SequenceRequiresResync {
		t.Fatalf("sequence action = %q, want %q", normalizedGap.SequenceResult.Action, ingestion.SequenceRequiresResync)
	}
}

func loadOrderBookFixture(t *testing.T, relativePath string) orderBookFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}

	var fixture orderBookFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	return fixture
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

func assertCanonicalFeedHealthMatchesFixture(t *testing.T, actual ingestion.CanonicalFeedHealthEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalFeedHealthEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected feed-health event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical feed-health mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}
