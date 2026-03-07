package features

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestContributorEligibility(t *testing.T) {
	snapshot, err := BuildCompositeSnapshot(testConfig(), CompositeGroupUSA, "BTC-USD", time.Unix(0, 0).UTC(), []ContributorInput{
		{Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 64000, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 64010, LiquidityScore: 80, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthStale},
	})
	if err != nil {
		t.Fatalf("build composite: %v", err)
	}
	if snapshot.EligibleContributorCount != 1 {
		t.Fatalf("eligible contributors = %d, want 1", snapshot.EligibleContributorCount)
	}
	if snapshot.Contributors[1].Status != ContributorStatusExcluded {
		t.Fatalf("kraken status = %q, want excluded", snapshot.Contributors[1].Status)
	}
	if !containsReason(snapshot.Contributors[1].ReasonCodes, ReasonFeedHealthExcluded) {
		t.Fatalf("kraken reasons = %v, want %q", snapshot.Contributors[1].ReasonCodes, ReasonFeedHealthExcluded)
	}
}

func TestStablecoinNormalization(t *testing.T) {
	config := testConfig()
	world, err := BuildCompositeSnapshot(config, CompositeGroupWorld, "ETH-USD", time.Unix(0, 0).UTC(), []ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3500, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDC", Price: 3501, LiquidityScore: 80, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, QuoteConfidenceDropped: true},
	})
	if err != nil {
		t.Fatalf("build composite: %v", err)
	}
	if world.EligibleContributorCount != 1 {
		t.Fatalf("eligible contributors = %d, want 1", world.EligibleContributorCount)
	}
	if world.QuoteNormalizationMode != QuoteNormalizationModeProxy {
		t.Fatalf("quote mode = %q, want %q", world.QuoteNormalizationMode, QuoteNormalizationModeProxy)
	}
	if !containsReason(world.Contributors[1].ReasonCodes, ReasonQuoteConfidenceLoss) {
		t.Fatalf("bybit reasons = %v, want %q", world.Contributors[1].ReasonCodes, ReasonQuoteConfidenceLoss)
	}

	config.QuoteProxies["USDT"] = QuoteProxyRule{Enabled: false, PenaltyMultiplier: 1}
	denied, err := BuildCompositeSnapshot(config, CompositeGroupWorld, "ETH-USD", time.Unix(0, 0).UTC(), []ContributorInput{{
		Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3500, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy,
	}})
	if err != nil {
		t.Fatalf("build denied composite: %v", err)
	}
	if !denied.Unavailable {
		t.Fatal("expected unavailable snapshot when quote proxy is denied")
	}
	if !containsReason(denied.Contributors[0].ReasonCodes, ReasonQuoteProxyDenied) {
		t.Fatalf("reasons = %v, want %q", denied.Contributors[0].ReasonCodes, ReasonQuoteProxyDenied)
	}
}

func TestCompositeWeighting(t *testing.T) {
	snapshot, err := BuildCompositeSnapshot(testConfig(), CompositeGroupWorld, "BTC-USD", time.Unix(0, 0).UTC(), []ContributorInput{
		{Symbol: "BTC-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 64000, LiquidityScore: 200, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 64020, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusDegraded, FeedHealthState: ingestion.FeedHealthDegraded},
	})
	if err != nil {
		t.Fatalf("build composite: %v", err)
	}
	if snapshot.CompositePrice == nil {
		t.Fatal("expected composite price")
	}
	weights := []float64{snapshot.Contributors[0].FinalWeight, snapshot.Contributors[1].FinalWeight}
	if !reflect.DeepEqual(weights, []float64{0.769231, 0.230769}) {
		t.Fatalf("weights = %v, want [0.769231 0.230769]", weights)
	}
	if snapshot.HealthScore >= 1 {
		t.Fatalf("health score = %v, want degraded score", snapshot.HealthScore)
	}
}

func TestCompositeClamping(t *testing.T) {
	snapshot, err := BuildCompositeSnapshot(testConfig(), CompositeGroupUSA, "ETH-USD", time.Unix(0, 0).UTC(), []ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 3500, LiquidityScore: 1000, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 3495, LiquidityScore: 10, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
	})
	if err != nil {
		t.Fatalf("build composite: %v", err)
	}
	if snapshot.Contributors[0].Status != ContributorStatusClamped {
		t.Fatalf("coinbase status = %q, want clamped", snapshot.Contributors[0].Status)
	}
	if snapshot.Contributors[0].FinalWeight != 0.7 {
		t.Fatalf("coinbase final weight = %v, want 0.7", snapshot.Contributors[0].FinalWeight)
	}
	if snapshot.Contributors[1].FinalWeight != 0.3 {
		t.Fatalf("kraken final weight = %v, want 0.3", snapshot.Contributors[1].FinalWeight)
	}
}

func testConfig() CompositeConfig {
	return CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties: PenaltyConfig{
			FeedHealthDegradedMultiplier: 0.8,
			TimestampDegradedMultiplier:  0.75,
		},
		QuoteProxies: map[string]QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
			"USDC": {Enabled: true, PenaltyMultiplier: 0.98},
		},
		Groups: map[CompositeGroup]GroupConfig{
			CompositeGroupWorld: {
				Members: []MemberConfig{
					{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			CompositeGroupUSA: {
				Members: []MemberConfig{
					{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: ClampConfig{MinWeight: 0.2, MaxWeight: 0.7},
			},
		},
	}
}

func containsReason(reasons []ReasonCode, target ReasonCode) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
