package integration

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestIngestionBinanceUSDMInfluenceKeepsAuxiliaryOutputStable(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	query := tradeableCurrentStateQueryFixture("BTC-USD")
	baseline, err := featureService.QueryCurrentState(query)
	if err != nil {
		t.Fatalf("baseline current state: %v", err)
	}
	signals, err := featureService.EvaluateUSDMInfluence(integrationUSDMInfluenceInputFixture())
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	query.SymbolRegime, query.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(query.SymbolRegime, signalBySymbol(t, signals, "BTC-USD"))
	if err != nil {
		t.Fatalf("apply influence: %v", err)
	}
	after, err := featureService.QueryCurrentState(query)
	if err != nil {
		t.Fatalf("current state after influence application: %v", err)
	}
	if after.Regime.Symbol.State != baseline.Regime.Symbol.State || after.Regime.EffectiveState != baseline.Regime.EffectiveState {
		t.Fatalf("auxiliary influence changed output\nbaseline: %+v\nafter: %+v", baseline.Regime, after.Regime)
	}
	if after.Provenance.USDMInfluence == nil || after.Provenance.USDMInfluence.AppliedCap {
		t.Fatalf("expected auxiliary provenance without cap, got %+v", after.Provenance.USDMInfluence)
	}
	if after.Provenance.USDMInfluence.Posture != features.USDMInfluencePostureAuxiliary {
		t.Fatalf("posture = %q", after.Provenance.USDMInfluence.Posture)
	}
}

func TestIngestionBinanceUSDMInfluenceAppliesBoundedWatchCap(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	query := tradeableCurrentStateQueryFixture("BTC-USD")
	signals, err := featureService.EvaluateUSDMInfluence(integrationDegradeCapUSDMInfluenceInputFixture())
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	query.SymbolRegime, query.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(query.SymbolRegime, signalBySymbol(t, signals, "BTC-USD"))
	if err != nil {
		t.Fatalf("apply influence: %v", err)
	}
	response, err := featureService.QueryCurrentState(query)
	if err != nil {
		t.Fatalf("query current state: %v", err)
	}
	if response.Regime.Symbol.State != features.RegimeStateWatch {
		t.Fatalf("symbol state = %q, want %q", response.Regime.Symbol.State, features.RegimeStateWatch)
	}
	if response.Regime.EffectiveState != features.RegimeStateWatch {
		t.Fatalf("effective state = %q, want %q", response.Regime.EffectiveState, features.RegimeStateWatch)
	}
	if !reflect.DeepEqual(response.Regime.Symbol.Reasons, []features.RegimeReasonCode{features.RegimeReasonUSDMInfluenceCap, features.RegimeReasonHealthy}) {
		t.Fatalf("symbol reasons = %+v", response.Regime.Symbol.Reasons)
	}
	if response.Provenance.USDMInfluence == nil || !response.Provenance.USDMInfluence.AppliedCap {
		t.Fatalf("expected applied cap provenance, got %+v", response.Provenance.USDMInfluence)
	}
	if response.Provenance.USDMInfluence.PrimaryReason != features.USDMInfluenceReasonBasisWide {
		t.Fatalf("primary reason = %q", response.Provenance.USDMInfluence.PrimaryReason)
	}
}

func TestIngestionBinanceUSDMInfluenceCanCapGlobalResponse(t *testing.T) {
	featureService := newCurrentStateFeatureService(t)
	regimeService := newCurrentStateRegimeService(t)
	signals, err := featureService.EvaluateUSDMInfluence(integrationDegradeCapUSDMInfluenceInputFixture())
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	btcQuery := tradeableCurrentStateQueryFixture("BTC-USD")
	ethQuery := tradeableCurrentStateQueryFixture("ETH-USD")
	btcQuery.SymbolRegime, btcQuery.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(btcQuery.SymbolRegime, signalBySymbol(t, signals, "BTC-USD"))
	if err != nil {
		t.Fatalf("apply btc influence: %v", err)
	}
	ethQuery.SymbolRegime, ethQuery.USDMInfluence, err = features.ApplyUSDMInfluenceToSymbolRegime(ethQuery.SymbolRegime, signalBySymbol(t, signals, "ETH-USD"))
	if err != nil {
		t.Fatalf("apply eth influence: %v", err)
	}
	globalRegime, err := features.EvaluateGlobalRegime(features.RegimeConfig{SchemaVersion: "v1", ConfigVersion: "regime-engine.market-state.v1", AlgorithmVersion: "symbol-global-regime.v1", Symbols: []string{"BTC-USD", "ETH-USD"}, Symbol: features.SymbolRegimeThresholds{CoverageWatchMax: 0.85, CoverageNoOperateMax: 0.60, CombinedTrustCapWatchMax: 0.65, CombinedTrustCapNoOperateMax: 0.35, TimestampFallbackWatchRatio: 0.10, TimestampFallbackNoOpRatio: 0.50, NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2}, Global: features.GlobalRegimeThresholds{NoOperateToWatchWindows: 2, WatchToTradeableWindows: 2}}, map[string]features.SymbolRegimeSnapshot{
		"BTC-USD": btcQuery.SymbolRegime,
		"ETH-USD": ethQuery.SymbolRegime,
	}, nil)
	if err != nil {
		t.Fatalf("evaluate global regime: %v", err)
	}
	btcQuery.GlobalRegime = globalRegime
	ethQuery.GlobalRegime = globalRegime
	btc, err := featureService.QueryCurrentState(btcQuery)
	if err != nil {
		t.Fatalf("query btc current state: %v", err)
	}
	eth, err := featureService.QueryCurrentState(ethQuery)
	if err != nil {
		t.Fatalf("query eth current state: %v", err)
	}
	global, err := regimeService.QueryCurrentGlobalState(features.GlobalCurrentStateQuery{GlobalRegime: globalRegime, Symbols: []features.MarketStateCurrentResponse{btc, eth}})
	if err != nil {
		t.Fatalf("query current global state: %v", err)
	}
	if global.Global.State != features.RegimeStateWatch {
		t.Fatalf("global state = %q, want %q", global.Global.State, features.RegimeStateWatch)
	}
	if !reflect.DeepEqual(global.Global.Reasons, []features.RegimeReasonCode{features.RegimeReasonGlobalSharedWatch, features.RegimeReasonUSDMInfluenceCap}) {
		t.Fatalf("global reasons = %+v", global.Global.Reasons)
	}
	if global.Symbols[0].EffectiveState != features.RegimeStateWatch || global.Symbols[1].EffectiveState != features.RegimeStateWatch {
		t.Fatalf("expected watch symbol summaries, got %+v", global.Symbols)
	}
}

func signalBySymbol(t *testing.T, signals features.USDMInfluenceSignalSet, symbol string) *features.USDMSymbolInfluenceSignal {
	t.Helper()
	for index := range signals.Signals {
		if signals.Signals[index].Symbol == symbol {
			copy := signals.Signals[index]
			return &copy
		}
	}
	t.Fatalf("missing signal for %s", symbol)
	return nil
}

func tradeableCurrentStateQueryFixture(symbol string) features.SymbolCurrentStateQuery {
	query := currentStateQueryFixture()
	query.Symbol = symbol
	query.World.Symbol = symbol
	query.USA.Symbol = symbol
	for index := range query.Buckets {
		query.Buckets[index].Symbol = symbol
	}
	for index := range query.RecentContext {
		query.RecentContext[index].Symbol = symbol
	}
	query.SymbolRegime.Symbol = symbol
	query.SymbolRegime.State = features.RegimeStateTradeable
	query.SymbolRegime.ObservedInstantaneous = features.RegimeStateTradeable
	query.SymbolRegime.Reasons = []features.RegimeReasonCode{features.RegimeReasonHealthy}
	query.SymbolRegime.PrimaryReason = features.RegimeReasonHealthy
	query.GlobalRegime.State = features.RegimeStateTradeable
	query.GlobalRegime.ObservedInstantaneous = features.RegimeStateTradeable
	query.GlobalRegime.Reasons = []features.RegimeReasonCode{features.RegimeReasonHealthy}
	query.GlobalRegime.PrimaryReason = features.RegimeReasonHealthy
	return query
}

func integrationUSDMInfluenceInputFixture() features.USDMInfluenceEvaluatorInput {
	return features.USDMInfluenceEvaluatorInput{
		SchemaVersion: features.USDMInfluenceInputSchema,
		ObservedAt:    "2026-03-15T12:00:00Z",
		Symbols: []features.USDMSymbolInfluenceInput{
			{
				Symbol:        "BTC-USD",
				SourceSymbol:  "BTCUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0002", NextFundingTs: "2026-03-15T16:00:00Z"},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64003", IndexPrice: "64000"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2500}, OpenInterest: "10659.509"},
			},
			{
				Symbol:        "ETH-USD",
				SourceSymbol:  "ETHUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, FundingRate: "0.0001", NextFundingTs: "2026-03-15T16:00:00Z"},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, MarkPrice: "3200.3", IndexPrice: "3200.1"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2600}, OpenInterest: "82450.5"},
			},
		},
	}
}

func integrationDegradeCapUSDMInfluenceInputFixture() features.USDMInfluenceEvaluatorInput {
	input := integrationUSDMInfluenceInputFixture()
	input.Symbols[0].Funding.FundingRate = "0.0009"
	input.Symbols[0].MarkIndex.MarkPrice = "64080"
	input.Symbols[0].Liquidation = features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, Side: "sell", Price: "64000", Size: "2"}
	input.Symbols[1].Funding.FundingRate = "0.0008"
	input.Symbols[1].MarkIndex.MarkPrice = "3210.0"
	input.Symbols[1].Liquidation = features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, Side: "sell", Price: "3200", Size: "40"}
	return input
}
