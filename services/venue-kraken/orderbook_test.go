package venuekraken

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
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

func TestParseOrderBookEventFeedsCanonicalNormalizationHappyPath(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/kraken/ETH-USD/happy-order-book-snapshot-delta-usd.fixture.v1.json")
	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueKraken,
		MarketType:    "spot",
	}
	state := &L2IntegrityState{}

	snapshotRaw := []byte(`{"channel":"book","type":"snapshot","pair":"ETH/USD","sequence":500,"bids":[["3499.10","1.20"]],"asks":[["3499.40","0.95"]],"exchangeTs":"2026-03-06T12:00:02.000Z"}`)
	deltaRaw := []byte(`{"channel":"book","type":"delta","pair":"ETH/USD","sequence":501,"bids":[["3499.20","1.00"]],"asks":[["3499.50","0.90"]],"exchangeTs":"2026-03-06T12:00:02.100Z"}`)

	snapshot, err := ParseOrderBookEvent(snapshotRaw, "2026-03-06T12:00:02.040Z")
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	delta, err := ParseOrderBookEvent(deltaRaw, "2026-03-06T12:00:02.130Z")
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}

	first, err := state.Normalize(metadata, snapshot.Message)
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := state.Normalize(metadata, delta.Message)
	if err != nil {
		t.Fatalf("normalize delta: %v", err)
	}

	if first.OrderBookEvent == nil || second.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events for snapshot and delta")
	}
	if first.FeedHealthEvent != nil || second.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health event for happy path")
	}
	assertCanonicalOrderBookMatchesFixture(t, *first.OrderBookEvent, fixture.ExpectedCanonical[0])
	assertCanonicalOrderBookMatchesFixture(t, *second.OrderBookEvent, fixture.ExpectedCanonical[1])
	if !state.Ready() {
		t.Fatal("expected l2 integrity state to be ready after sequential delta")
	}
}

func TestParseOrderBookEventMarksGapAsResync(t *testing.T) {
	fixture := loadOrderBookFixture(t, "tests/fixtures/events/kraken/ETH-USD/edge-sequence-gap-usd.fixture.v1.json")
	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueKraken,
		MarketType:    "spot",
	}
	state := &L2IntegrityState{}

	snapshotRaw := []byte(`{"channel":"book","type":"snapshot","pair":"ETH/USD","sequence":900,"bids":[["3498.80","1.40"]],"asks":[["3499.10","1.10"]],"exchangeTs":"2026-03-06T12:01:00.000Z"}`)
	gapRaw := []byte(`{"channel":"book","type":"delta","pair":"ETH/USD","sequence":902,"bids":[["3498.70","1.35"]],"asks":[["3499.20","1.05"]],"exchangeTs":"2026-03-06T12:01:00.100Z"}`)

	snapshot, err := ParseOrderBookEvent(snapshotRaw, "2026-03-06T12:01:00.030Z")
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	gap, err := ParseOrderBookEvent(gapRaw, "2026-03-06T12:01:00.120Z")
	if err != nil {
		t.Fatalf("parse gap delta: %v", err)
	}

	first, err := state.Normalize(metadata, snapshot.Message)
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := state.Normalize(metadata, gap.Message)
	if err != nil {
		t.Fatalf("normalize gap delta: %v", err)
	}

	if first.OrderBookEvent == nil {
		t.Fatal("expected snapshot event before gap")
	}
	if second.OrderBookEvent != nil {
		t.Fatal("did not expect order-book event on gap")
	}
	if second.FeedHealthEvent == nil {
		t.Fatal("expected feed-health event on gap")
	}
	assertCanonicalFeedHealthMatchesFixture(t, *second.FeedHealthEvent, fixture.ExpectedCanonical[0])
	if !state.ResyncRequired() {
		t.Fatal("expected l2 integrity state to require resync after gap")
	}
	if second.SequenceResult.Action != ingestion.SequenceRequiresResync {
		t.Fatalf("sequence action = %q, want %q", second.SequenceResult.Action, ingestion.SequenceRequiresResync)
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
