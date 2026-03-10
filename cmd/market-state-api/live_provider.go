package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

const (
	defaultConfigPath   = "configs/local/ingestion.v1.json"
	defaultBinanceURL   = "https://api.binance.com"
	defaultPollInterval = 5 * time.Second
)

type providerOptions struct {
	clock        func() time.Time
	client       *http.Client
	configPath   string
	binanceURL   string
	pollInterval time.Duration
}

type binanceSpotSnapshotReader struct {
	client       *http.Client
	baseURL      string
	pollInterval time.Duration

	mu     sync.Mutex
	states map[string]*spotSnapshotState
	order  []string
}

type spotSnapshotState struct {
	binding             spotBinding
	owner               *venuebinance.SpotDepthRecoveryOwner
	lastObservation     marketstateapi.SpotCurrentStateObservation
	haveObservation     bool
	lastRefresh         time.Time
	lastRequestErr      error
	consecutiveFailures int
}

type spotBinding struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
}

type noopSpotDepthSnapshotFetcher struct{}

func newProvider() (marketstateapi.Provider, error) {
	return newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       &http.Client{Timeout: 10 * time.Second},
		configPath:   configPath(),
		binanceURL:   binanceBaseURL(),
		pollInterval: pollInterval(),
	})
}

func newProviderWithOptions(options providerOptions) (marketstateapi.Provider, error) {
	if options.clock == nil {
		return nil, fmt.Errorf("clock is required")
	}
	if options.client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if options.configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}
	if options.binanceURL == "" {
		return nil, fmt.Errorf("binance base url is required")
	}
	if options.pollInterval <= 0 {
		return nil, fmt.Errorf("poll interval must be positive")
	}

	envConfig, err := ingestion.LoadEnvironmentConfig(options.configPath)
	if err != nil {
		return nil, fmt.Errorf("load environment config: %w", err)
	}
	runtimeConfig, err := envConfig.RuntimeConfigFor(ingestion.VenueBinance)
	if err != nil {
		return nil, fmt.Errorf("load binance runtime config: %w", err)
	}
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("create binance runtime: %w", err)
	}
	reader, err := newBinanceSpotSnapshotReader(runtimeConfig, runtime, options.client, options.binanceURL, options.pollInterval)
	if err != nil {
		return nil, err
	}
	provider, err := marketstateapi.NewLiveSpotProvider(reader, options.clock, nil)
	if err != nil {
		return nil, fmt.Errorf("create live spot provider: %w", err)
	}
	return provider, nil
}

func configPath() string {
	if value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_CONFIG_PATH")); value != "" {
		return value
	}
	return filepath.Clean(defaultConfigPath)
}

func binanceBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_BINANCE_BASE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return defaultBinanceURL
}

func pollInterval() time.Duration {
	value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_SPOT_POLL_INTERVAL"))
	if value == "" {
		return defaultPollInterval
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return defaultPollInterval
	}
	return parsed
}

func newBinanceSpotSnapshotReader(runtimeConfig ingestion.VenueRuntimeConfig, runtime *venuebinance.Runtime, client *http.Client, baseURL string, pollInterval time.Duration) (*binanceSpotSnapshotReader, error) {
	if runtime == nil {
		return nil, fmt.Errorf("binance runtime is required")
	}
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("binance base url is required")
	}
	if pollInterval <= 0 {
		return nil, fmt.Errorf("poll interval must be positive")
	}
	bindings, err := spotBindings(runtimeConfig.Symbols)
	if err != nil {
		return nil, err
	}
	states := make(map[string]*spotSnapshotState, len(bindings))
	order := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, noopSpotDepthSnapshotFetcher{})
		if err != nil {
			return nil, fmt.Errorf("create depth recovery owner for %s: %w", binding.Symbol, err)
		}
		states[binding.Symbol] = &spotSnapshotState{binding: binding, owner: owner}
		order = append(order, binding.Symbol)
	}
	return &binanceSpotSnapshotReader{
		client:       client,
		baseURL:      strings.TrimRight(baseURL, "/"),
		pollInterval: pollInterval,
		states:       states,
		order:        order,
	}, nil
}

func (r *binanceSpotSnapshotReader) Snapshot(ctx context.Context, now time.Time) (marketstateapi.SpotCurrentStateSnapshot, error) {
	if ctx == nil {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("context is required")
	}
	if now.IsZero() {
		return marketstateapi.SpotCurrentStateSnapshot{}, fmt.Errorf("current time is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, symbol := range r.order {
		if err := ctx.Err(); err != nil {
			return marketstateapi.SpotCurrentStateSnapshot{}, err
		}
		state := r.states[symbol]
		if state == nil || !state.shouldRefresh(now, r.pollInterval) {
			continue
		}
		state.refresh(ctx, now, r.client, r.baseURL)
	}

	observations := make([]marketstateapi.SpotCurrentStateObservation, 0, len(r.order))
	for _, symbol := range r.order {
		state := r.states[symbol]
		if state == nil {
			continue
		}
		observation, ok, err := state.observation(now)
		if err != nil {
			return marketstateapi.SpotCurrentStateSnapshot{}, err
		}
		if ok {
			observations = append(observations, observation)
		}
	}
	return marketstateapi.SpotCurrentStateSnapshot{Observations: observations}, nil
}

func (s *spotSnapshotState) shouldRefresh(now time.Time, interval time.Duration) bool {
	if s == nil {
		return false
	}
	if interval <= 0 || s.lastRefresh.IsZero() {
		return true
	}
	return !now.Before(s.lastRefresh.Add(interval))
}

func (s *spotSnapshotState) refresh(ctx context.Context, now time.Time, client *http.Client, baseURL string) {
	response, err := fetchDepthSnapshot(ctx, client, baseURL, s.binding.SourceSymbol)
	if err != nil {
		s.recordFailure(now, err)
		return
	}
	parsed, err := venuebinance.ParseOrderBookSnapshotWithSourceSymbol(response, s.binding.SourceSymbol, now)
	if err != nil {
		s.recordFailure(now, err)
		return
	}
	if err := s.owner.StartSynchronized(venuebinance.SpotDepthBootstrapSync{
		SourceSymbol: s.binding.SourceSymbol,
		Snapshot:     parsed,
	}); err != nil {
		s.recordFailure(now, err)
		return
	}
	bestBid, err := strconv.ParseFloat(parsed.Message.BestBidPrice, 64)
	if err != nil {
		s.recordFailure(now, fmt.Errorf("parse best bid price: %w", err))
		return
	}
	bestAsk, err := strconv.ParseFloat(parsed.Message.BestAskPrice, 64)
	if err != nil {
		s.recordFailure(now, fmt.Errorf("parse best ask price: %w", err))
		return
	}
	s.lastObservation = marketstateapi.SpotCurrentStateObservation{
		Symbol:        s.binding.Symbol,
		SourceSymbol:  s.binding.SourceSymbol,
		QuoteCurrency: s.binding.QuoteCurrency,
		BestBidPrice:  bestBid,
		BestAskPrice:  bestAsk,
		ExchangeTs:    time.Time{},
		RecvTs:        now.UTC(),
	}
	s.haveObservation = true
	s.lastRefresh = now.UTC()
	s.lastRequestErr = nil
	s.consecutiveFailures = 0
}

func (s *spotSnapshotState) recordFailure(now time.Time, err error) {
	if s == nil {
		return
	}
	s.lastRefresh = now.UTC()
	s.lastRequestErr = err
	s.consecutiveFailures++
	if !s.haveObservation {
		_ = s.owner.MarkBootstrapFailure(s.binding.SourceSymbol)
	}
}

func (s *spotSnapshotState) observation(now time.Time) (marketstateapi.SpotCurrentStateObservation, bool, error) {
	if s == nil || !s.haveObservation {
		return marketstateapi.SpotCurrentStateObservation{}, false, nil
	}
	connectionState := ingestion.ConnectionConnected
	if s.lastRequestErr != nil {
		connectionState = ingestion.ConnectionDisconnected
	}
	feedHealth, err := s.owner.HealthStatus(now.UTC(), connectionState, 0, s.consecutiveFailures)
	if err != nil {
		return marketstateapi.SpotCurrentStateObservation{}, false, err
	}
	observation := s.lastObservation
	observation.FeedHealth = feedHealth
	observation.DepthStatus = s.owner.Status()
	return observation, true, nil
}

func fetchDepthSnapshot(ctx context.Context, client *http.Client, baseURL string, sourceSymbol string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v3/depth?symbol="+sourceSymbol+"&limit=5", nil)
	if err != nil {
		return nil, fmt.Errorf("create depth request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request depth snapshot: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("depth snapshot status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read depth snapshot: %w", err)
	}
	return body, nil
}

func spotBindings(symbols []string) ([]spotBinding, error) {
	bindings := make([]spotBinding, 0, len(symbols))
	for _, symbol := range symbols {
		if !strings.HasSuffix(symbol, "-USD") {
			return nil, fmt.Errorf("unsupported spot symbol %q", symbol)
		}
		base := strings.TrimSuffix(symbol, "-USD")
		if base == "" {
			return nil, fmt.Errorf("unsupported spot symbol %q", symbol)
		}
		bindings = append(bindings, spotBinding{
			Symbol:        symbol,
			SourceSymbol:  base + "USDT",
			QuoteCurrency: "USDT",
		})
	}
	return bindings, nil
}

func (noopSpotDepthSnapshotFetcher) FetchSpotDepthSnapshot(context.Context, string) (venuebinance.SpotDepthSnapshotResponse, error) {
	return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("depth refresh fetcher is not wired for command snapshot reader")
}
