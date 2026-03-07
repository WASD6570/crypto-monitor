package features

import "testing"

func TestRegimeClassification(t *testing.T) {
	config := testRegimeConfig()
	tradeable, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.98, 0.97, 0.92, FragmentationSeverityLow, 0.0, false, false, false), nil)
	if err != nil {
		t.Fatalf("tradeable regime: %v", err)
	}
	if tradeable.State != RegimeStateTradeable || tradeable.PrimaryReason != RegimeReasonHealthy {
		t.Fatalf("tradeable snapshot = %+v", tradeable)
	}
	watch, err := EvaluateSymbolRegime(config, regimeBucket("ETH-USD", 0.82, 0.94, 0.62, FragmentationSeverityModerate, 0.15, false, false, false), nil)
	if err != nil {
		t.Fatalf("watch regime: %v", err)
	}
	if watch.State != RegimeStateWatch || watch.PrimaryReason != RegimeReasonFragmentationModerate {
		t.Fatalf("watch snapshot = %+v", watch)
	}
	noOperate, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.48, 0.75, 0.28, FragmentationSeveritySevere, 0.55, false, false, true), nil)
	if err != nil {
		t.Fatalf("no-operate regime: %v", err)
	}
	if noOperate.State != RegimeStateNoOperate || noOperate.PrimaryReason != RegimeReasonCompositeUnavailable {
		t.Fatalf("no-operate snapshot = %+v", noOperate)
	}
}

func TestRegimeDowngradePrecedence(t *testing.T) {
	config := testRegimeConfig()
	snapshot, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.55, 0.81, 0.30, FragmentationSeveritySevere, 0.40, false, true, false), nil)
	if err != nil {
		t.Fatalf("evaluate regime: %v", err)
	}
	if snapshot.Reasons[0] != RegimeReasonLateWindowIncomplete {
		t.Fatalf("primary reason = %q, want %q", snapshot.Reasons[0], RegimeReasonLateWindowIncomplete)
	}
	if snapshot.State != RegimeStateNoOperate {
		t.Fatalf("state = %q, want %q", snapshot.State, RegimeStateNoOperate)
	}
}

func TestRegimeRecoveryRequiresPersistence(t *testing.T) {
	config := testRegimeConfig()
	previous, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.40, 0.75, 0.28, FragmentationSeveritySevere, 0.60, false, false, false), nil)
	if err != nil {
		t.Fatalf("evaluate previous: %v", err)
	}
	first, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.98, 0.98, 0.92, FragmentationSeverityLow, 0.0, false, false, false), &previous)
	if err != nil {
		t.Fatalf("first recovery: %v", err)
	}
	if first.State != RegimeStateNoOperate || first.PrimaryReason != RegimeReasonRecoveryPersistencePending {
		t.Fatalf("first recovery snapshot = %+v", first)
	}
	second, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.98, 0.98, 0.92, FragmentationSeverityLow, 0.0, false, false, false), &first)
	if err != nil {
		t.Fatalf("second recovery: %v", err)
	}
	if second.State != RegimeStateWatch || second.TransitionKind != RegimeTransitionRecover {
		t.Fatalf("second recovery snapshot = %+v", second)
	}
	third, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.98, 0.98, 0.92, FragmentationSeverityLow, 0.0, false, false, false), &second)
	if err != nil {
		t.Fatalf("third recovery: %v", err)
	}
	if third.State != RegimeStateWatch || third.PrimaryReason != RegimeReasonRecoveryPersistencePending {
		t.Fatalf("third recovery snapshot = %+v", third)
	}
	fourth, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.98, 0.98, 0.92, FragmentationSeverityLow, 0.0, false, false, false), &third)
	if err != nil {
		t.Fatalf("fourth recovery: %v", err)
	}
	if fourth.State != RegimeStateTradeable || fourth.TransitionKind != RegimeTransitionRecover {
		t.Fatalf("fourth recovery snapshot = %+v", fourth)
	}
}

func TestRegimeThresholdEdgesAreDeterministic(t *testing.T) {
	config := testRegimeConfig()
	snapshot, err := EvaluateSymbolRegime(config, regimeBucket("ETH-USD", config.Symbol.CoverageWatchMax, 0.99, 0.90, FragmentationSeverityLow, config.Symbol.TimestampFallbackWatchRatio, false, false, false), nil)
	if err != nil {
		t.Fatalf("evaluate regime: %v", err)
	}
	if snapshot.State != RegimeStateWatch {
		t.Fatalf("state = %q, want %q", snapshot.State, RegimeStateWatch)
	}
}

func TestRegimeReasonsIncludeTriggerMetrics(t *testing.T) {
	config := testRegimeConfig()
	snapshot, err := EvaluateSymbolRegime(config, regimeBucket("BTC-USD", 0.79, 0.96, 0.61, FragmentationSeverityModerate, 0.12, false, false, false), nil)
	if err != nil {
		t.Fatalf("evaluate regime: %v", err)
	}
	if len(snapshot.TriggerMetrics) == 0 {
		t.Fatal("expected trigger metrics")
	}
	if snapshot.TriggerMetrics[0].Name == "" {
		t.Fatalf("invalid trigger metrics: %+v", snapshot.TriggerMetrics)
	}
}

func TestGlobalCeilingRules(t *testing.T) {
	config := testRegimeConfig()
	tradeable, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateTradeable, RegimeReasonHealthy),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateTradeable, RegimeReasonHealthy),
	}, nil)
	if err != nil {
		t.Fatalf("tradeable global: %v", err)
	}
	if tradeable.State != RegimeStateTradeable {
		t.Fatalf("tradeable global = %+v", tradeable)
	}
	watch, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
	}, nil)
	if err != nil {
		t.Fatalf("watch global: %v", err)
	}
	if watch.State != RegimeStateWatch || watch.PrimaryReason != RegimeReasonGlobalSharedWatch {
		t.Fatalf("watch global = %+v", watch)
	}
	noOperate, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateNoOperate, RegimeReasonTimestampTrustLoss),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateNoOperate, RegimeReasonTimestampTrustLoss),
	}, nil)
	if err != nil {
		t.Fatalf("no-operate global: %v", err)
	}
	if noOperate.State != RegimeStateNoOperate || noOperate.PrimaryReason != RegimeReasonGlobalSharedNoOperate {
		t.Fatalf("no-operate global = %+v", noOperate)
	}
}

func TestGlobalRecoveryRequiresPersistence(t *testing.T) {
	config := testRegimeConfig()
	previous, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateNoOperate, RegimeReasonTimestampTrustLoss),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateNoOperate, RegimeReasonTimestampTrustLoss),
	}, nil)
	if err != nil {
		t.Fatalf("initial global: %v", err)
	}
	first, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
	}, &previous)
	if err != nil {
		t.Fatalf("first global recovery: %v", err)
	}
	if first.State != RegimeStateNoOperate || first.PrimaryReason != RegimeReasonRecoveryPersistencePending {
		t.Fatalf("first global recovery = %+v", first)
	}
	second, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateWatch, RegimeReasonFragmentationModerate),
	}, &first)
	if err != nil {
		t.Fatalf("second global recovery: %v", err)
	}
	if second.State != RegimeStateWatch || second.TransitionKind != RegimeTransitionRecover {
		t.Fatalf("second global recovery = %+v", second)
	}
}

func TestGlobalTransitionReasonsAreDeterministic(t *testing.T) {
	config := testRegimeConfig()
	left, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateWatch, RegimeReasonCoverageLow, RegimeReasonTimestampTrustLoss),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateWatch, RegimeReasonTimestampTrustLoss, RegimeReasonCoverageLow),
	}, nil)
	if err != nil {
		t.Fatalf("left global: %v", err)
	}
	right, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": symbolSnapshot("BTC-USD", RegimeStateWatch, RegimeReasonTimestampTrustLoss, RegimeReasonCoverageLow),
		"ETH-USD": symbolSnapshot("ETH-USD", RegimeStateWatch, RegimeReasonCoverageLow, RegimeReasonTimestampTrustLoss),
	}, nil)
	if err != nil {
		t.Fatalf("right global: %v", err)
	}
	if len(left.Reasons) != len(right.Reasons) {
		t.Fatalf("reason lengths differ: %+v %+v", left.Reasons, right.Reasons)
	}
	for index := range left.Reasons {
		if left.Reasons[index] != right.Reasons[index] {
			t.Fatalf("reason ordering differs: %+v %+v", left.Reasons, right.Reasons)
		}
	}
}

func testRegimeConfig() RegimeConfig {
	return RegimeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "regime-engine.market-state.v1",
		AlgorithmVersion: "symbol-global-regime.v1",
		Symbols:          []string{"BTC-USD", "ETH-USD"},
		Symbol: SymbolRegimeThresholds{
			CoverageWatchMax:             0.85,
			CoverageNoOperateMax:         0.60,
			CombinedTrustCapWatchMax:     0.65,
			CombinedTrustCapNoOperateMax: 0.35,
			TimestampFallbackWatchRatio:  0.10,
			TimestampFallbackNoOpRatio:   0.50,
			NoOperateToWatchWindows:      2,
			WatchToTradeableWindows:      2,
		},
		Global: GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2},
	}

}

func regimeBucket(symbol string, coverage float64, health float64, cap float64, fragmentation FragmentationSeverity, fallbackRatio float64, recvSource bool, missing bool, unavailable bool) MarketQualityBucket {
	bucket := MarketQualityBucket{
		Symbol:         symbol,
		Window:         BucketWindowSummary{Family: BucketFamily5m, End: "2026-03-06T12:05:00Z", ClosedBucketCount: 10, ExpectedBucketCount: 10},
		Assignment:     BucketAssignment{BucketSource: BucketSourceExchangeTs},
		World:          CompositeBucketSide{Available: !unavailable, Unavailable: unavailable, CoverageRatio: coverage, HealthScore: health},
		USA:            CompositeBucketSide{Available: !unavailable, Unavailable: unavailable, CoverageRatio: coverage, HealthScore: health},
		Divergence:     DivergenceSummary{Available: !unavailable},
		Fragmentation:  FragmentationSummary{Severity: fragmentation, PersistenceCount: 2},
		TimestampTrust: TimestampTrustSummary{FallbackRatio: fallbackRatio, TrustCap: fallbackRatio >= 0.10},
		MarketQuality:  MarketQualitySummary{CombinedTrustCap: cap},
	}
	if recvSource {
		bucket.Assignment.BucketSource = BucketSourceRecvTs
	}
	if missing {
		bucket.Window.MissingBucketCount = 1
		bucket.Window.ClosedBucketCount = 9
	}
	return bucket
}

func symbolSnapshot(symbol string, state RegimeState, reasons ...RegimeReasonCode) SymbolRegimeSnapshot {
	return SymbolRegimeSnapshot{Symbol: symbol, State: state, Reasons: reasons, EffectiveBucketEnd: "2026-03-06T12:05:00Z"}
}
