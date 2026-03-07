package normalizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	venuecoinbase "github.com/crypto-market-copilot/alerts/services/venue-coinbase"
	venuekraken "github.com/crypto-market-copilot/alerts/services/venue-kraken"
)

type fixtureEnvelope struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
}

type fixtureMessage struct {
	RecvTs string `json:"recvTs"`
}

func TestServiceNormalizeTradePreservesCanonicalOutput(t *testing.T) {
	service := newService(t)
	fixture := loadFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])

	parsed, err := venuecoinbase.ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse trade event: %v", err)
	}

	actual, err := service.NormalizeTrade(TradeInput{
		Metadata: ingestion.TradeMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  parsed.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueCoinbase,
			MarketType:    "spot",
		},
		Message: parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize trade: %v", err)
	}

	assertCanonicalTradeMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
}

func TestServiceNormalizeOrderBookPreservesCanonicalOutput(t *testing.T) {
	service := newService(t)
	fixture := loadFixture(t, "tests/fixtures/events/kraken/ETH-USD/happy-order-book-snapshot-delta-usd.fixture.v1.json")
	sequencer := &ingestion.OrderBookSequencer{}

	snapshotRaw := []byte(`{"channel":"book","type":"snapshot","pair":"ETH/USD","sequence":500,"bids":[["3499.10","1.20"]],"asks":[["3499.40","0.95"]],"exchangeTs":"2026-03-06T12:00:02.000Z"}`)
	deltaRaw := []byte(`{"channel":"book","type":"delta","pair":"ETH/USD","sequence":501,"bids":[["3499.20","1.00"]],"asks":[["3499.50","0.90"]],"exchangeTs":"2026-03-06T12:00:02.100Z"}`)

	snapshot, err := venuekraken.ParseOrderBookEvent(snapshotRaw, "2026-03-06T12:00:02.040Z")
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	delta, err := venuekraken.ParseOrderBookEvent(deltaRaw, "2026-03-06T12:00:02.130Z")
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}

	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueKraken,
		MarketType:    "spot",
	}

	first, err := service.NormalizeOrderBook(OrderBookInput{Metadata: metadata, Message: snapshot.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := service.NormalizeOrderBook(OrderBookInput{Metadata: metadata, Message: delta.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize delta: %v", err)
	}

	if first.OrderBookEvent == nil || second.OrderBookEvent == nil {
		t.Fatal("expected canonical order-book events")
	}
	assertCanonicalOrderBookMatchesFixture(t, *first.OrderBookEvent, fixture.ExpectedCanonical[0])
	assertCanonicalOrderBookMatchesFixture(t, *second.OrderBookEvent, fixture.ExpectedCanonical[1])
}

func TestServiceNormalizeFeedHealthPreservesDegradationMetadata(t *testing.T) {
	service := newService(t)
	config := loadKrakenRuntimeConfig(t)
	runtime, err := venuekraken.NewRuntime(config)
	if err != nil {
		t.Fatalf("new kraken runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	status, err := runtime.EvaluateLoopState(venuekraken.AdapterLoopState{
		ConnectionState:       ingestion.ConnectionReconnecting,
		LastMessageAt:         now.Add(-5 * time.Second),
		LastSnapshotAt:        now.Add(-10 * time.Second),
		SequenceGapDetected:   true,
		LocalClockOffset:      250 * time.Millisecond,
		ConsecutiveReconnects: 3,
		ResyncCount:           2,
	}, now)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}

	actual, err := service.NormalizeFeedHealth(FeedHealthInput{
		Metadata: ingestion.FeedHealthMetadata{
			Symbol:        "ETH-USD",
			SourceSymbol:  "ETH/USD",
			QuoteCurrency: "USD",
			Venue:         ingestion.VenueKraken,
			MarketType:    "spot",
		},
		Message: ingestion.FeedHealthMessage{
			ExchangeTs:     "2026-03-06T12:02:05Z",
			RecvTs:         "2026-03-06T12:02:05.02Z",
			SourceRecordID: "runtime:kraken-loop",
			Status:         status,
		},
	})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}

	if actual.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	for _, reason := range []ingestion.DegradationReason{
		ingestion.ReasonClockDegraded,
		ingestion.ReasonConnectionNotReady,
		ingestion.ReasonReconnectLoop,
		ingestion.ReasonResyncLoop,
		ingestion.ReasonSequenceGap,
	} {
		if !containsReason(actual.DegradationReasons, reason) {
			t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, reason)
		}
	}
	if actual.SourceRecordID != "runtime:kraken-loop" {
		t.Fatalf("source record id = %q, want %q", actual.SourceRecordID, "runtime:kraken-loop")
	}
	if actual.ExchangeTs != "2026-03-06T12:02:05Z" {
		t.Fatalf("exchangeTs = %q, want %q", actual.ExchangeTs, "2026-03-06T12:02:05Z")
	}
}

func newService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

func loadFixture(t *testing.T, relativePath string) fixtureEnvelope {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}
	var fixture fixtureEnvelope
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatalf("decode fixture %s: %v", relativePath, err)
	}
	return fixture
}

func mustRecvTime(t *testing.T, raw json.RawMessage) time.Time {
	t.Helper()
	var message fixtureMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("decode fixture raw message: %v", err)
	}
	recvTime, err := time.Parse(time.RFC3339Nano, message.RecvTs)
	if err != nil {
		t.Fatalf("parse recv timestamp: %v", err)
	}
	return recvTime
}

func loadKrakenRuntimeConfig(t *testing.T) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}
	runtimeConfig, err := config.RuntimeConfigFor(ingestion.VenueKraken)
	if err != nil {
		t.Fatalf("load kraken runtime config: %v", err)
	}
	return runtimeConfig
}

func assertCanonicalTradeMatchesFixture(t *testing.T, actual ingestion.CanonicalTradeEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalTradeEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected trade event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical trade mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalOrderBookMatchesFixture(t *testing.T, actual ingestion.CanonicalOrderBookEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalOrderBookEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected order-book event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical order-book mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func containsReason(reasons []ingestion.DegradationReason, target ingestion.DegradationReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}
