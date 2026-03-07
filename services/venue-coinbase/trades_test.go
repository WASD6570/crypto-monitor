package venuecoinbase

import (
	"encoding/json"
	"os"
	"path/filepath"
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

type tradeFixtureMessage struct {
	RecvTs string `json:"recvTs"`
}

func TestParseTradeEventProducesSharedTradeMessage(t *testing.T) {
	fixture := loadTradeFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustFixtureRecvTime(t, fixture.RawMessages[0])

	parsed, err := ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	if parsed.SourceSymbol != fixture.SourceSymbol {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, fixture.SourceSymbol)
	}
	if parsed.Message.Type != "trade" {
		t.Fatalf("message type = %q, want %q", parsed.Message.Type, "trade")
	}
	if parsed.Message.TradeID != "1001" {
		t.Fatalf("trade id = %q, want %q", parsed.Message.TradeID, "1001")
	}
	if parsed.Message.Side != "buy" {
		t.Fatalf("side = %q, want %q", parsed.Message.Side, "buy")
	}
	if parsed.Message.ExchangeTs != "2026-03-06T12:00:00.1Z" {
		t.Fatalf("exchangeTs = %q, want %q", parsed.Message.ExchangeTs, "2026-03-06T12:00:00.1Z")
	}
	if parsed.Message.RecvTs != "2026-03-06T12:00:00.18Z" {
		t.Fatalf("recvTs = %q, want %q", parsed.Message.RecvTs, "2026-03-06T12:00:00.18Z")
	}
}

func TestParseTradeEventFeedsCanonicalNormalization(t *testing.T) {
	fixture := loadTradeFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustFixtureRecvTime(t, fixture.RawMessages[0])

	parsed, err := ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	event, err := ingestion.NormalizeTradeMessage(ingestion.TradeMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueCoinbase,
		MarketType:    "spot",
	}, parsed.Message, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed trade: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, event, fixture.ExpectedCanonical[0])
	if !event.CanonicalEventTime.Equal(time.Date(2026, time.March, 6, 12, 0, 0, 100000000, time.UTC)) {
		t.Fatalf("canonical event time = %s, want fixture exchange time", event.CanonicalEventTime)
	}
	if event.TimestampFallbackReason != ingestion.TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", event.TimestampFallbackReason)
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
	var message tradeFixtureMessage
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
	if actual != expected {
		t.Fatalf("canonical trade mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}
