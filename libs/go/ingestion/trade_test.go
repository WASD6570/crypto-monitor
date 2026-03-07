package ingestion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type tradeFixture struct {
	Symbol            string                `json:"symbol"`
	SourceSymbol      string                `json:"sourceSymbol"`
	QuoteCurrency     string                `json:"quoteCurrency"`
	Venue             Venue                 `json:"venue"`
	RawMessages       []TradeMessage        `json:"rawMessages"`
	ExpectedCanonical []CanonicalTradeEvent `json:"expectedCanonical"`
}

func TestNormalizeTradeMessageMatchesHappyFixture(t *testing.T) {
	fixture := loadTradeFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-trade-usdt.fixture.v1.json")
	actual, err := NormalizeTradeMessage(tradeMetadataFromFixture(fixture), fixture.RawMessages[0], StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize trade message: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	if !actual.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 0, 100000000, time.UTC)) {
		t.Fatalf("canonical event time = %s, want fixture exchange time", actual.CanonicalEventTime)
	}
	if actual.TimestampFallbackReason != TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", actual.TimestampFallbackReason)
	}
}

func TestNormalizeTradeMessageMatchesDegradedTimestampFixture(t *testing.T) {
	fixture := loadTradeFixture(t, "tests/fixtures/events/binance/ETH-USD/edge-timestamp-degraded-usdt.fixture.v1.json")
	actual, err := NormalizeTradeMessage(tradeMetadataFromFixture(fixture), fixture.RawMessages[0], StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize trade message: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	wantRecvTime := time.Date(2026, time.March, 6, 12, 2, 20, 30000000, time.UTC)
	if !actual.CanonicalEventTime.Equal(wantRecvTime) {
		t.Fatalf("canonical event time = %s, want recv time %s", actual.CanonicalEventTime, wantRecvTime)
	}
	if actual.TimestampFallbackReason != TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, TimestampReasonExchangeSkewExceeded)
	}
}

func loadTradeFixture(t *testing.T, relativePath string) tradeFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}

	var fixture tradeFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	if len(fixture.RawMessages) != 1 {
		t.Fatalf("fixture %s rawMessages = %d, want 1", relativePath, len(fixture.RawMessages))
	}
	if len(fixture.ExpectedCanonical) != 1 {
		t.Fatalf("fixture %s expectedCanonical = %d, want 1", relativePath, len(fixture.ExpectedCanonical))
	}
	return fixture
}

func tradeMetadataFromFixture(fixture tradeFixture) TradeMetadata {
	return TradeMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         fixture.Venue,
		MarketType:    "spot",
	}
}

func assertCanonicalTradeMatchesFixture(t *testing.T, actual, expected CanonicalTradeEvent) {
	t.Helper()
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical trade mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}
