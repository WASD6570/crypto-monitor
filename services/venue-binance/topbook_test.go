package venuebinance

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestParseTopOfBookEventProducesSharedBookMessage(t *testing.T) {
	recvTime := time.UnixMilli(1772798400180).UTC()
	raw := []byte(`{"u":9001,"s":"BTCUSDT","b":"64000.10","B":"1.25","a":"64000.20","A":"0.98"}`)

	parsed, err := ParseTopOfBookEvent(raw, recvTime)
	if err != nil {
		t.Fatalf("parse top-of-book event: %v", err)
	}

	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, "BTCUSDT")
	}
	if parsed.Message.Type != string(ingestion.BookUpdateTopOfBook) {
		t.Fatalf("message type = %q, want %q", parsed.Message.Type, ingestion.BookUpdateTopOfBook)
	}
	if parsed.Message.Sequence != 9001 {
		t.Fatalf("sequence = %d, want %d", parsed.Message.Sequence, 9001)
	}
	if parsed.Message.BestBidPrice != "64000.10" {
		t.Fatalf("best bid = %q, want %q", parsed.Message.BestBidPrice, "64000.10")
	}
	if parsed.Message.BestAskPrice != "64000.20" {
		t.Fatalf("best ask = %q, want %q", parsed.Message.BestAskPrice, "64000.20")
	}
	if parsed.Message.ExchangeTs != "" {
		t.Fatalf("exchangeTs = %q, want empty", parsed.Message.ExchangeTs)
	}
	if parsed.Message.RecvTs != "2026-03-06T12:00:00.18Z" {
		t.Fatalf("recvTs = %q, want %q", parsed.Message.RecvTs, "2026-03-06T12:00:00.18Z")
	}
}

func TestParseTopOfBookEventPreservesExchangeTimestampWhenPresent(t *testing.T) {
	recvTime := time.UnixMilli(1772798520130).UTC()
	raw := []byte(`{"u":9002,"E":1772798520100,"s":"ETHUSDT","b":"3501.10","B":"2.4","a":"3501.30","A":"1.7"}`)

	parsed, err := ParseTopOfBookEvent(raw, recvTime)
	if err != nil {
		t.Fatalf("parse top-of-book event: %v", err)
	}

	if parsed.Message.ExchangeTs != "2026-03-06T12:02:00.1Z" {
		t.Fatalf("exchangeTs = %q, want %q", parsed.Message.ExchangeTs, "2026-03-06T12:02:00.1Z")
	}
	if parsed.Message.Sequence != 9002 {
		t.Fatalf("sequence = %d, want %d", parsed.Message.Sequence, 9002)
	}
}

func TestParseTopOfBookFrameConsumesSupervisorAcceptedFrame(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new supervisor: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := supervisor.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := supervisor.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := supervisor.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	frame, err := supervisor.AcceptDataFrame([]byte(`{"u":9010,"s":"ETHUSDT","b":"3502.10","B":"5.0","a":"3502.40","A":"4.1"}`), base.Add(300*time.Millisecond))
	if err != nil {
		t.Fatalf("accept top-of-book frame: %v", err)
	}

	parsed, err := ParseTopOfBookFrame(frame)
	if err != nil {
		t.Fatalf("parse top-of-book frame: %v", err)
	}
	if parsed.SourceSymbol != "ETHUSDT" {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, "ETHUSDT")
	}
	if parsed.Message.Type != string(ingestion.BookUpdateTopOfBook) {
		t.Fatalf("message type = %q, want %q", parsed.Message.Type, ingestion.BookUpdateTopOfBook)
	}
	if parsed.Message.Sequence != 9010 {
		t.Fatalf("sequence = %d, want %d", parsed.Message.Sequence, 9010)
	}
	if parsed.Message.RecvTs != "2026-03-06T12:02:05.3Z" {
		t.Fatalf("recvTs = %q, want %q", parsed.Message.RecvTs, "2026-03-06T12:02:05.3Z")
	}
}
