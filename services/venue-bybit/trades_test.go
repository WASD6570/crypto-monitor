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

type tradeFixture struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

type fixtureMessage struct {
	RecvTs string `json:"recvTs"`
}

func TestParseTradeEventFeedsCanonicalNormalization(t *testing.T) {
	fixture := loadTradeFixture(t, "tests/fixtures/events/bybit/BTC-USD/happy-trade-usdt.fixture.v1.json")
	recvTime := mustFixtureRecvTime(t, fixture.RawMessages[0])

	parsed, err := ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	event, err := ingestion.NormalizeTradeMessage(ingestion.TradeMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueBybit,
		MarketType:    "spot",
	}, parsed.Message, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed trade: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, event, fixture.ExpectedCanonical[0])
	if !event.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 0, 100000000, time.UTC)) {
		t.Fatalf("canonical event time = %s, want fixture exchange time", event.CanonicalEventTime)
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
	return fixture
}

func mustFixtureRecvTime(t *testing.T, raw json.RawMessage) time.Time {
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
