package marketstateapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	regimeengine "github.com/crypto-market-copilot/alerts/services/regime-engine"
	slowcontext "github.com/crypto-market-copilot/alerts/services/slow-context"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

var (
	ErrUnsupportedSymbol        = errors.New("unsupported symbol")
	ErrRuntimeStatusUnsupported = errors.New("runtime status unsupported by provider")
)

type SymbolStateResponse struct {
	features.MarketStateCurrentResponse
	SlowContext slowcontext.AssetContextResponse `json:"slowContext"`
}

type Provider interface {
	CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error)
	CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error)
}

type RuntimeStatusReader interface {
	RuntimeStatus(ctx context.Context) (RuntimeStatusResponse, error)
}

type RuntimeStatusReadiness string

const (
	RuntimeStatusNotReady RuntimeStatusReadiness = "NOT_READY"
	RuntimeStatusReady    RuntimeStatusReadiness = "READY"
)

type RuntimeStatusResponse struct {
	GeneratedAt time.Time                     `json:"generatedAt"`
	Symbols     []RuntimeStatusSymbolResponse `json:"symbols"`
}

type RuntimeStatusSymbolResponse struct {
	Symbol                 string                           `json:"symbol"`
	SourceSymbol           string                           `json:"sourceSymbol"`
	QuoteCurrency          string                           `json:"quoteCurrency"`
	Readiness              RuntimeStatusReadiness           `json:"readiness"`
	FeedHealth             RuntimeStatusFeedHealthResponse  `json:"feedHealth"`
	ConnectionState        ingestion.ConnectionState        `json:"connectionState"`
	LocalClockOffsetMillis int64                            `json:"localClockOffsetMillis"`
	ConsecutiveReconnects  int                              `json:"consecutiveReconnects"`
	DepthStatus            RuntimeStatusDepthStatusResponse `json:"depthStatus"`
	LastAcceptedExchange   *time.Time                       `json:"lastAcceptedExchange"`
	LastAcceptedRecv       *time.Time                       `json:"lastAcceptedRecv"`
	LastMessageAt          *time.Time                       `json:"lastMessageAt"`
	LastSnapshotAt         *time.Time                       `json:"lastSnapshotAt"`
}

type RuntimeStatusFeedHealthResponse struct {
	State                 ingestion.FeedHealthState     `json:"state"`
	ConnectionState       ingestion.ConnectionState     `json:"connectionState"`
	MessageFreshness      ingestion.FreshnessState      `json:"messageFreshness"`
	SnapshotFreshness     ingestion.FreshnessState      `json:"snapshotFreshness"`
	SequenceGapDetected   bool                          `json:"sequenceGapDetected"`
	ConsecutiveReconnects int                           `json:"consecutiveReconnects"`
	ResyncCount           int                           `json:"resyncCount"`
	ClockState            ingestion.ClockState          `json:"clockState"`
	Reasons               []ingestion.DegradationReason `json:"reasons"`
}

type RuntimeStatusDepthStatusResponse struct {
	State                   venuebinance.SpotDepthRecoveryState   `json:"state"`
	Trigger                 venuebinance.SpotDepthRecoveryTrigger `json:"trigger"`
	SourceSymbol            string                                `json:"sourceSymbol"`
	LastAcceptedSequence    int64                                 `json:"lastAcceptedSequence"`
	BufferedDeltaCount      int                                   `json:"bufferedDeltaCount"`
	LastMessageAt           *time.Time                            `json:"lastMessageAt"`
	LastSnapshotAt          *time.Time                            `json:"lastSnapshotAt"`
	LastRecoveryAttemptAt   *time.Time                            `json:"lastRecoveryAttemptAt"`
	RemainingCooldownMillis int64                                 `json:"remainingCooldownMillis"`
	RetryAfterMillis        int64                                 `json:"retryAfterMillis"`
	ResyncCount             int                                   `json:"resyncCount"`
	SequenceGapDetected     bool                                  `json:"sequenceGapDetected"`
	RefreshDue              bool                                  `json:"refreshDue"`
	Synchronized            bool                                  `json:"synchronized"`
}

type SlowContextReader interface {
	QueryAsset(query slowcontext.AssetQuery) (slowcontext.AssetContextResponse, error)
}

type currentStateSource interface {
	Bundle(ctx context.Context, now time.Time) (currentStateBundle, error)
}

type Handler struct {
	provider Provider
}

type DeterministicProvider struct {
	provider *currentStateProvider
}

type LiveSpotProvider struct {
	provider *currentStateProvider
}

type currentStateProvider struct {
	clock  func() time.Time
	source currentStateSource
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

var runtimeStatusTrackedSymbols = []string{"BTC-USD", "ETH-USD"}

func NewHandler(provider Provider) (*Handler, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	return &Handler{provider: provider}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.HandleFunc("GET /api/runtime-status", h.handleRuntimeStatus)
	mux.HandleFunc("GET /api/market-state/global", h.handleCurrentGlobalState)
	mux.HandleFunc("GET /api/market-state/{symbol}", h.handleCurrentSymbolState)
	return mux
}

func NewDeterministicProvider() *DeterministicProvider {
	provider, err := NewDeterministicProviderWithClock(func() time.Time {
		return time.Now().UTC()
	})
	if err != nil {
		panic(err)
	}
	return provider
}

func NewDeterministicProviderWithClock(clock func() time.Time) (*DeterministicProvider, error) {
	provider, err := newCurrentStateProvider(deterministicSource{}, clock)
	if err != nil {
		return nil, err
	}
	return &DeterministicProvider{provider: provider}, nil
}

func (p *DeterministicProvider) CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error) {
	if p == nil || p.provider == nil {
		return features.MarketStateCurrentGlobalResponse{}, fmt.Errorf("provider is required")
	}
	return p.provider.CurrentGlobalState(ctx)
}

func (p *DeterministicProvider) CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error) {
	if p == nil || p.provider == nil {
		return SymbolStateResponse{}, fmt.Errorf("provider is required")
	}
	return p.provider.CurrentSymbolState(ctx, symbol)
}

func NewLiveSpotProvider(reader SpotCurrentStateReader, clock func() time.Time, slowContextReader SlowContextReader) (*LiveSpotProvider, error) {
	source, err := newSpotLiveCurrentStateSource(reader, slowContextReader)
	if err != nil {
		return nil, err
	}
	provider, err := newCurrentStateProvider(source, clock)
	if err != nil {
		return nil, err
	}
	return &LiveSpotProvider{provider: provider}, nil
}

func (p *LiveSpotProvider) CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error) {
	if p == nil || p.provider == nil {
		return features.MarketStateCurrentGlobalResponse{}, fmt.Errorf("provider is required")
	}
	return p.provider.CurrentGlobalState(ctx)
}

func (p *LiveSpotProvider) CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error) {
	if p == nil || p.provider == nil {
		return SymbolStateResponse{}, fmt.Errorf("provider is required")
	}
	return p.provider.CurrentSymbolState(ctx, symbol)
}

func newCurrentStateProvider(source currentStateSource, clock func() time.Time) (*currentStateProvider, error) {
	if source == nil {
		return nil, fmt.Errorf("source is required")
	}
	if clock == nil {
		return nil, fmt.Errorf("clock is required")
	}
	return &currentStateProvider{clock: clock, source: source}, nil
}

func (p *currentStateProvider) CurrentGlobalState(ctx context.Context) (features.MarketStateCurrentGlobalResponse, error) {
	bundle, err := p.currentStateBundle(ctx)
	if err != nil {
		return features.MarketStateCurrentGlobalResponse{}, err
	}
	return bundle.global, nil
}

func (p *currentStateProvider) CurrentSymbolState(ctx context.Context, symbol string) (SymbolStateResponse, error) {
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

func (p *currentStateProvider) currentStateBundle(ctx context.Context) (currentStateBundle, error) {
	if p == nil || p.source == nil {
		return currentStateBundle{}, fmt.Errorf("provider is required")
	}
	select {
	case <-ctx.Done():
		return currentStateBundle{}, ctx.Err()
	default:
	}
	return p.source.Bundle(ctx, p.clock().UTC().Truncate(time.Second))
}

type deterministicSource struct{}

func (deterministicSource) Bundle(ctx context.Context, now time.Time) (currentStateBundle, error) {
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
	signalsBySymbol, err := currentStateUSDMInfluenceSignalsFromInput(featureService, deterministicUSDMInfluenceInput(now))
	if err != nil {
		return currentStateBundle{}, err
	}

	btcQuery, btcSlowQuery, err := deterministicSymbolQueries("BTC-USD", now)
	if err != nil {
		return currentStateBundle{}, err
	}
	btcQuery.SymbolRegime, btcQuery.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(btcQuery.SymbolRegime, signalsBySymbol[btcQuery.Symbol])
	if err != nil {
		return currentStateBundle{}, err
	}
	ethQuery, ethSlowQuery, err := deterministicSymbolQueries("ETH-USD", now)
	if err != nil {
		return currentStateBundle{}, err
	}
	ethQuery.SymbolRegime, ethQuery.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(ethQuery.SymbolRegime, signalsBySymbol[ethQuery.Symbol])
	if err != nil {
		return currentStateBundle{}, err
	}
	globalRegime, err := features.EvaluateGlobalRegime(deterministicRegimeConfig(), map[string]features.SymbolRegimeSnapshot{
		btcQuery.Symbol: btcQuery.SymbolRegime,
		ethQuery.Symbol: ethQuery.SymbolRegime,
	}, nil)
	if err != nil {
		return currentStateBundle{}, err
	}
	btcQuery.GlobalRegime = globalRegime
	ethQuery.GlobalRegime = globalRegime
	btc, err := buildSymbolStateFromQuery(featureService, btcQuery, btcSlowQuery)
	if err != nil {
		return currentStateBundle{}, err
	}
	eth, err := buildSymbolStateFromQuery(featureService, ethQuery, ethSlowQuery)
	if err != nil {
		return currentStateBundle{}, err
	}
	global, err := regimeService.QueryCurrentGlobalState(features.GlobalCurrentStateQuery{
		AsOf:         now,
		GlobalRegime: globalRegime,
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

func (h *Handler) handleRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	reader, ok := h.provider.(RuntimeStatusReader)
	if !ok {
		writeError(w, http.StatusNotImplemented, ErrRuntimeStatusUnsupported)
		return
	}
	response, err := reader.RuntimeStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	response, err = normalizeRuntimeStatusResponse(response)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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

func NewRuntimeStatusFeedHealthResponse(status ingestion.FeedHealthStatus) RuntimeStatusFeedHealthResponse {
	return RuntimeStatusFeedHealthResponse{
		State:                 status.State,
		ConnectionState:       status.ConnectionState,
		MessageFreshness:      status.MessageFreshness,
		SnapshotFreshness:     status.SnapshotFreshness,
		SequenceGapDetected:   status.SequenceGapDetected,
		ConsecutiveReconnects: status.ConsecutiveReconnects,
		ResyncCount:           status.ResyncCount,
		ClockState:            status.ClockState,
		Reasons:               append([]ingestion.DegradationReason{}, status.Reasons...),
	}
}

func NewRuntimeStatusDepthStatusResponse(status venuebinance.SpotDepthRecoveryStatus) RuntimeStatusDepthStatusResponse {
	return RuntimeStatusDepthStatusResponse{
		State:                   status.State,
		Trigger:                 status.Trigger,
		SourceSymbol:            status.SourceSymbol,
		LastAcceptedSequence:    status.LastAcceptedSequence,
		BufferedDeltaCount:      status.BufferedDeltaCount,
		LastMessageAt:           nullableTime(status.LastMessageAt),
		LastSnapshotAt:          nullableTime(status.LastSnapshotAt),
		LastRecoveryAttemptAt:   nullableTime(status.LastRecoveryAttemptAt),
		RemainingCooldownMillis: status.RemainingCooldown.Milliseconds(),
		RetryAfterMillis:        status.RetryAfter.Milliseconds(),
		ResyncCount:             status.ResyncCount,
		SequenceGapDetected:     status.SequenceGapDetected,
		RefreshDue:              status.RefreshDue,
		Synchronized:            status.Synchronized,
	}
}

func nullableTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func normalizeRuntimeStatusResponse(response RuntimeStatusResponse) (RuntimeStatusResponse, error) {
	bySymbol := make(map[string]RuntimeStatusSymbolResponse, len(response.Symbols))
	for _, symbol := range response.Symbols {
		if symbol.Symbol == "" {
			return RuntimeStatusResponse{}, fmt.Errorf("runtime status symbol is required")
		}
		if _, exists := bySymbol[symbol.Symbol]; exists {
			return RuntimeStatusResponse{}, fmt.Errorf("duplicate runtime status symbol: %s", symbol.Symbol)
		}
		bySymbol[symbol.Symbol] = symbol
	}

	normalized := make([]RuntimeStatusSymbolResponse, 0, len(runtimeStatusTrackedSymbols))
	for _, symbol := range runtimeStatusTrackedSymbols {
		entry, ok := bySymbol[symbol]
		if !ok {
			return RuntimeStatusResponse{}, fmt.Errorf("missing runtime status symbol: %s", symbol)
		}
		normalized = append(normalized, entry)
		delete(bySymbol, symbol)
	}
	if len(bySymbol) != 0 {
		extra := make([]string, 0, len(bySymbol))
		for symbol := range bySymbol {
			extra = append(extra, symbol)
		}
		sort.Strings(extra)
		return RuntimeStatusResponse{}, fmt.Errorf("unsupported runtime status symbol: %s", extra[0])
	}

	response.Symbols = normalized
	return response, nil
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
	return buildSymbolStateFromQuery(service, query, slowQuery)
}

func buildSymbolStateFromQuery(service *featureengine.Service, query features.SymbolCurrentStateQuery, slowQuery slowcontext.AssetQuery) (SymbolStateResponse, error) {
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
