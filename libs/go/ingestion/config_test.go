package ingestion

import (
	"path/filepath"
	"runtime"
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

	if config.Environment != "local" {
		t.Fatalf("environment = %q, want %q", config.Environment, "local")
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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", "..", ".."))
}
