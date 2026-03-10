package integration

import (
	"context"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceSpotTradeRawAppendUsesWebsocketContext(t *testing.T) {
	service := newNormalizerService(t)
	writer := ingestion.NewInMemoryRawEventWriter()
	fixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, fixture.RawMessages[0]).Add(-180*time.Millisecond))

	frame, err := supervisor.AcceptDataFrame(fixture.RawMessages[0], mustRecvTime(t, fixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept trade frame: %v", err)
	}
	parsed, err := venuebinance.ParseTradeFrame(frame)
	if err != nil {
		t.Fatalf("parse trade frame: %v", err)
	}
	metadata := ingestion.TradeMetadata{
		Symbol:        fixture.Symbol,
		SourceSymbol:  parsed.SourceSymbol,
		QuoteCurrency: fixture.QuoteCurrency,
		Venue:         ingestion.VenueBinance,
		MarketType:    "spot",
	}
	event, err := service.NormalizeTrade(normalizer.TradeInput{Metadata: metadata, Message: parsed.Message})
	if err != nil {
		t.Fatalf("normalize trade: %v", err)
	}
	entry, err := venuebinance.BuildSpotTradeRawAppendEntry(supervisor, event, metadata, parsed.Message, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build raw append trade entry: %v", err)
	}
	if err := writer.Append(entry); err != nil {
		t.Fatalf("append trade raw entry: %v", err)
	}

	entries := writer.Entries()
	if len(entries) != 1 {
		t.Fatalf("raw entry count = %d, want 1", len(entries))
	}
	if entries[0].ConnectionRef != "binance-spot-ws" {
		t.Fatalf("connection ref = %q, want %q", entries[0].ConnectionRef, "binance-spot-ws")
	}
	if entries[0].SessionRef != supervisor.State().SessionRef {
		t.Fatalf("session ref = %q, want %q", entries[0].SessionRef, supervisor.State().SessionRef)
	}
	if entries[0].PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE" {
		t.Fatalf("partition key = %q, want %q", entries[0].PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE")
	}
	if entries[0].DuplicateAudit.IdentityKey != "message:trades:1001" {
		t.Fatalf("identity key = %q, want %q", entries[0].DuplicateAudit.IdentityKey, "message:trades:1001")
	}
}

func TestIngestionBinanceSpotDepthFeedHealthRawAppendRetainsDegradedLinkage(t *testing.T) {
	service := newNormalizerService(t)
	writer := ingestion.NewInMemoryRawEventWriter()
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
		t.Fatalf("accept depth recovery frame: %v", err)
	}
	if err := owner.MarkSequenceGap(frame); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	if _, err := owner.Recover(context.Background(), mustRecvTime(t, fixture.RawMessages[0])); err == nil {
		t.Fatal("expected cooldown-blocked recovery to stay degraded")
	}
	input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{
		Symbol:          fixture.Symbol,
		QuoteCurrency:   fixture.QuoteCurrency,
		Now:             mustRecvTime(t, fixture.RawMessages[0]),
		ConnectionState: ingestion.ConnectionConnected,
	})
	if err != nil {
		t.Fatalf("depth feed health input: %v", err)
	}
	event, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
	if err != nil {
		t.Fatalf("normalize depth feed health: %v", err)
	}
	entry, err := venuebinance.BuildSpotDepthFeedHealthRawAppendEntry(supervisor, event, input.Metadata.SourceSymbol, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build depth feed health raw entry: %v", err)
	}
	if err := writer.Append(entry); err != nil {
		t.Fatalf("append depth feed health raw entry: %v", err)
	}

	entries := writer.Entries()
	if len(entries) != 1 {
		t.Fatalf("raw entry count = %d, want 1", len(entries))
	}
	if entries[0].StreamFamily != ingestion.RawStreamFamilyFeedHealth {
		t.Fatalf("stream family = %q, want %q", entries[0].StreamFamily, ingestion.RawStreamFamilyFeedHealth)
	}
	if entries[0].DegradedFeedRef != event.SourceRecordID {
		t.Fatalf("degraded feed ref = %q, want %q", entries[0].DegradedFeedRef, event.SourceRecordID)
	}
	if entries[0].StreamKey != string(ingestion.StreamOrderBook) {
		t.Fatalf("stream key = %q, want %q", entries[0].StreamKey, ingestion.StreamOrderBook)
	}
	if entries[0].PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/feed-health" {
		t.Fatalf("partition key = %q, want %q", entries[0].PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/feed-health")
	}
}

func TestIngestionBinanceUSDMMixedSurfaceRawAppendDistinctProvenance(t *testing.T) {
	service := newNormalizerService(t)
	writer := ingestion.NewInMemoryRawEventWriter()
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
		t.Fatalf("new open interest poller: %v", err)
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
	if _, err := usdm.AcceptMarkPriceFrame([]byte(`{"stream":"btcusdt@markPrice@1s"}`), base.Add(500*time.Millisecond)); err != nil {
		t.Fatalf("accept fresh mark price: %v", err)
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
	fundingMetadata := ingestion.DerivativesMetadata{Symbol: fundingFixture.Symbol, SourceSymbol: fundingParsed.SourceSymbol, QuoteCurrency: fundingFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"}
	openInterestMetadata := ingestion.DerivativesMetadata{Symbol: openInterestFixture.Symbol, SourceSymbol: openInterestParsed.SourceSymbol, QuoteCurrency: openInterestFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "perpetual"}
	fundingEvent, err := service.NormalizeFunding(normalizer.FundingInput{Metadata: fundingMetadata, Message: fundingParsed.Funding})
	if err != nil {
		t.Fatalf("normalize funding: %v", err)
	}
	openInterestEvent, err := service.NormalizeOpenInterest(normalizer.OpenInterestInput{Metadata: openInterestMetadata, Message: openInterestParsed.Message})
	if err != nil {
		t.Fatalf("normalize open interest: %v", err)
	}
	fundingEntry, err := venuebinance.BuildUSDMFundingRawAppendEntry(usdm, fundingEvent, fundingMetadata, fundingParsed.Funding, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build funding raw entry: %v", err)
	}
	openInterestEntry, err := venuebinance.BuildUSDMOpenInterestRawAppendEntry(poller, openInterestEvent, openInterestMetadata, openInterestParsed.Message, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build open-interest raw entry: %v", err)
	}
	if err := writer.Append(fundingEntry); err != nil {
		t.Fatalf("append funding raw entry: %v", err)
	}
	if err := writer.Append(openInterestEntry); err != nil {
		t.Fatalf("append open-interest raw entry: %v", err)
	}

	wsInputs, err := usdm.FeedHealthInputs(base.Add(16 * time.Second))
	if err != nil {
		t.Fatalf("ws feed health inputs: %v", err)
	}
	restInputs, err := poller.FeedHealthInputs(base.Add(16 * time.Second))
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
	wsHealthEntry, err := venuebinance.BuildUSDMWebsocketFeedHealthRawAppendEntry(usdm, wsEvent, "BTCUSDT", ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build websocket feed health raw entry: %v", err)
	}
	restHealthEntry, err := venuebinance.BuildUSDMOpenInterestFeedHealthRawAppendEntry(poller, restEvent, "BTCUSDT", ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build rest feed health raw entry: %v", err)
	}

	if fundingEntry.ConnectionRef == openInterestEntry.ConnectionRef {
		t.Fatal("expected websocket and rest raw append connections to stay distinct")
	}
	if fundingEntry.PartitionKey == openInterestEntry.PartitionKey {
		t.Fatal("expected websocket and rest raw append partitions to stay distinct")
	}
	if wsHealthEntry.VenueMessageID == restHealthEntry.VenueMessageID {
		t.Fatal("expected websocket and rest feed-health source identities to stay distinct")
	}
	if wsHealthEntry.StreamKey == restHealthEntry.StreamKey {
		t.Fatal("expected websocket and rest feed-health stream keys to stay distinct")
	}
}

func TestIngestionBinanceRawAppendDuplicateIdentityStaysDeterministic(t *testing.T) {
	service := newNormalizerService(t)
	writer := ingestion.NewInMemoryRawEventWriter()
	tradeFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json")
	topBookFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-native-top-book-usdt.fixture.v1.json")
	supervisor := connectedSpotSupervisor(t, mustRecvTime(t, tradeFixture.RawMessages[0]).Add(-180*time.Millisecond))
	tradeFrame, err := supervisor.AcceptDataFrame(tradeFixture.RawMessages[0], mustRecvTime(t, tradeFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept trade frame: %v", err)
	}
	tradeParsed, err := venuebinance.ParseTradeFrame(tradeFrame)
	if err != nil {
		t.Fatalf("parse trade frame: %v", err)
	}
	topBookFrame, err := supervisor.AcceptDataFrame(topBookFixture.RawMessages[0], mustRecvTime(t, topBookFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept top-of-book frame: %v", err)
	}
	topBookParsed, err := venuebinance.ParseTopOfBookFrame(topBookFrame)
	if err != nil {
		t.Fatalf("parse top-of-book frame: %v", err)
	}
	tradeMetadata := ingestion.TradeMetadata{Symbol: tradeFixture.Symbol, SourceSymbol: tradeParsed.SourceSymbol, QuoteCurrency: tradeFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "spot"}
	topBookMetadata := ingestion.BookMetadata{Symbol: topBookFixture.Symbol, SourceSymbol: topBookParsed.SourceSymbol, QuoteCurrency: topBookFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "spot"}
	options := ingestion.RawWriteOptions{BuildVersion: "test-build"}
	for i := 0; i < 2; i++ {
		tradeEvent, err := service.NormalizeTrade(normalizer.TradeInput{Metadata: tradeMetadata, Message: tradeParsed.Message})
		if err != nil {
			t.Fatalf("normalize trade %d: %v", i, err)
		}
		tradeEntry, err := venuebinance.BuildSpotTradeRawAppendEntry(supervisor, tradeEvent, tradeMetadata, tradeParsed.Message, options)
		if err != nil {
			t.Fatalf("build trade raw entry %d: %v", i, err)
		}
		if err := writer.Append(tradeEntry); err != nil {
			t.Fatalf("append trade raw entry %d: %v", i, err)
		}
		topBookEvent, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: topBookMetadata, Message: topBookParsed.Message, Sequencer: &ingestion.OrderBookSequencer{}})
		if err != nil {
			t.Fatalf("normalize top-of-book %d: %v", i, err)
		}
		if topBookEvent.OrderBookEvent == nil {
			t.Fatalf("expected top-of-book event for iteration %d", i)
		}
		topBookEntry, err := venuebinance.BuildSpotOrderBookRawAppendEntry(supervisor, *topBookEvent.OrderBookEvent, topBookMetadata, topBookParsed.Message, options)
		if err != nil {
			t.Fatalf("build top-of-book raw entry %d: %v", i, err)
		}
		if err := writer.Append(topBookEntry); err != nil {
			t.Fatalf("append top-of-book raw entry %d: %v", i, err)
		}
	}

	tradeEntries := rawEntriesForStreamFamily(writer.Entries(), string(ingestion.StreamTrades))
	if len(tradeEntries) != 2 {
		t.Fatalf("trade raw entry count = %d, want 2", len(tradeEntries))
	}
	if !tradeEntries[1].DuplicateAudit.Duplicate || tradeEntries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("trade duplicate audit = %+v, want duplicate occurrence 2", tradeEntries[1].DuplicateAudit)
	}
	topBookEntries := rawEntriesForStreamFamily(writer.Entries(), string(ingestion.StreamTopOfBook))
	if len(topBookEntries) != 2 {
		t.Fatalf("top-of-book raw entry count = %d, want 2", len(topBookEntries))
	}
	if !topBookEntries[1].DuplicateAudit.Duplicate || topBookEntries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("top-of-book duplicate audit = %+v, want duplicate occurrence 2", topBookEntries[1].DuplicateAudit)
	}
	if tradeEntries[0].PartitionKey != tradeEntries[1].PartitionKey {
		t.Fatalf("trade partitions drifted: %+v vs %+v", tradeEntries[0].PartitionKey, tradeEntries[1].PartitionKey)
	}
	if topBookEntries[0].PartitionKey != topBookEntries[1].PartitionKey {
		t.Fatalf("top-of-book partitions drifted: %+v vs %+v", topBookEntries[0].PartitionKey, topBookEntries[1].PartitionKey)
	}
}
