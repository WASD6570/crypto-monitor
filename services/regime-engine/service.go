package regimeengine

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

type Service struct {
	config  features.RegimeConfig
	symbols map[string]features.SymbolRegimeSnapshot
	global  *features.GlobalRegimeSnapshot
}

func NewService(config features.RegimeConfig) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Service{config: config, symbols: map[string]features.SymbolRegimeSnapshot{}}, nil
}

func (s *Service) Observe(bucket features.MarketQualityBucket) (features.RegimeEvaluation, error) {
	if s == nil {
		return features.RegimeEvaluation{}, fmt.Errorf("regime engine service is required")
	}
	if bucket.Window.Family != features.BucketFamily5m {
		return features.RegimeEvaluation{}, fmt.Errorf("regime engine requires 5m buckets")
	}
	if !supportedSymbol(s.config.Symbols, bucket.Symbol) {
		return features.RegimeEvaluation{}, fmt.Errorf("symbol %q is not configured", bucket.Symbol)
	}
	var previousSymbol *features.SymbolRegimeSnapshot
	if snapshot, ok := s.symbols[bucket.Symbol]; ok {
		copy := snapshot
		previousSymbol = &copy
	}
	symbolSnapshot, err := features.EvaluateSymbolRegime(s.config, bucket, previousSymbol)
	if err != nil {
		return features.RegimeEvaluation{}, err
	}
	s.symbols[bucket.Symbol] = symbolSnapshot
	globalSnapshot, err := features.EvaluateGlobalRegime(s.config, s.symbols, s.global)
	if err != nil {
		return features.RegimeEvaluation{}, err
	}
	s.global = &globalSnapshot
	return features.RegimeEvaluation{
		Symbol:         symbolSnapshot,
		Global:         globalSnapshot,
		EffectiveState: features.EffectiveStates(s.config.Symbols, s.symbols, globalSnapshot),
	}, nil
}

func supportedSymbol(symbols []string, target string) bool {
	for _, symbol := range symbols {
		if symbol == target {
			return true
		}
	}
	return false
}

func (s *Service) QueryCurrentGlobalState(query features.GlobalCurrentStateQuery) (features.MarketStateCurrentGlobalResponse, error) {
	if s == nil {
		return features.MarketStateCurrentGlobalResponse{}, fmt.Errorf("regime engine service is required")
	}
	return features.BuildMarketStateCurrentGlobalResponse(query)
}
