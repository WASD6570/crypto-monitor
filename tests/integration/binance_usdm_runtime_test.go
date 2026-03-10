package integration

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceUSDMHappyPath(t *testing.T) {
	service := newNormalizerService(t)
	fundingFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-funding-usdt.fixture.v1.json")
	openInterestFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-open-interest-usdt.fixture.v1.json")
	openInterestETHFixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/happy-open-interest-usdt.fixture.v1.json")
	markFixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/happy-mark-index-usdt.fixture.v1.json")
	liquidationFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-liquidation-usdt.fixture.v1.json")

	openInterestParsed, err := venuebinance.ParseUSDMOpenInterest(openInterestFixture.RawMessages[0], mustRecvTime(t, openInterestFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse open interest: %v", err)
	}
	openInterestActual, err := service.NormalizeOpenInterest(normalizer.OpenInterestInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: openInterestFixture.Symbol, SourceSymbol: openInterestParsed.SourceSymbol, QuoteCurrency: openInterestFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  openInterestParsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize open interest: %v", err)
	}
	assertCanonicalOpenInterestMatchesFixture(t, openInterestActual, openInterestFixture.ExpectedCanonical[0])

	openInterestETHParsed, err := venuebinance.ParseUSDMOpenInterest(openInterestETHFixture.RawMessages[0], mustRecvTime(t, openInterestETHFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse eth open interest: %v", err)
	}
	openInterestETHActual, err := service.NormalizeOpenInterest(normalizer.OpenInterestInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: openInterestETHFixture.Symbol, SourceSymbol: openInterestETHParsed.SourceSymbol, QuoteCurrency: openInterestETHFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  openInterestETHParsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize eth open interest: %v", err)
	}
	assertCanonicalOpenInterestMatchesFixture(t, openInterestETHActual, openInterestETHFixture.ExpectedCanonical[0])

	fundingParsed, err := venuebinance.ParseUSDMMarkPrice(fundingFixture.RawMessages[0], mustRecvTime(t, fundingFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse funding mark price: %v", err)
	}
	fundingActual, err := service.NormalizeFunding(normalizer.FundingInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: fundingFixture.Symbol, SourceSymbol: fundingParsed.SourceSymbol, QuoteCurrency: fundingFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  fundingParsed.Funding,
	})
	if err != nil {
		t.Fatalf("normalize funding: %v", err)
	}
	assertCanonicalFundingMatchesFixture(t, fundingActual, fundingFixture.ExpectedCanonical[0])

	markParsed, err := venuebinance.ParseUSDMMarkPrice(markFixture.RawMessages[0], mustRecvTime(t, markFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse mark index mark price: %v", err)
	}
	markActual, err := service.NormalizeMarkIndex(normalizer.MarkIndexInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: markFixture.Symbol, SourceSymbol: markParsed.SourceSymbol, QuoteCurrency: markFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  markParsed.MarkIndex,
	})
	if err != nil {
		t.Fatalf("normalize mark index: %v", err)
	}
	assertCanonicalMarkIndexMatchesFixture(t, markActual, markFixture.ExpectedCanonical[0])

	liquidationParsed, err := venuebinance.ParseUSDMForceOrder(liquidationFixture.RawMessages[0], mustRecvTime(t, liquidationFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse liquidation: %v", err)
	}
	liquidationActual, err := service.NormalizeLiquidation(normalizer.LiquidationInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: liquidationFixture.Symbol, SourceSymbol: liquidationParsed.SourceSymbol, QuoteCurrency: liquidationFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  liquidationParsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize liquidation: %v", err)
	}
	assertCanonicalLiquidationMatchesFixture(t, liquidationActual, liquidationFixture.ExpectedCanonical[0])
}

func TestIngestionBinanceUSDMOpenInterestTimestampDegraded(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/ETH-USD/edge-timestamp-degraded-open-interest-usdt.fixture.v1.json")
	parsed, err := venuebinance.ParseUSDMOpenInterest(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse degraded open interest: %v", err)
	}
	actual, err := service.NormalizeOpenInterest(normalizer.OpenInterestInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: fixture.Symbol, SourceSymbol: parsed.SourceSymbol, QuoteCurrency: fixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.Message,
	})
	if err != nil {
		t.Fatalf("normalize degraded open interest: %v", err)
	}
	assertCanonicalOpenInterestMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	if actual.TimestampFallbackReason != ingestion.TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, ingestion.TimestampReasonExchangeMissingOrInvalid)
	}
}

func TestIngestionBinanceUSDMTimestampDegraded(t *testing.T) {
	service := newNormalizerService(t)
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/edge-timestamp-degraded-funding-usdt.fixture.v1.json")
	parsed, err := venuebinance.ParseUSDMMarkPrice(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse degraded mark price: %v", err)
	}
	actual, err := service.NormalizeFunding(normalizer.FundingInput{
		Metadata: ingestion.DerivativesMetadata{Symbol: fixture.Symbol, SourceSymbol: parsed.SourceSymbol, QuoteCurrency: fixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
		Message:  parsed.Funding,
	})
	if err != nil {
		t.Fatalf("normalize degraded funding: %v", err)
	}
	assertCanonicalFundingMatchesFixture(t, actual, fixture.ExpectedCanonical[0])
	if actual.TimestampFallbackReason != ingestion.TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, ingestion.TimestampReasonExchangeSkewExceeded)
	}
}

func TestIngestionBinanceUSDMReconnect(t *testing.T) {
	service := newNormalizerService(t)
	runtimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := venuebinance.NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept mark price: %v", err)
	}
	for range 4 {
		plan, err := usdm.HandleDisconnect(base.Add(time.Second), venuebinance.USDMReconnectCauseTransport)
		if err != nil {
			t.Fatalf("handle disconnect: %v", err)
		}
		if err := usdm.StartConnect(plan.RetryAt); err != nil {
			t.Fatalf("restart connect: %v", err)
		}
		command, err = usdm.CompleteConnect(plan.RetryAt.Add(50 * time.Millisecond))
		if err != nil {
			t.Fatalf("complete reconnect: %v", err)
		}
	}
	state := usdm.State()
	inputs, err := usdm.FeedHealthInputs(state.LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("feed health inputs: %v", err)
	}
	actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: inputs[0].Metadata, Message: inputs[0].Message})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	if actual.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	if !containsReason(actual.DegradationReasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, ingestion.ReasonReconnectLoop)
	}
}

func TestIngestionBinanceUSDMNoLiquidationStale(t *testing.T) {
	runtimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	usdm, err := venuebinance.NewUSDMRuntime(runtime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}

	base := time.UnixMilli(1772798525000).UTC()
	if err := usdm.StartConnect(base); err != nil {
		t.Fatalf("start connect: %v", err)
	}
	command, err := usdm.CompleteConnect(base.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("complete connect: %v", err)
	}
	if err := usdm.AckSubscribe(base.Add(200*time.Millisecond), command.ID); err != nil {
		t.Fatalf("ack subscribe: %v", err)
	}
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"ethusdt@markPrice@1s"}`), base.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("accept mark price: %v", err)
	}
	if _, err := usdm.AcceptForceOrderFrame([]byte(`{"stream":"ethusdt@forceOrder"}`), base.Add(12*time.Second)); err != nil {
		t.Fatalf("accept sparse liquidation: %v", err)
	}
	status, err := usdm.HealthStatus(base.Add(12*time.Second + 500*time.Millisecond))
	if err != nil {
		t.Fatalf("health status: %v", err)
	}
	if status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("state = %q, want %q", status.State, ingestion.FeedHealthHealthy)
	}
	if containsReason(status.Reasons, ingestion.ReasonMessageStale) {
		t.Fatalf("reasons = %v, did not want stale", status.Reasons)
	}
}

func TestIngestionBinanceUSDMOpenInterestPollHealth(t *testing.T) {
	service := newNormalizerService(t)
	runtimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	runtime, err := venuebinance.NewRuntime(runtimeConfig)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	poller, err := venuebinance.NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	base := time.UnixMilli(1772798404000).UTC()
	if _, err := poller.BeginPoll("BTCUSDT", base); err != nil {
		t.Fatalf("begin poll: %v", err)
	}
	if _, err := poller.AcceptSnapshot([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), base.Add(60*time.Millisecond)); err != nil {
		t.Fatalf("accept snapshot: %v", err)
	}
	inputs, err := poller.FeedHealthInputs(base.Add(16 * time.Second))
	if err != nil {
		t.Fatalf("feed health inputs: %v", err)
	}
	actual, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: inputs[0].Metadata, Message: inputs[0].Message})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	if actual.FeedHealthState != ingestion.FeedHealthStale {
		t.Fatalf("feed health state = %q, want %q", actual.FeedHealthState, ingestion.FeedHealthStale)
	}
	if !containsReason(actual.DegradationReasons, ingestion.ReasonMessageStale) {
		t.Fatalf("degradation reasons = %v, want %q", actual.DegradationReasons, ingestion.ReasonMessageStale)
	}
	if actual.SourceSymbol != "BTCUSDT" || actual.MarketType != "perpetual" {
		t.Fatalf("unexpected feed health provenance: %+v", actual)
	}
}

func assertCanonicalFundingMatchesFixture(t *testing.T, actual ingestion.CanonicalFundingRateEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalFundingRateEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected funding event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical funding mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalMarkIndexMatchesFixture(t *testing.T, actual ingestion.CanonicalMarkIndexEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalMarkIndexEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected mark-index event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical mark-index mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalOpenInterestMatchesFixture(t *testing.T, actual ingestion.CanonicalOpenInterestSnapshotEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalOpenInterestSnapshotEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected open-interest event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical open-interest mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func assertCanonicalLiquidationMatchesFixture(t *testing.T, actual ingestion.CanonicalLiquidationPrintEvent, expectedJSON json.RawMessage) {
	t.Helper()
	var expected ingestion.CanonicalLiquidationPrintEvent
	if err := json.Unmarshal(expectedJSON, &expected); err != nil {
		t.Fatalf("decode expected liquidation event: %v", err)
	}
	actual.CanonicalEventTime = time.Time{}
	actual.TimestampFallbackReason = ingestion.TimestampReasonNone
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("canonical liquidation mismatch\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}
