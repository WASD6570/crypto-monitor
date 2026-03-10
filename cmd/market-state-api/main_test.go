package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewProviderWithOptionsUsesLiveSnapshotReader(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		symbol := r.URL.Query().Get("symbol")
		switch symbol {
		case "BTCUSDT":
			_, _ = w.Write([]byte(`{"lastUpdateId":1001,"symbol":"BTCUSDT","bids":[["64000.10","1.2"]],"asks":[["64000.20","1.4"]]}`))
		case "ETHUSDT":
			_, _ = w.Write([]byte(`{"lastUpdateId":2002,"symbol":"ETHUSDT","bids":[["3200.10","2.2"]],"asks":[["3200.30","2.4"]]}`))
		default:
			http.Error(w, "unexpected symbol", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	provider, err := newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC) },
		client:       server.Client(),
		configPath:   testConfigPath(),
		binanceURL:   server.URL,
		pollInterval: time.Hour,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q, want BTC-USD", response.Symbol)
	}
	if response.SchemaVersion == "" {
		t.Fatalf("expected schema version in response")
	}
	if response.Composite.World.CompositePrice == nil {
		t.Fatalf("expected world composite price in live response")
	}
	if !response.Composite.USA.Unavailable {
		t.Fatalf("expected usa composite to remain unavailable: %+v", response.Composite.USA)
	}
	if requests.Load() != 2 {
		t.Fatalf("request count = %d, want 2", requests.Load())
	}

	_, err = provider.CurrentSymbolState(context.Background(), "ETH-USD")
	if err != nil {
		t.Fatalf("second current symbol state: %v", err)
	}
	if requests.Load() != 2 {
		t.Fatalf("request count after cached read = %d, want 2", requests.Load())
	}
}

func TestNewProviderWithOptionsReturnsUnavailableWhenSnapshotFetchFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	provider, err := newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC) },
		client:       server.Client(),
		configPath:   testConfigPath(),
		binanceURL:   server.URL,
		pollInterval: time.Second,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	response, err := provider.CurrentSymbolState(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("current symbol state: %v", err)
	}
	if response.Symbol != "BTC-USD" {
		t.Fatalf("symbol = %q, want BTC-USD", response.Symbol)
	}
	if response.Composite.World.Unavailable != true {
		t.Fatalf("expected unavailable world composite after upstream failure: %+v", response.Composite.World)
	}
	if response.Regime.EffectiveState == "" {
		t.Fatalf("expected stable regime payload after upstream failure")
	}
}

func TestNewProviderWithOptionsRejectsMissingConfig(t *testing.T) {
	_, err := newProviderWithOptions(providerOptions{
		clock:        func() time.Time { return time.Now().UTC() },
		client:       http.DefaultClient,
		configPath:   filepath.Join(t.TempDir(), "missing.json"),
		binanceURL:   defaultBinanceURL,
		pollInterval: time.Second,
	})
	if err == nil {
		t.Fatalf("expected config error")
	}
}

func TestServerAddressUsesDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	if got := serverAddress(); got != ":8080" {
		t.Fatalf("server address = %q, want :8080", got)
	}
}

func TestServerAddressUsesPortEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	if got := serverAddress(); got != ":9090" {
		t.Fatalf("server address = %q, want :9090", got)
	}
}

func testConfigPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "configs", "local", "ingestion.v1.json"))
}
