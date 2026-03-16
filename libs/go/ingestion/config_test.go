package ingestion

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestLoadEnvironmentConfigParsesLocalRuntimeConfig(t *testing.T) {
	repoRoot := repoRoot(t)
	config, err := LoadEnvironmentConfig(filepath.Join(repoRoot, "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}

	if config.Environment == "" {
		t.Fatal("expected non-empty environment")
	}

	runtimeConfig, err := config.RuntimeConfigFor(VenueKraken)
	if err != nil {
		t.Fatalf("load kraken runtime config: %v", err)
	}

	if runtimeConfig.ServicePath != "services/venue-kraken" {
		t.Fatalf("service path = %q, want %q", runtimeConfig.ServicePath, "services/venue-kraken")
	}
	if runtimeConfig.Adapter.MessageStaleAfter != 12*time.Second {
		t.Fatalf("message stale after = %s, want %s", runtimeConfig.Adapter.MessageStaleAfter, 12*time.Second)
	}
	if runtimeConfig.SnapshotRefreshInterval != 180*time.Second {
		t.Fatalf("snapshot refresh interval = %s, want %s", runtimeConfig.SnapshotRefreshInterval, 180*time.Second)
	}
	if len(runtimeConfig.Adapter.Streams) != 3 {
		t.Fatalf("stream count = %d, want %d", len(runtimeConfig.Adapter.Streams), 3)
	}
	if !runtimeConfig.ResubscribeOnReconnect {
		t.Fatal("expected resubscribe on reconnect to be true")
	}

	binanceRuntime, err := config.RuntimeConfigFor(VenueBinance)
	if err != nil {
		t.Fatalf("load binance runtime config: %v", err)
	}
	if binanceRuntime.OpenInterestPollInterval != 15*time.Second {
		t.Fatalf("open interest poll interval = %s, want %s", binanceRuntime.OpenInterestPollInterval, 15*time.Second)
	}
	if binanceRuntime.OpenInterestPollsPerMinuteLimit != 18 {
		t.Fatalf("open interest polls per-minute limit = %d, want %d", binanceRuntime.OpenInterestPollsPerMinuteLimit, 18)
	}
}

func TestLoadEnvironmentConfigKeepsCheckedInProfilesProdLikeAndIdentical(t *testing.T) {
	t.Parallel()

	repoRoot := repoRoot(t)
	baselinePath := filepath.Join(repoRoot, "configs/prod/ingestion.v1.json")
	baselineRaw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("read baseline config: %v", err)
	}
	baseline := normalizeEnvironmentLabel(string(baselineRaw))

	for _, environment := range []string{"local", "dev", "prod"} {
		environment := environment
		t.Run(environment, func(t *testing.T) {
			configPath := filepath.Join(repoRoot, "configs", environment, "ingestion.v1.json")
			config, err := LoadEnvironmentConfig(configPath)
			if err != nil {
				t.Fatalf("load environment config: %v", err)
			}
			if config.Environment != environment {
				t.Fatalf("environment = %q, want %q", config.Environment, environment)
			}
			if !slices.Equal(config.Symbols, []string{"BTC-USD", "ETH-USD"}) {
				t.Fatalf("symbols = %v, want [BTC-USD ETH-USD]", config.Symbols)
			}

			runtimeConfig, err := config.RuntimeConfigFor(VenueBinance)
			if err != nil {
				t.Fatalf("load binance runtime config: %v", err)
			}
			if runtimeConfig.ServicePath != "services/venue-binance" {
				t.Fatalf("service path = %q, want %q", runtimeConfig.ServicePath, "services/venue-binance")
			}
			if !slices.Equal(runtimeConfig.Symbols, []string{"BTC-USD", "ETH-USD"}) {
				t.Fatalf("runtime symbols = %v, want [BTC-USD ETH-USD]", runtimeConfig.Symbols)
			}
			if runtimeConfig.OpenInterestPollInterval != 15*time.Second {
				t.Fatalf("open interest poll interval = %s, want %s", runtimeConfig.OpenInterestPollInterval, 15*time.Second)
			}
			if runtimeConfig.OpenInterestPollsPerMinuteLimit != 18 {
				t.Fatalf("open interest polls per-minute limit = %d, want %d", runtimeConfig.OpenInterestPollsPerMinuteLimit, 18)
			}

			raw, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("read config file: %v", err)
			}
			if normalized := normalizeEnvironmentLabel(string(raw)); normalized != baseline {
				t.Fatalf("%s config diverges from prod baseline", environment)
			}
		})
	}
}

func TestEnvironmentConfigValidateRejectsInvalidSnapshotPolicy(t *testing.T) {
	config := validEnvironmentConfig()
	coinbase := config.Venues[VenueCoinbase]
	coinbase.SnapshotRefreshPolicy.RefreshIntervalMs = 1000
	config.Venues[VenueCoinbase] = coinbase

	err := config.Validate()
	if err == nil {
		t.Fatal("expected snapshot policy validation error")
	}
	if !strings.Contains(err.Error(), "snapshot refresh interval must be zero") {
		t.Fatalf("error = %q, want snapshot policy validation message", err)
	}
}

func TestRuntimeConfigForUnknownVenue(t *testing.T) {
	config := validEnvironmentConfig()

	_, err := config.RuntimeConfigFor(Venue("UNKNOWN"))
	if err == nil {
		t.Fatal("expected unknown venue error")
	}
}

func TestVenueRuntimeSourceRuntimeConfigRejectsInvalidOpenInterestSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*VenueRuntimeSource)
		wantErr string
	}{
		{
			name: "missing poll interval",
			mutate: func(source *VenueRuntimeSource) {
				source.Rest.OpenInterestPollIntervalMs = 0
			},
			wantErr: "open interest poll interval must be positive",
		},
		{
			name: "missing per-minute limit",
			mutate: func(source *VenueRuntimeSource) {
				source.Rest.OpenInterestPollsPerMinuteLimit = 0
			},
			wantErr: "open interest polls per-minute limit must be positive",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			source := validBinanceVenueRuntimeSource()
			tt.mutate(&source)

			_, err := source.RuntimeConfig(VenueBinance, []string{"BTC-USD", "ETH-USD"}, NormalizerHandoffConfig{
				Service:                  "services/normalizer",
				PreserveExchangeTs:       true,
				PreserveRecvTs:           true,
				PropagateDegradedReasons: true,
			})
			if err == nil {
				t.Fatal("expected open-interest validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func validEnvironmentConfig() EnvironmentConfig {
	return EnvironmentConfig{
		SchemaVersion: "v1",
		Environment:   "test",
		Symbols:       []string{"BTC-USD", "ETH-USD"},
		NormalizerHandoff: NormalizerHandoffConfig{
			Service:                  "services/normalizer",
			PreserveExchangeTs:       true,
			PreserveRecvTs:           true,
			PropagateDegradedReasons: true,
		},
		Venues: map[Venue]VenueRuntimeSource{
			VenueCoinbase: {
				ServicePath: "services/venue-coinbase",
				Websocket: WebsocketConfig{
					HeartbeatTimeoutMs:     12000,
					ReconnectBackoffMinMs:  1000,
					ReconnectBackoffMaxMs:  10000,
					ReconnectLoopThreshold: 3,
					ConnectsPerMinuteLimit: 8,
					ResubscribeOnReconnect: true,
				},
				Rest: RestConfig{
					SnapshotRecoveryPerMinuteLimit: 0,
					SnapshotCooldownMs:             0,
				},
				Health: HealthThresholds{
					MessageStaleAfterMs:   15000,
					SnapshotStaleAfterMs:  60000,
					ResyncLoopThreshold:   2,
					ClockOffsetWarningMs:  100,
					ClockOffsetDegradedMs: 250,
				},
				SnapshotRefreshPolicy: SnapshotRefreshPolicy{
					Required:          false,
					RefreshIntervalMs: 0,
				},
				Streams: []StreamDefinition{
					{Kind: StreamTrades, MarketType: "spot"},
					{Kind: StreamTopOfBook, MarketType: "spot"},
				},
			},
		},
	}
}

func validBinanceVenueRuntimeSource() VenueRuntimeSource {
	return VenueRuntimeSource{
		ServicePath: "services/venue-binance",
		Websocket: WebsocketConfig{
			HeartbeatTimeoutMs:     10000,
			ReconnectBackoffMinMs:  500,
			ReconnectBackoffMaxMs:  5000,
			ReconnectLoopThreshold: 4,
			ConnectsPerMinuteLimit: 12,
			ResubscribeOnReconnect: true,
		},
		Rest: RestConfig{
			SnapshotRecoveryPerMinuteLimit:  30,
			SnapshotCooldownMs:              1000,
			OpenInterestPollIntervalMs:      5000,
			OpenInterestPollsPerMinuteLimit: 30,
		},
		Health: HealthThresholds{
			MessageStaleAfterMs:   15000,
			SnapshotStaleAfterMs:  30000,
			ResyncLoopThreshold:   3,
			ClockOffsetWarningMs:  100,
			ClockOffsetDegradedMs: 250,
		},
		SnapshotRefreshPolicy: SnapshotRefreshPolicy{
			Required:          true,
			RefreshIntervalMs: 300000,
		},
		Streams: []StreamDefinition{
			{Kind: StreamTrades, MarketType: "spot", SnapshotRequired: false},
			{Kind: StreamTopOfBook, MarketType: "spot", SnapshotRequired: false},
			{Kind: StreamOrderBook, MarketType: "spot", SnapshotRequired: true},
			{Kind: StreamFundingRate, MarketType: "perpetual", SnapshotRequired: false},
			{Kind: StreamOpenInterest, MarketType: "perpetual", SnapshotRequired: false},
		},
	}
}

func normalizeEnvironmentLabel(raw string) string {
	replacer := strings.NewReplacer(
		`"environment": "local"`, `"environment": "normalized"`,
		`"environment": "dev"`, `"environment": "normalized"`,
		`"environment": "prod"`, `"environment": "normalized"`,
	)
	return replacer.Replace(raw)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", "..", ".."))
}
