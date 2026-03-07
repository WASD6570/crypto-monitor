package replay

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestWorldUSACompositeDeterminism(t *testing.T) {
	service := newReplayService(t)
	inputs := []features.ContributorInput{
		{Symbol: "BTC-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 64000.1, LiquidityScore: 180, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 64020.2, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded, FeedHealthReasons: []ingestion.DegradationReason{ingestion.ReasonReconnectLoop}},
		{Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 64005.3, LiquidityScore: 85, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 64002.4, LiquidityScore: 80, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
	}
	bucketTs := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	first, err := service.BuildWorldUSASnapshots("BTC-USD", bucketTs, inputs)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	second, err := service.BuildWorldUSASnapshots("BTC-USD", bucketTs, inputs)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("snapshot replay outputs differ\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestWorldUSAReplayTimestampFallback(t *testing.T) {
	service := newReplayService(t)
	snapshots, err := service.BuildWorldUSASnapshots("ETH-USD", time.Date(2026, 3, 6, 12, 2, 30, 0, time.UTC), []features.ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3501.2, LiquidityScore: 120, TimestampStatus: ingestion.TimestampStatusDegraded, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 3501.5, LiquidityScore: 115, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 3499.2, LiquidityScore: 80, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 3498.8, LiquidityScore: 60, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthStale},
	})
	if err != nil {
		t.Fatalf("build snapshots: %v", err)
	}
	if snapshots[0].TimestampFallbackContributorCount != 1 {
		t.Fatalf("world timestamp fallback contributors = %d, want 1", snapshots[0].TimestampFallbackContributorCount)
	}
	if !snapshots[0].Degraded {
		t.Fatal("expected world snapshot to remain degraded")
	}
	if snapshots[1].Unavailable {
		t.Fatal("did not expect usa snapshot to be unavailable")
	}
}

func newReplayService(t *testing.T) *featureengine.Service {
	t.Helper()
	service, err := featureengine.NewService(features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties:        features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75},
		QuoteProxies: map[string]features.QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
			"USDC": {Enabled: true, PenaltyMultiplier: 0.98},
		},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {
				Members: []features.MemberConfig{
					{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			features.CompositeGroupUSA: {
				Members: []features.MemberConfig{
					{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.7},
			},
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}
