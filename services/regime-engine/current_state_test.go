package regimeengine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func TestMarketStateCurrentRegimeSection(t *testing.T) {
	response := currentGlobalFixture(t)
	if response.Symbols[0].EffectiveState == "" {
		t.Fatalf("missing effective state: %+v", response.Symbols[0])
	}
	if response.Symbols[0].GlobalState != response.Global.State {
		t.Fatalf("global state mismatch: %+v", response)
	}
}

func TestMarketStateCurrentGlobalSchema(t *testing.T) {
	contents, err := os.ReadFile("../../schemas/json/features/market-state-current-global.v1.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(contents, &parsed); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	response := currentGlobalFixture(t)
	if response.SchemaVersion != features.MarketStateCurrentGlobalSchema {
		t.Fatalf("schema version = %q", response.SchemaVersion)
	}
	if response.Version.ConfigVersion == "" || response.AsOf == "" {
		t.Fatalf("missing global version metadata: %+v", response)
	}
	if len(response.Symbols) != 2 {
		t.Fatalf("symbols = %d", len(response.Symbols))
	}
}

func TestGlobalCurrentStateQuery(t *testing.T) {
	response := currentGlobalFixture(t)
	if response.Global.State != features.RegimeStateWatch {
		t.Fatalf("global state = %q", response.Global.State)
	}
	if response.Symbols[0].Availability == "" || response.Symbols[1].Availability == "" {
		t.Fatalf("missing symbol summaries: %+v", response.Symbols)
	}
}

func TestGlobalCeilingAppliedToSymbolResponse(t *testing.T) {
	response := currentGlobalFixture(t)
	for _, symbol := range response.Symbols {
		if symbol.GlobalState != features.RegimeStateWatch {
			t.Fatalf("expected watch ceiling on symbol summary: %+v", symbol)
		}
		if symbol.EffectiveState != features.RegimeStateWatch {
			t.Fatalf("expected effective watch state: %+v", symbol)
		}
	}
}

func TestCurrentStateTransitionReasons(t *testing.T) {
	response := currentGlobalFixture(t)
	if len(response.Global.Reasons) == 0 {
		t.Fatalf("missing global reasons: %+v", response.Global)
	}
	if len(response.Symbols[0].ReasonCodes) == 0 {
		t.Fatalf("missing symbol reason codes: %+v", response.Symbols[0])
	}
}

func currentGlobalFixture(t *testing.T) features.MarketStateCurrentGlobalResponse {
	t.Helper()
	service := newRegimeService(t)
	query := features.GlobalCurrentStateQuery{
		GlobalRegime: features.GlobalRegimeSnapshot{SchemaVersion: "v1", State: features.RegimeStateWatch, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}, PrimaryReason: features.RegimeReasonGlobalSharedWatch, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
		Symbols: []features.MarketStateCurrentResponse{
			{Symbol: "BTC-USD", AsOf: "2026-03-06T12:05:00Z", Version: features.MarketStateCurrentVersion{ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"}, Regime: features.MarketStateCurrentRegimeSection{Availability: features.CurrentStateAvailabilityDegraded, EffectiveState: features.RegimeStateWatch, Symbol: features.SymbolRegimeSnapshot{State: features.RegimeStateTradeable, Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}}, Global: features.GlobalRegimeSnapshot{State: features.RegimeStateWatch, Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}}}, Provenance: features.MarketStateCurrentProvenance{BucketRefs: []features.MarketStateCurrentBucketRef{{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}}, SymbolBucketEnd: "2026-03-06T12:05:00Z", GlobalBucketEnd: "2026-03-06T12:05:00Z", HistorySeam: features.MarketStateHistoryAuditSeam{ReservedSchemaFamily: "market-state-history-and-audit-reads"}}},
			{Symbol: "ETH-USD", AsOf: "2026-03-06T12:05:00Z", Version: features.MarketStateCurrentVersion{ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"}, Regime: features.MarketStateCurrentRegimeSection{Availability: features.CurrentStateAvailabilityDegraded, EffectiveState: features.RegimeStateWatch, Symbol: features.SymbolRegimeSnapshot{State: features.RegimeStateWatch, Reasons: []features.RegimeReasonCode{features.RegimeReasonFragmentationModerate}}, Global: features.GlobalRegimeSnapshot{State: features.RegimeStateWatch, Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}}}, Provenance: features.MarketStateCurrentProvenance{BucketRefs: []features.MarketStateCurrentBucketRef{{Family: features.BucketFamily5m, BucketStart: "2026-03-06T12:00:00Z", BucketEnd: "2026-03-06T12:05:00Z", ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"}}, SymbolBucketEnd: "2026-03-06T12:05:00Z", GlobalBucketEnd: "2026-03-06T12:05:00Z", HistorySeam: features.MarketStateHistoryAuditSeam{ReservedSchemaFamily: "market-state-history-and-audit-reads"}}},
		},
	}
	response, err := service.QueryCurrentGlobalState(query)
	if err != nil {
		t.Fatalf("query current global state: %v", err)
	}
	return response
}
