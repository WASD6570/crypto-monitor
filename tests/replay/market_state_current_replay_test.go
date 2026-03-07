package replay

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestMarketStateCurrentReplayDeterminism(t *testing.T) {
	first := replayCurrentState(t, "regime-engine.market-state.v1")
	second := replayCurrentState(t, "regime-engine.market-state.v1")
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("current-state replay outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestMarketStateCurrentVersionPinning(t *testing.T) {
	baseline := replayCurrentState(t, "regime-engine.market-state.v1")
	updated := replayCurrentState(t, "regime-engine.market-state.v2")
	if reflect.DeepEqual(baseline, updated) {
		t.Fatal("expected config-version change to alter current-state response")
	}
	if baseline.Version.ConfigVersion != "regime-engine.market-state.v1" {
		t.Fatalf("baseline config version = %q", baseline.Version.ConfigVersion)
	}
	if updated.Version.ConfigVersion != "regime-engine.market-state.v2" {
		t.Fatalf("updated config version = %q", updated.Version.ConfigVersion)
	}
}

func replayCurrentState(t *testing.T, configVersion string) features.MarketStateCurrentResponse {
	t.Helper()
	service, err := featureengine.NewService(features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}})
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	response, err := service.QueryCurrentState(features.SymbolCurrentStateQuery{
		Symbol: "BTC-USD",
		World:  features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupWorld, CompositePrice: floatPtr(64000), CoverageRatio: 1, HealthScore: 0.99, ConfiguredContributorCount: 1, EligibleContributorCount: 1, ContributingContributorCount: 1, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"},
		USA:    features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupUSA, CompositePrice: floatPtr(63990), CoverageRatio: 1, HealthScore: 0.99, ConfiguredContributorCount: 1, EligibleContributorCount: 1, ContributingContributorCount: 1, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"},
		Buckets: []features.MarketQualityBucket{
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily30s, Start: "2026-03-06T12:04:30Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 1, ExpectedBucketCount: 1, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily30s, BucketStart: "2026-03-06T12:04:30Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily2m, Start: "2026-03-06T12:03:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 4, ExpectedBucketCount: 4, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily2m, BucketStart: "2026-03-06T12:03:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily5m, Start: "2026-03-06T12:00:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
		},
		RecentContext: []features.MarketQualityBucket{{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily5m, Start: "2026-03-06T12:00:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}}},
		SymbolRegime:  features.SymbolRegimeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ConfigVersion: configVersion, AlgorithmVersion: "symbol-global-regime.v1"},
		GlobalRegime:  features.GlobalRegimeSnapshot{SchemaVersion: "v1", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ConfigVersion: configVersion, AlgorithmVersion: "symbol-global-regime.v1"},
	})
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	return response
}

func floatPtr(value float64) *float64 {
	return &value
}
