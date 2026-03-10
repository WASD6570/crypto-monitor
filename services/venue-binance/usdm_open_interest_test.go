package venuebinance

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestUSDMOpenInterestParsesSnapshot(t *testing.T) {
	recvTime := time.UnixMilli(1772798404060).UTC()
	parsed, err := ParseUSDMOpenInterest([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), recvTime)
	if err != nil {
		t.Fatalf("parse open interest: %v", err)
	}
	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q", parsed.SourceSymbol)
	}
	if parsed.Message.Type != "open-interest" || parsed.Message.OpenInterest != "10659.509" {
		t.Fatalf("unexpected parsed message: %+v", parsed.Message)
	}
	if parsed.Message.ExchangeTs != "2026-03-06T12:00:04Z" {
		t.Fatalf("exchangeTs = %q", parsed.Message.ExchangeTs)
	}
}

func TestUSDMOpenInterestAllowsMissingExchangeTime(t *testing.T) {
	parsed, err := ParseUSDMOpenInterest([]byte(`{"symbol":"ETHUSDT","openInterest":"82450.5"}`), time.UnixMilli(1772798404060).UTC())
	if err != nil {
		t.Fatalf("parse open interest: %v", err)
	}
	if parsed.Message.ExchangeTs != "" {
		t.Fatalf("exchangeTs = %q, want empty", parsed.Message.ExchangeTs)
	}
	if parsed.Message.RecvTs != "2026-03-06T12:00:04.06Z" {
		t.Fatalf("recvTs = %q", parsed.Message.RecvTs)
	}
}

func TestUSDMOpenInterestPollerSchedulesRequestsAndEmitsHealthInputs(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	poller, err := NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	base := time.UnixMilli(1772798404000).UTC()
	plans, err := poller.DuePolls(base)
	if err != nil {
		t.Fatalf("due polls: %v", err)
	}
	if len(plans) != 2 {
		t.Fatalf("due poll count = %d, want 2", len(plans))
	}
	if !plans[0].Allowed || plans[0].Request.Path == "" {
		t.Fatalf("unexpected first plan: %+v", plans[0])
	}

	request, err := poller.BeginPoll("BTCUSDT", base)
	if err != nil {
		t.Fatalf("begin poll: %v", err)
	}
	if request.Path != "/fapi/v1/openInterest?symbol=BTCUSDT" {
		t.Fatalf("request path = %q", request.Path)
	}
	if _, err := poller.AcceptSnapshot([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), base.Add(60*time.Millisecond)); err != nil {
		t.Fatalf("accept snapshot: %v", err)
	}

	status, err := poller.HealthStatus("BTCUSDT", base.Add(time.Second))
	if err != nil {
		t.Fatalf("health status: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	inputs, err := poller.FeedHealthInputs(base.Add(time.Second))
	if err != nil {
		t.Fatalf("feed health inputs: %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("feed health inputs = %d, want 2", len(inputs))
	}
	if inputs[0].Message.SourceRecordID == "" {
		t.Fatal("expected source record id")
	}
}

func TestUSDMOpenInterestPollerMarksStaleAndRateLimit(t *testing.T) {
	config := loadBinanceRuntimeConfig(t)
	config.OpenInterestPollInterval = time.Millisecond
	config.OpenInterestPollsPerMinuteLimit = 2
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	poller, err := NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	base := time.UnixMilli(1772798404000).UTC()
	if _, err := poller.BeginPoll("BTCUSDT", base); err != nil {
		t.Fatalf("begin first poll: %v", err)
	}
	if _, err := poller.AcceptSnapshot([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), base.Add(100*time.Microsecond)); err != nil {
		t.Fatalf("accept first snapshot: %v", err)
	}
	if _, err := poller.BeginPoll("BTCUSDT", base.Add(time.Millisecond)); err != nil {
		t.Fatalf("begin second poll: %v", err)
	}
	if _, err := poller.BeginPoll("BTCUSDT", base.Add(2*time.Millisecond)); err == nil {
		t.Fatal("expected poll rate limit error")
	}

	rateLimited, err := poller.HealthStatus("BTCUSDT", base.Add(2*time.Millisecond))
	if err != nil {
		t.Fatalf("rate-limited health status: %v", err)
	}
	if rateLimited.State != ingestion.FeedHealthDegraded {
		t.Fatalf("state = %q, want %q", rateLimited.State, ingestion.FeedHealthDegraded)
	}
	if !hasReason(rateLimited.Reasons, ingestion.ReasonRateLimit) {
		t.Fatalf("reasons = %v, want %q", rateLimited.Reasons, ingestion.ReasonRateLimit)
	}

	stale, err := poller.HealthStatus("BTCUSDT", base.Add(16*time.Second))
	if err != nil {
		t.Fatalf("stale health status: %v", err)
	}
	if stale.State != ingestion.FeedHealthStale {
		t.Fatalf("state = %q, want %q", stale.State, ingestion.FeedHealthStale)
	}
	if !hasReason(stale.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, want %q", stale.Reasons, ingestion.ReasonMessageStale)
	}
}
