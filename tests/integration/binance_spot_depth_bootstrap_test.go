package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

type depthBootstrapFixture struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	SnapshotRaw       json.RawMessage   `json:"snapshotRaw"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

func TestIngestionBinanceSpotDepthBootstrap(t *testing.T) {
	t.Run("happy bootstrap alignment", func(t *testing.T) {
		service := newNormalizerService(t)
		fixture := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json")
		runtime := newBinanceVenueRuntime(t)
		supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-200*time.Millisecond))
		owner, err := venuebinance.NewSpotDepthBootstrapOwner(runtime, integrationSpotDepthSnapshotFetcher{
			snapshotRaw: fixture.SnapshotRaw,
		})
		if err != nil {
			t.Fatalf("new bootstrap owner: %v", err)
		}

		for _, raw := range fixture.RawMessages {
			frame, err := supervisor.AcceptDataFrame(raw, mustRecvTime(t, raw))
			if err != nil {
				t.Fatalf("accept depth frame: %v", err)
			}
			if err := owner.BufferDelta(frame); err != nil {
				t.Fatalf("buffer depth frame: %v", err)
			}
		}

		sync, err := owner.Synchronize(context.Background())
		if err != nil {
			t.Fatalf("synchronize depth bootstrap: %v", err)
		}
		if len(sync.Deltas) != 2 {
			t.Fatalf("aligned delta count = %d, want %d", len(sync.Deltas), 2)
		}

		metadata := ingestion.BookMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  fixture.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		}
		sequencer := &ingestion.OrderBookSequencer{}
		results := make([]ingestion.CanonicalOrderBookEvent, 0, 1+len(sync.Deltas))

		snapshot, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: sync.Snapshot.Message, Sequencer: sequencer})
		if err != nil {
			t.Fatalf("normalize snapshot: %v", err)
		}
		if snapshot.OrderBookEvent == nil || snapshot.FeedHealthEvent != nil {
			t.Fatal("expected canonical snapshot event without feed-health degradation")
		}
		results = append(results, *snapshot.OrderBookEvent)

		for _, delta := range sync.Deltas {
			result, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: delta.Message, Sequencer: sequencer})
			if err != nil {
				t.Fatalf("normalize delta: %v", err)
			}
			if result.OrderBookEvent == nil || result.FeedHealthEvent != nil {
				t.Fatal("expected canonical delta event without feed-health degradation")
			}
			results = append(results, *result.OrderBookEvent)
		}

		if len(results) != len(fixture.ExpectedCanonical) {
			t.Fatalf("canonical event count = %d, want %d", len(results), len(fixture.ExpectedCanonical))
		}
		for index, actual := range results {
			assertCanonicalOrderBookMatchesFixture(t, actual, fixture.ExpectedCanonical[index])
		}
	})

	t.Run("missing bridging delta stays unsynchronized", func(t *testing.T) {
		fixture := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-depth-bootstrap-missing-bridge-usdt.fixture.v1.json")
		runtime := newBinanceVenueRuntime(t)
		supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-200*time.Millisecond))
		owner, err := venuebinance.NewSpotDepthBootstrapOwner(runtime, integrationSpotDepthSnapshotFetcher{
			snapshotRaw: fixture.SnapshotRaw,
		})
		if err != nil {
			t.Fatalf("new bootstrap owner: %v", err)
		}

		for _, raw := range fixture.RawMessages {
			frame, err := supervisor.AcceptDataFrame(raw, mustRecvTime(t, raw))
			if err != nil {
				t.Fatalf("accept depth frame: %v", err)
			}
			if err := owner.BufferDelta(frame); err != nil {
				t.Fatalf("buffer depth frame: %v", err)
			}
		}

		if _, err := owner.Synchronize(context.Background()); err == nil {
			t.Fatal("expected synchronize without bridging delta to fail")
		}
		status := owner.Status()
		if status.State != venuebinance.SpotDepthBootstrapFailed {
			t.Fatalf("state = %q, want %q", status.State, venuebinance.SpotDepthBootstrapFailed)
		}
		if status.FailureReason != venuebinance.SpotDepthBootstrapFailureBridgeMissing {
			t.Fatalf("failure reason = %q, want %q", status.FailureReason, venuebinance.SpotDepthBootstrapFailureBridgeMissing)
		}
		if status.Synchronized {
			t.Fatal("did not expect synchronized status after missing bridging delta")
		}
	})
}

type integrationSpotDepthSnapshotFetcher struct {
	snapshotRaw json.RawMessage
}

func (f integrationSpotDepthSnapshotFetcher) FetchSpotDepthSnapshot(_ context.Context, sourceSymbol string) (venuebinance.SpotDepthSnapshotResponse, error) {
	if sourceSymbol == "" {
		return venuebinance.SpotDepthSnapshotResponse{}, fmt.Errorf("source symbol is required")
	}
	recvTime := mustRecvTimeFromFixtureMessage(f.snapshotRaw)
	return venuebinance.SpotDepthSnapshotResponse{
		Payload:  append([]byte(nil), f.snapshotRaw...),
		RecvTime: recvTime,
	}, nil
}

func loadDepthBootstrapFixture(t *testing.T, relativePath string) depthBootstrapFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}
	var fixture depthBootstrapFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	return fixture
}

func mustRecvTimeFromFixtureMessage(raw json.RawMessage) time.Time {
	var message timeFixtureMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		panic(err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, message.RecvTs)
	if err != nil {
		panic(err)
	}
	return parsed
}

func newBinanceVenueRuntime(t *testing.T) *venuebinance.Runtime {
	t.Helper()
	runtime, err := venuebinance.NewRuntime(loadRuntimeConfig(t, ingestion.VenueBinance))
	if err != nil {
		t.Fatalf("new binance runtime: %v", err)
	}
	return runtime
}
