package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

type depthRecoveryFixture struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	SnapshotRaw       json.RawMessage   `json:"snapshotRaw"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
	ObserveAt         string            `json:"observeAt"`
}

func TestIngestionBinanceSpotDepthRecovery(t *testing.T) {
	t.Run("successful resync emits replacement depth", func(t *testing.T) {
		service := newNormalizerService(t)
		bootstrap := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json")
		recovery := loadDepthRecoveryFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-resync-usdt.fixture.v1.json")
		runtime, supervisor, sync := bootstrapSpotDepthPath(t, bootstrap)

		owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, integrationSpotDepthSnapshotFetcher{snapshotRaw: recovery.SnapshotRaw})
		if err != nil {
			t.Fatalf("new depth recovery owner: %v", err)
		}
		if err := owner.StartSynchronized(sync); err != nil {
			t.Fatalf("start synchronized recovery owner: %v", err)
		}

		gapFrame, err := supervisor.AcceptDataFrame(recovery.RawMessages[0], mustRecvTime(t, recovery.RawMessages[0]))
		if err != nil {
			t.Fatalf("accept gap frame: %v", err)
		}
		if err := owner.MarkSequenceGap(gapFrame); err != nil {
			t.Fatalf("mark sequence gap: %v", err)
		}
		deltaFrame, err := supervisor.AcceptDataFrame(recovery.RawMessages[1], mustRecvTime(t, recovery.RawMessages[1]))
		if err != nil {
			t.Fatalf("accept recovery delta frame: %v", err)
		}
		if err := owner.BufferRecoveryDelta(deltaFrame); err != nil {
			t.Fatalf("buffer recovery delta: %v", err)
		}

		recovered, err := owner.Recover(context.Background(), mustRecvTime(t, recovery.RawMessages[1]).Add(-220*time.Millisecond))
		if err != nil {
			t.Fatalf("recover synchronized depth: %v", err)
		}

		sequencer := &ingestion.OrderBookSequencer{}
		metadata := ingestion.BookMetadata{
			Symbol:        recovery.Symbol,
			SourceSymbol:  recovery.SourceSymbol,
			QuoteCurrency: recovery.QuoteCurrency,
			Venue:         ingestion.VenueBinance,
			MarketType:    "spot",
		}
		results := make([]ingestion.CanonicalOrderBookEvent, 0, 1+len(recovered.Deltas))
		snapshotResult, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: recovered.Snapshot.Message, Sequencer: sequencer})
		if err != nil {
			t.Fatalf("normalize replacement snapshot: %v", err)
		}
		results = append(results, *snapshotResult.OrderBookEvent)
		for _, delta := range recovered.Deltas {
			result, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: delta.Message, Sequencer: sequencer})
			if err != nil {
				t.Fatalf("normalize recovered delta: %v", err)
			}
			results = append(results, *result.OrderBookEvent)
		}
		if len(results) != len(recovery.ExpectedCanonical) {
			t.Fatalf("canonical event count = %d, want %d", len(results), len(recovery.ExpectedCanonical))
		}
		for idx, actual := range results {
			assertCanonicalOrderBookMatchesFixture(t, actual, recovery.ExpectedCanonical[idx])
		}
	})

	t.Run("cooldown blocked recovery stays degraded", func(t *testing.T) {
		service := newNormalizerService(t)
		bootstrap := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json")
		fixture := loadDepthRecoveryFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-depth-recovery-cooldown-blocked-usdt.fixture.v1.json")
		runtime, supervisor, sync := bootstrapSpotDepthPath(t, bootstrap)

		owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, integrationSpotDepthSnapshotFetcher{snapshotRaw: bootstrap.SnapshotRaw})
		if err != nil {
			t.Fatalf("new depth recovery owner: %v", err)
		}
		if err := owner.StartSynchronized(sync); err != nil {
			t.Fatalf("start synchronized recovery owner: %v", err)
		}
		frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
		if err != nil {
			t.Fatalf("accept cooldown frame: %v", err)
		}
		if err := owner.MarkSequenceGap(frame); err != nil {
			t.Fatalf("mark cooldown gap: %v", err)
		}
		if _, err := owner.Recover(context.Background(), mustRecvTime(t, fixture.RawMessages[0])); err == nil {
			t.Fatal("expected cooldown-blocked recovery to fail")
		}
		input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{
			Symbol:          fixture.Symbol,
			QuoteCurrency:   fixture.QuoteCurrency,
			Now:             mustRecvTime(t, fixture.RawMessages[0]),
			ConnectionState: ingestion.ConnectionConnected,
		})
		if err != nil {
			t.Fatalf("cooldown feed health input: %v", err)
		}
		actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
		if err != nil {
			t.Fatalf("normalize cooldown feed health: %v", err)
		}
		assertCanonicalFeedHealthMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	})

	t.Run("rate limit blocked recovery adds explicit reason", func(t *testing.T) {
		service := newNormalizerService(t)
		bootstrap := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json")
		fixture := loadDepthRecoveryFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-depth-recovery-rate-limit-usdt.fixture.v1.json")
		_, supervisor, sync := bootstrapSpotDepthPath(t, bootstrap)
		runtimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
		runtimeConfig.SnapshotRecoveryPerMinuteLimit = 1
		runtime, err := venuebinance.NewRuntime(runtimeConfig)
		if err != nil {
			t.Fatalf("new rate-limit runtime: %v", err)
		}

		owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, integrationSpotDepthSnapshotFetcher{snapshotRaw: bootstrap.SnapshotRaw})
		if err != nil {
			t.Fatalf("new depth recovery owner: %v", err)
		}
		if err := owner.StartSynchronized(sync); err != nil {
			t.Fatalf("start synchronized recovery owner: %v", err)
		}
		frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
		if err != nil {
			t.Fatalf("accept rate-limit frame: %v", err)
		}
		if err := owner.MarkSequenceGap(frame); err != nil {
			t.Fatalf("mark rate-limit gap: %v", err)
		}
		if _, err := owner.Recover(context.Background(), mustRecvTime(t, fixture.RawMessages[0])); err == nil {
			t.Fatal("expected rate-limit blocked recovery to fail")
		}
		input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{
			Symbol:          fixture.Symbol,
			QuoteCurrency:   fixture.QuoteCurrency,
			Now:             mustRecvTime(t, fixture.RawMessages[0]),
			ConnectionState: ingestion.ConnectionConnected,
		})
		if err != nil {
			t.Fatalf("rate-limit feed health input: %v", err)
		}
		actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
		if err != nil {
			t.Fatalf("normalize rate-limit feed health: %v", err)
		}
		assertCanonicalFeedHealthMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	})

	t.Run("snapshot stale remains visible while messages stay fresh", func(t *testing.T) {
		service := newNormalizerService(t)
		bootstrap := loadDepthBootstrapFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json")
		fixture := loadDepthRecoveryFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-depth-snapshot-stale-usdt.fixture.v1.json")
		runtime, supervisor, sync := bootstrapSpotDepthPath(t, bootstrap)

		owner, err := venuebinance.NewSpotDepthRecoveryOwner(runtime, integrationSpotDepthSnapshotFetcher{snapshotRaw: bootstrap.SnapshotRaw})
		if err != nil {
			t.Fatalf("new depth recovery owner: %v", err)
		}
		if err := owner.StartSynchronized(sync); err != nil {
			t.Fatalf("start synchronized recovery owner: %v", err)
		}
		frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
		if err != nil {
			t.Fatalf("accept fresh synchronized delta: %v", err)
		}
		if err := owner.AcceptSynchronizedDelta(frame); err != nil {
			t.Fatalf("accept synchronized delta: %v", err)
		}
		observeAt := mustParseTime(t, fixture.ObserveAt)
		input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{
			Symbol:          fixture.Symbol,
			QuoteCurrency:   fixture.QuoteCurrency,
			Now:             observeAt,
			ConnectionState: ingestion.ConnectionConnected,
		})
		if err != nil {
			t.Fatalf("snapshot-stale feed health input: %v", err)
		}
		actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
		if err != nil {
			t.Fatalf("normalize snapshot-stale feed health: %v", err)
		}
		assertCanonicalFeedHealthMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	})
}

func TestIngestionBinanceSpotDepthLive(t *testing.T) {
	if os.Getenv("BINANCE_LIVE_VALIDATION") != "1" {
		t.Skip("set BINANCE_LIVE_VALIDATION=1 to run live Binance depth validation")
	}
	response, err := http.Get("https://api.binance.com/api/v3/depth?symbol=BTCUSDT&limit=5")
	if err != nil {
		t.Fatalf("fetch live depth snapshot: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("live depth status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read live depth body: %v", err)
	}
	parsed, err := venuebinance.ParseOrderBookSnapshotWithSourceSymbol(body, "BTCUSDT", time.Now().UTC())
	if err != nil {
		t.Fatalf("parse live depth snapshot: %v", err)
	}
	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", parsed.SourceSymbol, "BTCUSDT")
	}
	if parsed.FinalSequence <= 0 {
		t.Fatalf("snapshot sequence = %d, want positive", parsed.FinalSequence)
	}
}

func loadDepthRecoveryFixture(t *testing.T, relativePath string) depthRecoveryFixture {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}
	var fixture depthRecoveryFixture
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	return fixture
}

func bootstrapSpotDepthPath(t *testing.T, fixture depthBootstrapFixture) (*venuebinance.Runtime, *venuebinance.SpotWebsocketSupervisor, venuebinance.SpotDepthBootstrapSync) {
	t.Helper()
	runtime := newBinanceVenueRuntime(t)
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-200*time.Millisecond))
	owner, err := venuebinance.NewSpotDepthBootstrapOwner(runtime, integrationSpotDepthSnapshotFetcher{snapshotRaw: fixture.SnapshotRaw})
	if err != nil {
		t.Fatalf("new depth bootstrap owner: %v", err)
	}
	for _, raw := range fixture.RawMessages {
		frame, err := supervisor.AcceptDataFrame(raw, mustRecvTime(t, raw))
		if err != nil {
			t.Fatalf("accept bootstrap frame: %v", err)
		}
		if err := owner.BufferDelta(frame); err != nil {
			t.Fatalf("buffer bootstrap frame: %v", err)
		}
	}
	sync, err := owner.Synchronize(context.Background())
	if err != nil {
		t.Fatalf("bootstrap synchronize: %v", err)
	}
	return runtime, supervisor, sync
}
