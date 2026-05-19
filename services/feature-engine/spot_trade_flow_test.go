package featureengine

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestSpotTradeFlowServiceWrapper(t *testing.T) {
	service := newSpotTradeFlowService(t)
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	first, err := service.ObserveSpotTradeFlow(serviceSpotTradeFlowObservation("BTC-USD", "trade:1001", features.SpotTradeFlowSideBuy, 64000, 0.5, start.Add(10*time.Second)))
	if err != nil {
		t.Fatalf("observe first trade: %v", err)
	}
	if !first.Accepted || first.Duplicate {
		t.Fatalf("first result = %+v, want accepted non-duplicate", first)
	}
	second, err := service.ObserveSpotTradeFlow(serviceSpotTradeFlowObservation("ETH-USD", "trade:2001", features.SpotTradeFlowSideSell, 3500, 2, start.Add(20*time.Second)))
	if err != nil {
		t.Fatalf("observe second trade: %v", err)
	}
	if !second.Accepted {
		t.Fatalf("second result = %+v, want accepted", second)
	}
	duplicate, err := service.ObserveSpotTradeFlow(serviceSpotTradeFlowObservation("BTC-USD", "trade:1001", features.SpotTradeFlowSideBuy, 64000, 0.5, start.Add(10*time.Second)))
	if err != nil {
		t.Fatalf("observe duplicate trade: %v", err)
	}
	if duplicate.Accepted || !duplicate.Duplicate {
		t.Fatalf("duplicate result = %+v, want rejected duplicate", duplicate)
	}

	snapshot, err := service.SpotTradeFlowSnapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot) != 6 {
		t.Fatalf("snapshot bucket count = %d, want 6", len(snapshot))
	}
	if snapshot[0].Symbol != "BTC-USD" || snapshot[0].Family != features.BucketFamily30s || snapshot[3].Symbol != "ETH-USD" {
		t.Fatalf("snapshot order = %+v, want deterministic symbol/family order", snapshot)
	}
	btcOnly, err := service.SpotTradeFlowSnapshot("BTC-USD")
	if err != nil {
		t.Fatalf("BTC snapshot: %v", err)
	}
	if len(btcOnly) != 3 || btcOnly[0].DuplicateCount != 1 {
		t.Fatalf("BTC snapshot = %+v, want three buckets with duplicate count", btcOnly)
	}
	again, err := service.SpotTradeFlowSnapshot()
	if err != nil {
		t.Fatalf("second snapshot: %v", err)
	}
	if !reflect.DeepEqual(snapshot, again) {
		t.Fatalf("repeated snapshots differ:\nfirst=%+v\nsecond=%+v", snapshot, again)
	}
}

func TestSpotTradeFlowServiceRequiresProcessor(t *testing.T) {
	service := newService(t)
	_, observeErr := service.ObserveSpotTradeFlow(serviceSpotTradeFlowObservation("BTC-USD", "trade:1001", features.SpotTradeFlowSideBuy, 64000, 0.5, time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC)))
	if observeErr == nil {
		t.Fatal("expected observe error when spot trade-flow processor is not configured")
	}
	if _, snapshotErr := service.SpotTradeFlowSnapshot(); snapshotErr == nil {
		t.Fatal("expected snapshot error when spot trade-flow processor is not configured")
	}
}

func newSpotTradeFlowService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties:        features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75},
		QuoteProxies: map[string]features.QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
			"USDC": {Enabled: true, PenaltyMultiplier: 0.98},
		},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {
				Members: []features.MemberConfig{{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			features.CompositeGroupUSA: {
				Members: []features.MemberConfig{{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.7},
			},
		},
	}, WithSpotTradeFlowConfig(features.DefaultSpotTradeFlowConfig()))
	if err != nil {
		t.Fatalf("new spot trade-flow service: %v", err)
	}
	return service
}

func serviceSpotTradeFlowObservation(symbol, sourceRecordID, side string, price, size float64, exchangeTs time.Time) features.SpotTradeFlowObservation {
	sourceSymbol := "BTCUSDT"
	if symbol == "ETH-USD" {
		sourceSymbol = "ETHUSDT"
	}
	return features.SpotTradeFlowObservation{
		Symbol:          symbol,
		Venue:           ingestion.VenueBinance,
		MarketType:      features.SpotTradeFlowMarketType,
		SourceSymbol:    sourceSymbol,
		SourceRecordID:  sourceRecordID,
		Side:            side,
		Price:           price,
		Size:            size,
		ExchangeTs:      exchangeTs,
		RecvTs:          exchangeTs.Add(100 * time.Millisecond),
		TimestampStatus: ingestion.TimestampStatusNormal,
		FeedHealthState: ingestion.FeedHealthHealthy,
	}
}
