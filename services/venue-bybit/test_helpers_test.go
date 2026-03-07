package venuebybit

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}

func loadBybitRuntimeConfig(t *testing.T) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}

	runtimeConfig, err := config.RuntimeConfigFor(ingestion.VenueBybit)
	if err != nil {
		t.Fatalf("load bybit runtime config: %v", err)
	}
	return runtimeConfig
}
