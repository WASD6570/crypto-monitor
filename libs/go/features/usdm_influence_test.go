package features

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestUSDMInfluenceEvaluatorInputValidateRequiresTrackedSymbolOrder(t *testing.T) {
	input := validUSDMInfluenceInputFixture()
	if err := input.Validate(); err != nil {
		t.Fatalf("validate input: %v", err)
	}

	input.Symbols[0], input.Symbols[1] = input.Symbols[1], input.Symbols[0]
	if err := input.Validate(); err == nil {
		t.Fatal("expected out-of-order symbols to fail")
	}
}

func TestUSDMInfluenceSignalSetValidateRequiresVersionAndPrimaryReason(t *testing.T) {
	signals := USDMInfluenceSignalSet{
		SchemaVersion:    USDMInfluenceSignalSchema,
		ObservedAt:       "2026-03-15T12:00:00Z",
		ConfigVersion:    "feature-engine.binance-usdm-influence.v1",
		AlgorithmVersion: "binance-usdm-policy.v1",
		Signals: []USDMSymbolInfluenceSignal{
			{
				SchemaVersion:    USDMInfluenceSignalSchema,
				Symbol:           "BTC-USD",
				SourceSymbol:     "BTCUSDT",
				QuoteCurrency:    "USDT",
				Posture:          USDMInfluencePostureAuxiliary,
				Reasons:          []USDMInfluenceReasonCode{USDMInfluenceReasonHealthy},
				PrimaryReason:    USDMInfluenceReasonHealthy,
				ConfigVersion:    "feature-engine.binance-usdm-influence.v1",
				AlgorithmVersion: "binance-usdm-policy.v1",
				ObservedAt:       "2026-03-15T12:00:00Z",
			},
			{
				SchemaVersion:    USDMInfluenceSignalSchema,
				Symbol:           "ETH-USD",
				SourceSymbol:     "ETHUSDT",
				QuoteCurrency:    "USDT",
				Posture:          USDMInfluencePostureNoContext,
				Reasons:          []USDMInfluenceReasonCode{USDMInfluenceReasonNoContext},
				PrimaryReason:    USDMInfluenceReasonNoContext,
				ConfigVersion:    "feature-engine.binance-usdm-influence.v1",
				AlgorithmVersion: "binance-usdm-policy.v1",
				ObservedAt:       "2026-03-15T12:00:00Z",
			},
		},
	}
	if err := signals.Validate(); err != nil {
		t.Fatalf("validate signals: %v", err)
	}

	signals.Signals[1].PrimaryReason = USDMInfluenceReasonHealthy
	if err := signals.Validate(); err == nil {
		t.Fatal("expected primary reason mismatch to fail")
	}
}

func TestDefaultUSDMInfluenceConfigTracksSupportedSymbols(t *testing.T) {
	config := DefaultUSDMInfluenceConfig()
	if err := config.Validate(); err != nil {
		t.Fatalf("validate default config: %v", err)
	}
}

func TestApplyUSDMInfluenceToSymbolRegimePreservesAuxiliaryState(t *testing.T) {
	symbol := SymbolRegimeSnapshot{
		SchemaVersion:         "v1",
		Symbol:                "BTC-USD",
		State:                 RegimeStateTradeable,
		EffectiveBucketEnd:    "2026-03-15T12:05:00Z",
		Reasons:               []RegimeReasonCode{RegimeReasonHealthy},
		PrimaryReason:         RegimeReasonHealthy,
		TransitionKind:        RegimeTransitionHold,
		ConfigVersion:         "regime-engine.market-state.v1",
		AlgorithmVersion:      "symbol-global-regime.v1",
		ObservedInstantaneous: RegimeStateTradeable,
	}
	signal := &USDMSymbolInfluenceSignal{
		SchemaVersion:    USDMInfluenceSignalSchema,
		Symbol:           "BTC-USD",
		SourceSymbol:     "BTCUSDT",
		QuoteCurrency:    "USDT",
		Posture:          USDMInfluencePostureAuxiliary,
		Reasons:          []USDMInfluenceReasonCode{USDMInfluenceReasonHealthy},
		PrimaryReason:    USDMInfluenceReasonHealthy,
		ConfigVersion:    "feature-engine.binance-usdm-influence.v1",
		AlgorithmVersion: "binance-usdm-policy.v1",
		ObservedAt:       "2026-03-15T12:05:00Z",
	}

	adjusted, provenance, err := ApplyUSDMInfluenceToSymbolRegime(symbol, signal)
	if err != nil {
		t.Fatalf("apply usdm influence: %v", err)
	}
	if !reflect.DeepEqual(adjusted, symbol) {
		t.Fatalf("symbol regime changed for auxiliary posture\nwant: %+v\ngot: %+v", symbol, adjusted)
	}
	if provenance == nil || provenance.AppliedCap {
		t.Fatalf("expected provenance without cap, got %+v", provenance)
	}
	if provenance.Posture != USDMInfluencePostureAuxiliary {
		t.Fatalf("posture = %q", provenance.Posture)
	}
}

func TestApplyUSDMInfluenceToSymbolRegimeCapsTradeableToWatch(t *testing.T) {
	symbol := SymbolRegimeSnapshot{
		SchemaVersion:         "v1",
		Symbol:                "BTC-USD",
		State:                 RegimeStateTradeable,
		EffectiveBucketEnd:    "2026-03-15T12:05:00Z",
		Reasons:               []RegimeReasonCode{RegimeReasonHealthy},
		PrimaryReason:         RegimeReasonHealthy,
		TransitionKind:        RegimeTransitionHold,
		ConfigVersion:         "regime-engine.market-state.v1",
		AlgorithmVersion:      "symbol-global-regime.v1",
		ObservedInstantaneous: RegimeStateTradeable,
	}
	signal := &USDMSymbolInfluenceSignal{
		SchemaVersion:    USDMInfluenceSignalSchema,
		Symbol:           "BTC-USD",
		SourceSymbol:     "BTCUSDT",
		QuoteCurrency:    "USDT",
		Posture:          USDMInfluencePostureDegradeCap,
		Reasons:          []USDMInfluenceReasonCode{USDMInfluenceReasonBasisWide},
		PrimaryReason:    USDMInfluenceReasonBasisWide,
		ConfigVersion:    "feature-engine.binance-usdm-influence.v1",
		AlgorithmVersion: "binance-usdm-policy.v1",
		ObservedAt:       "2026-03-15T12:05:00Z",
	}

	adjusted, provenance, err := ApplyUSDMInfluenceToSymbolRegime(symbol, signal)
	if err != nil {
		t.Fatalf("apply usdm influence: %v", err)
	}
	if adjusted.State != RegimeStateWatch {
		t.Fatalf("state = %q, want %q", adjusted.State, RegimeStateWatch)
	}
	if adjusted.PrimaryReason != RegimeReasonUSDMInfluenceCap {
		t.Fatalf("primary reason = %q", adjusted.PrimaryReason)
	}
	if !reflect.DeepEqual(adjusted.Reasons, []RegimeReasonCode{RegimeReasonUSDMInfluenceCap, RegimeReasonHealthy}) {
		t.Fatalf("reasons = %+v", adjusted.Reasons)
	}
	if adjusted.TransitionKind != RegimeTransitionDowngrade {
		t.Fatalf("transition = %q", adjusted.TransitionKind)
	}
	if adjusted.ObservedInstantaneous != RegimeStateWatch {
		t.Fatalf("observed instantaneous = %q", adjusted.ObservedInstantaneous)
	}
	if provenance == nil || !provenance.AppliedCap {
		t.Fatalf("expected applied cap provenance, got %+v", provenance)
	}
	if provenance.PrimaryReason != USDMInfluenceReasonBasisWide {
		t.Fatalf("primary reason = %q", provenance.PrimaryReason)
	}
}

func TestEvaluateGlobalRegimeTreatsSharedUSDMCapAsWatch(t *testing.T) {
	config := RegimeConfig{
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
	global, err := EvaluateGlobalRegime(config, map[string]SymbolRegimeSnapshot{
		"BTC-USD": {Symbol: "BTC-USD", State: RegimeStateWatch, EffectiveBucketEnd: "2026-03-15T12:05:00Z", Reasons: []RegimeReasonCode{RegimeReasonUSDMInfluenceCap}, PrimaryReason: RegimeReasonUSDMInfluenceCap},
		"ETH-USD": {Symbol: "ETH-USD", State: RegimeStateWatch, EffectiveBucketEnd: "2026-03-15T12:05:00Z", Reasons: []RegimeReasonCode{RegimeReasonUSDMInfluenceCap}, PrimaryReason: RegimeReasonUSDMInfluenceCap},
	}, nil)
	if err != nil {
		t.Fatalf("evaluate global regime: %v", err)
	}
	if global.State != RegimeStateWatch {
		t.Fatalf("state = %q, want %q", global.State, RegimeStateWatch)
	}
	if !reflect.DeepEqual(global.Reasons, []RegimeReasonCode{RegimeReasonGlobalSharedWatch, RegimeReasonUSDMInfluenceCap}) {
		t.Fatalf("reasons = %+v", global.Reasons)
	}
}

func validUSDMInfluenceInputFixture() USDMInfluenceEvaluatorInput {
	return USDMInfluenceEvaluatorInput{
		SchemaVersion: USDMInfluenceInputSchema,
		ObservedAt:    "2026-03-15T12:00:00Z",
		Symbols: []USDMSymbolInfluenceInput{
			{
				Symbol:        "BTC-USD",
				SourceSymbol:  "BTCUSDT",
				QuoteCurrency: "USDT",
				Funding:       USDMFundingInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Available: true, Freshness: USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy}},
				MarkIndex:     USDMMarkIndexInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Available: true, Freshness: USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy}},
				Liquidation:   USDMLiquidationInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Freshness: USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  USDMOpenInterestInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy}},
			},
			{
				Symbol:        "ETH-USD",
				SourceSymbol:  "ETHUSDT",
				QuoteCurrency: "USDT",
				Funding:       USDMFundingInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Freshness: USDMInfluenceFreshnessUnavailable}},
				MarkIndex:     USDMMarkIndexInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Freshness: USDMInfluenceFreshnessUnavailable}},
				Liquidation:   USDMLiquidationInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceWebsocket, Freshness: USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  USDMOpenInterestInput{Metadata: USDMInfluenceInputMetadata{Surface: USDMInfluenceSurfaceRESTPoll, Freshness: USDMInfluenceFreshnessUnavailable}},
			},
		},
	}
}
