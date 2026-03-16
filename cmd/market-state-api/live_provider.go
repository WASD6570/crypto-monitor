package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

const (
	defaultConfigPath              = "configs/local/ingestion.v1.json"
	defaultBinanceURL              = "https://api.binance.com"
	defaultBinanceWebsocketURL     = "wss://stream.binance.com:9443/ws"
	defaultBinanceUSDMURL          = "https://fapi.binance.com"
	defaultBinanceUSDMWebsocketURL = "wss://fstream.binance.com/ws"
)

var runtimeStatusSymbols = []string{"BTC-USD", "ETH-USD"}

type providerOptions struct {
	clock            func() time.Time
	client           *http.Client
	configPath       string
	binanceURL       string
	websocketURL     string
	binanceUSDMURL   string
	usdmWebsocketURL string
}

type providerWithRuntime struct {
	marketstateapi.Provider
	owner     *binanceSpotRuntimeOwner
	usdmOwner *binanceUSDMInfluenceOwner
}

func (p *providerWithRuntime) RuntimeStatus(ctx context.Context) (marketstateapi.RuntimeStatusResponse, error) {
	if p == nil || p.owner == nil {
		return marketstateapi.RuntimeStatusResponse{}, fmt.Errorf("binance runtime health owner is required")
	}
	snapshot, err := p.RuntimeHealthSnapshot(ctx, p.owner.now().UTC())
	if err != nil {
		return marketstateapi.RuntimeStatusResponse{}, err
	}
	return snapshot.runtimeStatusResponse(), nil
}

func (p *providerWithRuntime) RuntimeHealthSnapshot(ctx context.Context, now time.Time) (binanceRuntimeHealthSnapshot, error) {
	if p == nil || p.owner == nil {
		return binanceRuntimeHealthSnapshot{}, fmt.Errorf("binance runtime health owner is required")
	}
	return p.owner.RuntimeHealthSnapshot(ctx, now)
}

func (p *providerWithRuntime) Close(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var firstErr error
	if p.usdmOwner != nil {
		if err := p.usdmOwner.Stop(ctx); err != nil {
			firstErr = err
		}
	}
	if p.owner != nil {
		if err := p.owner.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func newProvider() (marketstateapi.Provider, error) {
	return newProviderWithOptions(providerOptions{
		clock:            func() time.Time { return time.Now().UTC() },
		client:           &http.Client{Timeout: 10 * time.Second},
		configPath:       configPath(),
		binanceURL:       binanceBaseURL(),
		websocketURL:     binanceWebsocketURL(),
		binanceUSDMURL:   binanceUSDMBaseURL(),
		usdmWebsocketURL: binanceUSDMWebsocketURL(),
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
	if options.websocketURL == "" {
		return nil, fmt.Errorf("binance websocket url is required")
	}
	if options.binanceUSDMURL == "" {
		if options.binanceURL != defaultBinanceURL {
			return nil, fmt.Errorf("binance usdm base url is required when spot base url is overridden")
		}
		options.binanceUSDMURL = defaultBinanceUSDMURL
	}
	if options.usdmWebsocketURL == "" {
		if options.websocketURL != defaultBinanceWebsocketURL {
			return nil, fmt.Errorf("binance usdm websocket url is required when spot websocket url is overridden")
		}
		options.usdmWebsocketURL = defaultBinanceUSDMWebsocketURL
	}

	envConfig, err := ingestion.LoadEnvironmentConfig(options.configPath)
	if err != nil {
		return nil, fmt.Errorf("load environment config: %w", err)
	}
	runtimeConfig, err := envConfig.RuntimeConfigFor(ingestion.VenueBinance)
	if err != nil {
		return nil, fmt.Errorf("load binance runtime config: %w", err)
	}
	if !slices.Equal(runtimeConfig.Symbols, runtimeStatusSymbols) {
		return nil, fmt.Errorf("binance runtime symbols must stay %v for runtime-status support, got %v", runtimeStatusSymbols, runtimeConfig.Symbols)
	}
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("create binance runtime: %w", err)
	}
	reader, err := newBinanceSpotRuntimeOwner(runtimeConfig, runtime, binanceSpotRuntimeOwnerOptions{
		client:       options.client,
		baseURL:      options.binanceURL,
		websocketURL: options.websocketURL,
		now:          options.clock,
	})
	if err != nil {
		return nil, err
	}
	if err := reader.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("start binance spot runtime owner: %w", err)
	}
	usdmOwner, err := newBinanceUSDMInfluenceOwner(runtimeConfig, runtime, binanceUSDMInfluenceOwnerOptions{
		client:           options.client,
		baseURL:          options.binanceUSDMURL,
		websocketURL:     options.usdmWebsocketURL,
		heartbeatTimeout: runtimeConfig.HeartbeatTimeout,
		now:              options.clock,
	})
	if err != nil {
		_ = reader.Stop(context.Background())
		return nil, fmt.Errorf("create usdm influence owner: %w", err)
	}
	if err := usdmOwner.Start(context.Background()); err != nil {
		_ = reader.Stop(context.Background())
		return nil, fmt.Errorf("start usdm influence owner: %w", err)
	}
	provider, err := marketstateapi.NewLiveSpotProvider(&combinedCurrentStateReader{spot: reader, usdm: usdmOwner}, options.clock, nil)
	if err != nil {
		_ = usdmOwner.Stop(context.Background())
		_ = reader.Stop(context.Background())
		return nil, fmt.Errorf("create live spot provider: %w", err)
	}
	return &providerWithRuntime{Provider: provider, owner: reader, usdmOwner: usdmOwner}, nil
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

func binanceWebsocketURL() string {
	if value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_BINANCE_WS_URL")); value != "" {
		return value
	}
	return defaultBinanceWebsocketURL
}

func binanceUSDMBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_BINANCE_USDM_BASE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return defaultBinanceUSDMURL
}

func binanceUSDMWebsocketURL() string {
	if value := strings.TrimSpace(os.Getenv("MARKET_STATE_API_BINANCE_USDM_WS_URL")); value != "" {
		return value
	}
	return defaultBinanceUSDMWebsocketURL
}
