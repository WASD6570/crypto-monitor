package replayengine

import "github.com/crypto-market-copilot/alerts/libs/go/features"

func QueryMarketStateAuditProvenance(query features.MarketStateAuditQuery) (features.MarketStateAuditProvenanceResponse, error) {
	return features.BuildMarketStateAuditProvenance(query)
}
