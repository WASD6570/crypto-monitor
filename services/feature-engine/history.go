package featureengine

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func (s *Service) QueryHistoricalState(query features.SymbolHistoricalStateQuery) (features.MarketStateHistorySymbolResponse, error) {
	if s == nil {
		return features.MarketStateHistorySymbolResponse{}, fmt.Errorf("feature engine service is required")
	}
	return features.BuildMarketStateHistorySymbolResponse(query)
}
