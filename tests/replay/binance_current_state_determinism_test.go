package replay

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestReplayBinanceMarketStateDeterminism(t *testing.T) {
	first := replayBinanceCurrentState(t)
	second := replayBinanceCurrentState(t)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("binance current-state outputs differ\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func replayBinanceCurrentState(t *testing.T) marketstateapi.SymbolStateResponse {
	t.Helper()
	provider, err := marketstateapi.NewLiveSpotProvider(replaySpotReader{snapshot: replaySpotSnapshot()}, func() time.Time {
		return time.Date(2026, time.March, 8, 23, 35, 25, 0, time.UTC)
	}, nil)
	if err != nil {
		t.Fatalf("new live provider: %v", err)
	}
	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	return response
}

type replaySpotReader struct {
	snapshot marketstateapi.SpotCurrentStateSnapshot
}

func (s replaySpotReader) Snapshot(context.Context, time.Time) (marketstateapi.SpotCurrentStateSnapshot, error) {
	return s.snapshot, nil
}

func replaySpotSnapshot() marketstateapi.SpotCurrentStateSnapshot {
	start := time.Date(2026, time.March, 8, 23, 30, 0, 0, time.UTC)
	return marketstateapi.SpotCurrentStateSnapshot{
		Observations: append(
			replayObservationSeries("BTC-USD", "BTCUSDT", 64000, start),
			replayObservationSeries("ETH-USD", "ETHUSDT", 3200, start)...,
		),
	}
}

func replayObservationSeries(symbol string, sourceSymbol string, basePrice float64, start time.Time) []marketstateapi.SpotCurrentStateObservation {
	observations := make([]marketstateapi.SpotCurrentStateObservation, 0, 10)
	for step := 0; step < 10; step++ {
		recv := start.Add(time.Duration(step) * 30 * time.Second).Add(500 * time.Millisecond)
		observations = append(observations, marketstateapi.SpotCurrentStateObservation{
			Symbol:        symbol,
			SourceSymbol:  sourceSymbol,
			QuoteCurrency: "USDT",
			BestBidPrice:  basePrice + float64(step),
			BestAskPrice:  basePrice + float64(step) + 1,
			ExchangeTs:    recv.Add(-500 * time.Millisecond),
			RecvTs:        recv,
			FeedHealth: ingestion.FeedHealthStatus{
				State:             ingestion.FeedHealthHealthy,
				ConnectionState:   ingestion.ConnectionConnected,
				MessageFreshness:  ingestion.FreshnessFresh,
				SnapshotFreshness: ingestion.FreshnessFresh,
				ClockState:        ingestion.ClockNormal,
			},
			DepthStatus: venuebinance.SpotDepthRecoveryStatus{State: venuebinance.SpotDepthRecoverySynchronized, Synchronized: true},
		})
	}
	return observations
}
