package replay

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	regimeengine "github.com/crypto-market-copilot/alerts/services/regime-engine"
)

func TestWorldUSAReplayDeterminism(t *testing.T) {
	first := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), false)
	second := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), false)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("replay outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestWorldUSARegimeReplayDeterminism(t *testing.T) {
	first := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), true)
	second := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), true)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("regime replay outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestWorldUSARegimeConfigVersionPinning(t *testing.T) {
	baseline := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), false)
	updated := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v2"), false)
	if reflect.DeepEqual(baseline, updated) {
		t.Fatal("expected config-version change to alter replay outputs")
	}
	if baseline[len(baseline)-1].Symbol.ConfigVersion != "regime-engine.market-state.v1" {
		t.Fatalf("baseline config version = %q", baseline[len(baseline)-1].Symbol.ConfigVersion)
	}
	if updated[len(updated)-1].Symbol.ConfigVersion != "regime-engine.market-state.v2" {
		t.Fatalf("updated config version = %q", updated[len(updated)-1].Symbol.ConfigVersion)
	}
}

func TestWorldUSALateEventReplayCorrectionDoesNotChangeLiveTransitionRules(t *testing.T) {
	first := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), true)
	second := replayRegimeSequence(t, replayRegimeService(t, "regime-engine.market-state.v1"), true)
	if first[1].Symbol.State != second[1].Symbol.State || first[1].Symbol.PrimaryReason != second[1].Symbol.PrimaryReason {
		t.Fatalf("late replay correction changed transition rules\nfirst: %+v\nsecond: %+v", first[1].Symbol, second[1].Symbol)
	}
}

func replayRegimeService(t *testing.T, configVersion string) *regimeengine.Service {
	t.Helper()
	noOperateCap := 0.35
	if configVersion == "regime-engine.market-state.v2" {
		noOperateCap = 0.40
	}
	service, err := regimeengine.NewService(features.RegimeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    configVersion,
		AlgorithmVersion: "symbol-global-regime.v1",
		Symbols:          []string{"BTC-USD", "ETH-USD"},
		Symbol:           features.SymbolRegimeThresholds{CoverageWatchMax: 0.85, CoverageNoOperateMax: 0.60, CombinedTrustCapWatchMax: 0.65, CombinedTrustCapNoOperateMax: noOperateCap, TimestampFallbackWatchRatio: 0.10, TimestampFallbackNoOpRatio: 0.50, NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
		Global:           features.GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
	})
	if err != nil {
		t.Fatalf("new replay regime service: %v", err)
	}
	return service
}

func replayRegimeSequence(t *testing.T, service *regimeengine.Service, late bool) []features.RegimeEvaluation {
	t.Helper()
	sequence := []features.MarketQualityBucket{
		replayRegimeBucket("BTC-USD", 0.45, 0.88, 0.30, features.FragmentationSeveritySevere, 0.55, false, false),
		replayRegimeBucket("BTC-USD", 0.98, 0.98, 0.92, features.FragmentationSeverityLow, 0.0, late, false),
		replayRegimeBucket("ETH-USD", 0.45, 0.88, 0.30, features.FragmentationSeveritySevere, 0.55, false, false),
		replayRegimeBucket("ETH-USD", 0.98, 0.98, 0.92, features.FragmentationSeverityLow, 0.0, false, false),
	}
	result := make([]features.RegimeEvaluation, 0, len(sequence))
	for _, bucket := range sequence {
		evaluation, err := service.Observe(bucket)
		if err != nil {
			t.Fatalf("observe replay regime: %v", err)
		}
		result = append(result, evaluation)
	}
	return result
}

func replayRegimeBucket(symbol string, coverage float64, health float64, cap float64, fragmentation features.FragmentationSeverity, fallbackRatio float64, missing bool, unavailable bool) features.MarketQualityBucket {
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
	if missing {
		bucket.Window.MissingBucketCount = 1
		bucket.Window.ClosedBucketCount = 9
	}
	return bucket
}
