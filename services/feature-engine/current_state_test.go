package featureengine

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	slowcontext "github.com/crypto-market-copilot/alerts/services/slow-context"
)

func TestMarketStateCurrentSchema(t *testing.T) {
	for _, path := range []string{
		"../../schemas/json/features/market-state-current-symbol.v1.schema.json",
		"../../schemas/json/features/market-state-current-response.v1.schema.json",
		"../../schemas/json/features/market-state-recent-context.v1.schema.json",
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
}

func TestMarketStateCurrentResponseShape(t *testing.T) {
	response := currentStateResponseFixture(t)
	if response.SchemaVersion != features.MarketStateCurrentResponseSchema {
		t.Fatalf("schema version = %q", response.SchemaVersion)
	}
	if response.Symbol == "" || response.AsOf == "" || response.Version.ConfigVersion == "" || response.Version.AlgorithmVersion == "" {
		t.Fatalf("missing current-state identity: %+v", response)
	}
	if response.Composite.Availability == "" || response.Buckets.FiveMinutes.Availability == "" || response.Regime.EffectiveState == "" {
		t.Fatalf("missing assembled sections: %+v", response)
	}
	if response.RecentContext.SchemaVersion != features.MarketStateRecentContextSchema {
		t.Fatalf("recent context schema = %q", response.RecentContext.SchemaVersion)
	}
	if response.Provenance.HistorySeam.ReservedSchemaFamily != "market-state-history-and-audit-reads" {
		t.Fatalf("history seam = %+v", response.Provenance.HistorySeam)
	}
}

func TestCompositeSnapshotSchemaCompatibility(t *testing.T) {
	response := currentStateResponseFixture(t)
	if response.Composite.World.SchemaVersion != "v1" || response.Composite.USA.SchemaVersion != "v1" {
		t.Fatalf("expected upstream composite schema versions, got %+v", response.Composite)
	}
	if response.Composite.World.ConfigVersion == "" || response.Composite.USA.AlgorithmVersion == "" {
		t.Fatalf("missing upstream composite version metadata: %+v", response.Composite)
	}
}

func TestSymbolCurrentStateQuery(t *testing.T) {
	response := currentStateResponseFixture(t)
	if response.Regime.EffectiveState != features.RegimeStateWatch {
		t.Fatalf("effective state = %q", response.Regime.EffectiveState)
	}
	if response.Buckets.ThirtySeconds.Bucket.Window.Family != features.BucketFamily30s || response.Buckets.FiveMinutes.Bucket.Window.Family != features.BucketFamily5m {
		t.Fatalf("bucket families not assembled: %+v", response.Buckets)
	}
	if len(response.Provenance.BucketRefs) < 3 {
		t.Fatalf("expected bucket refs for current state provenance: %+v", response.Provenance)
	}
}

func TestCurrentStateRecentContext(t *testing.T) {
	response := currentStateResponseFixture(t)
	if len(response.RecentContext.ThirtySeconds.Buckets) != 1 || len(response.RecentContext.TwoMinutes.Buckets) != 1 || len(response.RecentContext.FiveMinutes.Buckets) != 1 {
		t.Fatalf("expected bounded recent context, got %+v", response.RecentContext)
	}
	if !response.RecentContext.FiveMinutes.Complete {
		t.Fatalf("expected complete recent 5m context: %+v", response.RecentContext.FiveMinutes)
	}
}

func TestCurrentStateUnavailableSections(t *testing.T) {
	service := newBucketService(t)
	response, err := service.QueryCurrentState(features.SymbolCurrentStateQuery{
		Symbol: "ETH-USD",
		World:  testBucketSnapshot(features.CompositeGroupWorld, 0, 0, 0, 0, 0, true, ""),
		USA:    testBucketSnapshot(features.CompositeGroupUSA, 3500, 1, 0.99, 0.7, 0, false, "coinbase"),
		Buckets: []features.MarketQualityBucket{{
			Symbol:        "ETH-USD",
			Window:        features.BucketWindowSummary{Family: features.BucketFamily30s, Start: "2026-03-06T12:00:00Z", End: "2026-03-06T12:00:30Z", ClosedBucketCount: 0, ExpectedBucketCount: 1, MissingBucketCount: 1, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"},
			World:         features.CompositeBucketSide{Unavailable: true},
			USA:           features.CompositeBucketSide{Available: true},
			Fragmentation: features.FragmentationSummary{Severity: features.FragmentationSeveritySevere, PrimaryCauses: []features.ReasonCode{features.ReasonCompositeUnavailable}},
			MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 0, DowngradedReasons: []features.ReasonCode{features.ReasonCompositeUnavailable}},
		}},
		SymbolRegime: features.SymbolRegimeSnapshot{State: features.RegimeStateNoOperate, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonCompositeUnavailable}, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
		GlobalRegime: features.GlobalRegimeSnapshot{State: features.RegimeStateWatch, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
	})
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	if response.Composite.Availability != features.CurrentStateAvailabilityPartial {
		t.Fatalf("composite availability = %q", response.Composite.Availability)
	}
	if response.Buckets.ThirtySeconds.Availability != features.CurrentStateAvailabilityUnavailable {
		t.Fatalf("30s availability = %q", response.Buckets.ThirtySeconds.Availability)
	}
}

func TestCurrentStateSucceedsWhenSlowContextFails(t *testing.T) {
	service := newBucketServiceWithSlowContext(t, failingSlowContextReader{err: errors.New("slow-context store offline")})
	baseline := currentStateResponseFixture(t)

	response, err := service.QueryCurrentStateWithSlowContext(currentStateQueryFixture(), slowcontext.AssetQuery{
		Asset: "BTC",
		Now:   mustTimeRFC3339(t, "2026-03-10T21:05:00Z"),
	})
	if err != nil {
		t.Fatalf("query current state with slow context: %v", err)
	}
	if response.CurrentState.Symbol != baseline.Symbol || response.CurrentState.AsOf != baseline.AsOf {
		t.Fatalf("current state identity changed: %+v vs %+v", response.CurrentState, baseline)
	}
	if response.CurrentState.Regime.EffectiveState != baseline.Regime.EffectiveState {
		t.Fatalf("effective state changed: %q vs %q", response.CurrentState.Regime.EffectiveState, baseline.Regime.EffectiveState)
	}
	context, ok := response.SlowContext.Context(slowcontext.MetricFamilyCMEVolume)
	if !ok {
		t.Fatal("expected slow context for cme volume")
	}
	if context.Availability != slowcontext.AvailabilityUnavailable {
		t.Fatalf("availability = %q, want %q", context.Availability, slowcontext.AvailabilityUnavailable)
	}
	if context.Error != "slow-context store offline" {
		t.Fatalf("error = %q, want %q", context.Error, "slow-context store offline")
	}
	if response.CurrentState.Buckets.FiveMinutes.Availability != baseline.Buckets.FiveMinutes.Availability {
		t.Fatalf("bucket availability changed: %q vs %q", response.CurrentState.Buckets.FiveMinutes.Availability, baseline.Buckets.FiveMinutes.Availability)
	}
}

func TestSlowContextResponseExplicitlyUnavailable(t *testing.T) {
	slowService, err := slowcontext.NewService()
	if err != nil {
		t.Fatalf("new slow context service: %v", err)
	}
	service := newBucketServiceWithSlowContext(t, slowService)

	response, err := service.QueryCurrentStateWithSlowContext(currentStateQueryFixture(), slowcontext.AssetQuery{
		Asset: "ETH",
		Now:   mustTimeRFC3339(t, "2026-03-10T12:00:00Z"),
	})
	if err != nil {
		t.Fatalf("query current state with slow context: %v", err)
	}
	context, ok := response.SlowContext.Context(slowcontext.MetricFamilyCMEOpenInterest)
	if !ok {
		t.Fatal("expected cme open interest context")
	}
	if context.Availability != slowcontext.AvailabilityUnavailable {
		t.Fatalf("availability = %q, want %q", context.Availability, slowcontext.AvailabilityUnavailable)
	}
	if context.Freshness != slowcontext.FreshnessUnavailable {
		t.Fatalf("freshness = %q, want %q", context.Freshness, slowcontext.FreshnessUnavailable)
	}
	if context.MessageKey != "cme_open_interest_unavailable" {
		t.Fatalf("message key = %q", context.MessageKey)
	}
	if response.CurrentState.Symbol != "BTC-USD" {
		t.Fatalf("current state symbol = %q", response.CurrentState.Symbol)
	}
}

func currentStateResponseFixture(t *testing.T) features.MarketStateCurrentResponse {
	t.Helper()
	service := newBucketService(t)
	response, err := service.QueryCurrentState(currentStateQueryFixture())
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	return response
}

func currentStateQueryFixture() features.SymbolCurrentStateQuery {
	world := testBucketSnapshot(features.CompositeGroupWorld, 64000, 1, 0.99, 0.6, 0, false, "binance")
	world.SchemaVersion = "v1"
	world.BucketTs = "2026-03-06T12:05:00Z"
	world.ConfigVersion = "composite-config.v1"
	world.AlgorithmVersion = "world-usa-composite.v1"
	usa := testBucketSnapshot(features.CompositeGroupUSA, 63990, 0.8, 0.82, 0.7, 1, false, "coinbase")
	usa.SchemaVersion = "v1"
	usa.BucketTs = "2026-03-06T12:05:00Z"
	usa.Degraded = true
	usa.DegradedReasons = []features.ReasonCode{features.ReasonTimestampFallback}
	usa.ConfigVersion = "composite-config.v1"
	usa.AlgorithmVersion = "world-usa-composite.v1"
	buckets := []features.MarketQualityBucket{
		currentBucketFixture(features.BucketFamily30s, "2026-03-06T12:04:30Z", "2026-03-06T12:05:00Z", 1, 0, 0.92),
		currentBucketFixture(features.BucketFamily2m, "2026-03-06T12:03:00Z", "2026-03-06T12:05:00Z", 4, 0, 0.78),
		currentBucketFixture(features.BucketFamily5m, "2026-03-06T12:00:00Z", "2026-03-06T12:05:00Z", 10, 0, 0.78),
	}
	return features.SymbolCurrentStateQuery{
		Symbol:        "BTC-USD",
		World:         world,
		USA:           usa,
		Buckets:       buckets,
		RecentContext: buckets,
		SymbolRegime:  features.SymbolRegimeSnapshot{SchemaVersion: "v1", Symbol: "BTC-USD", State: features.RegimeStateTradeable, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonHealthy}, PrimaryReason: features.RegimeReasonHealthy, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
		GlobalRegime:  features.GlobalRegimeSnapshot{SchemaVersion: "v1", State: features.RegimeStateWatch, EffectiveBucketEnd: "2026-03-06T12:05:00Z", Reasons: []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch}, PrimaryReason: features.RegimeReasonGlobalSharedWatch, ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1"},
	}
}

func currentBucketFixture(family features.BucketFamily, start string, end string, closed int, missing int, cap float64) features.MarketQualityBucket {
	return features.MarketQualityBucket{
		Symbol:         "BTC-USD",
		Window:         features.BucketWindowSummary{Family: family, Start: start, End: end, ClosedBucketCount: closed, ExpectedBucketCount: closed + missing, MissingBucketCount: missing, ConfigVersion: "market-quality.v1", AlgorithmVersion: "market-quality-buckets.v1"},
		Assignment:     features.BucketAssignment{Family: family, BucketStart: start, BucketEnd: end, BucketSource: features.BucketSourceExchangeTs},
		World:          features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99},
		USA:            features.CompositeBucketSide{Available: true, CoverageRatio: 0.8, HealthScore: 0.82, TimestampFallbackContributorCount: 1},
		Divergence:     features.DivergenceSummary{Available: true, ReasonCodes: []features.ReasonCode{features.ReasonTimestampTrustLoss}},
		Fragmentation:  features.FragmentationSummary{Severity: features.FragmentationSeverityModerate, PersistenceCount: 1, PrimaryCauses: []features.ReasonCode{features.ReasonTimestampTrustLoss}},
		TimestampTrust: features.TimestampTrustSummary{RecvFallbackCount: 1, FallbackRatio: 0.10, TrustCap: family != features.BucketFamily30s},
		MarketQuality:  features.MarketQualitySummary{CombinedTrustCap: cap, DowngradedReasons: []features.ReasonCode{features.ReasonTimestampTrustLoss}, ReplayProvenance: "market-quality.v1"},
	}
}

var _ = ingestion.VenueBinance
var _ = time.RFC3339

type failingSlowContextReader struct {
	err error
}

func (f failingSlowContextReader) QueryAsset(query slowcontext.AssetQuery) (slowcontext.AssetContextResponse, error) {
	return slowcontext.AssetContextResponse{}, f.err
}

func newBucketServiceWithSlowContext(t *testing.T, reader SlowContextReader) *Service {
	t.Helper()
	service := newBucketService(t)
	service.slowContext = reader
	return service
}

func mustTimeRFC3339(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}
