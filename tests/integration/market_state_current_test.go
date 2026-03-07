package integration

import (
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	regimeengine "github.com/crypto-market-copilot/alerts/services/regime-engine"
)

func TestMarketStateCurrentSymbolQuery(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	response, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current symbol state: %v", err)
	}
	if response.Composite.Availability == "" || response.Regime.EffectiveState == "" {
		t.Fatalf("incomplete current symbol response: %+v", response)
	}
	if response.Regime.EffectiveState != features.RegimeStateWatch {
		t.Fatalf("effective state = %q", response.Regime.EffectiveState)
	}
}

func TestMarketStateCurrentRecentContextOrdering(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	query := currentStateQueryFixture()
	query.RecentContext = []features.MarketQualityBucket{
		integrationCurrentBucket(features.BucketFamily30s, "2026-03-06T12:03:30Z", "2026-03-06T12:04:00Z", 1, 0),
		integrationCurrentBucket(features.BucketFamily30s, "2026-03-06T12:04:30Z", "2026-03-06T12:05:00Z", 1, 0),
	}
	response, err := featureService.QueryCurrentState(query)
	if err != nil {
		t.Fatalf("query current symbol state: %v", err)
	}
	if len(response.RecentContext.ThirtySeconds.Buckets) != 1 {
		t.Fatalf("expected bounded recent context: %+v", response.RecentContext.ThirtySeconds)
	}
	if response.RecentContext.ThirtySeconds.Buckets[0].Window.End != "2026-03-06T12:05:00Z" {
		t.Fatalf("latest recent context not selected: %+v", response.RecentContext.ThirtySeconds)
	}
}

func TestMarketStateCurrentGlobalQuery(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	regimeService := newCurrentStateRegimeService(t)
	btc, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query btc current state: %v", err)
	}
	ethQuery := currentStateQueryFixture()
	ethQuery.Symbol = "ETH-USD"
	ethQuery.SymbolRegime.Symbol = "ETH-USD"
	ethQuery.SymbolRegime.State = features.RegimeStateWatch
	ethQuery.SymbolRegime.Reasons = []features.RegimeReasonCode{features.RegimeReasonFragmentationModerate}
	eth, err := featureService.QueryCurrentState(ethQuery)
	if err != nil {
		t.Fatalf("query eth current state: %v", err)
	}
	global, err := regimeService.QueryCurrentGlobalState(features.GlobalCurrentStateQuery{GlobalRegime: currentStateQueryFixture().GlobalRegime, Symbols: []features.MarketStateCurrentResponse{btc, eth}})
	if err != nil {
		t.Fatalf("query current global state: %v", err)
	}
	if global.Global.State != features.RegimeStateWatch {
		t.Fatalf("global state = %q", global.Global.State)
	}
	if len(global.Symbols) != 2 {
		t.Fatalf("global symbol summary size = %d", len(global.Symbols))
	}
}

func TestMarketStateCurrentConsumerContractSeam(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	response, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current symbol state: %v", err)
	}
	if response.Provenance.HistorySeam.ReservedSchemaFamily == "" {
		t.Fatalf("missing history seam: %+v", response.Provenance)
	}
	if len(response.Provenance.HistorySeam.BucketRefs) == 0 {
		t.Fatalf("missing bucket refs for consumers: %+v", response.Provenance)
	}
	if response.RecentContext.FiveMinutes.Buckets[0].Window.Family != features.BucketFamily5m {
		t.Fatalf("consumer recent context seam missing 5m bucket: %+v", response.RecentContext)
	}
}

func TestMarketStateCurrentConfigVersionContext(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	query := currentStateQueryFixture()
	query.SymbolRegime.ConfigVersion = "regime-engine.market-state.v2"
	response, err := featureService.QueryCurrentState(query)
	if err != nil {
		t.Fatalf("query current symbol state: %v", err)
	}
	if response.Version.ConfigVersion != "regime-engine.market-state.v2" {
		t.Fatalf("config version = %q", response.Version.ConfigVersion)
	}
}

func newCurrentStateFeatureService(t *testing.T) *featureengine.Service {
	t.Helper()
	service, err := featureengine.NewService(testCompositeConfig())
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	return service
}

func newCurrentStateRegimeService(t *testing.T) *regimeengine.Service {
	t.Helper()
	service, err := regimeengine.NewService(features.RegimeConfig{SchemaVersion: "v1", ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1", Symbols: []string{"BTC-USD", "ETH-USD"}, Symbol: features.SymbolRegimeThresholds{CoverageWatchMax: 0.85, CoverageNoOperateMax: 0.60, CombinedTrustCapWatchMax: 0.65, CombinedTrustCapNoOperateMax: 0.35, TimestampFallbackWatchRatio: 0.10, TimestampFallbackNoOpRatio: 0.50, NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2}, Global: features.GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2}})
	if err != nil {
		t.Fatalf("new regime service: %v", err)
	}
	return service
}

func currentStateQueryFixture() features.SymbolCurrentStateQuery {
	world := features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupWorld, CompositePrice: floatPtr(64000), CoverageRatio: 1, HealthScore: 0.99, ConfiguredContributorCount: 2, EligibleContributorCount: 2, ContributingContributorCount: 2, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"}
	usa := features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupUSA, CompositePrice: floatPtr(63990), CoverageRatio: 0.8, HealthScore: 0.82, ConfiguredContributorCount: 2, EligibleContributorCount: 2, ContributingContributorCount: 2, Degraded: true, DegradedReasons: []features.ReasonCode{features.ReasonTimestampFallback}, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"}
	buckets := []features.MarketQualityBucket{
		integrationCurrentBucket(features.BucketFamily30s, "2026-03-06T12:04:30Z", "2026-03-06T12:05:00Z", 1, 0),
		integrationCurrentBucket(features.BucketFamily2m, "2026-03-06T12:03:00Z", "2026-03-06T12:05:00Z", 4, 0),
		integrationCurrentBucket(features.BucketFamily5m, "2026-03-06T12:00:00Z", "2026-03-06T12:05:00Z", 10, 0),
	}
	return features.SymbolCurrentStateQuery{Symbol: "BTC-USD", World: world, USA: usa, Buckets: buckets, RecentContext: buckets, SymbolRegime: features.SymbolRegimeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"}, GlobalRegime: features.GlobalRegimeSnapshot{SchemaVersion: "v1", State: features.RegimeStateWatch, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}, PrimaryReason: features.RegimeReasonGlobalSharedWatch, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"}}
}

func integrationCurrentBucket(family features.BucketFamily, start string, end string, closed int, missing int) features.MarketQualityBucket {
	return features.MarketQualityBucket{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: family, Start: start, End: end, ClosedBucketCount: closed, ExpectedBucketCount: closed + missing, MissingBucketCount: missing, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: family, BucketStart: start, BucketEnd: end, BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 0.8, HealthScore: 0.82}, Divergence: features.DivergenceSummary{Available: true, ReasonCodes: []features.ReasonCode{features.ReasonTimestampTrustLoss}}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityModerate, PersistenceCount: 1, PrimaryCauses: []features.ReasonCode{features.ReasonTimestampTrustLoss}}, TimestampTrust: features.TimestampTrustSummary{RecvFallbackCount: 1, FallbackRatio: 0.10, TrustCap: family != features.BucketFamily30s}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 0.78, DowngradedReasons: []features.ReasonCode{features.ReasonTimestampTrustLoss}}}
}

func testCompositeConfig() features.CompositeConfig {
	return features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}}
}

func floatPtr(value float64) *float64 {
	return &value
}
