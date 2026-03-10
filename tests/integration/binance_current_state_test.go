package integration

import (
	"context"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceCurrentStateSymbolResponse(t *testing.T) {
	provider := newIntegrationLiveProvider(t, integrationSpotReader{snapshot: integrationSpotSnapshot()})
	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.SchemaVersion != features.MarketStateCurrentResponseSchema {
		t.Fatalf("schema version = %q", response.SchemaVersion)
	}
	if response.Composite.Availability != features.CurrentStateAvailabilityPartial {
		t.Fatalf("composite availability = %q", response.Composite.Availability)
	}
	if !response.Composite.USA.Unavailable {
		t.Fatalf("expected unavailable usa composite: %+v", response.Composite.USA)
	}
	if response.Provenance.HistorySeam.ReservedSchemaFamily == "" {
		t.Fatalf("missing history seam: %+v", response.Provenance)
	}
	if len(response.Provenance.HistorySeam.BucketRefs) == 0 {
		t.Fatalf("missing bucket refs: %+v", response.Provenance)
	}
}

func TestIngestionBinanceCurrentStateDegradation(t *testing.T) {
	snapshot := integrationSpotSnapshot()
	for index := range snapshot.Observations {
		if snapshot.Observations[index].Symbol != "ETH-USD" {
			continue
		}
		snapshot.Observations[index].DepthStatus = venuebinance.SpotDepthRecoveryStatus{
			State:               venuebinance.SpotDepthRecoveryResyncing,
			SequenceGapDetected: true,
		}
		snapshot.Observations[index].FeedHealth = ingestion.FeedHealthStatus{
			State:             ingestion.FeedHealthDegraded,
			ConnectionState:   ingestion.ConnectionConnected,
			MessageFreshness:  ingestion.FreshnessFresh,
			SnapshotFreshness: ingestion.FreshnessFresh,
			ClockState:        ingestion.ClockNormal,
			Reasons:           []ingestion.DegradationReason{ingestion.ReasonSequenceGap},
		}
	}
	provider := newIntegrationLiveProvider(t, integrationSpotReader{snapshot: snapshot})
	response, err := provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if !response.Composite.World.Degraded {
		t.Fatalf("expected degraded world composite: %+v", response.Composite.World)
	}
	if response.Buckets.ThirtySeconds.Availability == features.CurrentStateAvailabilityAvailable {
		t.Fatalf("expected degraded 30s bucket: %+v", response.Buckets.ThirtySeconds)
	}
	if response.Regime.Symbol.State != features.RegimeStateNoOperate {
		t.Fatalf("symbol regime = %q", response.Regime.Symbol.State)
	}
}

type integrationSpotReader struct {
	snapshot marketstateapi.SpotCurrentStateSnapshot
	terr     error
}

func (s integrationSpotReader) Snapshot(context.Context, time.Time) (marketstateapi.SpotCurrentStateSnapshot, error) {
	if s.terr != nil {
		return marketstateapi.SpotCurrentStateSnapshot{}, s.terr
	}
	return s.snapshot, nil
}

func newIntegrationLiveProvider(t *testing.T, reader marketstateapi.SpotCurrentStateReader) *marketstateapi.LiveSpotProvider {
	t.Helper()
	provider, err := marketstateapi.NewLiveSpotProvider(reader, func() time.Time {
		return time.Date(2026, time.March, 8, 23, 35, 25, 0, time.UTC)
	}, nil)
	if err != nil {
		t.Fatalf("new live provider: %v", err)
	}
	return provider
}

func integrationSpotSnapshot() marketstateapi.SpotCurrentStateSnapshot {
	start := time.Date(2026, time.March, 8, 23, 30, 0, 0, time.UTC)
	return marketstateapi.SpotCurrentStateSnapshot{
		Observations: append(
			integrationObservationSeries("BTC-USD", "BTCUSDT", 64000, start),
			integrationObservationSeries("ETH-USD", "ETHUSDT", 3200, start)...,
		),
	}
}

func integrationObservationSeries(symbol string, sourceSymbol string, basePrice float64, start time.Time) []marketstateapi.SpotCurrentStateObservation {
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
