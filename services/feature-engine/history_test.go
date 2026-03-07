package featureengine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func TestMarketStateHistoryReadByBucketKey(t *testing.T) {
	for _, path := range []string{
		"../../schemas/json/features/market-state-history-symbol.v1.schema.json",
		"../../schemas/json/features/market-state-audit-provenance.v1.schema.json",
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read schema %s: %v", path, err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(contents, &parsed); err != nil {
			t.Fatalf("parse schema %s: %v", path, err)
		}
	}
	service := newBucketService(t)
	current := currentStateResponseFixture(t)
	history, err := service.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-history-v1",
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical symbol state: %v", err)
	}
	if history.SchemaVersion != features.MarketStateHistorySymbolSchema {
		t.Fatalf("schema version = %q", history.SchemaVersion)
	}
	if history.Lookup.ResolutionStatus != features.MarketStateHistoryResolutionExact {
		t.Fatalf("resolution status = %q", history.Lookup.ResolutionStatus)
	}
	if history.State == nil || history.State.Version.ConfigVersion != current.Version.ConfigVersion {
		t.Fatalf("missing historical state payload: %+v", history)
	}
	if history.Audit.Status != features.MarketStateAuditStatusAuthoritativeOriginal {
		t.Fatalf("audit status = %q", history.Audit.Status)
	}
}

func TestMarketStateHistoryUnavailableOnPinMismatch(t *testing.T) {
	service := newBucketService(t)
	current := currentStateResponseFixture(t)
	history, err := service.QueryHistoricalState(features.SymbolHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           current.Symbol,
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             current.AsOf,
			ConfigVersion:    "regime-engine.market-state.v2",
			AlgorithmVersion: current.Version.AlgorithmVersion,
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical symbol state: %v", err)
	}
	if history.Availability.Code != features.MarketStateHistoryAvailabilityPinMismatch {
		t.Fatalf("availability = %q", history.Availability.Code)
	}
	if history.Lookup.ResolutionStatus != features.MarketStateHistoryResolutionUnavailable {
		t.Fatalf("resolution status = %q", history.Lookup.ResolutionStatus)
	}
	if history.State != nil {
		t.Fatalf("expected unavailable historical state, got %+v", history.State)
	}
	if history.Audit.Status != features.MarketStateAuditStatusUnavailable {
		t.Fatalf("audit status = %q", history.Audit.Status)
	}
}
