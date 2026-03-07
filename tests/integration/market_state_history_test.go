package integration

import (
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestMarketStateHistoryClosedWindowLookup(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	current, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	history, err := featureService.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-history-integration",
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical state: %v", err)
	}
	if history.State == nil || history.State.Regime.EffectiveState != current.Regime.EffectiveState {
		t.Fatalf("historical state mismatch: %+v", history)
	}
	if history.Lookup.ResolutionStatus != features.MarketStateHistoryResolutionExact {
		t.Fatalf("resolution status = %q", history.Lookup.ResolutionStatus)
	}
}

func TestMarketStateHistoryVersionPinnedBucketContext(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	current, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	history, err := featureService.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    "regime-engine.market-state.v2",
			AlgorithmVersion: current.Version.AlgorithmVersion,
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical state: %v", err)
	}
	if history.Availability.Code != features.MarketStateHistoryAvailabilityPinMismatch {
		t.Fatalf("availability = %q", history.Availability.Code)
	}
}

func TestMarketStateHistoryAuditProvenanceConsistency(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	current, err := featureService.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	history, err := featureService.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-history-audit",
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical state: %v", err)
	}
	audit, err := replayengine.QueryMarketStateAuditProvenance(features.MarketStateAuditQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        current.Provenance.SymbolBucketEnd,
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-history-audit",
		},
		Provenance: current.Provenance,
		Available:  true,
	})
	if err != nil {
		t.Fatalf("query audit provenance: %v", err)
	}
	if len(history.Audit.AuthoritativeLineage.BucketRefs) != len(audit.AuthoritativeLineage.BucketRefs) {
		t.Fatalf("audit lineage drift: history=%d audit=%d", len(history.Audit.AuthoritativeLineage.BucketRefs), len(audit.AuthoritativeLineage.BucketRefs))
	}
	if history.Audit.Status != audit.Status {
		t.Fatalf("audit status drift: history=%q audit=%q", history.Audit.Status, audit.Status)
	}
}
