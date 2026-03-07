package regimeengine

import (
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func TestRegimeClassification(t *testing.T) {
	service := newRegimeService(t)
	evaluation, err := service.Observe(regimeServiceBucket("BTC-USD", 0.99, 0.99, 0.92, features.FragmentationSeverityLow, 0.0, false, false, false))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if evaluation.Symbol.State != features.RegimeStateTradeable {
		t.Fatalf("symbol state = %q, want %q", evaluation.Symbol.State, features.RegimeStateTradeable)
	}
}

func TestFragmentedMarketDowngrade(t *testing.T) {
	service := newRegimeService(t)
	evaluation, err := service.Observe(regimeServiceBucket("ETH-USD", 0.78, 0.90, 0.60, features.FragmentationSeverityModerate, 0.12, false, false, false))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if evaluation.Symbol.State != features.RegimeStateWatch {
		t.Fatalf("symbol state = %q, want %q", evaluation.Symbol.State, features.RegimeStateWatch)
	}
}

func TestGlobalCeilingRules(t *testing.T) {
	service := newRegimeService(t)
	if _, err := service.Observe(regimeServiceBucket("BTC-USD", 0.45, 0.80, 0.30, features.FragmentationSeveritySevere, 0.55, false, false, false)); err != nil {
		t.Fatalf("observe btc: %v", err)
	}
	evaluation, err := service.Observe(regimeServiceBucket("ETH-USD", 0.42, 0.82, 0.30, features.FragmentationSeveritySevere, 0.55, false, false, false))
	if err != nil {
		t.Fatalf("observe eth: %v", err)
	}
	if evaluation.Global.State != features.RegimeStateNoOperate {
		t.Fatalf("global state = %q, want %q", evaluation.Global.State, features.RegimeStateNoOperate)
	}
	if evaluation.EffectiveState["BTC-USD"] != features.RegimeStateNoOperate || evaluation.EffectiveState["ETH-USD"] != features.RegimeStateNoOperate {
		t.Fatalf("effective states = %+v", evaluation.EffectiveState)
	}
}

func TestGlobalCeilingDoesNotHideSymbolSpecificDifferences(t *testing.T) {
	service := newRegimeService(t)
	if _, err := service.Observe(regimeServiceBucket("BTC-USD", 0.99, 0.99, 0.92, features.FragmentationSeverityLow, 0.0, false, false, false)); err != nil {
		t.Fatalf("observe btc: %v", err)
	}
	evaluation, err := service.Observe(regimeServiceBucket("ETH-USD", 0.78, 0.92, 0.60, features.FragmentationSeverityModerate, 0.12, false, false, false))
	if err != nil {
		t.Fatalf("observe eth: %v", err)
	}
	if evaluation.Global.State != features.RegimeStateTradeable {
		t.Fatalf("global state = %q, want %q", evaluation.Global.State, features.RegimeStateTradeable)
	}
	if evaluation.EffectiveState["BTC-USD"] != features.RegimeStateTradeable || evaluation.EffectiveState["ETH-USD"] != features.RegimeStateWatch {
		t.Fatalf("effective states = %+v", evaluation.EffectiveState)
	}
}

func TestWorldUSAMarketStateTransitionReasons(t *testing.T) {
	service := newRegimeService(t)
	evaluation, err := service.Observe(regimeServiceBucket("BTC-USD", 0.55, 0.90, 0.34, features.FragmentationSeveritySevere, 0.20, false, true, false))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if evaluation.Symbol.PrimaryReason != features.RegimeReasonLateWindowIncomplete {
		t.Fatalf("primary reason = %q, want %q", evaluation.Symbol.PrimaryReason, features.RegimeReasonLateWindowIncomplete)
	}
	if evaluation.Symbol.ConfigVersion == "" || evaluation.Symbol.AlgorithmVersion == "" {
		t.Fatalf("missing provenance: %+v", evaluation.Symbol)
	}
}

func newRegimeService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(features.RegimeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "regime-engine.market-state.v1",
		AlgorithmVersion: "symbol-global-regime.v1",
		Symbols:          []string{"BTC-USD", "ETH-USD"},
		Symbol: features.SymbolRegimeThresholds{
			CoverageWatchMax:             0.85,
			CoverageNoOperateMax:         0.60,
			CombinedTrustCapWatchMax:     0.65,
			CombinedTrustCapNoOperateMax: 0.35,
			TimestampFallbackWatchRatio:  0.10,
			TimestampFallbackNoOpRatio:   0.50,
			NoOperateToWatchWindows:      2,
			WatchToTradeableWindows:      2,
		},
		Global: features.GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
	})
	if err != nil {
		t.Fatalf("new regime service: %v", err)
	}
	return service
}

func regimeServiceBucket(symbol string, coverage float64, health float64, cap float64, fragmentation features.FragmentationSeverity, fallbackRatio float64, recvSource bool, missing bool, unavailable bool) features.MarketQualityBucket {
	bucket := features.MarketQualityBucket{
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
	if recvSource {
		bucket.Assignment.BucketSource = features.BucketSourceRecvTs
	}
	if missing {
		bucket.Window.MissingBucketCount = 1
		bucket.Window.ClosedBucketCount = 9
	}
	return bucket
}
