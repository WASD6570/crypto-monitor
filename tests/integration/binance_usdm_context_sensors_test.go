package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceUSDMMixedSurfaceDistinctHealthVisibility(t *testing.T) {
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
	poller, err := venuebinance.NewUSDMOpenInterestPoller(runtime)
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	base := time.UnixMilli(1772798404000).UTC()
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
	if _, err := poller.BeginPoll("BTCUSDT", base); err != nil {
		t.Fatalf("begin poll: %v", err)
	}
	if _, err := poller.AcceptSnapshot([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798404000}`), base.Add(60*time.Millisecond)); err != nil {
		t.Fatalf("accept snapshot: %v", err)
	}

	observe := base.Add(16 * time.Second)
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), observe.Add(-500*time.Millisecond)); err != nil {
		t.Fatalf("accept fresh mark price: %v", err)
	}

	wsInputs, err := usdm.FeedHealthInputs(observe)
	if err != nil {
		t.Fatalf("ws feed health inputs: %v", err)
	}
	restInputs, err := poller.FeedHealthInputs(observe)
	if err != nil {
		t.Fatalf("rest feed health inputs: %v", err)
	}
	wsEvent, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: mustUSDMFeedHealthInput(t, wsInputs, "BTCUSDT").Metadata, Message: mustUSDMFeedHealthInput(t, wsInputs, "BTCUSDT").Message})
	if err != nil {
		t.Fatalf("normalize websocket feed health: %v", err)
	}
	restEvent, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: mustUSDMFeedHealthInput(t, restInputs, "BTCUSDT").Metadata, Message: mustUSDMFeedHealthInput(t, restInputs, "BTCUSDT").Message})
	if err != nil {
		t.Fatalf("normalize rest feed health: %v", err)
	}

	if wsEvent.FeedHealthState != ingestion.FeedHealthHealthy {
		t.Fatalf("websocket feed health state = %q, want %q", wsEvent.FeedHealthState, ingestion.FeedHealthHealthy)
	}
	if restEvent.FeedHealthState != ingestion.FeedHealthStale {
		t.Fatalf("rest feed health state = %q, want %q", restEvent.FeedHealthState, ingestion.FeedHealthStale)
	}
	if !containsReason(restEvent.DegradationReasons, ingestion.ReasonMessageStale) {
		t.Fatalf("rest degradation reasons = %v, want %q", restEvent.DegradationReasons, ingestion.ReasonMessageStale)
	}
	if wsEvent.SourceRecordID == restEvent.SourceRecordID {
		t.Fatal("expected websocket and rest source record ids to stay distinct")
	}
	if !strings.HasPrefix(wsEvent.SourceRecordID, "runtime:binance-usdm-ws:") {
		t.Fatalf("websocket source record id = %q", wsEvent.SourceRecordID)
	}
	if !strings.HasPrefix(restEvent.SourceRecordID, "runtime:binance-usdm-open-interest:") {
		t.Fatalf("rest source record id = %q", restEvent.SourceRecordID)
	}
}

func TestIngestionBinanceUSDMMixedSurfaceDegradedReasonsStayDistinct(t *testing.T) {
	service := newNormalizerService(t)
	wsRuntimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	wsRuntime, err := venuebinance.NewRuntime(wsRuntimeConfig)
	if err != nil {
		t.Fatalf("new websocket runtime: %v", err)
	}
	usdm, err := venuebinance.NewUSDMRuntime(wsRuntime)
	if err != nil {
		t.Fatalf("new usdm runtime: %v", err)
	}
	restRuntimeConfig := loadRuntimeConfig(t, ingestion.VenueBinance)
	restRuntimeConfig.OpenInterestPollInterval = time.Millisecond
	restRuntimeConfig.OpenInterestPollsPerMinuteLimit = 1
	restRuntime, err := venuebinance.NewRuntime(restRuntimeConfig)
	if err != nil {
		t.Fatalf("new rest runtime: %v", err)
	}
	poller, err := venuebinance.NewUSDMOpenInterestPoller(restRuntime)
	if err != nil {
		t.Fatalf("new poller: %v", err)
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
	if _, err := poller.BeginPoll("BTCUSDT", base); err != nil {
		t.Fatalf("begin first poll: %v", err)
	}
	if _, err := poller.AcceptSnapshot([]byte(`{"symbol":"BTCUSDT","openInterest":"10659.509","time":1772798525000}`), base.Add(100*time.Microsecond)); err != nil {
		t.Fatalf("accept first snapshot: %v", err)
	}
	if _, err := poller.BeginPoll("BTCUSDT", base.Add(time.Millisecond)); err == nil {
		t.Fatal("expected second poll to hit the rate limit")
	}

	wsInputs, err := usdm.FeedHealthInputs(usdm.State().LastFrameAt.Add(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("ws feed health inputs: %v", err)
	}
	restInputs, err := poller.FeedHealthInputs(base.Add(time.Millisecond))
	if err != nil {
		t.Fatalf("rest feed health inputs: %v", err)
	}
	wsEvent, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: mustUSDMFeedHealthInput(t, wsInputs, "BTCUSDT").Metadata, Message: mustUSDMFeedHealthInput(t, wsInputs, "BTCUSDT").Message})
	if err != nil {
		t.Fatalf("normalize websocket feed health: %v", err)
	}
	restEvent, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: mustUSDMFeedHealthInput(t, restInputs, "BTCUSDT").Metadata, Message: mustUSDMFeedHealthInput(t, restInputs, "BTCUSDT").Message})
	if err != nil {
		t.Fatalf("normalize rest feed health: %v", err)
	}

	if wsEvent.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("websocket feed health state = %q, want %q", wsEvent.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	if !containsReason(wsEvent.DegradationReasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("websocket degradation reasons = %v, want %q", wsEvent.DegradationReasons, ingestion.ReasonReconnectLoop)
	}
	if containsReason(wsEvent.DegradationReasons, ingestion.ReasonRateLimit) {
		t.Fatalf("websocket degradation reasons = %v, did not want %q", wsEvent.DegradationReasons, ingestion.ReasonRateLimit)
	}
	if restEvent.FeedHealthState != ingestion.FeedHealthDegraded {
		t.Fatalf("rest feed health state = %q, want %q", restEvent.FeedHealthState, ingestion.FeedHealthDegraded)
	}
	if !containsReason(restEvent.DegradationReasons, ingestion.ReasonRateLimit) {
		t.Fatalf("rest degradation reasons = %v, want %q", restEvent.DegradationReasons, ingestion.ReasonRateLimit)
	}
	if containsReason(restEvent.DegradationReasons, ingestion.ReasonReconnectLoop) {
		t.Fatalf("rest degradation reasons = %v, did not want %q", restEvent.DegradationReasons, ingestion.ReasonReconnectLoop)
	}
}

func TestIngestionBinanceUSDMMixedSurfaceDuplicateSourceIdentitiesStayStable(t *testing.T) {
	writer := ingestion.NewInMemoryRawEventWriter()
	service, err := normalizer.NewService(
		ingestion.StrictTimestampPolicy(),
		normalizer.WithRawEventWriter(writer, ingestion.RawWriteOptions{
			NormalizerService: "services/normalizer",
			BuildVersion:      "test-build",
		}),
	)
	if err != nil {
		t.Fatalf("new normalizer service: %v", err)
	}

	fundingFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-funding-usdt.fixture.v1.json")
	openInterestFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-open-interest-usdt.fixture.v1.json")
	fundingParsed, err := venuebinance.ParseUSDMMarkPrice(fundingFixture.RawMessages[0], mustRecvTime(t, fundingFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse funding mark price: %v", err)
	}
	openInterestParsed, err := venuebinance.ParseUSDMOpenInterest(openInterestFixture.RawMessages[0], mustRecvTime(t, openInterestFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse open interest: %v", err)
	}

	var fundingID string
	var openInterestID string
	for i := 0; i < 2; i++ {
		fundingActual, err := service.NormalizeFunding(normalizer.FundingInput{
			Metadata: ingestion.DerivativesMetadata{Symbol: fundingFixture.Symbol, SourceSymbol: fundingParsed.SourceSymbol, QuoteCurrency: fundingFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
			Message:  fundingParsed.Funding,
			Raw: ingestion.RawWriteContext{
				ConnectionRef: "binance-usdm-ws",
				SessionRef:    "binance-usdm-ws-session",
			},
		})
		if err != nil {
			t.Fatalf("normalize funding: %v", err)
		}
		openInterestActual, err := service.NormalizeOpenInterest(normalizer.OpenInterestInput{
			Metadata: ingestion.DerivativesMetadata{Symbol: openInterestFixture.Symbol, SourceSymbol: openInterestParsed.SourceSymbol, QuoteCurrency: openInterestFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"},
			Message:  openInterestParsed.Message,
			Raw: ingestion.RawWriteContext{
				ConnectionRef: "binance-usdm-rest",
				SessionRef:    "binance-usdm-rest-session",
			},
		})
		if err != nil {
			t.Fatalf("normalize open interest: %v", err)
		}
		if i == 0 {
			fundingID = fundingActual.SourceRecordID
			openInterestID = openInterestActual.SourceRecordID
			continue
		}
		if fundingActual.SourceRecordID != fundingID {
			t.Fatalf("funding sourceRecordId = %q, want %q", fundingActual.SourceRecordID, fundingID)
		}
		if openInterestActual.SourceRecordID != openInterestID {
			t.Fatalf("open interest sourceRecordId = %q, want %q", openInterestActual.SourceRecordID, openInterestID)
		}
	}

	entries := writer.Entries()
	if len(entries) != 4 {
		t.Fatalf("raw entry count = %d, want 4", len(entries))
	}
	fundingEntries := rawEntriesForStreamFamily(entries, string(ingestion.StreamFundingRate))
	if len(fundingEntries) != 2 {
		t.Fatalf("funding raw entry count = %d, want 2", len(fundingEntries))
	}
	if !fundingEntries[1].DuplicateAudit.Duplicate || fundingEntries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("funding duplicate audit = %+v, want duplicate occurrence 2", fundingEntries[1].DuplicateAudit)
	}
	if fundingEntries[0].PartitionKey.StreamFamily != string(ingestion.StreamFundingRate) {
		t.Fatalf("funding partition stream family = %q, want %q", fundingEntries[0].PartitionKey.StreamFamily, ingestion.StreamFundingRate)
	}
	openInterestEntries := rawEntriesForStreamFamily(entries, string(ingestion.StreamOpenInterest))
	if len(openInterestEntries) != 2 {
		t.Fatalf("open-interest raw entry count = %d, want 2", len(openInterestEntries))
	}
	if !openInterestEntries[1].DuplicateAudit.Duplicate || openInterestEntries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("open-interest duplicate audit = %+v, want duplicate occurrence 2", openInterestEntries[1].DuplicateAudit)
	}
	if openInterestEntries[0].PartitionKey.StreamFamily != string(ingestion.StreamOpenInterest) {
		t.Fatalf("open-interest partition stream family = %q, want %q", openInterestEntries[0].PartitionKey.StreamFamily, ingestion.StreamOpenInterest)
	}
	if fundingEntries[0].DuplicateAudit.IdentityKey == openInterestEntries[0].DuplicateAudit.IdentityKey {
		t.Fatal("expected websocket and rest duplicate identities to remain distinct")
	}
}

func mustUSDMFeedHealthInput(t *testing.T, inputs []venuebinance.USDMFeedHealthInput, sourceSymbol string) venuebinance.USDMFeedHealthInput {
	t.Helper()
	for _, input := range inputs {
		if input.Metadata.SourceSymbol == sourceSymbol {
			return input
		}
	}
	t.Fatalf("missing feed health input for %s", sourceSymbol)
	return venuebinance.USDMFeedHealthInput{}
}

func rawEntriesForStreamFamily(entries []ingestion.RawAppendEntry, streamFamily string) []ingestion.RawAppendEntry {
	filtered := make([]ingestion.RawAppendEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.StreamFamily == streamFamily {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
