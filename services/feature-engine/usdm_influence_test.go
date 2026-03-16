package featureengine

import (
	"reflect"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestEvaluateUSDMInfluenceNoContextPosture(t *testing.T) {
	service := newService(t)
	actual, err := service.EvaluateUSDMInfluence(usdmInfluenceInputFixture())
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	if actual.Signals[0].Posture != features.USDMInfluencePostureNoContext {
		t.Fatalf("btc posture = %q", actual.Signals[0].Posture)
	}
	if actual.Signals[0].PrimaryReason != features.USDMInfluenceReasonNoContext {
		t.Fatalf("btc reason = %q", actual.Signals[0].PrimaryReason)
	}
}

func TestEvaluateUSDMInfluenceMixedContextYieldsStableDegradedReasons(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0].Funding = features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessDegraded, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded}, FundingRate: "0.0008"}
	input.Symbols[0].MarkIndex = features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessDegraded, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded}, MarkPrice: "64020", IndexPrice: "64000"}
	input.Symbols[0].OpenInterest = features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessDegraded, TimestampStatus: ingestion.TimestampStatusDegraded, FeedHealthState: ingestion.FeedHealthHealthy}, OpenInterest: "10659.509"}

	actual, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	btc := actual.Signals[0]
	if btc.Posture != features.USDMInfluencePostureDegradedContext {
		t.Fatalf("btc posture = %q", btc.Posture)
	}
	wantReasons := []features.USDMInfluenceReasonCode{
		features.USDMInfluenceReasonWebsocketDegraded,
		features.USDMInfluenceReasonOpenInterestDegraded,
		features.USDMInfluenceReasonTimestampDegraded,
	}
	if !reflect.DeepEqual(btc.Reasons, wantReasons) {
		t.Fatalf("btc reasons = %v, want %v", btc.Reasons, wantReasons)
	}
}

func TestEvaluateUSDMInfluenceFreshContextCanEmitDegradeCap(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0] = features.USDMSymbolInfluenceInput{
		Symbol:        "BTC-USD",
		SourceSymbol:  "BTCUSDT",
		QuoteCurrency: "USDT",
		Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0009", NextFundingTs: "2026-03-15T16:00:00Z"},
		MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64080", IndexPrice: "64000"},
		Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, Side: "sell", Price: "64000", Size: "2"},
		OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2000}, OpenInterest: "10659.509"},
	}
	input.Symbols[1] = input.Symbols[0]
	input.Symbols[1].Symbol = "ETH-USD"
	input.Symbols[1].SourceSymbol = "ETHUSDT"

	actual, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	if actual.Signals[0].Posture != features.USDMInfluencePostureDegradeCap {
		t.Fatalf("btc posture = %q", actual.Signals[0].Posture)
	}
	if !reflect.DeepEqual(actual.Signals[0].Reasons, []features.USDMInfluenceReasonCode{features.USDMInfluenceReasonBasisWide, features.USDMInfluenceReasonFundingElevated, features.USDMInfluenceReasonRecentLiquidation}) {
		t.Fatalf("btc reasons = %v", actual.Signals[0].Reasons)
	}
}

func TestEvaluateUSDMInfluenceRepeatedInputIsDeterministic(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0].Funding = features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0001"}
	input.Symbols[0].MarkIndex = features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64001", IndexPrice: "64000"}
	input.Symbols[0].OpenInterest = features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2000}, OpenInterest: "10659.509"}
	input.Symbols[1] = input.Symbols[0]
	input.Symbols[1].Symbol = "ETH-USD"
	input.Symbols[1].SourceSymbol = "ETHUSDT"

	first, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("first evaluate influence: %v", err)
	}
	second, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("second evaluate influence: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated influence signals differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestEvaluateUSDMInfluenceOnlyLiquidationStillReturnsNoContext(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0].Liquidation = features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 500}, Side: "sell", Price: "64000", Size: "1"}
	actual, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	if actual.Signals[0].Posture != features.USDMInfluencePostureNoContext {
		t.Fatalf("btc posture = %q, want %q", actual.Signals[0].Posture, features.USDMInfluencePostureNoContext)
	}
}

func TestEvaluateUSDMInfluenceRejectsNonFiniteNumericInputs(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0].Funding = features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh}, FundingRate: "NaN"}
	if _, err := service.EvaluateUSDMInfluence(input); err == nil {
		t.Fatal("expected non-finite numeric input to fail")
	}
}

func TestEvaluateUSDMInfluenceUnknownLiquidationAgeDoesNotTriggerDegradeCap(t *testing.T) {
	service := newService(t)
	input := usdmInfluenceInputFixture()
	input.Symbols[0].Funding = features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0001"}
	input.Symbols[0].MarkIndex = features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64001", IndexPrice: "64000"}
	input.Symbols[0].OpenInterest = features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 2000}, OpenInterest: "10659.509"}
	input.Symbols[0].Liquidation = features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: -1}, Side: "sell", Price: "64000", Size: "2"}
	actual, err := service.EvaluateUSDMInfluence(input)
	if err != nil {
		t.Fatalf("evaluate influence: %v", err)
	}
	if actual.Signals[0].Posture != features.USDMInfluencePostureAuxiliary {
		t.Fatalf("btc posture = %q, want %q", actual.Signals[0].Posture, features.USDMInfluencePostureAuxiliary)
	}
}

func usdmInfluenceInputFixture() features.USDMInfluenceEvaluatorInput {
	return features.USDMInfluenceEvaluatorInput{
		SchemaVersion: features.USDMInfluenceInputSchema,
		ObservedAt:    "2026-03-15T12:00:00Z",
		Symbols: []features.USDMSymbolInfluenceInput{
			{
				Symbol:        "BTC-USD",
				SourceSymbol:  "BTCUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Freshness: features.USDMInfluenceFreshnessUnavailable}},
			},
			{
				Symbol:        "ETH-USD",
				SourceSymbol:  "ETHUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Freshness: features.USDMInfluenceFreshnessUnavailable}},
			},
		},
	}
}
