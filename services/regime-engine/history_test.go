package regimeengine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func TestMarketStateHistoryGlobalLookup(t *testing.T) {
	contents, err := os.ReadFile("../../schemas/json/features/market-state-history-global.v1.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(contents, &parsed); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	service := newRegimeService(t)
	current := currentGlobalFixture(t)
	history, err := service.QueryHistoricalGlobalState(features.GlobalHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "global",
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: current.Version.AlgorithmVersion,
			ReplayRunID:      "run-global-history-v1",
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical global state: %v", err)
	}
	if history.Lookup.ResolutionStatus != features.MarketStateHistoryResolutionExact {
		t.Fatalf("resolution status = %q", history.Lookup.ResolutionStatus)
	}
	if history.State == nil || len(history.State.Symbols) != 2 {
		t.Fatalf("missing global historical state: %+v", history)
	}
	if history.Audit.Status != features.MarketStateAuditStatusAuthoritativeOriginal {
		t.Fatalf("audit status = %q", history.Audit.Status)
	}
}

func TestMarketStateHistoryVersionPinnedContext(t *testing.T) {
	service := newRegimeService(t)
	current := currentGlobalFixture(t)
	history, err := service.QueryHistoricalGlobalState(features.GlobalHistoricalStateQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "global",
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             current.AsOf,
			ConfigVersion:    current.Version.ConfigVersion,
			AlgorithmVersion: "symbol-global-regime.v2",
		},
		Current: current,
	})
	if err != nil {
		t.Fatalf("query historical global state: %v", err)
	}
	if history.Availability.Code != features.MarketStateHistoryAvailabilityPinMismatch {
		t.Fatalf("availability = %q", history.Availability.Code)
	}
	if history.State != nil {
		t.Fatalf("expected unavailable global history state, got %+v", history.State)
	}
}
