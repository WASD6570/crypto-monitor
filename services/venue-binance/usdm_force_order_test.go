package venuebinance

import (
	"testing"
	"time"
)

func TestUSDMForceOrderParsesLiquidationCandidate(t *testing.T) {
	recvTime := time.UnixMilli(1772798406020).UTC()
	parsed, err := ParseUSDMForceOrder([]byte(`{"e":"forceOrder","E":1772798406000,"o":{"s":"BTCUSDT","S":"SELL","q":"0.014","p":"9910","T":1772798406000}}`), recvTime)
	if err != nil {
		t.Fatalf("parse usdm force order: %v", err)
	}
	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q", parsed.SourceSymbol)
	}
	if parsed.Message.Type != "liquidation-print" || parsed.Message.Side != "sell" {
		t.Fatalf("unexpected liquidation message: %+v", parsed.Message)
	}
	if parsed.Message.ExchangeTs != "2026-03-06T12:00:06Z" {
		t.Fatalf("exchangeTs = %q", parsed.Message.ExchangeTs)
	}
	if parsed.Message.RecvTs != "2026-03-06T12:00:06.02Z" {
		t.Fatalf("recvTs = %q", parsed.Message.RecvTs)
	}
}

func TestUSDMForceOrderRejectsMalformedPayload(t *testing.T) {
	_, err := ParseUSDMForceOrder([]byte(`{"e":"forceOrder","E":1772798406000,"o":{"s":"BTCUSDT","q":"0.014","p":"9910"}}`), time.UnixMilli(1772798406020).UTC())
	if err == nil {
		t.Fatal("expected missing side to fail")
	}
}
