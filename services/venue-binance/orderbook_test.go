package venuebinance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestParseOrderBookSnapshotProducesSharedBookMessage(t *testing.T) {
	recvTime := time.UnixMilli(1772798460030).UTC()
	raw := []byte(`{"lastUpdateId":900,"symbol":"BTCUSDT","bids":[["64010.00","1.25"]],"asks":[["64010.50","0.90"]]}`)

	parsed, err := ParseOrderBookSnapshot(raw, recvTime)
	if err != nil {
		t.Fatalf("parse order-book snapshot: %v", err)
	}

	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, "BTCUSDT")
	}
	if parsed.Message.Type != "snapshot" {
		t.Fatalf("message type = %q, want %q", parsed.Message.Type, "snapshot")
	}
	if parsed.Message.Sequence != 900 {
		t.Fatalf("sequence = %d, want %d", parsed.Message.Sequence, 900)
	}
	if parsed.Message.BestBidPrice != "64010.00" {
		t.Fatalf("best bid = %q, want %q", parsed.Message.BestBidPrice, "64010.00")
	}
	if parsed.Message.BestAskPrice != "64010.50" {
		t.Fatalf("best ask = %q, want %q", parsed.Message.BestAskPrice, "64010.50")
	}
	if parsed.Message.ExchangeTs != "" {
		t.Fatalf("exchangeTs = %q, want empty for REST snapshot", parsed.Message.ExchangeTs)
	}
	if parsed.Message.RecvTs != "2026-03-06T12:01:00.03Z" {
		t.Fatalf("recvTs = %q, want %q", parsed.Message.RecvTs, "2026-03-06T12:01:00.03Z")
	}
}

func TestParseOrderBookEventsFeedCanonicalNormalizationHappyPath(t *testing.T) {
	snapshotRecvTime := time.UnixMilli(1772798520040).UTC()
	deltaRecvTime := time.UnixMilli(1772798520130).UTC()
	snapshotRaw := []byte(`{"lastUpdateId":700,"symbol":"BTCUSDT","bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`)
	deltaRaw := []byte(`{"e":"depthUpdate","E":1772798520100,"s":"BTCUSDT","U":701,"u":701,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`)

	snapshot, err := ParseOrderBookSnapshot(snapshotRaw, snapshotRecvTime)
	if err != nil {
		t.Fatalf("parse order-book snapshot: %v", err)
	}
	delta, err := ParseOrderBookDelta(deltaRaw, deltaRecvTime)
	if err != nil {
		t.Fatalf("parse order-book delta: %v", err)
	}

	metadata := ingestion.BookMetadata{
		Symbol:        "BTC-USD",
		SourceSymbol:  snapshot.SourceSymbol,
		QuoteCurrency: "USDT",
		Venue:         ingestion.VenueBinance,
		MarketType:    "spot",
	}
	sequencer := &ingestion.OrderBookSequencer{}

	first, err := ingestion.NormalizeOrderBookMessage(metadata, snapshot.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed snapshot: %v", err)
	}
	second, err := ingestion.NormalizeOrderBookMessage(metadata, delta.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed delta: %v", err)
	}

	if first.SequenceResult.Action != ingestion.SequenceAcceptedSnapshot {
		t.Fatalf("snapshot action = %q, want %q", first.SequenceResult.Action, ingestion.SequenceAcceptedSnapshot)
	}
	if second.SequenceResult.Action != ingestion.SequenceAcceptedDelta {
		t.Fatalf("delta action = %q, want %q", second.SequenceResult.Action, ingestion.SequenceAcceptedDelta)
	}
	if first.OrderBookEvent == nil || second.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events for snapshot and delta")
	}
	if first.FeedHealthEvent != nil || second.FeedHealthEvent != nil {
		t.Fatal("did not expect feed-health events for happy order-book flow")
	}
	if first.OrderBookEvent.TimestampStatus != ingestion.TimestampStatusDegraded {
		t.Fatalf("snapshot timestamp status = %q, want %q", first.OrderBookEvent.TimestampStatus, ingestion.TimestampStatusDegraded)
	}
	if first.OrderBookEvent.TimestampFallbackReason != ingestion.TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("snapshot fallback reason = %q, want %q", first.OrderBookEvent.TimestampFallbackReason, ingestion.TimestampReasonExchangeMissingOrInvalid)
	}
	if !first.OrderBookEvent.CanonicalEventTime.Equal(snapshotRecvTime) {
		t.Fatalf("snapshot canonical event time = %s, want recv time %s", first.OrderBookEvent.CanonicalEventTime, snapshotRecvTime)
	}
	if first.OrderBookEvent.SourceRecordID != "book:700" {
		t.Fatalf("snapshot sourceRecordId = %q, want %q", first.OrderBookEvent.SourceRecordID, "book:700")
	}
	if second.OrderBookEvent.TimestampStatus != ingestion.TimestampStatusNormal {
		t.Fatalf("delta timestamp status = %q, want %q", second.OrderBookEvent.TimestampStatus, ingestion.TimestampStatusNormal)
	}
	if second.OrderBookEvent.SourceSymbol != "BTCUSDT" {
		t.Fatalf("delta source symbol = %q, want %q", second.OrderBookEvent.SourceSymbol, "BTCUSDT")
	}
	if second.OrderBookEvent.ExchangeTs != "2026-03-06T12:02:00.1Z" {
		t.Fatalf("delta exchangeTs = %q, want %q", second.OrderBookEvent.ExchangeTs, "2026-03-06T12:02:00.1Z")
	}
	if !second.OrderBookEvent.CanonicalEventTime.Equal(time.UnixMilli(1772798520100).UTC()) {
		t.Fatalf("delta canonical event time = %s, want exchange time %s", second.OrderBookEvent.CanonicalEventTime, time.UnixMilli(1772798520100).UTC())
	}
}

func TestParseOrderBookEventsFeedCanonicalNormalizationGapPath(t *testing.T) {
	fixture := loadBinanceGapFixture(t)
	snapshotRecvTime := time.UnixMilli(1772798460030).UTC()
	deltaRecvTime := time.UnixMilli(1772798460120).UTC()
	snapshotRaw := []byte(`{"lastUpdateId":900,"symbol":"BTCUSDT","bids":[["64010.00","1.25"]],"asks":[["64010.50","0.90"]]}`)
	deltaRaw := []byte(`{"e":"depthUpdate","E":1772798460100,"s":"BTCUSDT","U":902,"u":902,"b":[["64009.80","1.10"]],"a":[["64010.40","0.80"]]}`)

	snapshot, err := ParseOrderBookSnapshot(snapshotRaw, snapshotRecvTime)
	if err != nil {
		t.Fatalf("parse order-book snapshot: %v", err)
	}
	delta, err := ParseOrderBookDelta(deltaRaw, deltaRecvTime)
	if err != nil {
		t.Fatalf("parse order-book delta: %v", err)
	}

	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  snapshot.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         fixture.Venue,
		MarketType:    "spot",
	}
	sequencer := &ingestion.OrderBookSequencer{}

	first, err := ingestion.NormalizeOrderBookMessage(metadata, snapshot.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed snapshot: %v", err)
	}
	second, err := ingestion.NormalizeOrderBookMessage(metadata, delta.Message, sequencer, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed delta: %v", err)
	}

	if first.SequenceResult.Action != ingestion.SequenceAcceptedSnapshot {
		t.Fatalf("snapshot action = %q, want %q", first.SequenceResult.Action, ingestion.SequenceAcceptedSnapshot)
	}
	if second.SequenceResult.Action != ingestion.SequenceRequiresResync {
		t.Fatalf("delta action = %q, want %q", second.SequenceResult.Action, ingestion.SequenceRequiresResync)
	}
	if second.OrderBookEvent != nil {
		t.Fatal("did not expect order-book event after sequence gap")
	}
	if second.FeedHealthEvent == nil {
		t.Fatal("expected feed-health event after sequence gap")
	}

	assertCanonicalFeedHealthMatchesFixture(t, *second.FeedHealthEvent, fixture.ExpectedCanonical[0])
	if !second.FeedHealthEvent.CanonicalEventTime.Equal(time.UnixMilli(1772798460100).UTC()) {
		t.Fatalf("gap canonical event time = %s, want exchange time %s", second.FeedHealthEvent.CanonicalEventTime, time.UnixMilli(1772798460100).UTC())
	}
	if second.FeedHealthEvent.TimestampFallbackReason != ingestion.TimestampReasonNone {
		t.Fatalf("gap fallback reason = %q, want none", second.FeedHealthEvent.TimestampFallbackReason)
	}
}

type binanceGapFixture struct {
	Symbol            string            `json:"symbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	Venue             ingestion.Venue   `json:"venue"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

func loadBinanceGapFixture(t *testing.T) binanceGapFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), "tests/fixtures/events/binance/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture binanceGapFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func assertCanonicalFeedHealthMatchesFixture(t *testing.T, actual ingestion.CanonicalFeedHealthEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalFeedHealthEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected feed-health event: %v", err)
	}
	assertTimestampStringsEqual(t, actual.ExchangeTs, expected.ExchangeTs)
	assertTimestampStringsEqual(t, actual.RecvTs, expected.RecvTs)
	actual.ExchangeTs = ""
	actual.RecvTs = ""
	expected.ExchangeTs = ""
	expected.RecvTs = ""
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical feed-health mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertTimestampStringsEqual(t *testing.T, actual, expected string) {
	t.Helper()
	actualTime, err := time.Parse(time.RFC3339Nano, actual)
	if err != nil {
		t.Fatalf("parse actual timestamp %q: %v", actual, err)
	}
	expectedTime, err := time.Parse(time.RFC3339Nano, expected)
	if err != nil {
		t.Fatalf("parse expected timestamp %q: %v", expected, err)
	}
	if !actualTime.Equal(expectedTime) {
		t.Fatalf("timestamp = %q, want same instant as %q", actual, expected)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}
