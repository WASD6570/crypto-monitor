package venuebinance

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestSpotRawAppendUsesSupervisorSession(t *testing.T) {
	supervisor := connectedSpotRawAppendSupervisor(t)
	context, err := SpotWebsocketRawWriteContext(supervisor)
	if err != nil {
		t.Fatalf("spot raw write context: %v", err)
	}
	if context.ConnectionRef != spotWebsocketRawConnectionRef {
		t.Fatalf("connection ref = %q, want %q", context.ConnectionRef, spotWebsocketRawConnectionRef)
	}
	if context.SessionRef != supervisor.State().SessionRef {
		t.Fatalf("session ref = %q, want %q", context.SessionRef, supervisor.State().SessionRef)
	}
}

func TestSpotDepthFeedHealthRawAppendRetainsDepthDegradedReference(t *testing.T) {
	supervisor := connectedSpotRawAppendSupervisor(t)
	entry, err := BuildSpotDepthFeedHealthRawAppendEntry(
		supervisor,
		ingestion.CanonicalFeedHealthEvent{
			SchemaVersion:      "v1",
			EventType:          "feed-health",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              ingestion.VenueBinance,
			MarketType:         "spot",
			ExchangeTs:         "2026-03-06T12:04:00Z",
			RecvTs:             "2026-03-06T12:04:00.050Z",
			TimestampStatus:    ingestion.TimestampStatusNormal,
			FeedHealthState:    ingestion.FeedHealthDegraded,
			DegradationReasons: []ingestion.DegradationReason{ingestion.ReasonSequenceGap},
			SourceRecordID:     "runtime:binance-spot-depth:BTCUSDT",
			CanonicalEventTime: mustRawAppendTime(t, "2026-03-06T12:04:00Z"),
		},
		"BTCUSDT",
		ingestion.RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build spot depth feed-health entry: %v", err)
	}
	if entry.ConnectionRef != spotWebsocketRawConnectionRef {
		t.Fatalf("connection ref = %q, want %q", entry.ConnectionRef, spotWebsocketRawConnectionRef)
	}
	if entry.SessionRef != supervisor.State().SessionRef {
		t.Fatalf("session ref = %q, want %q", entry.SessionRef, supervisor.State().SessionRef)
	}
	if entry.StreamKey != string(ingestion.StreamOrderBook) {
		t.Fatalf("stream key = %q, want %q", entry.StreamKey, ingestion.StreamOrderBook)
	}
	if entry.DegradedFeedRef != "runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("degraded feed ref = %q, want %q", entry.DegradedFeedRef, "runtime:binance-spot-depth:BTCUSDT")
	}
}

func TestUSDMFundingRawAppendUsesWebsocketProvenance(t *testing.T) {
	runtime := connectedUSDMRawAppendRuntime(t)
	entry, err := BuildUSDMFundingRawAppendEntry(
		runtime,
		ingestion.CanonicalFundingRateEvent{
			SchemaVersion:      "v1",
			EventType:          "funding-rate",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              ingestion.VenueBinance,
			MarketType:         "perpetual",
			FundingRate:        "0.00038167",
			NextFundingTs:      "2026-03-06T16:00:00Z",
			ExchangeTs:         "2026-03-06T12:00:05Z",
			RecvTs:             "2026-03-06T12:00:05.040Z",
			TimestampStatus:    ingestion.TimestampStatusNormal,
			SourceRecordID:     "funding:2026-03-06T12:00:05Z",
			CanonicalEventTime: mustRawAppendTime(t, "2026-03-06T12:00:05Z"),
		},
		ingestion.DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		ingestion.FundingRateMessage{Type: "funding-rate", FundingRate: "0.00038167", NextFundingTs: "2026-03-06T16:00:00Z", ExchangeTs: "2026-03-06T12:00:05Z", RecvTs: "2026-03-06T12:00:05.040Z"},
		ingestion.RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build funding entry: %v", err)
	}
	if entry.ConnectionRef != usdmWebsocketRawConnectionRef {
		t.Fatalf("connection ref = %q, want %q", entry.ConnectionRef, usdmWebsocketRawConnectionRef)
	}
	if entry.SessionRef != runtime.State().SessionRef {
		t.Fatalf("session ref = %q, want %q", entry.SessionRef, runtime.State().SessionRef)
	}
	if entry.DuplicateAudit.IdentityKey != "message:funding-rate:2026-03-06T12:00:05Z" {
		t.Fatalf("identity key = %q, want %q", entry.DuplicateAudit.IdentityKey, "message:funding-rate:2026-03-06T12:00:05Z")
	}
}

func TestUSDMOpenInterestRawAppendUsesRESTProvenance(t *testing.T) {
	runtimeConfig := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	poller, err := NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new open interest poller: %v", err)
	}
	entry, err := BuildUSDMOpenInterestRawAppendEntry(
		poller,
		ingestion.CanonicalOpenInterestSnapshotEvent{
			SchemaVersion:      "v1",
			EventType:          "open-interest-snapshot",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              ingestion.VenueBinance,
			MarketType:         "perpetual",
			OpenInterest:       "10659.509",
			ExchangeTs:         "2026-03-06T12:00:04Z",
			RecvTs:             "2026-03-06T12:00:04.060Z",
			TimestampStatus:    ingestion.TimestampStatusNormal,
			SourceRecordID:     "open-interest:2026-03-06T12:00:04Z",
			CanonicalEventTime: mustRawAppendTime(t, "2026-03-06T12:00:04Z"),
		},
		ingestion.DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		ingestion.OpenInterestMessage{Type: "open-interest", OpenInterest: "10659.509", ExchangeTs: "2026-03-06T12:00:04Z", RecvTs: "2026-03-06T12:00:04.060Z"},
		ingestion.RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build open-interest entry: %v", err)
	}
	if entry.ConnectionRef != usdmOpenInterestRawConnectionRef {
		t.Fatalf("connection ref = %q, want %q", entry.ConnectionRef, usdmOpenInterestRawConnectionRef)
	}
	if entry.SessionRef != usdmOpenInterestRawSessionPrefix+"btcusdt" {
		t.Fatalf("session ref = %q, want %q", entry.SessionRef, usdmOpenInterestRawSessionPrefix+"btcusdt")
	}
	if entry.StreamFamily != string(ingestion.StreamOpenInterest) {
		t.Fatalf("stream family = %q, want %q", entry.StreamFamily, ingestion.StreamOpenInterest)
	}
	if entry.PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/open-interest" {
		t.Fatalf("partition key = %q, want %q", entry.PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/open-interest")
	}
}

func TestUSDMOpenInterestFeedHealthRawAppendUsesDistinctRESTSurface(t *testing.T) {
	runtimeConfig := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	poller, err := NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new open interest poller: %v", err)
	}
	entry, err := BuildUSDMOpenInterestFeedHealthRawAppendEntry(
		poller,
		ingestion.CanonicalFeedHealthEvent{
			SchemaVersion:      "v1",
			EventType:          "feed-health",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              ingestion.VenueBinance,
			MarketType:         "perpetual",
			ExchangeTs:         "2026-03-06T12:04:00Z",
			RecvTs:             "2026-03-06T12:04:00.030Z",
			TimestampStatus:    ingestion.TimestampStatusNormal,
			FeedHealthState:    ingestion.FeedHealthDegraded,
			DegradationReasons: []ingestion.DegradationReason{ingestion.ReasonRateLimit},
			SourceRecordID:     "runtime:binance-usdm-open-interest:BTCUSDT",
			CanonicalEventTime: mustRawAppendTime(t, "2026-03-06T12:04:00Z"),
		},
		"BTCUSDT",
		ingestion.RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build open-interest feed-health entry: %v", err)
	}
	if entry.ConnectionRef != usdmOpenInterestRawConnectionRef {
		t.Fatalf("connection ref = %q, want %q", entry.ConnectionRef, usdmOpenInterestRawConnectionRef)
	}
	if entry.StreamKey != string(ingestion.StreamOpenInterest) {
		t.Fatalf("stream key = %q, want %q", entry.StreamKey, ingestion.StreamOpenInterest)
	}
	if entry.DegradedFeedRef != "runtime:binance-usdm-open-interest:BTCUSDT" {
		t.Fatalf("degraded feed ref = %q, want %q", entry.DegradedFeedRef, "runtime:binance-usdm-open-interest:BTCUSDT")
	}
	if entry.DuplicateAudit.IdentityKey != "message:open-interest:runtime:binance-usdm-open-interest:BTCUSDT" {
		t.Fatalf("identity key = %q, want %q", entry.DuplicateAudit.IdentityKey, "message:open-interest:runtime:binance-usdm-open-interest:BTCUSDT")
	}
}

func connectedSpotRawAppendSupervisor(t *testing.T) *SpotWebsocketSupervisor {
	t.Helper()
	runtimeConfig := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	supervisor, err := NewSpotWebsocketSupervisor(runtime)
	if err != nil {
		t.Fatalf("new spot websocket supervisor: %v", err)
	}
	base := time.UnixMilli(1772798400000).UTC()
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
	return supervisor
}

func connectedUSDMRawAppendRuntime(t *testing.T) *USDMRuntime {
	t.Helper()
	runtimeConfig := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}
	base := time.UnixMilli(1772798400000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	return usdm
}

func mustRawAppendTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse raw append time %q: %v", value, err)
	}
	return parsed
}
