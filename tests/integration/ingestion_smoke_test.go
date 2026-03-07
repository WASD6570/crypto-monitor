package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
	venuebybit "github.com/crypto-market-copilot/alerts/services/venue-bybit"
	venuecoinbase "github.com/crypto-market-copilot/alerts/services/venue-coinbase"
	venuekraken "github.com/crypto-market-copilot/alerts/services/venue-kraken"
)

type fixtureEnvelope struct {
	Symbol            string            `json:"symbol"`
	SourceSymbol      string            `json:"sourceSymbol"`
	QuoteCurrency     string            `json:"quoteCurrency"`
	RawMessages       []json.RawMessage `json:"rawMessages"`
	ExpectedCanonical []json.RawMessage `json:"expectedCanonical"`
}

type timeFixtureMessage struct {
	ExchangeTs string `json:"exchangeTs"`
	RecvTs     string `json:"recvTs"`
}

func TestIngestionAdapterHappyPath(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/coinbase/BTC-USD/happy-trade-usd.fixture.v1.json")
	recvTime := mustRecvTime(t, fixture.RawMessages[0])

	parsed, err := venuecoinbase.ParseTradeEvent(fixture.RawMessages[0], recvTime)
	if err != nil {
		t.Fatalf("parse coinbase trade: %v", err)
	}

	actual, err := service.NormalizeTrade(normalizer.TradeInput{
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

	config := loadRuntimeConfig(t, ingestion.VenueCoinbase)
	runtime, err := venuecoinbase.NewRuntime(config)
	if err != nil {
		t.Fatalf("new coinbase runtime: %v", err)
	}
	status, err := runtime.EvaluateLoopState(venuecoinbase.AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   recvTime.Add(-5 * time.Second),
	}, recvTime)
	if err != nil {
		t.Fatalf("evaluate loop state: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("feed health state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	if len(status.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", status.Reasons)
	}
}

func TestIngestionGapPathEmitsDeterministicResyncOutput(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/bybit/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json")
	sequencer := &ingestion.OrderBookSequencer{}
	metadata := ingestion.BookMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  fixture.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueBybit,
		MarketType:    "spot",
	}

	snapshot, err := venuebybit.ParseOrderBookEvent(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse bybit snapshot: %v", err)
	}
	gap, err := venuebybit.ParseOrderBookEvent(fixture.RawMessages[1], mustRecvTime(t, fixture.RawMessages[1]))
	if err != nil {
		t.Fatalf("parse bybit gap delta: %v", err)
	}

	first, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: snapshot.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	second, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: metadata, Message: gap.Message, Sequencer: sequencer})
	if err != nil {
		t.Fatalf("normalize gap delta: %v", err)
	}

	if first.OrderBookEvent == nil {
		t.Fatal("expected canonical snapshot event before gap")
	}
	if second.OrderBookEvent != nil {
		t.Fatal("did not expect canonical order-book event after gap")
	}
	if second.FeedHealthEvent == nil {
		t.Fatal("expected feed-health event after gap")
	}
	assertCanonicalFeedHealthMatchesFixture(t, *second.FeedHealthEvent, fixture.ExpectedCanonical[0])
	if !containsReason(second.FeedHealthEvent.DegradationReasons, ingestion.ReasonSequenceGap) {
		t.Fatalf("degradation reasons = %v, want %q", second.FeedHealthEvent.DegradationReasons, ingestion.ReasonSequenceGap)
	}
}

func TestIngestionStalePathPreservesFeedHealthOutput(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/kraken/ETH-USD/edge-stale-feed-usd.fixture.v1.json")
	timeMessage := mustTimeMessage(t, fixture.RawMessages[0])
	now := mustParseTime(t, timeMessage.RecvTs)
	lastMessageAt := mustParseTime(t, timeMessage.ExchangeTs)

	config := loadRuntimeConfig(t, ingestion.VenueKraken)
	runtime, err := venuekraken.NewRuntime(config)
	if err != nil {
		t.Fatalf("new kraken runtime: %v", err)
	}
	status, err := runtime.EvaluateLoopState(venuekraken.AdapterLoopState{
		ConnectionState: ingestion.ConnectionConnected,
		LastMessageAt:   lastMessageAt,
	}, now)
	if err != nil {
		t.Fatalf("evaluate stale loop state: %v", err)
	}

	actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{
		Metadata: ingestion.FeedHealthMetadata{
			Symbol:        fixture.Symbol,
			SourceSymbol:  fixture.SourceSymbol,
			QuoteCurrency: fixture.QuoteCurrency,
			Venue:         ingestion.VenueKraken,
			MarketType:    "spot",
		},
		Message: ingestion.FeedHealthMessage{
			ExchangeTs:     timeMessage.ExchangeTs,
			RecvTs:         timeMessage.RecvTs,
			SourceRecordID: "heartbeat:" + timeMessage.ExchangeTs,
			Status:         status,
		},
	})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	assertCanonicalFeedHealthMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	if !containsReason(actual.DegradationReasons, ingestion.ReasonMessageStale) {
		t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, ingestion.ReasonMessageStale)
	}
}

func TestIngestionRetrySafetyStaysBounded(t *testing.T) {
	config := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(config)
	if err != nil {
		t.Fatalf("new binance runtime: %v", err)
	}

	now := time.UnixMilli(1772798525000).UTC()
	delay, err := runtime.ReconnectDelay(6)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}
	if delay != 5*time.Second {
		t.Fatalf("reconnect delay = %s, want %s", delay, 5*time.Second)
	}

	cooldown, err := runtime.SnapshotRecoveryStatus(now, now.Add(-500*time.Millisecond))
	if err != nil {
		t.Fatalf("snapshot recovery status: %v", err)
	}
	if cooldown.Ready {
		t.Fatal("expected snapshot recovery cooldown to remain active")
	}
	if cooldown.RemainingCooldown != 500*time.Millisecond {
		t.Fatalf("remaining cooldown = %s, want %s", cooldown.RemainingCooldown, 500*time.Millisecond)
	}

	rateLimit, err := runtime.SnapshotRecoveryRateLimitStatus(now, []time.Time{
		now.Add(-59 * time.Second),
		now.Add(-58 * time.Second),
		now.Add(-57 * time.Second),
		now.Add(-56 * time.Second),
		now.Add(-55 * time.Second),
		now.Add(-54 * time.Second),
		now.Add(-53 * time.Second),
		now.Add(-52 * time.Second),
		now.Add(-51 * time.Second),
		now.Add(-50 * time.Second),
		now.Add(-49 * time.Second),
		now.Add(-48 * time.Second),
		now.Add(-47 * time.Second),
		now.Add(-46 * time.Second),
		now.Add(-45 * time.Second),
		now.Add(-44 * time.Second),
		now.Add(-43 * time.Second),
		now.Add(-42 * time.Second),
		now.Add(-41 * time.Second),
		now.Add(-40 * time.Second),
		now.Add(-39 * time.Second),
		now.Add(-38 * time.Second),
		now.Add(-37 * time.Second),
		now.Add(-36 * time.Second),
		now.Add(-35 * time.Second),
		now.Add(-34 * time.Second),
		now.Add(-33 * time.Second),
		now.Add(-32 * time.Second),
		now.Add(-31 * time.Second),
		now.Add(-30 * time.Second),
	})
	if err != nil {
		t.Fatalf("snapshot recovery rate limit: %v", err)
	}
	if rateLimit.Allowed {
		t.Fatal("expected rate limit to block extra recovery attempt")
	}
	if rateLimit.RetryAfter != time.Second {
		t.Fatalf("retry after = %s, want %s", rateLimit.RetryAfter, time.Second)
	}
}

func TestIngestionRunbookAlignmentUsesSharedHealthVocabulary(t *testing.T) {
	paths := []string{
		"docs/runbooks/ingestion-feed-health-ops.md",
		"docs/runbooks/degraded-feed-investigation.md",
	}
	requiredTerms := []string{
		"HEALTHY",
		"DEGRADED",
		"STALE",
		"connection-not-ready",
		"message-stale",
		"snapshot-stale",
		"sequence-gap",
		"reconnect-loop",
		"resync-loop",
		"clock-degraded",
	}

	for _, relativePath := range paths {
		contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
		if err != nil {
			t.Fatalf("read runbook %s: %v", relativePath, err)
		}
		text := string(contents)
		for _, term := range requiredTerms {
			if !strings.Contains(text, term) {
				t.Fatalf("runbook %s missing term %q", relativePath, term)
			}
		}
	}
}

func newNormalizerService(t *testing.T) *normalizer.Service {
	t.Helper()
	service, err := normalizer.NewService(ingestion.StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("new normalizer service: %v", err)
	}
	return service
}

func loadRuntimeConfig(t *testing.T, venue ingestion.Venue) ingestion.VenueRuntimeConfig {
	t.Helper()
	config, err := ingestion.LoadEnvironmentConfig(filepath.Join(repoRoot(t), "configs/local/ingestion.v1.json"))
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}
	runtimeConfig, err := config.RuntimeConfigFor(venue)
	if err != nil {
		t.Fatalf("load runtime config for %s: %v", venue, err)
	}
	return runtimeConfig
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
	return mustParseTime(t, mustTimeMessage(t, raw).RecvTs)
}

func mustTimeMessage(t *testing.T, raw json.RawMessage) timeFixtureMessage {
	t.Helper()
	var message timeFixtureMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("decode time fixture message: %v", err)
	}
	return message
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
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

func assertCanonicalFeedHealthMatchesFixture(t *testing.T, actual ingestion.CanonicalFeedHealthEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalFeedHealthEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected feed-health event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical feed-health mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
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
