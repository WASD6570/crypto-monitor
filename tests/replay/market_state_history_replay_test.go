package replay

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestMarketStateHistoryReplayLateEventCorrection(t *testing.T) {
	current := replayCurrentState(t, "regime-engine.market-state.v1")
	service, err := featureengine.NewService(features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}})
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	history, err := service.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-corrected",
		},
		Current: current,
		Correction: &features.MarketStateAuditCorrection{
			CorrectionCause:                "late_event_rebucket",
			ReasonCodes:                    []string{"late_event_rebucket"},
			AuthoritativeReplayRunID:       "run-corrected",
			AuthoritativeReplayManifestRef: "manifest:corrected",
			AuthoritativeArtifactIDs:       []string{"artifact:corrected"},
			SupersededReplayRunID:          "run-live",
			SupersededReplayManifestRef:    "manifest:live",
			SupersededArtifactIDs:          []string{"artifact:live"},
		},
	})
	if err != nil {
		t.Fatalf("query historical replay state: %v", err)
	}
	if history.Lookup.ResolutionStatus != features.MarketStateHistoryResolutionSuperseded {
		t.Fatalf("resolution status = %q", history.Lookup.ResolutionStatus)
	}
	if history.Audit.Status != features.MarketStateAuditStatusReplayCorrected {
		t.Fatalf("audit status = %q", history.Audit.Status)
	}
}

func TestMarketStateHistoryReplayConfigVersionPinnedLookup(t *testing.T) {
	current := replayCurrentState(t, "regime-engine.market-state.v2")
	service, err := featureengine.NewService(features.CompositeConfig{SchemaVersion: "v1", ConfigVersion: "composite-config.v1", AlgorithmVersion: "world-usa-composite.v1", Penalties: features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75}, QuoteProxies: map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}}, Groups: map[features.CompositeGroup]features.GroupConfig{features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}, features.CompositeGroupUSA: {Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}}}})
	if err != nil {
		t.Fatalf("new feature service: %v", err)
	}
	history, err := service.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    "regime-engine.market-state.v1",
			AlgorithmVersion: current.Version.AlgorithmVersion,
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical replay state: %v", err)
	}
	if history.Availability.Code != features.MarketStateHistoryAvailabilityPinMismatch {
		t.Fatalf("availability = %q", history.Availability.Code)
	}
}

func TestMarketStateHistoryReplayDeterministicAuditLineage(t *testing.T) {
	current := replayCurrentState(t, "regime-engine.market-state.v1")
	query := features.MarketStateAuditQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-history-deterministic",
		},
		Provenance: current.Provenance,
		Available:  true,
	}
	first, err := replayengine.QueryMarketStateAuditProvenance(query)
	if err != nil {
		t.Fatalf("first audit provenance: %v", err)
	}
	second, err := replayengine.QueryMarketStateAuditProvenance(query)
	if err != nil {
		t.Fatalf("second audit provenance: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("audit lineage differs\nfirst: %+v\nsecond: %+v", first, second)
	}
}
