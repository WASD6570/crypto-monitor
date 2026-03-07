package regimeengine

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

func (s *Service) QueryHistoricalGlobalState(query features.GlobalHistoricalStateQuery) (features.MarketStateHistoryGlobalResponse, error) {
	if s == nil {
		return features.MarketStateHistoryGlobalResponse{}, fmt.Errorf("regime engine service is required")
	}
	return features.BuildMarketStateHistoryGlobalResponse(query)
}
