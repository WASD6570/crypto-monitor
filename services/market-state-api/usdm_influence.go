package marketstateapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func currentStateUSDMInfluenceSignals(ctx context.Context, featureService *featureengine.Service, reader USDMInfluenceInputReader, now time.Time) (map[string]*features.USDMSymbolInfluenceSignal, error) {
	if reader == nil {
		return map[string]*features.USDMSymbolInfluenceSignal{}, nil
	}
	input, err := reader.SnapshotUSDMInfluenceInput(ctx, now)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return map[string]*features.USDMSymbolInfluenceSignal{}, nil
	}
	signals, err := currentStateUSDMInfluenceSignalsFromInput(featureService, input)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return map[string]*features.USDMSymbolInfluenceSignal{}, nil
	}
	return signals, nil
}

func currentStateUSDMInfluenceSignalsFromInput(featureService *featureengine.Service, input features.USDMInfluenceEvaluatorInput) (map[string]*features.USDMSymbolInfluenceSignal, error) {
	if featureService == nil {
		return nil, fmt.Errorf("feature service is required")
	}
	set, err := featureService.EvaluateUSDMInfluence(input)
	if err != nil {
		return nil, err
	}
	bySymbol := make(map[string]*features.USDMSymbolInfluenceSignal, len(set.Signals))
	for index := range set.Signals {
		signal := set.Signals[index]
		copy := signal
		bySymbol[signal.Symbol] = &copy
	}
	return bySymbol, nil
}

func deterministicUSDMInfluenceInput(now time.Time) features.USDMInfluenceEvaluatorInput {
	observedAt := now.UTC().Format(time.RFC3339Nano)
	return features.USDMInfluenceEvaluatorInput{
		SchemaVersion: features.USDMInfluenceInputSchema,
		ObservedAt:    observedAt,
		Symbols: []features.USDMSymbolInfluenceInput{
			{
				Symbol:        "BTC-USD",
				SourceSymbol:  "BTCUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0008", NextFundingTs: now.Add(4 * time.Hour).UTC().Format(time.RFC3339Nano)},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "64080", IndexPrice: "64000"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, Side: "sell", Price: "64000", Size: "2"},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, OpenInterest: "10659.509"},
			},
			{
				Symbol:        "ETH-USD",
				SourceSymbol:  "ETHUSDT",
				QuoteCurrency: "USDT",
				Funding:       features.USDMFundingInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, FundingRate: "0.0001", NextFundingTs: now.Add(4 * time.Hour).UTC().Format(time.RFC3339Nano)},
				MarkIndex:     features.USDMMarkIndexInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1000}, MarkPrice: "3200.3", IndexPrice: "3200.1"},
				Liquidation:   features.USDMLiquidationInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceWebsocket, Freshness: features.USDMInfluenceFreshnessUnavailable}},
				OpenInterest:  features.USDMOpenInterestInput{Metadata: features.USDMInfluenceInputMetadata{Surface: features.USDMInfluenceSurfaceRESTPoll, Available: true, Freshness: features.USDMInfluenceFreshnessFresh, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, AgeMillis: 1500}, OpenInterest: "82450.5"},
			},
		},
	}
}
