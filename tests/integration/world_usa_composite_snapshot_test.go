package integration

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestWorldUSACompositeSnapshotSeams(t *testing.T) {
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

	snapshots, err := service.BuildWorldUSASnapshots("ETH-USD", time.Date(2026, 3, 6, 12, 6, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3500, LiquidityScore: 110, TimestampStatus: ingestion.TimestampStatusDegraded, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 3502, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded, FeedHealthReasons: []ingestion.DegradationReason{ingestion.ReasonReconnectLoop}},
		{Symbol: "ETH-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 3501, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 3499, LiquidityScore: 70, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthStale},
	})
	if err != nil {
		t.Fatalf("build snapshots: %v", err)
	}
	world := snapshots[0]
	usa := snapshots[1]
	if !world.Degraded || !usa.Degraded {
		t.Fatalf("expected degraded snapshots, got world=%v usa=%v", world.Degraded, usa.Degraded)
	}
	if world.CompositePrice == nil || usa.CompositePrice == nil {
		t.Fatal("expected composite price seams for both groups")
	}
	if world.TimestampFallbackContributorCount != 1 {
		t.Fatalf("world timestamp fallback contributors = %d, want 1", world.TimestampFallbackContributorCount)
	}
	if usa.ContributingContributorCount != 1 || usa.ConfiguredContributorCount != 2 {
		t.Fatalf("usa contributor counts = %d/%d, want 1/2", usa.ContributingContributorCount, usa.ConfiguredContributorCount)
	}
	if world.MaxContributorWeight == 0 || usa.MaxContributorWeight == 0 {
		t.Fatalf("expected concentration seam fields, got world=%v usa=%v", world.MaxContributorWeight, usa.MaxContributorWeight)
	}
}
