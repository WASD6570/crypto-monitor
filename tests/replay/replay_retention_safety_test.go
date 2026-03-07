package replay

import (
	"fmt"
	"reflect"
	"testing"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestReplayDeterminismAcrossRetentionTiers(t *testing.T) {
	entries := replayRetentionEntries()
	hot := executeRetentionTier(t, ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeRebuild, entries, nil)
	cold := executeRetentionTier(t, ingestion.RawStorageStateCold, "cold://archive/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeRebuild, entries, nil)
	restored := executeRetentionTier(t, ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeRebuild, entries, nil)

	if hot.Result.OutputDigest != cold.Result.OutputDigest || hot.Result.OutputDigest != restored.Result.OutputDigest {
		t.Fatalf("output digest drift across tiers: hot=%q cold=%q restored=%q", hot.Result.OutputDigest, cold.Result.OutputDigest, restored.Result.OutputDigest)
	}
	if !reflect.DeepEqual(replayOrderedIDs(hot.OrderedEntries), replayOrderedIDs(cold.OrderedEntries)) || !reflect.DeepEqual(replayOrderedIDs(hot.OrderedEntries), replayOrderedIDs(restored.OrderedEntries)) {
		t.Fatalf("ordered event ids drift across tiers")
	}
	if hot.Result.InputCounters != cold.Result.InputCounters || hot.Result.InputCounters != restored.Result.InputCounters {
		t.Fatalf("input counter drift across tiers: hot=%+v cold=%+v restored=%+v", hot.Result.InputCounters, cold.Result.InputCounters, restored.Result.InputCounters)
	}
}

func TestReplayRetentionPreservesLateAndDegradedEvidence(t *testing.T) {
	entries := replayRetentionEntries()
	hot := executeRetentionTier(t, ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeInspect, entries, nil)
	cold := executeRetentionTier(t, ingestion.RawStorageStateCold, "cold://archive/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeInspect, entries, nil)

	if hot.Result.InputCounters.LateEvents != 1 || hot.Result.InputCounters.DegradedTimestampEvents != 1 {
		t.Fatalf("unexpected hot counters: %+v", hot.Result.InputCounters)
	}
	if hot.Result.InputCounters != cold.Result.InputCounters {
		t.Fatalf("late/degraded evidence drift across tiers: hot=%+v cold=%+v", hot.Result.InputCounters, cold.Result.InputCounters)
	}
}

func TestReplayCompareCapturesLateEventRepairCandidates(t *testing.T) {
	entries := replayRetentionEntries()
	baselineEntries := append([]ingestion.RawAppendEntry(nil), entries[:2]...)
	targetDigest, err := replayRetentionDigest(baselineEntries)
	if err != nil {
		t.Fatalf("baseline digest: %v", err)
	}
	output := executeRetentionTier(t, ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeCompare, entries, &replayengine.CompareTarget{ID: "baseline-without-late", Digest: targetDigest})
	if output.CompareSummary == nil {
		t.Fatal("expected compare summary")
	}
	if output.CompareSummary.DriftClassification != "drift" || output.CompareSummary.FirstMismatch == "" {
		t.Fatalf("expected deterministic repair candidate drift, got %+v", output.CompareSummary)
	}
}

func TestReplayModesDoNotEmitSideEffectsByDefault(t *testing.T) {
	entries := replayRetentionEntries()
	inspect := executeRetentionTier(t, ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeInspect, entries, nil)
	rebuild := executeRetentionTier(t, ingestion.RawStorageStateCold, "cold://archive/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeRebuild, entries, nil)
	targetDigest, err := replayRetentionDigest(entries)
	if err != nil {
		t.Fatalf("target digest: %v", err)
	}
	compare := executeRetentionTier(t, ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades", contracts.ReplayRuntimeModeCompare, entries, &replayengine.CompareTarget{ID: "baseline", Digest: targetDigest})

	if inspect.AuditTrail.Outcome.CrossedApplyGate || rebuild.AuditTrail.Outcome.CrossedApplyGate || compare.AuditTrail.Outcome.CrossedApplyGate {
		t.Fatalf("non-apply mode crossed apply gate: inspect=%v rebuild=%v compare=%v", inspect.AuditTrail.Outcome.CrossedApplyGate, rebuild.AuditTrail.Outcome.CrossedApplyGate, compare.AuditTrail.Outcome.CrossedApplyGate)
	}
	if inspect.AuditTrail.Outcome.TerminalStatus != "no-op" || rebuild.AuditTrail.Outcome.TerminalStatus != "isolated-rebuild" || compare.AuditTrail.Outcome.TerminalStatus != "compare-only" {
		t.Fatalf("unexpected mode outcomes: inspect=%q rebuild=%q compare=%q", inspect.AuditTrail.Outcome.TerminalStatus, rebuild.AuditTrail.Outcome.TerminalStatus, compare.AuditTrail.Outcome.TerminalStatus)
	}
}

func executeRetentionTier(t *testing.T, state ingestion.RawStorageState, location string, mode contracts.ReplayRuntimeMode, entries []ingestion.RawAppendEntry, compareTarget *replayengine.CompareTarget) replayengine.ExecutionOutput {
	t.Helper()
	record := replayRetentionManifestRecord(state, location)
	builder, err := replayengine.NewManifestBuilder(replayRetentionManifestReader{records: []ingestion.RawPartitionManifestRecord{record}}, replayRetentionSnapshotLoader{}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	manifest, err := builder.Build(replayRetentionRequest(mode))
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.RawPartitions[0].StorageState != string(state) {
		t.Fatalf("storage state = %q, want %q", manifest.RawPartitions[0].StorageState, state)
	}
	engine, err := replayengine.NewEngine(replayRetentionManifestReader{records: []ingestion.RawPartitionManifestRecord{record}}, replayRetentionSnapshotLoader{}, replayRetentionEntryLoader{entries: entries}, replayRetentionArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, compareTarget)
	if err != nil {
		t.Fatalf("execute replay: %v", err)
	}
	return output
}

func replayRetentionRequest(mode contracts.ReplayRuntimeMode) contracts.ReplayRunRequest {
	return contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         fmt.Sprintf("run-%s", mode),
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{"coinbase"},
			StreamFamilies: []string{"trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      mode,
		ConfigSnapshot:   contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}
}

func replayRetentionManifestRecord(state ingestion.RawStorageState, location string) ingestion.RawPartitionManifestRecord {
	return ingestion.RawPartitionManifestRecord{
		SchemaVersion:         "v1",
		LogicalPartition:      ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
		StorageState:          state,
		Location:              location,
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            3,
		FirstCanonicalEventID: "ce:a",
		LastCanonicalEventID:  "ce:late",
		ContinuityChecksum:    "sha256:continuity",
	}
}

func replayRetentionEntries() []ingestion.RawAppendEntry {
	degraded := replayRetentionEntry("2026-03-06T11:59:59.900Z", 2, "trade:degraded", "ce:degraded", ingestion.RawBucketTimestampSourceRecv)
	degraded.TimestampDegradationReason = ingestion.TimestampReasonExchangeSkewExceeded
	late := replayRetentionEntry("2026-03-06T12:00:00.500Z", 3, "trade:late", "ce:late", ingestion.RawBucketTimestampSourceExchange)
	late.Late = true
	return []ingestion.RawAppendEntry{
		replayRetentionEntry("2026-03-06T12:00:00Z", 1, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange),
		degraded,
		late,
	}
}

func replayRetentionEntry(bucketTimestamp string, sequence int64, sourceID string, canonicalID string, source ingestion.RawBucketTimestampSource) ingestion.RawAppendEntry {
	return ingestion.RawAppendEntry{
		SchemaVersion:          "v1",
		CanonicalSchemaVersion: "v1",
		CanonicalEventType:     "market-trade",
		CanonicalEventID:       canonicalID,
		VenueMessageID:         sourceID,
		VenueSequence:          sequence,
		StreamKey:              "trades",
		Symbol:                 "BTC-USD",
		Venue:                  ingestion.VenueCoinbase,
		MarketType:             "spot",
		StreamFamily:           "trades",
		ExchangeTs:             bucketTimestamp,
		RecvTs:                 bucketTimestamp,
		BucketTimestamp:        bucketTimestamp,
		BucketTimestampSource:  source,
		NormalizerService:      "normalizer",
		ConnectionRef:          "conn-1",
		SessionRef:             "session-1",
		BuildVersion:           "test",
		DuplicateAudit:         ingestion.RawDuplicateAudit{IdentityKey: sourceID, Occurrence: 1},
		PartitionKey:           ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
	}
}

func replayOrderedIDs(entries []ingestion.RawAppendEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.CanonicalEventID)
	}
	return ids
}

func replayRetentionDigest(entries []ingestion.RawAppendEntry) (string, error) {
	type replayDigestInput struct {
		IDs []string `json:"ids"`
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.CanonicalEventID+"|"+entry.VenueMessageID)
	}
	return contracts.ReplayValueDigest(replayDigestInput{IDs: ids})
}

type replayRetentionManifestReader struct {
	records []ingestion.RawPartitionManifestRecord
}

func (r replayRetentionManifestReader) ResolveRawPartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error) {
	return append([]ingestion.RawPartitionManifestRecord(nil), r.records...), nil
}

type replayRetentionSnapshotLoader struct{}

func (replayRetentionSnapshotLoader) LoadConfigSnapshot(ref contracts.ReplaySnapshotRef) (replayengine.ConfigSnapshot, error) {
	return replayengine.ConfigSnapshot{Ref: ref, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}, nil
}

type replayRetentionEntryLoader struct{ entries []ingestion.RawAppendEntry }

func (l replayRetentionEntryLoader) LoadRawEntries(partitions []contracts.ReplayPartitionRef) ([]ingestion.RawAppendEntry, error) {
	return append([]ingestion.RawAppendEntry(nil), l.entries...), nil
}

type replayRetentionArtifactWriter struct{}

func (replayRetentionArtifactWriter) WriteArtifact(runID string, mode contracts.ReplayRuntimeMode, kind string, payload any) (contracts.ReplayArtifactRef, error) {
	digest, err := contracts.ReplayValueDigest(payload)
	if err != nil {
		return contracts.ReplayArtifactRef{}, err
	}
	return contracts.ReplayArtifactRef{Kind: kind, Namespace: "runs/" + runID + "/" + string(mode), Digest: digest}, nil
}
