package venuecoinbase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type topOfBookFixture struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

func TestParseTopOfBookEventFeedsCanonicalNormalizationHappyPath(t *testing.T) {
	fixture := loadTopOfBookFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-top-book-usd.fixture.v1.json")
	recvTime := mustFixtureRecvTime(t, fixture.RawMessages[0])

	parsed, err := ParseTopOfBookEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse top-of-book event: %v", err)
	}

	result, err := ingestion.NormalizeOrderBookMessage(ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueCoinbase,
		MarketType:    "spot",
	}, parsed.Message, &ingestion.OrderBookSequencer{}, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed top-of-book: %v", err)
	}

	if result.SequenceResult.Action != ingestion.SequenceAcceptedTopOfBook {
		t.Fatalf("sequence action = %q, want %q", result.SequenceResult.Action, ingestion.SequenceAcceptedTopOfBook)
	}
	if result.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	if result.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health event for happy top-of-book flow")
	}
	assertCanonicalOrderBookMatchesFixture(t, *result.OrderBookEvent, fixture.ExpectedCanonical[0])
	if !result.OrderBookEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 1, 0, time.UTC)) {
		t.Fatalf("canonical event time = %s, want fixture exchange time", result.OrderBookEvent.CanonicalEventTime)
	}
	if result.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", result.OrderBookEvent.TimestampFallbackReason)
	}
}

func TestParseTopOfBookEventPreservesQuoteVariantMetadata(t *testing.T) {
	fixture := loadTopOfBookFixture(t, "tests/fixtures/events/coinbase/ETH-USD/edge-quote-variant-usdc.fixture.v1.json")
	recvTime := mustFixtureRecvTime(t, fixture.RawMessages[0])

	parsed, err := ParseTopOfBookEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse top-of-book event: %v", err)
	}

	result, err := ingestion.NormalizeOrderBookMessage(ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueCoinbase,
		MarketType:    "spot",
	}, parsed.Message, &ingestion.OrderBookSequencer{}, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed top-of-book: %v", err)
	}

	if result.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	assertCanonicalOrderBookMatchesFixture(t, *result.OrderBookEvent, fixture.ExpectedCanonical[0])
}

func loadTopOfBookFixture(t *testing.T, relativePath string) topOfBookFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}

	var fixture topOfBookFixture
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
