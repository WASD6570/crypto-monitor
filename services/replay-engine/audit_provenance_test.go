package replayengine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func TestMarketStateAuditProvenanceAuthoritativeOriginal(t *testing.T) {
	contents, err := os.ReadFile("../../schemas/json/replay/market-state-audit-lineage.v1.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(contents, &parsed); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	audit, err := QueryMarketStateAuditProvenance(features.MarketStateAuditQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           "BTC-USD",
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             "2026-03-06T12:05:00Z",
			ConfigVersion:    "regime-engine.market-state.v1",
			AlgorithmVersion: "symbol-global-regime.v1",
			ReplayRunID:      "run-authoritative",
		},
		Provenance: features.MarketStateCurrentProvenance{
			CompositeBucketTs: []string{"2026-03-06T12:05:00Z"},
			BucketRefs: []features.MarketStateCurrentBucketRef{{
				Family:           features.BucketFamily5m,
				BucketStart:      "2026-03-06T12:00:00Z",
				BucketEnd:        "2026-03-06T12:05:00Z",
				ConfigVersion:    "regime-engine.market-state.v1",
				AlgorithmVersion: "symbol-global-regime.v1",
			}},
			SymbolBucketEnd: "2026-03-06T12:05:00Z",
			GlobalBucketEnd: "2026-03-06T12:05:00Z",
		},
		Available: true,
	})
	if err != nil {
		t.Fatalf("query audit provenance: %v", err)
	}
	if audit.Status != features.MarketStateAuditStatusAuthoritativeOriginal {
		t.Fatalf("status = %q", audit.Status)
	}
	if audit.AuthoritativeLineage.ReplayRunID != "run-authoritative" {
		t.Fatalf("replay run id = %q", audit.AuthoritativeLineage.ReplayRunID)
	}
}

func TestMarketStateAuditProvenanceReplayCorrected(t *testing.T) {
	audit, err := QueryMarketStateAuditProvenance(features.MarketStateAuditQuery{
		Lookup: features.MarketStateHistoryLookupQuery{
			Scope:            "symbol",
			Symbol:           "BTC-USD",
			BucketFamily:     features.BucketFamily5m,
			BucketEnd:        "2026-03-06T12:05:00Z",
			AsOf:             "2026-03-06T12:05:00Z",
			ConfigVersion:    "regime-engine.market-state.v1",
			AlgorithmVersion: "symbol-global-regime.v1",
			ReplayRunID:      "run-current",
		},
		Provenance: features.MarketStateCurrentProvenance{
			BucketRefs: []features.MarketStateCurrentBucketRef{{
				Family:           features.BucketFamily5m,
				BucketStart:      "2026-03-06T12:00:00Z",
				BucketEnd:        "2026-03-06T12:05:00Z",
				ConfigVersion:    "regime-engine.market-state.v1",
				AlgorithmVersion: "symbol-global-regime.v1",
			}},
		},
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
		Available: true,
	})
	if err != nil {
		t.Fatalf("query audit provenance: %v", err)
	}
	if audit.Status != features.MarketStateAuditStatusReplayCorrected {
		t.Fatalf("status = %q", audit.Status)
	}
	if audit.SupersededLineage == nil || audit.SupersededLineage.ReplayRunID != "run-live" {
		t.Fatalf("missing superseded lineage: %+v", audit)
	}
	if audit.CorrectionCause != "late_event_rebucket" {
		t.Fatalf("correction cause = %q", audit.CorrectionCause)
	}
}
