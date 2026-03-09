package marketstateapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	regimeengine "github.com/crypto-market-copilot/alerts/services/regime-engine"
	slowcontext "github.com/crypto-market-copilot/alerts/services/slow-context"
)

var ErrUnsupportedSymbol = errors.New("unsupported symbol")

type SymbolStateResponse struct {
	features.MarketStateCurrentResponse
	SlowContext slowcontext.AssetContextResponse `json:"slowContext"`
}

type Provider interface {
	CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error)
	CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error)
}

type Handler struct {
	provider Provider
}

type DeterministicProvider struct {
	clock func() time.Time
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type currentStateBundle struct {
	global  features.MarketStateCurrentGlobalResponse
	symbols map[string]SymbolStateResponse
}

func NewHandler(provider Provider) (*Handler, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	return &Handler{provider: provider}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.HandleFunc("GET /api/market-state/global", h.handleCurrentGlobalState)
	mux.HandleFunc("GET /api/market-state/{symbol}", h.handleCurrentSymbolState)
	return mux
}

func NewDeterministicProvider() *DeterministicProvider {
	return &DeterministicProvider{
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func NewDeterministicProviderWithClock(clock func() time.Time) (*DeterministicProvider, error) {
	if clock == nil {
		return nil, fmt.Errorf("clock is required")
	}
	return &DeterministicProvider{clock: clock}, nil
}

func (p *DeterministicProvider) CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error) {
	bundle, err := p.currentStateBundle(ctx)
	if err != nil {
		return features.MarketStateCurrentGlobalResponse{}, err
	}
	return bundle.global, nil
}

func (p *DeterministicProvider) CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error) {
	bundle, err := p.currentStateBundle(ctx)
	if err != nil {
		return SymbolStateResponse{}, err
	}
	response, ok := bundle.symbols[strings.ToUpper(symbol)]
	if !ok {
		return SymbolStateResponse{}, fmt.Errorf("%w: %s", ErrUnsupportedSymbol, symbol)
	}
	return response, nil
}

func (p *DeterministicProvider) currentStateBundle(ctx context.Context) (currentStateBundle, error) {
	if p == nil {
		return currentStateBundle{}, fmt.Errorf("provider is required")
	}
	select {
	case <-ctx.Done():
		return currentStateBundle{}, ctx.Err()
	default:
	}

	now := p.clock().UTC().Truncate(time.Second)
	slowService, err := seededSlowContextService(now)
	if err != nil {
		return currentStateBundle{}, err
	}
	featureService, err := featureengine.NewService(deterministicCompositeConfig(), featureengine.WithSlowContextReader(slowService))
	if err != nil {
		return currentStateBundle{}, err
	}
	regimeService, err := regimeengine.NewService(deterministicRegimeConfig())
	if err != nil {
		return currentStateBundle{}, err
	}

	btc, err := buildSymbolState(featureService, "BTC-USD", now)
	if err != nil {
		return currentStateBundle{}, err
	}
	eth, err := buildSymbolState(featureService, "ETH-USD", now)
	if err != nil {
		return currentStateBundle{}, err
	}
	global, err := regimeService.QueryCurrentGlobalState(features.GlobalCurrentStateQuery{
		AsOf:         now,
		GlobalRegime: currentGlobalRegime(now),
		Symbols:      []features.MarketStateCurrentResponse{btc.MarketStateCurrentResponse, eth.MarketStateCurrentResponse},
	})
	if err != nil {
		return currentStateBundle{}, err
	}

	return currentStateBundle{
		global: global,
		symbols: map[string]SymbolStateResponse{
			"BTC-USD": btc,
			"ETH-USD": eth,
		},
	}, nil

}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	_ = r
}

func (h *Handler) handleCurrentGlobalState(w http.ResponseWriter, r *http.Request) {
	response, err := h.provider.CurrentGlobalState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCurrentSymbolState(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	response, err := h.provider.CurrentSymbolState(r.Context(), symbol)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrUnsupportedSymbol) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func deterministicCompositeConfig() features.CompositeConfig {
	return features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties: features.PenaltyConfig{
			FeedHealthDegradedMultiplier: 0.8,
			TimestampDegradedMultiplier:  0.75,
		},
		QuoteProxies: map[string]features.QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
		},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {
				Members: []features.MemberConfig{{Venue: "BINANCE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			features.CompositeGroupUSA: {
				Members: []features.MemberConfig{{Venue: "COINBASE", MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
		},
	}
}

func deterministicRegimeConfig() features.RegimeConfig {
	return features.RegimeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "regime-engine.market-state.v1",
		AlgorithmVersion: "symbol-global-regime.v1",
		Symbols:          []string{"BTC-USD", "ETH-USD"},
		Symbol: features.SymbolRegimeThresholds{
			CoverageWatchMax:             0.85,
			CoverageNoOperateMax:         0.60,
			CombinedTrustCapWatchMax:     0.65,
			CombinedTrustCapNoOperateMax: 0.35,
			TimestampFallbackWatchRatio:  0.10,
			TimestampFallbackNoOpRatio:   0.50,
			NoOperateToWatchWindows:      2,
			WatchToTradeableWindows:      2,
		},
		Global: features.GlobalRegimeThresholds{
			NoOperateToWatchWindows: 2,
			WatchToTradeableWindows: 2,
		},
	}
}

func buildSymbolState(service *featureengine.Service, symbol string, now time.Time) (SymbolStateResponse, error) {
	query, slowQuery, err := deterministicSymbolQueries(symbol, now)
	if err != nil {
		return SymbolStateResponse{}, err
	}
	response, err := service.QueryCurrentStateWithSlowContext(query, slowQuery)
	if err != nil {
		return SymbolStateResponse{}, err
	}
	return SymbolStateResponse{
		MarketStateCurrentResponse: response.CurrentState,
		SlowContext:                response.SlowContext,
	}, nil
}

func deterministicSymbolQueries(symbol string, now time.Time) (features.SymbolCurrentStateQuery, slowcontext.AssetQuery, error) {
	asOf := now.Add(-16 * time.Second)
	asset := "BTC"
	worldPrice := 64000.0
	usaPrice := 63996.0
	usaCoverage := 0.98
	usaHealth := 0.97
	compositeReasonCodes := []features.ReasonCode{}
	degradedUSA := false
	symbolRegime := features.SymbolRegimeSnapshot{
		SchemaVersion:         "v1",
		Symbol:                symbol,
		State:                 features.RegimeStateTradeable,
		EffectiveBucketEnd:    asOf.Format(time.RFC3339),
		Reasons:               []features.RegimeReasonCode{features.RegimeReasonHealthy},
		PrimaryReason:         features.RegimeReasonHealthy,
		TransitionKind:        features.RegimeTransitionHold,
		ConfigVersion:         "regime-engine.market-state.v1",
		AlgorithmVersion:      "symbol-global-regime.v1",
		ObservedInstantaneous: features.RegimeStateTradeable,
	}
	if symbol == "ETH-USD" {
		asOf = now.Add(-20 * time.Second)
		asset = "ETH"
		worldPrice = 3242.0
		usaPrice = 3236.0
		usaCoverage = 0.86
		usaHealth = 0.84
		compositeReasonCodes = []features.ReasonCode{features.ReasonTimestampFallback}
		degradedUSA = true
		symbolRegime = features.SymbolRegimeSnapshot{
			SchemaVersion:         "v1",
			Symbol:                symbol,
			State:                 features.RegimeStateWatch,
			EffectiveBucketEnd:    asOf.Format(time.RFC3339),
			Reasons:               []features.RegimeReasonCode{features.RegimeReasonTimestampTrustLoss},
			PrimaryReason:         features.RegimeReasonTimestampTrustLoss,
			TransitionKind:        features.RegimeTransitionHold,
			ConfigVersion:         "regime-engine.market-state.v1",
			AlgorithmVersion:      "symbol-global-regime.v1",
			ObservedInstantaneous: features.RegimeStateWatch,
		}
	}

	query := features.SymbolCurrentStateQuery{
		Symbol: symbol,
		AsOf:   asOf,
		World: features.CompositeSnapshot{
			SchemaVersion:                "v1",
			Symbol:                       symbol,
			BucketTs:                     asOf.Format(time.RFC3339),
			CompositeGroup:               features.CompositeGroupWorld,
			CompositePrice:               floatPtr(worldPrice),
			CoverageRatio:                1,
			HealthScore:                  0.99,
			ConfiguredContributorCount:   2,
			EligibleContributorCount:     2,
			ContributingContributorCount: 2,
			ConfigVersion:                "composite-config.v1",
			AlgorithmVersion:             "world-usa-composite.v1",
		},
		USA: features.CompositeSnapshot{
			SchemaVersion:                     "v1",
			Symbol:                            symbol,
			BucketTs:                          asOf.Format(time.RFC3339),
			CompositeGroup:                    features.CompositeGroupUSA,
			CompositePrice:                    floatPtr(usaPrice),
			CoverageRatio:                     usaCoverage,
			HealthScore:                       usaHealth,
			ConfiguredContributorCount:        2,
			EligibleContributorCount:          2,
			ContributingContributorCount:      2,
			TimestampFallbackContributorCount: boolToInt(degradedUSA),
			Degraded:                          degradedUSA,
			DegradedReasons:                   compositeReasonCodes,
			ConfigVersion:                     "composite-config.v1",
			AlgorithmVersion:                  "world-usa-composite.v1",
		},
		Buckets: []features.MarketQualityBucket{
			currentBucket(symbol, features.BucketFamily30s, asOf.Add(-30*time.Second), asOf, 1, 0, 0.92, degradedUSA),
			currentBucket(symbol, features.BucketFamily2m, asOf.Add(-2*time.Minute), asOf, 4, 0, 0.78, degradedUSA),
			currentBucket(symbol, features.BucketFamily5m, asOf.Add(-5*time.Minute), asOf, 10, 0, 0.78, degradedUSA),
		},
		SymbolRegime: symbolRegime,
		GlobalRegime: currentGlobalRegime(asOf),
	}
	query.RecentContext = append([]features.MarketQualityBucket(nil), query.Buckets...)
	return query, slowcontext.AssetQuery{
		Asset: asset,
		MetricFamilies: []slowcontext.MetricFamily{
			slowcontext.MetricFamilyCMEVolume,
			slowcontext.MetricFamilyCMEOpenInterest,
			slowcontext.MetricFamilyETFDailyFlow,
		},
		Now: now,
	}, nil
}

func currentGlobalRegime(now time.Time) features.GlobalRegimeSnapshot {
	return features.GlobalRegimeSnapshot{
		SchemaVersion:           "v1",
		State:                   features.RegimeStateTradeable,
		EffectiveBucketEnd:      now.UTC().Format(time.RFC3339),
		Reasons:                 []features.RegimeReasonCode{features.RegimeReasonHealthy},
		PrimaryReason:           features.RegimeReasonHealthy,
		AppliedCeilingToSymbols: nil,
		TransitionKind:          features.RegimeTransitionHold,
		ConfigVersion:           "regime-engine.market-state.v1",
		AlgorithmVersion:        "symbol-global-regime.v1",
		ObservedInstantaneous:   features.RegimeStateTradeable,
	}
}

func currentBucket(symbol string, family features.BucketFamily, start time.Time, end time.Time, closed int, missing int, trustCap float64, degraded bool) features.MarketQualityBucket {
	reasons := []features.ReasonCode{}
	if degraded {
		reasons = []features.ReasonCode{features.ReasonTimestampTrustLoss}
	}
	return features.MarketQualityBucket{
		Symbol: symbol,
		Window: features.BucketWindowSummary{
			Family:              family,
			Start:               start.UTC().Format(time.RFC3339),
			End:                 end.UTC().Format(time.RFC3339),
			ClosedBucketCount:   closed,
			ExpectedBucketCount: closed + missing,
			MissingBucketCount:  missing,
			ConfigVersion:       "market-quality.v1",
			AlgorithmVersion:    "market-quality-buckets.v1",
		},
		Assignment: features.BucketAssignment{
			Family:       family,
			BucketStart:  start.UTC().Format(time.RFC3339),
			BucketEnd:    end.UTC().Format(time.RFC3339),
			BucketSource: features.BucketSourceExchangeTs,
		},
		World:          features.CompositeBucketSide{Available: true, CoverageRatio: 1, HealthScore: 0.99},
		USA:            features.CompositeBucketSide{Available: true, CoverageRatio: ternaryFloat(degraded, 0.86, 0.98), HealthScore: ternaryFloat(degraded, 0.84, 0.97), TimestampFallbackContributorCount: boolToInt(degraded)},
		Divergence:     features.DivergenceSummary{Available: true, ReasonCodes: reasons},
		Fragmentation:  features.FragmentationSummary{Severity: ternaryFragmentation(degraded), PersistenceCount: boolToInt(degraded), PrimaryCauses: reasons},
		TimestampTrust: features.TimestampTrustSummary{RecvFallbackCount: boolToInt(degraded), FallbackRatio: ternaryFloat(degraded, 0.10, 0), TrustCap: !degraded || family != features.BucketFamily30s},
		MarketQuality:  features.MarketQualitySummary{CombinedTrustCap: trustCap, DowngradedReasons: reasons},
	}
}

func seededSlowContextService(now time.Time) (*slowcontext.Service, error) {
	service, err := slowcontext.NewService()
	if err != nil {
		return nil, err
	}
	for _, record := range deterministicSlowContextPolls(now) {
		if _, err := service.RecordAccepted(record.result, record.expectedCadence, record.acceptedAt); err != nil {
			return nil, err
		}
	}
	return service, nil
}

type deterministicSlowPoll struct {
	result          slowcontext.PollResult
	expectedCadence string
	acceptedAt      time.Time
}

func deterministicSlowContextPolls(now time.Time) []deterministicSlowPoll {
	cmeAsOf := currentSlowAsOf(now, slowcontext.SourceFamilyCME)
	etfAsOf := currentSlowAsOf(now, slowcontext.SourceFamilyETF)
	cmePublished := cmeAsOf.Add(20*time.Hour + 45*time.Minute)
	etfPublished := etfAsOf.Add(22*time.Hour + 15*time.Minute)
	return []deterministicSlowPoll{
		{
			result: slowcontext.PollResult{
				SourceFamily: slowcontext.SourceFamilyCME,
				MetricFamily: slowcontext.MetricFamilyCMEVolume,
				Status:       slowcontext.PublicationStatusNew,
				SourceKey:    "btc-cme-volume",
				Asset:        "BTC",
				AsOfTs:       cmeAsOf,
				PublishedTs:  cmePublished,
				IngestTs:     cmePublished.Add(3 * time.Minute),
				DedupeKey:    slowcontext.BuildDedupeKey(slowcontext.SourceFamilyCME, slowcontext.MetricFamilyCMEVolume, "BTC", "btc-cme-volume", cmeAsOf),
				Revision:     "1",
				Value:        &slowcontext.CandidateValue{Amount: "18342.55", Unit: "contracts"},
			},
			expectedCadence: "session",
			acceptedAt:      cmePublished.Add(4 * time.Minute),
		},
		{
			result: slowcontext.PollResult{
				SourceFamily: slowcontext.SourceFamilyCME,
				MetricFamily: slowcontext.MetricFamilyCMEOpenInterest,
				Status:       slowcontext.PublicationStatusNew,
				SourceKey:    "btc-cme-oi",
				Asset:        "BTC",
				AsOfTs:       cmeAsOf,
				PublishedTs:  cmePublished,
				IngestTs:     cmePublished.Add(3 * time.Minute),
				DedupeKey:    slowcontext.BuildDedupeKey(slowcontext.SourceFamilyCME, slowcontext.MetricFamilyCMEOpenInterest, "BTC", "btc-cme-oi", cmeAsOf),
				Revision:     "1",
				Value:        &slowcontext.CandidateValue{Amount: "9284.00", Unit: "contracts"},
			},
			expectedCadence: "session",
			acceptedAt:      cmePublished.Add(4 * time.Minute),
		},
		{
			result: slowcontext.PollResult{
				SourceFamily: slowcontext.SourceFamilyETF,
				MetricFamily: slowcontext.MetricFamilyETFDailyFlow,
				Status:       slowcontext.PublicationStatusNew,
				SourceKey:    "btc-etf-flow",
				Asset:        "BTC",
				AsOfTs:       etfAsOf,
				PublishedTs:  etfPublished,
				IngestTs:     etfPublished.Add(3 * time.Minute),
				DedupeKey:    slowcontext.BuildDedupeKey(slowcontext.SourceFamilyETF, slowcontext.MetricFamilyETFDailyFlow, "BTC", "btc-etf-flow", etfAsOf),
				Revision:     "1",
				Value:        &slowcontext.CandidateValue{Amount: "245000000.00", Unit: "usd"},
			},
			expectedCadence: "daily",
			acceptedAt:      etfPublished.Add(4 * time.Minute),
		},
		{
			result: slowcontext.PollResult{
				SourceFamily: slowcontext.SourceFamilyCME,
				MetricFamily: slowcontext.MetricFamilyCMEVolume,
				Status:       slowcontext.PublicationStatusNew,
				SourceKey:    "eth-cme-volume",
				Asset:        "ETH",
				AsOfTs:       cmeAsOf,
				PublishedTs:  cmePublished,
				IngestTs:     cmePublished.Add(3 * time.Minute),
				DedupeKey:    slowcontext.BuildDedupeKey(slowcontext.SourceFamilyCME, slowcontext.MetricFamilyCMEVolume, "ETH", "eth-cme-volume", cmeAsOf),
				Revision:     "1",
				Value:        &slowcontext.CandidateValue{Amount: "9284.00", Unit: "contracts"},
			},
			expectedCadence: "session",
			acceptedAt:      cmePublished.Add(4 * time.Minute),
		},
		{
			result: slowcontext.PollResult{
				SourceFamily: slowcontext.SourceFamilyCME,
				MetricFamily: slowcontext.MetricFamilyCMEOpenInterest,
				Status:       slowcontext.PublicationStatusNew,
				SourceKey:    "eth-cme-oi",
				Asset:        "ETH",
				AsOfTs:       cmeAsOf,
				PublishedTs:  cmePublished,
				IngestTs:     cmePublished.Add(3 * time.Minute),
				DedupeKey:    slowcontext.BuildDedupeKey(slowcontext.SourceFamilyCME, slowcontext.MetricFamilyCMEOpenInterest, "ETH", "eth-cme-oi", cmeAsOf),
				Revision:     "1",
				Value:        &slowcontext.CandidateValue{Amount: "6120.25", Unit: "contracts"},
			},
			expectedCadence: "session",
			acceptedAt:      cmePublished.Add(4 * time.Minute),
		},
	}
}

func currentSlowAsOf(now time.Time, sourceFamily slowcontext.SourceFamily) time.Time {
	config, err := slowcontext.DefaultScheduleConfig(sourceFamily)
	if err != nil {
		return startOfUTC(now).Add(-24 * time.Hour)
	}
	minuteOfDay := now.UTC().Hour()*60 + now.UTC().Minute()
	asOf := startOfUTC(now)
	if minuteOfDay < config.PublishWindowEndMinute {
		asOf = asOf.Add(-24 * time.Hour)
	}
	return asOf
}

func startOfUTC(now time.Time) time.Time {
	return time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func floatPtr(value float64) *float64 {
	return &value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ternaryFloat(condition bool, whenTrue float64, whenFalse float64) float64 {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func ternaryFragmentation(degraded bool) features.FragmentationSeverity {
	if degraded {
		return features.FragmentationSeverityModerate
	}
	return features.FragmentationSeverityLow
}
