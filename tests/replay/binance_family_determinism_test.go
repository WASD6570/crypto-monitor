package replay

import (
	"reflect"
	"testing"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestReplayBinanceSharedAndDedicatedPartitionsStayDeterministic(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{
		binanceManifestRecord("", "hot://raw/BTC-USD/2026-03-06", 3, "ce:trade", "ce:topbook", "sha256:shared"),
		binanceManifestRecord(ingestion.RawStreamFamilyFeedHealth, "hot://raw/BTC-USD/2026-03-06/feed-health", 1, "ce:depth-health", "ce:depth-health", "sha256:feed-health"),
		binanceManifestRecord(string(ingestion.StreamFundingRate), "hot://raw/BTC-USD/2026-03-06/funding-rate", 1, "ce:funding", "ce:funding", "sha256:funding"),
		binanceManifestRecord(string(ingestion.StreamOpenInterest), "hot://raw/BTC-USD/2026-03-06/open-interest", 1, "ce:oi", "ce:oi", "sha256:oi"),
	}
	entries := []ingestion.RawAppendEntry{
		binanceReplayEntry("market-trade", "ce:trade", "1001", 0, string(ingestion.StreamTrades), string(ingestion.StreamTrades), "2026-03-06T12:00:00.100Z", ingestion.RawBucketTimestampSourceExchange),
		binanceReplayEntry("market-trade", "ce:trade", "1001", 0, string(ingestion.StreamTrades), string(ingestion.StreamTrades), "2026-03-06T12:00:00.100Z", ingestion.RawBucketTimestampSourceExchange),
		binanceReplayEntry("order-book-top", "ce:topbook", "", 9001, string(ingestion.StreamTopOfBook), string(ingestion.StreamTopOfBook), "2026-03-06T12:00:00.180Z", ingestion.RawBucketTimestampSourceRecv),
		binanceReplayEntry("feed-health", "ce:depth-health", "runtime:binance-spot-depth:BTCUSDT", 0, string(ingestion.StreamOrderBook), ingestion.RawStreamFamilyFeedHealth, "2026-03-06T12:00:00.600Z", ingestion.RawBucketTimestampSourceExchange),
		binanceReplayEntry("funding-rate", "ce:funding", "2026-03-06T12:00:05Z", 0, string(ingestion.StreamFundingRate), string(ingestion.StreamFundingRate), "2026-03-06T12:00:05Z", ingestion.RawBucketTimestampSourceExchange),
		binanceReplayEntry("open-interest-snapshot", "ce:oi", "2026-03-06T12:00:04Z", 0, string(ingestion.StreamOpenInterest), string(ingestion.StreamOpenInterest), "2026-03-06T12:00:04Z", ingestion.RawBucketTimestampSourceExchange),
	}
	entries[1].DuplicateAudit = ingestion.RawDuplicateAudit{IdentityKey: "message:trades:1001", Occurrence: 2, Duplicate: true}
	entries[2].TimestampDegradationReason = ingestion.TimestampReasonExchangeMissingOrInvalid
	entries[3].DegradedFeedRef = "runtime:binance-spot-depth:BTCUSDT"

	first := executeBinanceReplay(t, records, entries, contracts.ReplayRuntimeModeRebuild, nil)
	second := executeBinanceReplay(t, records, entries, contracts.ReplayRuntimeModeRebuild, nil)

	if first.Result.OutputDigest != second.Result.OutputDigest {
		t.Fatalf("output digest drift: %q vs %q", first.Result.OutputDigest, second.Result.OutputDigest)
	}
	if !reflect.DeepEqual(replayOrderedIDs(first.OrderedEntries), replayOrderedIDs(second.OrderedEntries)) {
		t.Fatalf("ordered canonical ids drift: %v vs %v", replayOrderedIDs(first.OrderedEntries), replayOrderedIDs(second.OrderedEntries))
	}
	if first.Result.InputCounters.Duplicates != 1 {
		t.Fatalf("duplicate counter = %d, want 1", first.Result.InputCounters.Duplicates)
	}
	if first.Result.InputCounters.DegradedTimestampEvents != 1 {
		t.Fatalf("degraded timestamp counter = %d, want 1", first.Result.InputCounters.DegradedTimestampEvents)
	}
	depthHealth := replayEntryByCanonicalID(first.OrderedEntries, "ce:depth-health")
	if depthHealth.DegradedFeedRef != "runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("degraded feed ref = %q, want %q", depthHealth.DegradedFeedRef, "runtime:binance-spot-depth:BTCUSDT")
	}
}

func TestReplayBinanceCompareDetectsDegradedEvidenceDrift(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{
		binanceManifestRecord("", "hot://raw/BTC-USD/2026-03-06", 1, "ce:topbook", "ce:topbook", "sha256:shared"),
		binanceManifestRecord(ingestion.RawStreamFamilyFeedHealth, "hot://raw/BTC-USD/2026-03-06/feed-health", 1, "ce:depth-health", "ce:depth-health", "sha256:feed-health"),
	}
	baseline := []ingestion.RawAppendEntry{
		binanceReplayEntry("order-book-top", "ce:topbook", "", 9001, string(ingestion.StreamTopOfBook), string(ingestion.StreamTopOfBook), "2026-03-06T12:00:00.180Z", ingestion.RawBucketTimestampSourceRecv),
		binanceReplayEntry("feed-health", "ce:depth-health", "runtime:binance-spot-depth:BTCUSDT", 0, string(ingestion.StreamOrderBook), ingestion.RawStreamFamilyFeedHealth, "2026-03-06T12:00:00.600Z", ingestion.RawBucketTimestampSourceExchange),
	}
	baseline[0].TimestampDegradationReason = ingestion.TimestampReasonExchangeMissingOrInvalid
	baseline[1].DegradedFeedRef = "runtime:binance-spot-depth:BTCUSDT"
	baselineOutput := executeBinanceReplay(t, records, baseline, contracts.ReplayRuntimeModeRebuild, nil)

	changed := append([]ingestion.RawAppendEntry(nil), baseline...)
	changed[1] = baseline[1]
	changed[1].DegradedFeedRef = "runtime:binance-spot-depth:BTCUSDT:drift"
	compare := executeBinanceReplay(t, records, changed, contracts.ReplayRuntimeModeCompare, &replayengine.CompareTarget{ID: "baseline-binance", Digest: baselineOutput.Result.OutputDigest})

	if compare.CompareSummary == nil {
		t.Fatal("expected compare summary")
	}
	if compare.CompareSummary.DriftClassification != "drift" {
		t.Fatalf("drift classification = %q, want drift", compare.CompareSummary.DriftClassification)
	}
	if compare.CompareSummary.FirstMismatch == "" {
		t.Fatalf("expected compare mismatch details, got %+v", compare.CompareSummary)
	}
	if compare.Result.InputCounters.DegradedTimestampEvents != 1 {
		t.Fatalf("degraded timestamp counter = %d, want 1", compare.Result.InputCounters.DegradedTimestampEvents)
	}
}

func executeBinanceReplay(t *testing.T, records []ingestion.RawPartitionManifestRecord, entries []ingestion.RawAppendEntry, mode contracts.ReplayRuntimeMode, compareTarget *replayengine.CompareTarget) replayengine.ExecutionOutput {
	t.Helper()
	builder, err := replayengine.NewManifestBuilder(replayRetentionManifestReader{records: records}, replayRetentionSnapshotLoader{}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	manifest, err := builder.Build(binanceReplayRequest(mode))
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	engine, err := replayengine.NewEngine(replayRetentionManifestReader{records: records}, replayRetentionSnapshotLoader{}, replayRetentionEntryLoader{entries: entries}, replayRetentionArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, compareTarget)
	if err != nil {
		t.Fatalf("execute replay: %v", err)
	}
	return output
}

func binanceReplayRequest(mode contracts.ReplayRuntimeMode) contracts.ReplayRunRequest {
	return contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "binance-replay-run",
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{"binance"},
			StreamFamilies: []string{string(ingestion.StreamTrades), string(ingestion.StreamTopOfBook), ingestion.RawStreamFamilyFeedHealth, string(ingestion.StreamFundingRate), string(ingestion.StreamOpenInterest)},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      mode,
		ConfigSnapshot:   contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}
}

func binanceManifestRecord(streamFamily, location string, entryCount int, firstID, lastID, checksum string) ingestion.RawPartitionManifestRecord {
	return ingestion.RawPartitionManifestRecord{
		SchemaVersion:         "v1",
		LogicalPartition:      ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueBinance, StreamFamily: streamFamily},
		StorageState:          ingestion.RawStorageStateHot,
		Location:              location,
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            entryCount,
		FirstCanonicalEventID: firstID,
		LastCanonicalEventID:  lastID,
		ContinuityChecksum:    checksum,
	}
}

func binanceReplayEntry(eventType, canonicalID, messageID string, sequence int64, streamKey, streamFamily, bucketTimestamp string, source ingestion.RawBucketTimestampSource) ingestion.RawAppendEntry {
	marketType := "spot"
	if streamFamily == string(ingestion.StreamFundingRate) || streamFamily == string(ingestion.StreamOpenInterest) {
		marketType = "perpetual"
	}
	return ingestion.RawAppendEntry{
		SchemaVersion:          "v1",
		CanonicalSchemaVersion: "v1",
		CanonicalEventType:     eventType,
		CanonicalEventID:       canonicalID,
		VenueMessageID:         messageID,
		VenueSequence:          sequence,
		StreamKey:              streamKey,
		Symbol:                 "BTC-USD",
		Venue:                  ingestion.VenueBinance,
		MarketType:             marketType,
		StreamFamily:           streamFamily,
		ExchangeTs:             bucketTimestamp,
		RecvTs:                 bucketTimestamp,
		BucketTimestamp:        bucketTimestamp,
		BucketTimestampSource:  source,
		NormalizerService:      "services/normalizer",
		ConnectionRef:          "binance-test",
		SessionRef:             "binance-session",
		BuildVersion:           "test-build",
		DuplicateAudit:         ingestion.RawDuplicateAudit{IdentityKey: messageID, Occurrence: 1},
		PartitionKey:           ingestion.RouteRawPartition(ingestion.RawAppendEntry{Symbol: "BTC-USD", Venue: ingestion.VenueBinance, StreamFamily: streamFamily, BucketTimestamp: bucketTimestamp}),
	}
}

func replayEntryByCanonicalID(entries []ingestion.RawAppendEntry, canonicalID string) ingestion.RawAppendEntry {
	for _, entry := range entries {
		if entry.CanonicalEventID == canonicalID {
			return entry
		}
	}
	return ingestion.RawAppendEntry{}
}
