package integration

import (
	"reflect"
	"testing"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestReplayRetentionContinuityAcrossHotAndCold(t *testing.T) {
	entries := integrationRetentionEntries()
	hot := runIntegrationRetentionReplay(t, integrationRetentionRecord(ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild, entries, nil)
	cold := runIntegrationRetentionReplay(t, integrationRetentionRecord(ingestion.RawStorageStateCold, "cold://archive/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild, entries, nil)

	if hot.Result.OutputDigest != cold.Result.OutputDigest {
		t.Fatalf("output digest drift across hot/cold retention: hot=%q cold=%q", hot.Result.OutputDigest, cold.Result.OutputDigest)
	}
	if !reflect.DeepEqual(integrationRetentionIDs(hot.OrderedEntries), integrationRetentionIDs(cold.OrderedEntries)) {
		t.Fatalf("ordered event ids drift across hot/cold retention")
	}
	if hot.Result.InputCounters != cold.Result.InputCounters {
		t.Fatalf("input counters drift across hot/cold retention: hot=%+v cold=%+v", hot.Result.InputCounters, cold.Result.InputCounters)
	}
}

func TestReplayColdRestoreIsExplicitAndDeterministic(t *testing.T) {
	record := integrationRetentionRecord(ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades")
	first := runIntegrationRetentionReplay(t, record, contracts.ReplayRuntimeModeRebuild, integrationRetentionEntries(), nil)
	second := runIntegrationRetentionReplay(t, record, contracts.ReplayRuntimeModeRebuild, integrationRetentionEntries(), nil)

	if first.Result.OutputDigest != second.Result.OutputDigest {
		t.Fatalf("restore replay digest drift: %q vs %q", first.Result.OutputDigest, second.Result.OutputDigest)
	}
	if first.AuditTrail.Execution.ResolvedPartitions[0] != second.AuditTrail.Execution.ResolvedPartitions[0] {
		t.Fatalf("restore partition drift: %v vs %v", first.AuditTrail.Execution.ResolvedPartitions, second.AuditTrail.Execution.ResolvedPartitions)
	}
}

func TestReplaySafetyMatrixAcrossModes(t *testing.T) {
	entries := integrationRetentionEntries()
	inspect := runIntegrationRetentionReplay(t, integrationRetentionRecord(ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeInspect, entries, nil)
	rebuild := runIntegrationRetentionReplay(t, integrationRetentionRecord(ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild, entries, nil)
	targetDigest, err := integrationRetentionDigest(entries)
	if err != nil {
		t.Fatalf("target digest: %v", err)
	}
	compare := runIntegrationRetentionReplay(t, integrationRetentionRecord(ingestion.RawStorageStateCold, "cold://archive/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeCompare, entries, &replayengine.CompareTarget{ID: "baseline", Digest: targetDigest})

	if inspect.AuditTrail.Outcome.CrossedApplyGate || rebuild.AuditTrail.Outcome.CrossedApplyGate || compare.AuditTrail.Outcome.CrossedApplyGate {
		t.Fatalf("unexpected apply gate crossing in safety matrix")
	}
	if inspect.AuditTrail.Outcome.TerminalStatus != "no-op" || rebuild.AuditTrail.Outcome.TerminalStatus != "isolated-rebuild" || compare.AuditTrail.Outcome.TerminalStatus != "compare-only" {
		t.Fatalf("unexpected safety matrix outcomes: inspect=%q rebuild=%q compare=%q", inspect.AuditTrail.Outcome.TerminalStatus, rebuild.AuditTrail.Outcome.TerminalStatus, compare.AuditTrail.Outcome.TerminalStatus)
	}
}

func TestBackfillResumeAfterRetentionRestore(t *testing.T) {
	record := integrationRetentionRecord(ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades")
	output := runIntegrationRetentionReplay(t, record, contracts.ReplayRuntimeModeRebuild, integrationRetentionEntries(), nil)
	decision, err := replayengine.ReplayResumeFromCheckpoint(integrationRetentionManifest(t, record, contracts.ReplayRuntimeModeRebuild), *output.Checkpoint)
	if err != nil {
		t.Fatalf("resume from restored checkpoint: %v", err)
	}
	if decision.LogicalPartition == "" || decision.ResumeFromEventID == "" || decision.NextCheckpointSequence != 2 {
		t.Fatalf("unexpected resume decision: %+v", decision)
	}
}

func TestReplayApplyGateRejectsWithoutApproval(t *testing.T) {
	manifest := integrationRetentionManifest(t, integrationRetentionRecord(ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild)
	manifest.ApplyIntent = true
	engine, err := replayengine.NewEngine(&integrationManifestReader{records: []ingestion.RawPartitionManifestRecord{integrationRetentionRecord(ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades")}}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: integrationRetentionEntries()}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected apply rejection without approval")
	}
	if output.Result.FailureCategory != "approval-required" || output.AuditTrail.Outcome.TerminalStatus != "rejected-apply" {
		t.Fatalf("unexpected apply rejection output: %+v", output)
	}
}

func TestReplayApplyIsIdempotentAcrossRetries(t *testing.T) {
	manifest := integrationRetentionManifest(t, integrationRetentionRecord(ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild)
	manifest.ApplyIntent = true
	manifest.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}
	artifacts := []contracts.ReplayArtifactRef{{Kind: "rebuild-output", Namespace: "runs/" + manifest.RunID + "/rebuild", Digest: "sha256:artifact"}}
	outcomes := make(map[string]contracts.ReplayOutcomeAuditRecord)
	first, err := replayengine.ReplayApplyGate(manifest, artifacts, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("first apply gate: %v", err)
	}
	second, err := replayengine.ReplayApplyGate(manifest, artifacts, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("second apply gate: %v", err)
	}
	if !reflect.DeepEqual(first, second) || !second.CrossedApplyGate {
		t.Fatalf("unexpected idempotent apply outcome: first=%+v second=%+v", first, second)
	}
}

func TestReplayOverlapHandlingRemainsDeterministic(t *testing.T) {
	active := integrationRetentionManifest(t, integrationRetentionRecord(ingestion.RawStorageStateHot, "hot://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild)
	active.RequestID = "sha256:active"
	active.ConflictKey = "sha256:active-conflict"
	active.ApplyIntent = true
	active.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}
	candidate := integrationRetentionManifest(t, integrationRetentionRecord(ingestion.RawStorageStateTransition, "restore://raw/BTC-USD/2026-03-06/trades"), contracts.ReplayRuntimeModeRebuild)
	candidate.RequestID = "sha256:candidate"
	candidate.ApplyIntent = true
	candidate.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-2", PromotionToken: "promo-2"}

	key, conflicted, err := replayengine.ReplayRequestsConflict(candidate, []contracts.ReplayRunManifest{active})
	if err != nil {
		t.Fatalf("detect conflict: %v", err)
	}
	if !conflicted || key != "sha256:active-conflict" {
		t.Fatalf("unexpected overlap handling result: conflicted=%v key=%q", conflicted, key)
	}
}

func runIntegrationRetentionReplay(t *testing.T, record ingestion.RawPartitionManifestRecord, mode contracts.ReplayRuntimeMode, entries []ingestion.RawAppendEntry, compareTarget *replayengine.CompareTarget) replayengine.ExecutionOutput {
	t.Helper()
	manifest := integrationRetentionManifest(t, record, mode)
	engine, err := replayengine.NewEngine(&integrationManifestReader{records: []ingestion.RawPartitionManifestRecord{record}}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: entries}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, compareTarget)
	if err != nil {
		t.Fatalf("execute replay: %v", err)
	}
	return output
}

func integrationRetentionManifest(t *testing.T, record ingestion.RawPartitionManifestRecord, mode contracts.ReplayRuntimeMode) contracts.ReplayRunManifest {
	t.Helper()
	builder, err := replayengine.NewManifestBuilder(&integrationManifestReader{records: []ingestion.RawPartitionManifestRecord{record}}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	manifest, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "run-" + string(mode), Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"coinbase"}, StreamFamilies: []string{"trades"}, WindowStart: "2026-03-06T00:00:00Z", WindowEnd: "2026-03-06T23:59:59Z"}, RuntimeMode: mode, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	return manifest
}

func integrationRetentionRecord(state ingestion.RawStorageState, location string) ingestion.RawPartitionManifestRecord {
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

func integrationRetentionEntries() []ingestion.RawAppendEntry {
	degraded := integrationRetentionEntry("2026-03-06T11:59:59.900Z", 2, "trade:degraded", "ce:degraded", ingestion.RawBucketTimestampSourceRecv)
	degraded.TimestampDegradationReason = ingestion.TimestampReasonExchangeSkewExceeded
	late := integrationRetentionEntry("2026-03-06T12:00:00.500Z", 3, "trade:late", "ce:late", ingestion.RawBucketTimestampSourceExchange)
	late.Late = true
	return []ingestion.RawAppendEntry{
		integrationRetentionEntry("2026-03-06T12:00:00Z", 1, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange),
		degraded,
		late,
	}
}

func integrationRetentionEntry(bucketTimestamp string, sequence int64, sourceID string, canonicalID string, source ingestion.RawBucketTimestampSource) ingestion.RawAppendEntry {
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

func integrationRetentionIDs(entries []ingestion.RawAppendEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.CanonicalEventID)
	}
	return ids
}

func integrationRetentionDigest(entries []ingestion.RawAppendEntry) (string, error) {
	type replayDigestInput struct {
		IDs []string `json:"ids"`
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.CanonicalEventID+"|"+entry.VenueMessageID)
	}
	return contracts.ReplayValueDigest(replayDigestInput{IDs: ids})
}
