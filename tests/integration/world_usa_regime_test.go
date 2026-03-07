package integration

import (
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	regimeengine "github.com/crypto-market-copilot/alerts/services/regime-engine"
)

func TestWorldUSAMarketStateTransitions(t *testing.T) {
	service := newRegimeIntegrationService(t)
	first, err := service.Observe(integrationRegimeBucket("BTC-USD", 0.45, 0.88, 0.30, features.FragmentationSeveritySevere, 0.55, false))
	if err != nil {
		t.Fatalf("observe first: %v", err)
	}
	if first.Symbol.State != features.RegimeStateNoOperate {
		t.Fatalf("first symbol state = %q", first.Symbol.State)
	}
	second, err := service.Observe(integrationRegimeBucket("BTC-USD", 0.98, 0.98, 0.92, features.FragmentationSeverityLow, 0.0, false))
	if err != nil {
		t.Fatalf("observe second: %v", err)
	}
	if second.Symbol.State != features.RegimeStateNoOperate {
		t.Fatalf("expected hysteresis hold, got %+v", second.Symbol)
	}
}

func TestWorldUSAGlobalCeilingTransitions(t *testing.T) {
	service := newRegimeIntegrationService(t)
	if _, err := service.Observe(integrationRegimeBucket("BTC-USD", 0.45, 0.88, 0.30, features.FragmentationSeveritySevere, 0.55, false)); err != nil {
		t.Fatalf("observe btc: %v", err)
	}
	evaluation, err := service.Observe(integrationRegimeBucket("ETH-USD", 0.45, 0.88, 0.30, features.FragmentationSeveritySevere, 0.55, false))
	if err != nil {
		t.Fatalf("observe eth: %v", err)
	}
	if evaluation.Global.State != features.RegimeStateNoOperate {
		t.Fatalf("global state = %q", evaluation.Global.State)
	}
	if len(evaluation.Global.AppliedCeilingToSymbols) == 0 {
		t.Fatalf("expected applied ceiling: %+v", evaluation.Global)
	}
}

func TestWorldUSASymbolSpecificDegradeWithoutGlobalStop(t *testing.T) {
	service := newRegimeIntegrationService(t)
	if _, err := service.Observe(integrationRegimeBucket("BTC-USD", 0.99, 0.99, 0.92, features.FragmentationSeverityLow, 0.0, false)); err != nil {
		t.Fatalf("observe btc: %v", err)
	}
	evaluation, err := service.Observe(integrationRegimeBucket("ETH-USD", 0.79, 0.94, 0.62, features.FragmentationSeverityModerate, 0.12, false))
	if err != nil {
		t.Fatalf("observe eth: %v", err)
	}
	if evaluation.Global.State != features.RegimeStateTradeable {
		t.Fatalf("global state = %q", evaluation.Global.State)
	}
	if evaluation.EffectiveState["BTC-USD"] != features.RegimeStateTradeable {
		t.Fatalf("btc effective state = %q", evaluation.EffectiveState["BTC-USD"])
	}
}

func newRegimeIntegrationService(t *testing.T) *regimeengine.Service {
	t.Helper()
	service, err := regimeengine.NewService(features.RegimeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "regime-engine.market-state.v1",
		AlgorithmVersion: "symbol-global-regime.v1",
		Symbols:          []string{"BTC-USD", "ETH-USD"},
		Symbol:           features.SymbolRegimeThresholds{CoverageWatchMax: 0.85, CoverageNoOperateMax: 0.60, CombinedTrustCapWatchMax: 0.65, CombinedTrustCapNoOperateMax: 0.35, TimestampFallbackWatchRatio: 0.10, TimestampFallbackNoOpRatio: 0.50, NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
		Global:           features.GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
	})
	if err != nil {
		t.Fatalf("new regime integration service: %v", err)
	}
	return service
}

func integrationRegimeBucket(symbol string, coverage float64, health float64, cap float64, fragmentation features.FragmentationSeverity, fallbackRatio float64, unavailable bool) features.MarketQualityBucket {
	return features.MarketQualityBucket{
		Symbol:         symbol,
		Window:         features.BucketWindowSummary{Family: features.BucketFamily5m, End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10},
		Assignment:     features.BucketAssignment{BucketSource: features.BucketSourceExchangeTs},
		World:          features.CompositeBucketSide{Available: !unavailable, Unavailable: unavailable, CoverageRatio: coverage, HealthScore: health},
		USA:            features.CompositeBucketSide{Available: !unavailable, Unavailable: unavailable, CoverageRatio: coverage, HealthScore: health},
		Divergence:     features.DivergenceSummary{Available: !unavailable},
		Fragmentation:  features.FragmentationSummary{Severity: fragmentation, PersistenceCount: 2},
		TimestampTrust: features.TimestampTrustSummary{FallbackRatio: fallbackRatio, TrustCap: fallbackRatio >= 0.10},
		MarketQuality:  features.MarketQualitySummary{CombinedTrustCap: cap},
	}
}
