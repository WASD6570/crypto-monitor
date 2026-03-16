package replay

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestReplayBinanceUSDMInfluenceDeterminism(t *testing.T) {
	service, err := featureengine.NewService(features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}})
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	input := replayUSDMInfluenceInputFixture()
	first, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("first evaluate influence: %v", err)
	}
	second, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("second evaluate influence: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("replay influence outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestReplayBinanceUSDMInfluenceApplicationDeterminism(t *testing.T) {
	service, err := featureengine.NewService(features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}})
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	signals, err := service.EvaluateUSDMInfluence(replayUSDMInfluenceCapInputFixture())
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	run := func() features.MarketStateCurrentResponse {
		query := replayUSDMCurrentStateQueryFixture()
		query.SymbolRegime, query.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(query.SymbolRegime, replaySignalBySymbol(t, signals, "BTC-USD"))
		if err != nil {
			t.Fatalf("apply influence: %v", err)
		}
		response, err := service.QueryCurrentState(query)
		if err != nil {
			t.Fatalf("query current state: %v", err)
		}
		return response
	}
	first := run()
	second := run()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("replay current-state outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
	if first.Regime.Symbol.State != features.RegimeStateWatch {
		t.Fatalf("symbol state = %q, want %q", first.Regime.Symbol.State, features.RegimeStateWatch)
	}
	if first.Provenance.USDMInfluence == nil || !first.Provenance.USDMInfluence.AppliedCap {
		t.Fatalf("expected applied cap provenance, got %+v", first.Provenance.USDMInfluence)
	}
}

func replayUSDMInfluenceInputFixture() features.USDMInfluenceEvaluatorInput {
	return features.USDMInfluenceEvaluatorInput{
		SchemaVersion: features.USDMInfluenceInputSchema,
		ObservedAt:    "2026-03-15T12:00:00Z",
		Symbols: []features.USDMSymbolInfluenceInput{
			{
				Symbol:        "BTC-USD",
				SourceSymbol:  "BTCUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0003", NextFundingTs: "2026-03-15T16:00:00Z"},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64002", IndexPrice: "64000"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2500}, OpenInterest: "10659.509"},
			},
			{
				Symbol:        "ETH-USD",
				SourceSymbol:  "ETHUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, FundingRate: "0.0002", NextFundingTs: "2026-03-15T16:00:00Z"},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, MarkPrice: "3200.5", IndexPrice: "3200.2"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2600}, OpenInterest: "82450.5"},
			},
		},
	}
}

func replayUSDMInfluenceCapInputFixture() features.USDMInfluenceEvaluatorInput {
	input := replayUSDMInfluenceInputFixture()
	input.Symbols[0].Funding.FundingRate = "0.0009"
	input.Symbols[0].MarkIndex.MarkPrice = "64080"
	input.Symbols[0].Liquidation = features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, Side: "sell", Price: "64000", Size: "2"}
	return input
}

func replayUSDMCurrentStateQueryFixture() features.SymbolCurrentStateQuery {
	return features.SymbolCurrentStateQuery{
		Symbol: "BTC-USD",
		World:  features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupWorld, CompositePrice: replayFloatPtr(64000), CoverageRatio: 1, HealthScore: 0.99, ConfiguredContributorCount: 1, EligibleContributorCount: 1, ContributingContributorCount: 1, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"},
		USA:    features.CompositeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", BucketTs: "2026-03-06T12:05:00Z", CompositeGroup: features.CompositeGroupUSA, CompositePrice: replayFloatPtr(63990), CoverageRatio: 1, HealthScore: 0.99, ConfiguredContributorCount: 1, EligibleContributorCount: 1, ContributingContributorCount: 1, ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1"},
		Buckets: []features.MarketQualityBucket{
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily30s, Start: "2026-03-06T12:04:30Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 1, ExpectedBucketCount: 1, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily30s, BucketStart: "2026-03-06T12:04:30Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily2m, Start: "2026-03-06T12:03:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 4, ExpectedBucketCount: 4, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily2m, BucketStart: "2026-03-06T12:03:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
			{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily5m, Start: "2026-03-06T12:00:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}},
		},
		RecentContext: []features.MarketQualityBucket{{Symbol: "BTC-USD", Window: features.BucketWindowSummary{Family: features.BucketFamily5m, Start: "2026-03-06T12:00:00Z", End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}, Assignment: features.BucketAssignment{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", BucketSource: features.BucketSourceExchangeTs}, World: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, USA: features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99}, Divergence: features.DivergenceSummary{Available: true}, Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeverityLow, PersistenceCount: 1}, TimestampTrust: features.TimestampTrustSummary{}, MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 1}}},
		SymbolRegime:  features.SymbolRegimeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ObservedInstantaneous: features.RegimeStateTradeable, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
		GlobalRegime:  features.GlobalRegimeSnapshot{SchemaVersion: "v1", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ObservedInstantaneous: features.RegimeStateTradeable, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
	}
}

func replaySignalBySymbol(t *testing.T, signals features.USDMInfluenceSignalSet, symbol string) *features.USDMSymbolInfluenceSignal {
	t.Helper()
	for index := range signals.Signals {
		if signals.Signals[index].Symbol == symbol {
			copy := signals.Signals[index]
			return &copy
		}
	}
	t.Fatalf("missing signal for %s", symbol)
	return nil
}

func replayFloatPtr(value float64) *float64 {
	return &value
}
