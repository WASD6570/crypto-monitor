package marketstateapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
	slowcontext "github.com/crypto-market-copilot/alerts/services/slow-context"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestDeterministicProviderCurrentSymbolState(t *testing.T) {
	provider := fixedProvider(t)
	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q", response.Symbol)
	}
	if response.Regime.EffectiveState != features.RegimeStateTradeable {
		t.Fatalf("effective state = %q", response.Regime.EffectiveState)
	}
	if response.SlowContext.Asset != "BTC" {
		t.Fatalf("slow context asset = %q", response.SlowContext.Asset)
	}
	if len(response.SlowContext.Contexts) != 3 {
		t.Fatalf("slow context count = %d", len(response.SlowContext.Contexts))
	}
	contextEntry, ok := response.SlowContext.Context(slowcontext.MetricFamilyETFDailyFlow)
	if !ok {
		t.Fatal("expected etf daily flow slow context")
	}
	if contextEntry.Availability != slowcontext.AvailabilityAvailable {
		t.Fatalf("slow context availability = %q", contextEntry.Availability)
	}
}

func TestDeterministicProviderCurrentGlobalState(t *testing.T) {
	provider := fixedProvider(t)
	response, err := provider.CurrentGlobalState(context.Background())
	if err != nil {
		t.Fatalf("current global state: %v", err)
	}
	if response.SchemaVersion != features.MarketStateCurrentGlobalSchema {
		t.Fatalf("schema version = %q", response.SchemaVersion)
	}
	if len(response.Symbols) != 2 {
		t.Fatalf("symbol summary size = %d", len(response.Symbols))
	}
	if response.Symbols[0].Symbol != "BTC-USD" {
		t.Fatalf("first symbol = %q", response.Symbols[0].Symbol)
	}
	if response.Provenance.HistorySeam.ReservedSchemaFamily != "market-state-history-and-audit-reads" {
		t.Fatalf("history seam = %+v", response.Provenance.HistorySeam)
	}
}

func TestDeterministicProviderRejectsUnsupportedSymbols(t *testing.T) {
	provider := fixedProvider(t)
	_, err := provider.CurrentSymbolState(context.Background(), "SOL-USD")
	if !errors.Is(err, ErrUnsupportedSymbol) {
		t.Fatalf("error = %v, want unsupported symbol", err)
	}
}

func TestHandlerServesCurrentStateRoutes(t *testing.T) {
	handler, err := NewHandler(fixedProvider(t))
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	t.Cleanup(server.Close)

	globalResponse := decodeJSON[features.MarketStateCurrentGlobalResponse](t, httpGet(t, server.URL+"/api/market-state/global", http.StatusOK))
	if globalResponse.SchemaVersion != features.MarketStateCurrentGlobalSchema {
		t.Fatalf("global schema version = %q", globalResponse.SchemaVersion)
	}

	symbolResponse := decodeJSON[SymbolStateResponse](t, httpGet(t, server.URL+"/api/market-state/BTC-USD", http.StatusOK))
	if symbolResponse.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q", symbolResponse.Symbol)
	}
	if symbolResponse.SlowContext.Asset != "BTC" {
		t.Fatalf("slow context asset = %q", symbolResponse.SlowContext.Asset)
	}

	healthResponse := decodeJSON[healthResponse](t, httpGet(t, server.URL+"/healthz", http.StatusOK))
	if healthResponse.Status != "ok" {
		t.Fatalf("health status = %q", healthResponse.Status)
	}
}

func TestHandlerReturnsNotFoundForUnsupportedSymbol(t *testing.T) {
	handler, err := NewHandler(fixedProvider(t))
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	t.Cleanup(server.Close)

	response := decodeJSON[errorResponse](t, httpGet(t, server.URL+"/api/market-state/SOL-USD", http.StatusNotFound))
	if response.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestHandlerReturnsProviderFailures(t *testing.T) {
	handler, err := NewHandler(failingProvider{err: errors.New("provider offline")})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	server := httptest.NewServer(handler.Routes())
	t.Cleanup(server.Close)

	response := decodeJSON[errorResponse](t, httpGet(t, server.URL+"/api/market-state/global", http.StatusInternalServerError))
	if response.Error != "provider offline" {
		t.Fatalf("error = %q", response.Error)
	}
}

func TestDeterministicProviderUsesFeatureAssemblySurface(t *testing.T) {
	provider := fixedProvider(t)
	response, err := provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("current eth state: %v", err)
	}
	var _ featureengine.CurrentStateWithSlowContextResponse
	if response.Version.ConfigVersion != "regime-engine.market-state.v1" {
		t.Fatalf("config version = %q", response.Version.ConfigVersion)
	}
	if response.Regime.Symbol.State != features.RegimeStateWatch {
		t.Fatalf("symbol state = %q", response.Regime.Symbol.State)
	}
	entry, ok := response.SlowContext.Context(slowcontext.MetricFamilyETFDailyFlow)
	if !ok {
		t.Fatal("expected etf context")
	}
	if entry.Availability != slowcontext.AvailabilityUnavailable {
		t.Fatalf("availability = %q", entry.Availability)
	}
}

func TestNewLiveSpotProviderRequiresReaderAndClock(t *testing.T) {
	if _, err := NewLiveSpotProvider(nil, func() time.Time { return time.Now().UTC() }, nil); err == nil {
		t.Fatal("expected reader requirement error")
	}
	if _, err := NewLiveSpotProvider(staticSpotReader{}, nil, nil); err == nil {
		t.Fatal("expected clock requirement error")
	}
}

func TestLiveSpotProviderCurrentSymbolState(t *testing.T) {
	provider := fixedLiveProvider(t, staticSpotReader{snapshot: testSpotSnapshot()})
	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q", response.Symbol)
	}
	if response.Composite.World.Unavailable {
		t.Fatalf("expected world composite to be available: %+v", response.Composite.World)
	}
	if !response.Composite.USA.Unavailable {
		t.Fatalf("expected usa composite to stay unavailable: %+v", response.Composite.USA)
	}
	if response.Composite.Availability != features.CurrentStateAvailabilityPartial {
		t.Fatalf("composite availability = %q", response.Composite.Availability)
	}
	if response.Regime.Symbol.State != features.RegimeStateNoOperate {
		t.Fatalf("symbol regime = %q", response.Regime.Symbol.State)
	}
	if response.SlowContext.Asset != "BTC" {
		t.Fatalf("slow context asset = %q", response.SlowContext.Asset)
	}
	if response.SlowContext.Contexts[0].Availability == slowcontext.AvailabilityAvailable {
		t.Fatalf("expected unavailable slow context without reader: %+v", response.SlowContext)
	}
}

func TestLiveSpotProviderCurrentGlobalState(t *testing.T) {
	provider := fixedLiveProvider(t, staticSpotReader{snapshot: testSpotSnapshot()})
	response, err := provider.CurrentGlobalState(context.Background())
	if err != nil {
		t.Fatalf("current global state: %v", err)
	}
	if len(response.Symbols) != 2 {
		t.Fatalf("symbol summary size = %d", len(response.Symbols))
	}
	if response.Provenance.HistorySeam.ReservedSchemaFamily != "market-state-history-and-audit-reads" {
		t.Fatalf("history seam = %+v", response.Provenance.HistorySeam)
	}
	if response.Global.State == "" {
		t.Fatalf("expected global regime state: %+v", response.Global)
	}
}

func TestLiveSpotProviderRejectsUnsupportedSymbols(t *testing.T) {
	provider := fixedLiveProvider(t, staticSpotReader{snapshot: testSpotSnapshot()})
	_, err := provider.CurrentSymbolState(context.Background(), "SOL-USD")
	if !errors.Is(err, ErrUnsupportedSymbol) {
		t.Fatalf("error = %v, want unsupported symbol", err)
	}
}

func TestLiveSpotProviderReturnsUnavailableWhenObservationsMissing(t *testing.T) {
	provider := fixedLiveProvider(t, staticSpotReader{})
	response, err := provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if !response.Composite.World.Unavailable || !response.Composite.USA.Unavailable {
		t.Fatalf("expected unavailable composites: %+v", response.Composite)
	}
	if response.Buckets.FiveMinutes.Availability != features.CurrentStateAvailabilityUnavailable {
		t.Fatalf("expected unavailable five-minute bucket: %+v", response.Buckets.FiveMinutes)
	}
	if response.Regime.Symbol.State != features.RegimeStateNoOperate {
		t.Fatalf("symbol regime = %q", response.Regime.Symbol.State)
	}
}

func TestLiveSpotProviderReflectsDepthDegradation(t *testing.T) {
	snapshot := testSpotSnapshot()
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
	provider := fixedLiveProvider(t, staticSpotReader{snapshot: snapshot})
	response, err := provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if !response.Composite.World.Degraded {
		t.Fatalf("expected degraded world composite: %+v", response.Composite.World)
	}
	if response.Buckets.ThirtySeconds.Availability == features.CurrentStateAvailabilityAvailable {
		t.Fatalf("expected degraded bucket availability: %+v", response.Buckets.ThirtySeconds)
	}
}

type failingProvider struct {
	err error
}

func (f failingProvider) CurrentGlobalState(context.Context) (features.MarketStateCurrentGlobalResponse, error) {
	return features.MarketStateCurrentGlobalResponse{}, f.err
}

func (f failingProvider) CurrentSymbolState(context.Context, string) (SymbolStateResponse, error) {
	return SymbolStateResponse{}, f.err
}

func fixedProvider(t *testing.T) *DeterministicProvider {
	t.Helper()
	provider, err := NewDeterministicProviderWithClock(func() time.Time {
		return time.Date(2026, time.March, 8, 23, 33, 25, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	return provider
}

func fixedLiveProvider(t *testing.T, reader SpotCurrentStateReader) *LiveSpotProvider {
	t.Helper()
	provider, err := NewLiveSpotProvider(reader, func() time.Time {
		return time.Date(2026, time.March, 8, 23, 35, 25, 0, time.UTC)
	}, nil)
	if err != nil {
		t.Fatalf("new live provider: %v", err)
	}
	return provider
}

type staticSpotReader struct {
	snapshot SpotCurrentStateSnapshot
	err      error
}

func (s staticSpotReader) Snapshot(context.Context, time.Time) (SpotCurrentStateSnapshot, error) {
	if s.err != nil {
		return SpotCurrentStateSnapshot{}, s.err
	}
	return s.snapshot, nil
}

func testSpotSnapshot() SpotCurrentStateSnapshot {
	start := time.Date(2026, time.March, 8, 23, 30, 0, 0, time.UTC)
	return SpotCurrentStateSnapshot{
		Observations: append(
			spotObservationSeries("BTC-USD", "BTCUSDT", 64000, start),
			spotObservationSeries("ETH-USD", "ETHUSDT", 3200, start)...,
		),
	}
}

func spotObservationSeries(symbol string, sourceSymbol string, basePrice float64, start time.Time) []SpotCurrentStateObservation {
	observations := make([]SpotCurrentStateObservation, 0, 10)
	for step := 0; step < 10; step++ {
		recv := start.Add(time.Duration(step) * 30 * time.Second).Add(500 * time.Millisecond)
		observations = append(observations, SpotCurrentStateObservation{
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

func httpGet(t *testing.T, url string, wantStatus int) *http.Response {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	if response.StatusCode != wantStatus {
		defer response.Body.Close()
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	return response
}

func decodeJSON[T any](t *testing.T, response *http.Response) T {
	t.Helper()
	defer response.Body.Close()
	var payload T
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return payload
}
