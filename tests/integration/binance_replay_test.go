package integration

import (
	"context"
	"testing"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	"github.com/crypto-market-copilot/alerts/services/normalizer"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

func TestIngestionBinanceSpotReplayPreservesSharedPartitionDeterminism(t *testing.T) {
	service := newNormalizerService(t)
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
	tradeMetadata := ingestion.TradeMetadata{Symbol: tradeFixture.Symbol, SourceSymbol: tradeParsed.SourceSymbol, QuoteCurrency: tradeFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "spot"}
	tradeEvent, err := service.NormalizeTrade(normalizer.TradeInput{Metadata: tradeMetadata, Message: tradeParsed.Message})
	if err != nil {
		t.Fatalf("normalize trade: %v", err)
	}
	tradeEntry, err := venuebinance.BuildSpotTradeRawAppendEntry(supervisor, tradeEvent, tradeMetadata, tradeParsed.Message, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build trade raw entry: %v", err)
	}

	topBookFrame, err := supervisor.AcceptDataFrame(topBookFixture.RawMessages[0], mustRecvTime(t, topBookFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("accept top-of-book frame: %v", err)
	}
	topBookParsed, err := venuebinance.ParseTopOfBookFrame(topBookFrame)
	if err != nil {
		t.Fatalf("parse top-of-book frame: %v", err)
	}
	topBookMetadata := ingestion.BookMetadata{Symbol: topBookFixture.Symbol, SourceSymbol: topBookParsed.SourceSymbol, QuoteCurrency: topBookFixture.QuoteCurrency, Venue: ingestion.VenueBinance, MarketType: "spot"}
	topBookEvent, err := service.NormalizeOrderBook(normalizer.OrderBookInput{Metadata: topBookMetadata, Message: topBookParsed.Message, Sequencer: &ingestion.OrderBookSequencer{}})
	if err != nil {
		t.Fatalf("normalize top-of-book: %v", err)
	}
	if topBookEvent.OrderBookEvent == nil {
		t.Fatal("expected top-of-book event")
	}
	topBookEntry, err := venuebinance.BuildSpotOrderBookRawAppendEntry(supervisor, *topBookEvent.OrderBookEvent, topBookMetadata, topBookParsed.Message, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build top-of-book raw entry: %v", err)
	}

	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion:         "v1",
		LogicalPartition:      tradeEntry.PartitionKey,
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            2,
		FirstCanonicalEventID: tradeEntry.CanonicalEventID,
		LastCanonicalEventID:  topBookEntry.CanonicalEventID,
		ContinuityChecksum:    "sha256:binance-spot-shared",
	}}
	first := executeIntegrationBinanceReplay(t, records, []ingestion.RawAppendEntry{tradeEntry, topBookEntry}, []string{string(ingestion.StreamTrades), string(ingestion.StreamTopOfBook)}, contracts.ReplayRuntimeModeRebuild, nil)
	second := executeIntegrationBinanceReplay(t, records, []ingestion.RawAppendEntry{tradeEntry, topBookEntry}, []string{string(ingestion.StreamTrades), string(ingestion.StreamTopOfBook)}, contracts.ReplayRuntimeModeRebuild, nil)

	if len(first.OrderedEntries) != 2 {
		t.Fatalf("ordered entry count = %d, want 2", len(first.OrderedEntries))
	}
	if first.OrderedEntries[0].StreamFamily != string(ingestion.StreamTrades) || first.OrderedEntries[1].StreamFamily != string(ingestion.StreamTopOfBook) {
		t.Fatalf("ordered stream families = %q, %q", first.OrderedEntries[0].StreamFamily, first.OrderedEntries[1].StreamFamily)
	}
	if first.Result.OutputDigest != second.Result.OutputDigest {
		t.Fatalf("output digest drift: %q vs %q", first.Result.OutputDigest, second.Result.OutputDigest)
	}
}

func TestIngestionBinanceSpotDepthReplayRetainsDegradedFeedEvidence(t *testing.T) {
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
		t.Fatalf("accept depth frame: %v", err)
	}
	if err := owner.MarkSequenceGap(frame); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	if _, err := owner.Recover(context.Background(), mustRecvTime(t, fixture.RawMessages[0])); err == nil {
		t.Fatal("expected cooldown-blocked recovery to stay degraded")
	}
	input, err := owner.FeedHealthInput(venuebinance.SpotDepthFeedHealthOptions{Symbol: fixture.Symbol, QuoteCurrency: fixture.QuoteCurrency, Now: mustRecvTime(t, fixture.RawMessages[0]), ConnectionState: ingestion.ConnectionConnected})
	if err != nil {
		t.Fatalf("feed health input: %v", err)
	}
	event, err := service.NormalizeFeedHealth(normalizer.FeedHealthInput{Metadata: input.Metadata, Message: input.Message})
	if err != nil {
		t.Fatalf("normalize feed health: %v", err)
	}
	entry, err := venuebinance.BuildSpotDepthFeedHealthRawAppendEntry(supervisor, event, input.Metadata.SourceSymbol, ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build depth replay raw entry: %v", err)
	}
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion:         "v1",
		LogicalPartition:      entry.PartitionKey,
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06/feed-health",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            1,
		FirstCanonicalEventID: entry.CanonicalEventID,
		LastCanonicalEventID:  entry.CanonicalEventID,
		ContinuityChecksum:    "sha256:binance-depth-health",
	}}
	output := executeIntegrationBinanceReplay(t, records, []ingestion.RawAppendEntry{entry}, []string{ingestion.RawStreamFamilyFeedHealth}, contracts.ReplayRuntimeModeInspect, nil)

	if len(output.OrderedEntries) != 1 {
		t.Fatalf("ordered entry count = %d, want 1", len(output.OrderedEntries))
	}
	if output.OrderedEntries[0].DegradedFeedRef != entry.DegradedFeedRef {
		t.Fatalf("degraded feed ref = %q, want %q", output.OrderedEntries[0].DegradedFeedRef, entry.DegradedFeedRef)
	}
	if output.OrderedEntries[0].VenueMessageID != entry.VenueMessageID {
		t.Fatalf("venue message id = %q, want %q", output.OrderedEntries[0].VenueMessageID, entry.VenueMessageID)
	}
}

func TestIngestionBinanceUSDMReplayPreservesMixedSurfaceIdentity(t *testing.T) {
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
		t.Fatalf("accept mark price frame: %v", err)
	}

	fundingFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-funding-usdt.fixture.v1.json")
	openInterestFixture := loadFixture(t, "tests/fixtures/events/binance/BTC-USD/happy-open-interest-usdt.fixture.v1.json")
	fundingParsed, err := venuebinance.ParseUSDMMarkPrice(fundingFixture.RawMessages[0], mustRecvTime(t, fundingFixture.RawMessages[0]))
	if err != nil {
		t.Fatalf("parse funding: %v", err)
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
		t.Fatalf("build websocket health raw entry: %v", err)
	}
	restHealthEntry, err := venuebinance.BuildUSDMOpenInterestFeedHealthRawAppendEntry(poller, restEvent, "BTCUSDT", ingestion.RawWriteOptions{BuildVersion: "test-build"})
	if err != nil {
		t.Fatalf("build rest health raw entry: %v", err)
	}
	records := []ingestion.RawPartitionManifestRecord{
		{SchemaVersion: "v1", LogicalPartition: fundingEntry.PartitionKey, StorageState: ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/funding-rate", HotRetentionUntil: "2026-04-05T00:00:00Z", ColdRetentionUntil: "2027-03-06T00:00:00Z", EntryCount: 1, FirstCanonicalEventID: fundingEntry.CanonicalEventID, LastCanonicalEventID: fundingEntry.CanonicalEventID, ContinuityChecksum: "sha256:funding"},
		{SchemaVersion: "v1", LogicalPartition: openInterestEntry.PartitionKey, StorageState: ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/open-interest", HotRetentionUntil: "2026-04-05T00:00:00Z", ColdRetentionUntil: "2027-03-06T00:00:00Z", EntryCount: 1, FirstCanonicalEventID: openInterestEntry.CanonicalEventID, LastCanonicalEventID: openInterestEntry.CanonicalEventID, ContinuityChecksum: "sha256:oi"},
		{SchemaVersion: "v1", LogicalPartition: wsHealthEntry.PartitionKey, StorageState: ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/feed-health", HotRetentionUntil: "2026-04-05T00:00:00Z", ColdRetentionUntil: "2027-03-06T00:00:00Z", EntryCount: 2, FirstCanonicalEventID: wsHealthEntry.CanonicalEventID, LastCanonicalEventID: restHealthEntry.CanonicalEventID, ContinuityChecksum: "sha256:feed-health"},
	}
	entries := []ingestion.RawAppendEntry{fundingEntry, openInterestEntry, wsHealthEntry, restHealthEntry}
	first := executeIntegrationBinanceReplay(t, records, entries, []string{string(ingestion.StreamFundingRate), string(ingestion.StreamOpenInterest), ingestion.RawStreamFamilyFeedHealth}, contracts.ReplayRuntimeModeRebuild, nil)
	second := executeIntegrationBinanceReplay(t, records, entries, []string{string(ingestion.StreamFundingRate), string(ingestion.StreamOpenInterest), ingestion.RawStreamFamilyFeedHealth}, contracts.ReplayRuntimeModeRebuild, nil)

	if fundingEntry.ConnectionRef == openInterestEntry.ConnectionRef {
		t.Fatal("expected websocket and rest connection refs to remain distinct")
	}
	if wsHealthEntry.VenueMessageID == restHealthEntry.VenueMessageID {
		t.Fatal("expected websocket and rest feed health ids to remain distinct")
	}
	if first.Result.OutputDigest != second.Result.OutputDigest {
		t.Fatalf("output digest drift: %q vs %q", first.Result.OutputDigest, second.Result.OutputDigest)
	}
}

func executeIntegrationBinanceReplay(t *testing.T, records []ingestion.RawPartitionManifestRecord, entries []ingestion.RawAppendEntry, streamFamilies []string, mode contracts.ReplayRuntimeMode, compareTarget *replayengine.CompareTarget) replayengine.ExecutionOutput {
	t.Helper()
	reader := &filteredIntegrationManifestReader{records: records}
	builder, err := replayengine.NewManifestBuilder(reader, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	manifest, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "binance-replay-run", Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"binance"}, StreamFamilies: streamFamilies, WindowStart: "2026-03-06T00:00:00Z", WindowEnd: "2026-03-06T23:59:59Z"}, RuntimeMode: mode, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	engine, err := replayengine.NewEngine(reader, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: entries}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new replay engine: %v", err)
	}
	output, err := engine.Execute(manifest, compareTarget)
	if err != nil {
		t.Fatalf("execute replay: %v", err)
	}
	return output
}

type filteredIntegrationManifestReader struct {
	records []ingestion.RawPartitionManifestRecord
}

func (r *filteredIntegrationManifestReader) ResolveRawPartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error) {
	resolved := make([]ingestion.RawPartitionManifestRecord, 0, len(r.records))
	for _, record := range r.records {
		if record.LogicalPartition.Symbol != scope.Symbol || record.LogicalPartition.Venue != scope.Venue {
			continue
		}
		if scope.StreamFamily != "" && record.LogicalPartition.StreamFamily != scope.StreamFamily {
			continue
		}
		resolved = append(resolved, record)
	}
	return resolved, nil
}
