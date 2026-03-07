package venuebinance

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestParseTradeEventProducesSharedTradeMessage(t *testing.T) {
	recvTime := time.UnixMilli(1772798400180).UTC()
	raw := []byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.10","q":"0.015","T":1772798400100,"m":false,"M":true}`)

	parsed, err := ParseTradeEvent(raw, recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, "BTCUSDT")
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
	recvTime := time.UnixMilli(1772798400180).UTC()
	raw := []byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1001,"p":"64000.10","q":"0.015","T":1772798400100,"m":false,"M":true}`)

	parsed, err := ParseTradeEvent(raw, recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	event, err := ingestion.NormalizeTradeMessage(ingestion.TradeMetadata{
		Symbol:        "BTC-USD",
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: "USDT",
		Venue:         ingestion.VenueBinance,
		MarketType:    "spot",
	}, parsed.Message, ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize parsed trade: %v", err)
	}

	if event.EventType != "market-trade" {
		t.Fatalf("event type = %q, want %q", event.EventType, "market-trade")
	}
	if event.SourceRecordID != "trade:1001" {
		t.Fatalf("sourceRecordId = %q, want %q", event.SourceRecordID, "trade:1001")
	}
	if event.TimestampStatus != ingestion.TimestampStatusNormal {
		t.Fatalf("timestamp status = %q, want %q", event.TimestampStatus, ingestion.TimestampStatusNormal)
	}
	if event.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", event.SourceSymbol, "BTCUSDT")
	}
	if event.CanonicalEventTime.Format(time.RFC3339Nano) != "2026-03-06T12:00:00.1Z" {
		t.Fatalf("canonical event time = %q, want %q", event.CanonicalEventTime.Format(time.RFC3339Nano), "2026-03-06T12:00:00.1Z")
	}
}

func TestParseTradeEventFallsBackToEventTimeWhenTradeTimeMissing(t *testing.T) {
	recvTime := time.UnixMilli(1772798400180).UTC()
	raw := []byte(`{"e":"trade","E":1772798400180,"s":"BTCUSDT","t":1002,"p":"64000.20","q":"0.020","T":0,"m":true,"M":true}`)

	parsed, err := ParseTradeEvent(raw, recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	if parsed.Message.ExchangeTs != "2026-03-06T12:00:00.18Z" {
		t.Fatalf("exchangeTs = %q, want %q", parsed.Message.ExchangeTs, "2026-03-06T12:00:00.18Z")
	}
	if parsed.Message.Side != "sell" {
		t.Fatalf("side = %q, want %q", parsed.Message.Side, "sell")
	}
}
