package ingestion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type orderBookFixture struct {
	Symbol            string             `json:"symbol"`
	SourceSymbol      string             `json:"sourceSymbol"`
	QuoteCurrency     string             `json:"quoteCurrency"`
	Venue             Venue              `json:"venue"`
	RawMessages       []OrderBookMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage  `json:"expectedCanonical"`
}

func TestNormalizeOrderBookMessageMatchesHappyFixture(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/kraken/ETH-USD/happy-order-book-snapshot-delta-usd.fixture.v1.json")
	sequencer := &OrderBookSequencer{}
	metadata := bookMetadataFromFixture(fixture)

	first, err := NormalizeOrderBookMessage(metadata, fixture.RawMessages[0], sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := NormalizeOrderBookMessage(metadata, fixture.RawMessages[1], sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize delta: %v", err)
	}

	if first.OrderBookEvent == nil || second.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events for snapshot and delta")
	}
	if first.FeedHealthEvent != nil || second.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health event for happy order-book flow")
	}

	assertCanonicalOrderBookMatchesFixture(t, *first.OrderBookEvent, fixture.ExpectedCanonical[0])
	assertCanonicalOrderBookMatchesFixture(t, *second.OrderBookEvent, fixture.ExpectedCanonical[1])
	if !first.OrderBookEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 2, 0, time.UTC)) {
		t.Fatalf("snapshot canonical event time = %s, want fixture exchange time", first.OrderBookEvent.CanonicalEventTime)
	}
	if !second.OrderBookEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 2, 100000000, time.UTC)) {
		t.Fatalf("delta canonical event time = %s, want fixture exchange time", second.OrderBookEvent.CanonicalEventTime)
	}
}

func TestNormalizeOrderBookMessageEmitsDegradedFeedHealthOnGap(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json")
	sequencer := &OrderBookSequencer{}
	metadata := bookMetadataFromFixture(fixture)

	first, err := NormalizeOrderBookMessage(metadata, fixture.RawMessages[0], sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := NormalizeOrderBookMessage(metadata, fixture.RawMessages[1], sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize gap delta: %v", err)
	}

	if first.OrderBookEvent == nil {
		t.Fatal("expected snapshot event before gap")
	}
	if second.OrderBookEvent != nil {
		t.Fatal("did not expect order-book event on gap-driven resync")
	}
	if second.FeedHealthEvent == nil {
		t.Fatal("expected feed-health event on gap-driven resync")
	}
	if second.SequenceResult.Action != SequenceRequiresResync {
		t.Fatalf("sequence action = %q, want %q", second.SequenceResult.Action, SequenceRequiresResync)
	}

	assertCanonicalFeedHealthMatchesFixture(t, *second.FeedHealthEvent, fixture.ExpectedCanonical[0])
	if second.FeedHealthEvent.TimestampFallbackReason != TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", second.FeedHealthEvent.TimestampFallbackReason)
	}
	if !second.FeedHealthEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 1, 0, 100000000, time.UTC)) {
		t.Fatalf("gap canonical event time = %s, want fixture exchange time", second.FeedHealthEvent.CanonicalEventTime)
	}
	if second.FeedHealthEvent.SourceRecordID != "gap:900-902" {
		t.Fatalf("sourceRecordId = %q, want %q", second.FeedHealthEvent.SourceRecordID, "gap:900-902")
	}
}

func TestNormalizeOrderBookMessageUsesSequenceBackedIdentityForTopOfBookWithoutExchangeTs(t *testing.T) {
	metadata := BookMetadata{
		Symbol:        "BTC-USD",
		SourceSymbol:  "BTCUSDT",
		QuoteCurrency: "USDT",
		Venue:         VenueBinance,
		MarketType:    "spot",
	}
	message := OrderBookMessage{
		Type:         string(BookUpdateTopOfBook),
		Sequence:     9001,
		BestBidPrice: "64000.10",
		BestAskPrice: "64000.20",
		RecvTs:       "2026-03-06T12:00:00.18Z",
	}

	result, err := NormalizeOrderBookMessage(metadata, message, &OrderBookSequencer{}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize top-of-book: %v", err)
	}
	if result.SequenceResult.Action != SequenceAcceptedTopOfBook {
		t.Fatalf("sequence action = %q, want %q", result.SequenceResult.Action, SequenceAcceptedTopOfBook)
	}
	if result.OrderBookEvent == nil {
		t.Fatal("expected canonical top-of-book event")
	}
	if result.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health event for accepted top-of-book")
	}
	if result.OrderBookEvent.SourceRecordID != "ticker:9001" {
		t.Fatalf("sourceRecordId = %q, want %q", result.OrderBookEvent.SourceRecordID, "ticker:9001")
	}
	if result.OrderBookEvent.TimestampStatus != TimestampStatusDegraded {
		t.Fatalf("timestamp status = %q, want %q", result.OrderBookEvent.TimestampStatus, TimestampStatusDegraded)
	}
	if result.OrderBookEvent.TimestampFallbackReason != TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", result.OrderBookEvent.TimestampFallbackReason, TimestampReasonExchangeMissingOrInvalid)
	}
	if !result.OrderBookEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 0, 180000000, time.UTC)) {
		t.Fatalf("canonical event time = %s, want recv time fallback", result.OrderBookEvent.CanonicalEventTime)
	}
}

func TestNormalizeOrderBookMessageAcceptsWindowedBinanceDeltaAfterSnapshot(t *testing.T) {
	metadata := BookMetadata{
		Symbol:        "BTC-USD",
		SourceSymbol:  "BTCUSDT",
		QuoteCurrency: "USDT",
		Venue:         VenueBinance,
		MarketType:    "spot",
	}
	sequencer := &OrderBookSequencer{}

	snapshot, err := NormalizeOrderBookMessage(metadata, OrderBookMessage{
		Type:          "snapshot",
		FirstSequence: 700,
		Sequence:      700,
		BestBidPrice:  "64020.00",
		BestAskPrice:  "64020.50",
		RecvTs:        "2026-03-06T12:02:00.25Z",
	}, sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	delta, err := NormalizeOrderBookMessage(metadata, OrderBookMessage{
		Type:          "delta",
		FirstSequence: 700,
		Sequence:      701,
		BestBidPrice:  "64020.10",
		BestAskPrice:  "64020.40",
		ExchangeTs:    "2026-03-06T12:02:00.2Z",
		RecvTs:        "2026-03-06T12:02:00.3Z",
	}, sequencer, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize bridging delta: %v", err)
	}

	if snapshot.OrderBookEvent == nil || delta.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events for snapshot and bridging delta")
	}
	if delta.SequenceResult.Action != SequenceAcceptedDelta {
		t.Fatalf("sequence action = %q, want %q", delta.SequenceResult.Action, SequenceAcceptedDelta)
	}
	if delta.OrderBookEvent.SourceRecordID != "book:701" {
		t.Fatalf("sourceRecordId = %q, want %q", delta.OrderBookEvent.SourceRecordID, "book:701")
	}
	if delta.OrderBookEvent.Symbol != "BTC-USD" || delta.OrderBookEvent.SourceSymbol != "BTCUSDT" {
		t.Fatalf("unexpected provenance: %+v", delta.OrderBookEvent)
	}
	if delta.OrderBookEvent.TimestampStatus != TimestampStatusNormal {
		t.Fatalf("timestamp status = %q, want %q", delta.OrderBookEvent.TimestampStatus, TimestampStatusNormal)
	}
	if !delta.OrderBookEvent.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 2, 0, 200000000, time.UTC)) {
		t.Fatalf("canonical event time = %s, want exchange time", delta.OrderBookEvent.CanonicalEventTime)
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

func bookMetadataFromFixture(fixture orderBookFixture) BookMetadata {
	return BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         fixture.Venue,
		MarketType:    "spot",
	}
}

func assertCanonicalOrderBookMatchesFixture(t *testing.T, actual CanonicalOrderBookEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected CanonicalOrderBookEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected order-book event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical order-book mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalFeedHealthMatchesFixture(t *testing.T, actual CanonicalFeedHealthEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected CanonicalFeedHealthEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected feed-health event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical feed-health mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}
