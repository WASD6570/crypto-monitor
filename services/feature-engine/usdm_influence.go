package featureengine

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func (s *Service) EvaluateUSDMInfluence(input features.USDMInfluenceEvaluatorInput) (features.USDMInfluenceSignalSet, error) {
	if s == nil {
		return features.USDMInfluenceSignalSet{}, fmt.Errorf("feature engine service is required")
	}
	if err := input.Validate(); err != nil {
		return features.USDMInfluenceSignalSet{}, err
	}
	if err := s.usdmInfluence.Validate(); err != nil {
		return features.USDMInfluenceSignalSet{}, err
	}
	observedAt, err := time.Parse(time.RFC3339Nano, input.ObservedAt)
	if err != nil {
		return features.USDMInfluenceSignalSet{}, err
	}
	signals := make([]features.USDMSymbolInfluenceSignal, 0, len(input.Symbols))
	for _, symbol := range input.Symbols {
		signal, err := evaluateUSDMSymbolInfluence(s.usdmInfluence, observedAt.UTC(), symbol)
		if err != nil {
			return features.USDMInfluenceSignalSet{}, err
		}
		signals = append(signals, signal)
	}
	set := features.USDMInfluenceSignalSet{
		SchemaVersion:    features.USDMInfluenceSignalSchema,
		ObservedAt:       observedAt.UTC().Format(time.RFC3339Nano),
		Signals:          signals,
		ConfigVersion:    s.usdmInfluence.ConfigVersion,
		AlgorithmVersion: s.usdmInfluence.AlgorithmVersion,
	}
	if err := set.Validate(); err != nil {
		return features.USDMInfluenceSignalSet{}, err
	}
	return set, nil
}

func evaluateUSDMSymbolInfluence(config features.USDMInfluenceConfig, observedAt time.Time, input features.USDMSymbolInfluenceInput) (features.USDMSymbolInfluenceSignal, error) {
	if observedAt.IsZero() {
		return features.USDMSymbolInfluenceSignal{}, fmt.Errorf("observedAt is required")
	}
	metrics, err := buildUSDMInfluenceMetrics(config, input)
	if err != nil {
		return features.USDMSymbolInfluenceSignal{}, err
	}
	posture := features.USDMInfluencePostureAuxiliary
	reasons := []features.USDMInfluenceReasonCode{features.USDMInfluenceReasonHealthy}

	if inputMissingAllContext(input) {
		posture = features.USDMInfluencePostureNoContext
		reasons = []features.USDMInfluenceReasonCode{features.USDMInfluenceReasonNoContext}
	} else if degradedReasons := degradedContextReasons(input); len(degradedReasons) > 0 {
		posture = features.USDMInfluencePostureDegradedContext
		reasons = degradedReasons
	} else if degradeCapReasons := degradeCapReasons(config, input, metrics); len(degradeCapReasons) > 0 {
		posture = features.USDMInfluencePostureDegradeCap
		reasons = degradeCapReasons
	}

	signal := features.USDMSymbolInfluenceSignal{
		SchemaVersion:    features.USDMInfluenceSignalSchema,
		Symbol:           input.Symbol,
		SourceSymbol:     input.SourceSymbol,
		QuoteCurrency:    input.QuoteCurrency,
		Posture:          posture,
		Reasons:          reasons,
		PrimaryReason:    reasons[0],
		TriggerMetrics:   metrics,
		ConfigVersion:    config.ConfigVersion,
		AlgorithmVersion: config.AlgorithmVersion,
		ObservedAt:       observedAt.Format(time.RFC3339Nano),
	}
	return signal, nil
}

func inputMissingAllContext(input features.USDMSymbolInfluenceInput) bool {
	return !input.Funding.Metadata.Available && !input.MarkIndex.Metadata.Available && !input.OpenInterest.Metadata.Available
}

func degradedContextReasons(input features.USDMSymbolInfluenceInput) []features.USDMInfluenceReasonCode {
	reasonSet := map[features.USDMInfluenceReasonCode]struct{}{}
	if !input.Funding.Metadata.Available {
		reasonSet[features.USDMInfluenceReasonFundingMissing] = struct{}{}
	}
	if !input.MarkIndex.Metadata.Available {
		reasonSet[features.USDMInfluenceReasonMarkIndexMissing] = struct{}{}
	}
	if !input.OpenInterest.Metadata.Available {
		reasonSet[features.USDMInfluenceReasonOpenInterestMissing] = struct{}{}
	}
	if websocketStale(input) {
		reasonSet[features.USDMInfluenceReasonWebsocketStale] = struct{}{}
	} else if websocketDegraded(input) {
		reasonSet[features.USDMInfluenceReasonWebsocketDegraded] = struct{}{}
	}
	if input.OpenInterest.Metadata.Freshness == features.USDMInfluenceFreshnessStale {
		reasonSet[features.USDMInfluenceReasonOpenInterestStale] = struct{}{}
	} else if input.OpenInterest.Metadata.Freshness == features.USDMInfluenceFreshnessDegraded {
		reasonSet[features.USDMInfluenceReasonOpenInterestDegraded] = struct{}{}
	}
	if timestampDegraded(input) {
		reasonSet[features.USDMInfluenceReasonTimestampDegraded] = struct{}{}
	}
	return sortedUSDMInfluenceReasons(reasonSet)
}

func degradeCapReasons(config features.USDMInfluenceConfig, input features.USDMSymbolInfluenceInput, metrics []features.USDMInfluenceMetric) []features.USDMInfluenceReasonCode {
	reasonSet := map[features.USDMInfluenceReasonCode]struct{}{}
	if metricValue(metrics, "basisAbsBps") >= config.BasisAbsBpsDegradeCap {
		reasonSet[features.USDMInfluenceReasonBasisWide] = struct{}{}
	}
	if metricValue(metrics, "fundingAbsBps") >= config.FundingAbsBpsDegradeCap {
		reasonSet[features.USDMInfluenceReasonFundingElevated] = struct{}{}
	}
	ageMillis := metricInteger(metrics, "recentLiquidationAgeMillis")
	if input.Liquidation.Metadata.Available && ageMillis >= 0 && metricValue(metrics, "recentLiquidationNotionalUsd") >= config.LiquidationNotionalDegradeCapUSD && ageMillis <= int64(config.LiquidationRecentWindowSeconds)*int64(time.Second/time.Millisecond) {
		reasonSet[features.USDMInfluenceReasonRecentLiquidation] = struct{}{}
	}
	return sortedUSDMInfluenceReasons(reasonSet)
}

func websocketStale(input features.USDMSymbolInfluenceInput) bool {
	return input.Funding.Metadata.Freshness == features.USDMInfluenceFreshnessStale || input.MarkIndex.Metadata.Freshness == features.USDMInfluenceFreshnessStale
}

func websocketDegraded(input features.USDMSymbolInfluenceInput) bool {
	return input.Funding.Metadata.Freshness == features.USDMInfluenceFreshnessDegraded || input.MarkIndex.Metadata.Freshness == features.USDMInfluenceFreshnessDegraded
}

func timestampDegraded(input features.USDMSymbolInfluenceInput) bool {
	return input.Funding.Metadata.TimestampStatus == ingestion.TimestampStatusDegraded || input.MarkIndex.Metadata.TimestampStatus == ingestion.TimestampStatusDegraded || input.OpenInterest.Metadata.TimestampStatus == ingestion.TimestampStatusDegraded || input.Liquidation.Metadata.TimestampStatus == ingestion.TimestampStatusDegraded
}

func buildUSDMInfluenceMetrics(config features.USDMInfluenceConfig, input features.USDMSymbolInfluenceInput) ([]features.USDMInfluenceMetric, error) {
	basisAbsBps, err := basisAbsBps(input.MarkIndex)
	if err != nil {
		return nil, err
	}
	fundingAbsBps, err := fundingAbsBps(input.Funding)
	if err != nil {
		return nil, err
	}
	openInterestValue, err := numericString(input.OpenInterest.OpenInterest)
	if err != nil {
		return nil, err
	}
	liquidationNotional, err := liquidationNotional(input.Liquidation)
	if err != nil {
		return nil, err
	}
	metrics := []features.USDMInfluenceMetric{
		{Name: "basisAbsBps", Value: basisAbsBps},
		{Name: "fundingAbsBps", Value: fundingAbsBps},
		{Name: "openInterest", Value: openInterestValue},
		{Name: "websocketFreshness", Text: string(worstFreshness(input.Funding.Metadata.Freshness, input.MarkIndex.Metadata.Freshness))},
		{Name: "openInterestFreshness", Text: string(input.OpenInterest.Metadata.Freshness)},
		{Name: "fundingAgeMillis", Integer: input.Funding.Metadata.AgeMillis},
		{Name: "markIndexAgeMillis", Integer: input.MarkIndex.Metadata.AgeMillis},
		{Name: "openInterestAgeMillis", Integer: input.OpenInterest.Metadata.AgeMillis},
		{Name: "recentLiquidationAgeMillis", Integer: input.Liquidation.Metadata.AgeMillis},
		{Name: "recentLiquidationNotionalUsd", Value: liquidationNotional},
		{Name: "liquidationRecentWindowMillis", Integer: int64(config.LiquidationRecentWindowSeconds) * int64(time.Second/time.Millisecond)},
	}
	return metrics, nil
}

func basisAbsBps(input features.USDMMarkIndexInput) (float64, error) {
	mark, err := numericString(input.MarkPrice)
	if err != nil {
		return 0, err
	}
	index, err := numericString(input.IndexPrice)
	if err != nil {
		return 0, err
	}
	if mark == 0 || index == 0 {
		return 0, nil
	}
	return math.Abs(mark-index) / index * 10000, nil
}

func fundingAbsBps(input features.USDMFundingInput) (float64, error) {
	rate, err := numericString(input.FundingRate)
	if err != nil {
		return 0, err
	}
	return math.Abs(rate) * 10000, nil
}

func liquidationNotional(input features.USDMLiquidationInput) (float64, error) {
	price, err := numericString(input.Price)
	if err != nil {
		return 0, err
	}
	size, err := numericString(input.Size)
	if err != nil {
		return 0, err
	}
	return math.Abs(price * size), nil
}

func numericString(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse numeric value %q: %w", value, err)
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("numeric value %q must be finite", value)
	}
	return parsed, nil
}

func metricValue(metrics []features.USDMInfluenceMetric, name string) float64 {
	for _, metric := range metrics {
		if metric.Name == name {
			return metric.Value
		}
	}
	return 0
}

func metricInteger(metrics []features.USDMInfluenceMetric, name string) int64 {
	for _, metric := range metrics {
		if metric.Name == name {
			return metric.Integer
		}
	}
	return 0
}

func worstFreshness(values ...features.USDMInfluenceFreshness) features.USDMInfluenceFreshness {
	result := features.USDMInfluenceFreshnessFresh
	for _, value := range values {
		switch value {
		case features.USDMInfluenceFreshnessStale:
			return value
		case features.USDMInfluenceFreshnessDegraded:
			result = value
		case features.USDMInfluenceFreshnessUnavailable:
			if result == features.USDMInfluenceFreshnessFresh {
				result = value
			}
		}
	}
	return result
}

func sortedUSDMInfluenceReasons(reasonSet map[features.USDMInfluenceReasonCode]struct{}) []features.USDMInfluenceReasonCode {
	if len(reasonSet) == 0 {
		return nil
	}
	ordered := make([]features.USDMInfluenceReasonCode, 0, len(reasonSet))
	for reason := range reasonSet {
		ordered = append(ordered, reason)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return usdmInfluenceReasonPriority(ordered[i]) < usdmInfluenceReasonPriority(ordered[j])
	})
	return ordered
}

func usdmInfluenceReasonPriority(reason features.USDMInfluenceReasonCode) int {
	priorities := map[features.USDMInfluenceReasonCode]int{
		features.USDMInfluenceReasonNoContext:            0,
		features.USDMInfluenceReasonFundingMissing:       1,
		features.USDMInfluenceReasonMarkIndexMissing:     2,
		features.USDMInfluenceReasonOpenInterestMissing:  3,
		features.USDMInfluenceReasonWebsocketStale:       4,
		features.USDMInfluenceReasonWebsocketDegraded:    5,
		features.USDMInfluenceReasonOpenInterestStale:    6,
		features.USDMInfluenceReasonOpenInterestDegraded: 7,
		features.USDMInfluenceReasonTimestampDegraded:    8,
		features.USDMInfluenceReasonBasisWide:            9,
		features.USDMInfluenceReasonFundingElevated:      10,
		features.USDMInfluenceReasonRecentLiquidation:    11,
		features.USDMInfluenceReasonHealthy:              12,
	}
	if priority, ok := priorities[reason]; ok {
		return priority
	}
	return 100
}
