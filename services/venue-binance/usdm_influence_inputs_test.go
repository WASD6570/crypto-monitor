package venuebinance

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestUSDMInfluenceInputOwnerSnapshotDefaultsToNoContextInStableOrder(t *testing.T) {
	owner := newUSDMInfluenceInputOwnerForTest(t)
	snapshot, err := owner.Snapshot(time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
	if snapshot.Symbols[0].Symbol != "BTC-USD" || snapshot.Symbols[1].Symbol != "ETH-USD" {
		t.Fatalf("unexpected symbol order: %+v", snapshot.Symbols)
	}
	for _, symbol := range snapshot.Symbols {
		if symbol.Funding.Metadata.Freshness != features.USDMInfluenceFreshnessUnavailable {
			t.Fatalf("%s funding freshness = %q", symbol.Symbol, symbol.Funding.Metadata.Freshness)
		}
		if symbol.OpenInterest.Metadata.Surface != features.USDMInfluenceSurfaceRESTPoll {
			t.Fatalf("%s open interest surface = %q", symbol.Symbol, symbol.OpenInterest.Metadata.Surface)
		}
	}
}

func TestUSDMInfluenceInputOwnerSnapshotPreservesMixedFreshAndDegradedInputs(t *testing.T) {
	owner := newUSDMInfluenceInputOwnerForTest(t)
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	if err := owner.AcceptFunding(ingestion.CanonicalFundingRateEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", FundingRate: "0.0008", ExchangeTs: "2026-03-15T11:59:55Z", RecvTs: "2026-03-15T11:59:55.050Z", TimestampStatus: ingestion.TimestampStatusNormal, SourceRecordID: "funding:btc"}); err != nil {
		t.Fatalf("accept funding: %v", err)
	}
	if err := owner.AcceptMarkIndex(ingestion.CanonicalMarkIndexEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", MarkPrice: "64012", IndexPrice: "64000", ExchangeTs: "2026-03-15T11:59:55Z", RecvTs: "2026-03-15T11:59:55.050Z", TimestampStatus: ingestion.TimestampStatusNormal, SourceRecordID: "mark:btc"}); err != nil {
		t.Fatalf("accept mark index: %v", err)
	}
	if err := owner.AcceptOpenInterest(ingestion.CanonicalOpenInterestSnapshotEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", OpenInterest: "10659.509", ExchangeTs: "2026-03-15T11:59:50Z", RecvTs: "2026-03-15T11:59:50.060Z", TimestampStatus: ingestion.TimestampStatusDegraded, SourceRecordID: "oi:btc"}); err != nil {
		t.Fatalf("accept open interest: %v", err)
	}
	if err := owner.ObserveWebsocketFeedHealth(ingestion.CanonicalFeedHealthEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", FeedHealthState: ingestion.FeedHealthDegraded, DegradationReasons: []ingestion.DegradationReason{ingestion.ReasonReconnectLoop}, SourceRecordID: "ws-health:btc"}); err != nil {
		t.Fatalf("observe websocket health: %v", err)
	}
	if err := owner.ObserveOpenInterestFeedHealth(ingestion.CanonicalFeedHealthEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", FeedHealthState: ingestion.FeedHealthHealthy, SourceRecordID: "oi-health:btc"}); err != nil {
		t.Fatalf("observe open interest health: %v", err)
	}

	snapshot, err := owner.Snapshot(now)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	btc := snapshot.Symbols[0]
	if btc.Funding.Metadata.Freshness != features.USDMInfluenceFreshnessDegraded {
		t.Fatalf("funding freshness = %q", btc.Funding.Metadata.Freshness)
	}
	if btc.MarkIndex.Metadata.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("mark feed health state = %q", btc.MarkIndex.Metadata.FeedHealthState)
	}
	if btc.OpenInterest.Metadata.Freshness != features.USDMInfluenceFreshnessDegraded {
		t.Fatalf("open interest freshness = %q", btc.OpenInterest.Metadata.Freshness)
	}
	if btc.OpenInterest.Metadata.AgeMillis == 0 {
		t.Fatal("expected open interest age millis")
	}
}

func TestUSDMInfluenceInputOwnerIgnoresOlderOutOfOrderEvent(t *testing.T) {
	owner := newUSDMInfluenceInputOwnerForTest(t)
	newer := ingestion.CanonicalFundingRateEvent{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: ingestion.VenueBinance, MarketType: "perpetual", FundingRate: "0.0008", ExchangeTs: "2026-03-15T11:59:55Z", RecvTs: "2026-03-15T11:59:55.050Z", TimestampStatus: ingestion.TimestampStatusNormal, SourceRecordID: "funding:newer"}
	older := newer
	older.FundingRate = "0.0001"
	older.ExchangeTs = "2026-03-15T11:59:50Z"
	older.RecvTs = "2026-03-15T11:59:50.050Z"
	older.SourceRecordID = "funding:older"
	if err := owner.AcceptFunding(newer); err != nil {
		t.Fatalf("accept newer funding: %v", err)
	}
	if err := owner.AcceptFunding(older); err != nil {
		t.Fatalf("accept older funding: %v", err)
	}
	snapshot, err := owner.Snapshot(time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Symbols[0].Funding.FundingRate != "0.0008" {
		t.Fatalf("funding rate = %q, want latest event preserved", snapshot.Symbols[0].Funding.FundingRate)
	}
}

func newUSDMInfluenceInputOwnerForTest(t *testing.T) *USDMInfluenceInputOwner {
	t.Helper()
	runtime, err := NewRuntime(loadBinanceRuntimeConfig(t))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	owner, err := NewUSDMInfluenceInputOwner(runtime)
	if err != nil {
		t.Fatalf("new influence input owner: %v", err)
	}
	return owner
}
