package marketstateapi

import (
	"context"
	"fmt"
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

var supportedCurrentStateSymbols = []string{"BTC-USD", "ETH-USD"}

type SpotCurrentStateReader interface {
	Snapshot(ctx context.Context, now time.Time) (SpotCurrentStateSnapshot, error)
}

type SpotCurrentStateSnapshot struct {
	Observations []SpotCurrentStateObservation
}

type SpotCurrentStateObservation struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	BestBidPrice  float64
	BestAskPrice  float64
	ExchangeTs    time.Time
	RecvTs        time.Time
	FeedHealth    ingestion.FeedHealthStatus
	DepthStatus   venuebinance.SpotDepthRecoveryStatus
}

type spotLiveCurrentStateSource struct {
	reader      SpotCurrentStateReader
	slowContext SlowContextReader
}

type symbolAssemblyState struct {
	world        features.CompositeSnapshot
	usa          features.CompositeSnapshot
	buckets      []features.MarketQualityBucket
	symbolRegime features.SymbolRegimeSnapshot
	hasData      bool
}

func newSpotLiveCurrentStateSource(reader SpotCurrentStateReader, slowContextReader SlowContextReader) (*spotLiveCurrentStateSource, error) {
	if reader == nil {
		return nil, fmt.Errorf("spot current-state reader is required")
	}
	return &spotLiveCurrentStateSource{reader: reader, slowContext: slowContextReader}, nil
}

func (s *spotLiveCurrentStateSource) Bundle(ctx context.Context, now time.Time) (currentStateBundle, error) {
	if now.IsZero() {
		return currentStateBundle{}, fmt.Errorf("current time is required")
	}
	snapshot, err := s.reader.Snapshot(ctx, now)
	if err != nil {
		return currentStateBundle{}, err
	}
	featureService, err := newLiveFeatureService(s.slowContext)
	if err != nil {
		return currentStateBundle{}, err
	}
	regimeService, err := regimeengine.NewService(deterministicRegimeConfig())
	if err != nil {
		return currentStateBundle{}, err
	}

	grouped, err := groupSpotObservations(snapshot.Observations)
	if err != nil {
		return currentStateBundle{}, err
	}
	assembled := make(map[string]symbolAssemblyState, len(supportedCurrentStateSymbols))
	regimes := make(map[string]features.SymbolRegimeSnapshot, len(supportedCurrentStateSymbols))
	for _, symbol := range supportedCurrentStateSymbols {
		state, err := assembleSymbolState(featureService, symbol, grouped[symbol], now)
		if err != nil {
			return currentStateBundle{}, err
		}
		assembled[symbol] = state
		regimes[symbol] = state.symbolRegime
	}

	globalRegime, err := features.EvaluateGlobalRegime(deterministicRegimeConfig(), regimes, nil)
	if err != nil {
		return currentStateBundle{}, err
	}

	symbols := make(map[string]SymbolStateResponse, len(supportedCurrentStateSymbols))
	globalInputs := make([]features.MarketStateCurrentResponse, 0, len(supportedCurrentStateSymbols))
	for _, symbol := range supportedCurrentStateSymbols {
		state := assembled[symbol]
		query := features.SymbolCurrentStateQuery{
			Symbol:        symbol,
			AsOf:          now,
			World:         state.world,
			USA:           state.usa,
			Buckets:       append([]features.MarketQualityBucket(nil), state.buckets...),
			RecentContext: append([]features.MarketQualityBucket(nil), state.buckets...),
			SymbolRegime:  state.symbolRegime,
			GlobalRegime:  globalRegime,
		}
		response, err := featureService.QueryCurrentStateWithSlowContext(query, currentStateSlowContextQuery(symbol, now))
		if err != nil {
			return currentStateBundle{}, err
		}
		symbols[symbol] = SymbolStateResponse{
			MarketStateCurrentResponse: response.CurrentState,
			SlowContext:                response.SlowContext,
		}
		globalInputs = append(globalInputs, response.CurrentState)
	}

	global, err := regimeService.QueryCurrentGlobalState(features.GlobalCurrentStateQuery{
		AsOf:         now,
		GlobalRegime: globalRegime,
		Symbols:      globalInputs,
	})
	if err != nil {
		return currentStateBundle{}, err
	}
	return currentStateBundle{global: global, symbols: symbols}, nil
}

func newLiveFeatureService(slowContextReader SlowContextReader) (*featureengine.Service, error) {
	options := []featureengine.ServiceOption{featureengine.WithBucketConfig(currentStateBucketConfig())}
	if slowContextReader != nil {
		options = append(options, featureengine.WithSlowContextReader(slowContextReader))
	}
	return featureengine.NewService(deterministicCompositeConfig(), options...)
}

func groupSpotObservations(observations []SpotCurrentStateObservation) (map[string][]SpotCurrentStateObservation, error) {
	grouped := make(map[string][]SpotCurrentStateObservation, len(supportedCurrentStateSymbols))
	for _, observation := range observations {
		if !isSupportedCurrentStateSymbol(observation.Symbol) {
			continue
		}
		if err := validateSpotObservation(observation); err != nil {
			return nil, err
		}
		grouped[observation.Symbol] = append(grouped[observation.Symbol], observation)
	}
	for symbol := range grouped {
		sort.Slice(grouped[symbol], func(i, j int) bool {
			left := grouped[symbol][i]
			right := grouped[symbol][j]
			if !left.RecvTs.Equal(right.RecvTs) {
				return left.RecvTs.Before(right.RecvTs)
			}
			if !left.ExchangeTs.Equal(right.ExchangeTs) {
				return left.ExchangeTs.Before(right.ExchangeTs)
			}
			return left.SourceSymbol < right.SourceSymbol
		})
	}
	return grouped, nil
}

func assembleSymbolState(service *featureengine.Service, symbol string, observations []SpotCurrentStateObservation, now time.Time) (symbolAssemblyState, error) {
	if len(observations) == 0 {
		return unavailableSymbolAssemblyState(service, symbol, now, features.RegimeReasonCompositeUnavailable)
	}
	state := symbolAssemblyState{}
	var previous *features.SymbolRegimeSnapshot
	for _, observation := range observations {
		world, usa, exchangeTs, recvTs, err := buildLiveCompositeSnapshots(service, observation)
		if err != nil {
			return symbolAssemblyState{}, err
		}
		state.world = world
		state.usa = usa
		state.hasData = true
		result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
			Symbol:     symbol,
			ExchangeTs: exchangeTs,
			RecvTs:     recvTs,
			World:      world,
			USA:        usa,
			Now:        recvTs,
		})
		if err != nil {
			return symbolAssemblyState{}, err
		}
		state.buckets = append(state.buckets, result.Emitted...)
		for _, bucket := range result.Emitted {
			if bucket.Window.Family != features.BucketFamily5m {
				continue
			}
			snapshot, err := features.EvaluateSymbolRegime(deterministicRegimeConfig(), bucket, previous)
			if err != nil {
				return symbolAssemblyState{}, err
			}
			state.symbolRegime = snapshot
			copy := snapshot
			previous = &copy
		}
	}
	advanced, err := service.AdvanceWorldUSABuckets(symbol, now)
	if err != nil {
		return symbolAssemblyState{}, err
	}
	state.buckets = append(state.buckets, advanced...)
	for _, bucket := range advanced {
		if bucket.Window.Family != features.BucketFamily5m {
			continue
		}
		snapshot, err := features.EvaluateSymbolRegime(deterministicRegimeConfig(), bucket, previous)
		if err != nil {
			return symbolAssemblyState{}, err
		}
		state.symbolRegime = snapshot
		copy := snapshot
		previous = &copy
	}
	if state.symbolRegime.State == "" {
		pending, err := unavailableSymbolAssemblyState(service, symbol, now, features.RegimeReasonLateWindowIncomplete)
		if err != nil {
			return symbolAssemblyState{}, err
		}
		pending.world = state.world
		pending.usa = state.usa
		pending.buckets = append(state.buckets, pending.buckets...)
		pending.hasData = state.hasData
		return pending, nil
	}
	return state, nil
}

func buildLiveCompositeSnapshots(service *featureengine.Service, observation SpotCurrentStateObservation) (features.CompositeSnapshot, features.CompositeSnapshot, time.Time, time.Time, error) {
	canonical, err := ingestion.ResolveCanonicalTimestamp(observation.ExchangeTs, observation.RecvTs, ingestion.StrictTimestampPolicy())
	if err != nil {
		return features.CompositeSnapshot{}, features.CompositeSnapshot{}, time.Time{}, time.Time{}, err
	}
	health := mergedFeedHealth(observation)
	world, err := service.BuildCompositeSnapshot(features.CompositeGroupWorld, observation.Symbol, canonical.EventTime, []features.ContributorInput{spotContributorInput(observation, canonical.Status, health)})
	if err != nil {
		return features.CompositeSnapshot{}, features.CompositeSnapshot{}, time.Time{}, time.Time{}, err
	}
	usa, err := service.BuildCompositeSnapshot(features.CompositeGroupUSA, observation.Symbol, canonical.EventTime, nil)
	if err != nil {
		return features.CompositeSnapshot{}, features.CompositeSnapshot{}, time.Time{}, time.Time{}, err
	}
	return world, usa, canonical.ExchangeTime, canonical.RecvTime, nil
}

func spotContributorInput(observation SpotCurrentStateObservation, timestampStatus ingestion.CanonicalTimestampStatus, health ingestion.FeedHealthStatus) features.ContributorInput {
	return features.ContributorInput{
		Symbol:          observation.Symbol,
		Venue:           ingestion.VenueBinance,
		MarketType:      "spot",
		QuoteCurrency:   observation.QuoteCurrency,
		Price:           midPrice(observation.BestBidPrice, observation.BestAskPrice),
		LiquidityScore:  100,
		TimestampStatus: timestampStatus,
		FeedHealthState: health.State,
		FeedHealthReasons: append([]ingestion.DegradationReason(nil),
			health.Reasons...),
	}
}

func mergedFeedHealth(observation SpotCurrentStateObservation) ingestion.FeedHealthStatus {
	health := observation.FeedHealth
	if health.State == "" {
		health.State = ingestion.FeedHealthHealthy
		health.ConnectionState = ingestion.ConnectionConnected
		health.MessageFreshness = ingestion.FreshnessFresh
		health.SnapshotFreshness = ingestion.FreshnessNotApplicable
		health.ClockState = ingestion.ClockNormal
	}
	if observation.DepthStatus.State == venuebinance.SpotDepthRecoverySynchronized {
		return health
	}
	reasonSet := make(map[ingestion.DegradationReason]struct{}, len(health.Reasons)+2)
	for _, reason := range health.Reasons {
		reasonSet[reason] = struct{}{}
	}
	if observation.DepthStatus.SequenceGapDetected {
		reasonSet[ingestion.ReasonSequenceGap] = struct{}{}
	}
	if observation.DepthStatus.RefreshDue || observation.DepthStatus.State == venuebinance.SpotDepthRecoveryBootstrapFailed {
		reasonSet[ingestion.ReasonSnapshotStale] = struct{}{}
	}
	health.Reasons = health.Reasons[:0]
	for reason := range reasonSet {
		health.Reasons = append(health.Reasons, reason)
	}
	sort.Slice(health.Reasons, func(i, j int) bool { return health.Reasons[i] < health.Reasons[j] })
	if health.State != ingestion.FeedHealthStale {
		health.State = ingestion.FeedHealthDegraded
	}
	return health
}

func unavailableSymbolAssemblyState(service *featureengine.Service, symbol string, now time.Time, reason features.RegimeReasonCode) (symbolAssemblyState, error) {
	bucketTs := floorBucketTime(now, 30*time.Second)
	world, err := service.BuildCompositeSnapshot(features.CompositeGroupWorld, symbol, bucketTs, nil)
	if err != nil {
		return symbolAssemblyState{}, err
	}
	usa, err := service.BuildCompositeSnapshot(features.CompositeGroupUSA, symbol, bucketTs, nil)
	if err != nil {
		return symbolAssemblyState{}, err
	}
	buckets := unavailableBuckets(symbol, now)
	effectiveEnd := buckets[len(buckets)-1].Window.End
	return symbolAssemblyState{
		world:   world,
		usa:     usa,
		buckets: buckets,
		symbolRegime: features.SymbolRegimeSnapshot{
			SchemaVersion:         "v1",
			Symbol:                symbol,
			State:                 features.RegimeStateNoOperate,
			EffectiveBucketEnd:    effectiveEnd,
			Reasons:               []features.RegimeReasonCode{reason},
			PrimaryReason:         reason,
			TransitionKind:        features.RegimeTransitionHold,
			ConfigVersion:         deterministicRegimeConfig().ConfigVersion,
			AlgorithmVersion:      deterministicRegimeConfig().AlgorithmVersion,
			ObservedInstantaneous: features.RegimeStateNoOperate,
		},
	}, nil
}

func unavailableBuckets(symbol string, now time.Time) []features.MarketQualityBucket {
	config := currentStateBucketConfig()
	return []features.MarketQualityBucket{
		unavailableBucket(config, symbol, features.BucketFamily30s, now),
		unavailableBucket(config, symbol, features.BucketFamily2m, now),
		unavailableBucket(config, symbol, features.BucketFamily5m, now),
	}
}

func unavailableBucket(config features.BucketConfig, symbol string, family features.BucketFamily, now time.Time) features.MarketQualityBucket {
	interval := time.Duration(config.Families[family].IntervalSeconds) * time.Second
	end := floorBucketTime(now, interval)
	if end.Equal(now) {
		end = end.Add(interval)
	}
	start := end.Add(-interval)
	expected := int(interval / (30 * time.Second))
	if expected <= 0 {
		expected = 1
	}
	return features.MarketQualityBucket{
		SchemaVersion: config.SchemaVersion,
		Symbol:        symbol,
		Window: features.BucketWindowSummary{
			Family:              family,
			Start:               start.UTC().Format(time.RFC3339Nano),
			End:                 end.UTC().Format(time.RFC3339Nano),
			ExpectedBucketCount: expected,
			MissingBucketCount:  expected,
			ConfigVersion:       config.ConfigVersion,
			AlgorithmVersion:    config.AlgorithmVersion,
		},
		Assignment: features.BucketAssignment{
			Symbol:       symbol,
			Family:       family,
			BucketStart:  start.UTC().Format(time.RFC3339Nano),
			BucketEnd:    end.UTC().Format(time.RFC3339Nano),
			BucketSource: features.BucketSourceRecvTs,
		},
		World:         features.CompositeBucketSide{Unavailable: true},
		USA:           features.CompositeBucketSide{Unavailable: true},
		Divergence:    features.DivergenceSummary{},
		Fragmentation: features.FragmentationSummary{UnavailableSideCount: 2},
		TimestampTrust: features.TimestampTrustSummary{
			TrustCap: true,
		},
		MarketQuality: features.MarketQualitySummary{CombinedTrustCap: 0, DowngradedReasons: []features.ReasonCode{features.ReasonMissingInput}},
	}
}

func currentStateBucketConfig() features.BucketConfig {
	return features.BucketConfig{
		SchemaVersion:        "v1",
		ConfigVersion:        "market-quality.v1",
		AlgorithmVersion:     "market-quality-buckets.v1",
		TimestampSkewSeconds: 2,
		Families: map[features.BucketFamily]features.BucketFamilyConfig{
			features.BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1},
			features.BucketFamily2m:  {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 0.75},
			features.BucketFamily5m:  {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 0.8},
		},
		Thresholds: features.BucketThresholdConfig{
			Divergence: map[features.BucketFamily]features.DivergenceThresholds{
				features.BucketFamily30s: {PriceDistanceModerateBps: 2, PriceDistanceSevereBps: 8, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				features.BucketFamily2m:  {PriceDistanceModerateBps: 3, PriceDistanceSevereBps: 10, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				features.BucketFamily5m:  {PriceDistanceModerateBps: 4, PriceDistanceSevereBps: 12, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
			},
			Quality: features.MarketQualityThresholds{ConcentrationSoftCap: 0.7, ModerateCap: 0.65, SevereCap: 0.35, TimestampTrustCap: 0.55, IncompleteCap: 0.6},
		},
	}
}

func currentStateSlowContextQuery(symbol string, now time.Time) slowcontext.AssetQuery {
	asset := currentStateAssetFromSymbol(symbol)
	return slowcontext.AssetQuery{
		Asset: asset,
		MetricFamilies: []slowcontext.MetricFamily{
			slowcontext.MetricFamilyCMEVolume,
			slowcontext.MetricFamilyCMEOpenInterest,
			slowcontext.MetricFamilyETFDailyFlow,
		},
		Now: now,
	}
}

func validateSpotObservation(observation SpotCurrentStateObservation) error {
	if observation.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if observation.SourceSymbol == "" {
		return fmt.Errorf("source symbol is required")
	}
	if observation.QuoteCurrency == "" {
		return fmt.Errorf("quote currency is required")
	}
	if observation.BestBidPrice <= 0 || observation.BestAskPrice <= 0 {
		return fmt.Errorf("best bid and ask prices must be positive")
	}
	if observation.RecvTs.IsZero() {
		return fmt.Errorf("recv time is required")
	}
	return nil
}

func isSupportedCurrentStateSymbol(symbol string) bool {
	for _, supported := range supportedCurrentStateSymbols {
		if supported == symbol {
			return true
		}
	}
	return false
}

func floorBucketTime(now time.Time, interval time.Duration) time.Time {
	if interval <= 0 {
		return now.UTC()
	}
	unix := now.UTC().Unix()
	seconds := int64(interval / time.Second)
	return time.Unix((unix/seconds)*seconds, 0).UTC()
}

func midPrice(bid, ask float64) float64 {
	return (bid + ask) / 2
}

func currentStateAssetFromSymbol(symbol string) string {
	if symbol == "" {
		return ""
	}
	parts := strings.Split(symbol, "-")
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}
