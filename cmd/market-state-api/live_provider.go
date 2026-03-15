package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

const (
	defaultConfigPath          = "configs/local/ingestion.v1.json"
	defaultBinanceURL          = "https://api.binance.com"
	defaultBinanceWebsocketURL = "wss://stream.binance.com:9443/ws"
)

type providerOptions struct {
	clock        func() time.Time
	client       *http.Client
	configPath   string
	binanceURL   string
	websocketURL string
}

type providerWithRuntime struct {
	marketstateapi.Provider
	owner *binanceSpotRuntimeOwner
}

func (p *providerWithRuntime) Close(ctx context.Context) error {
	if p == nil || p.owner == nil {
		return nil
	}
	return p.owner.Stop(ctx)
}

func newProvider() (marketstateapi.Provider, error) {
	return newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       &http.Client{Timeout: 10 * time.Second},
		configPath:   configPath(),
		binanceURL:   binanceBaseURL(),
		websocketURL: binanceWebsocketURL(),
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
	provider, err := marketstateapi.NewLiveSpotProvider(reader, options.clock, nil)
	if err != nil {
		_ = reader.Stop(context.Background())
		return nil, fmt.Errorf("create live spot provider: %w", err)
	}
	return &providerWithRuntime{Provider: provider, owner: reader}, nil
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
