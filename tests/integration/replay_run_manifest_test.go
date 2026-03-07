package integration

import (
	"reflect"
	"testing"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestReplayManifestUsesResolvedRawPartitionRefs(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion: "v1",
		LogicalPartition: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.VenueCoinbase,
			StreamFamily: "trades",
		},
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06/trades",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            2,
		FirstCanonicalEventID: "ce:first",
		LastCanonicalEventID:  "ce:last",
		ContinuityChecksum:    "sha256:abc",
	}}
	builder, err := replayengine.NewManifestBuilder(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}

	manifest, err := builder.Build(contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "run-123",
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{string(ingestion.VenueCoinbase)},
			StreamFamilies: []string{"trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      contracts.ReplayRuntimeModeInspect,
		ConfigSnapshot:   contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if len(manifest.RawPartitions) != 1 {
		t.Fatalf("raw partitions = %d, want 1", len(manifest.RawPartitions))
	}
	if manifest.RawPartitions[0].Location != "hot://raw/BTC-USD/2026-03-06/trades" {
		t.Fatalf("location = %q, want resolved manifest location", manifest.RawPartitions[0].Location)
	}
}

func TestReplayOneSymbolOneDayDeterministicManifestExecution(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion: "v1",
		LogicalPartition: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.VenueCoinbase,
			StreamFamily: "trades",
		},
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06/trades",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            3,
		FirstCanonicalEventID: "ce:a",
		LastCanonicalEventID:  "ce:c",
		ContinuityChecksum:    "sha256:abc",
	}}
	builder, err := replayengine.NewManifestBuilder(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	manifest, err := builder.Build(contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "run-123",
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{string(ingestion.VenueCoinbase)},
			StreamFamilies: []string{"trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      contracts.ReplayRuntimeModeRebuild,
		ConfigSnapshot:   contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}

	entries := []ingestion.RawAppendEntry{
		integrationReplayEntry("2026-03-06T12:00:00Z", 2, "trade:b", "ce:b"),
		integrationReplayEntry("2026-03-06T12:00:00Z", 1, "trade:a", "ce:a"),
		integrationReplayEntry("2026-03-06T12:00:01Z", 0, "trade:c", "ce:c"),
	}
	engine, err := replayengine.NewEngine(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: entries}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	first, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("first execute: %v", err)
	}
	second, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("second execute: %v", err)
	}
	if first.Result.OutputDigest != second.Result.OutputDigest {
		t.Fatalf("output digests differ: %q vs %q", first.Result.OutputDigest, second.Result.OutputDigest)
	}
	if !reflect.DeepEqual(integrationSourceIDs(first.OrderedEntries), []string{"trade:a", "trade:b", "trade:c"}) {
		t.Fatalf("ordered ids = %v", integrationSourceIDs(first.OrderedEntries))
	}
}

func TestBackfillResumeAfterFailure(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion: "v1",
		LogicalPartition: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.VenueCoinbase,
			StreamFamily: "trades",
		},
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06/trades",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            1,
		FirstCanonicalEventID: "ce:a",
		LastCanonicalEventID:  "ce:a",
		ContinuityChecksum:    "sha256:abc",
	}}
	builder, err := replayengine.NewManifestBuilder(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	request := contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "run-123",
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{string(ingestion.VenueCoinbase)},
			StreamFamilies: []string{"trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      contracts.ReplayRuntimeModeRebuild,
		ConfigSnapshot:   contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}
	manifest, err := builder.Build(request)
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	engine, err := replayengine.NewEngine(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: []ingestion.RawAppendEntry{integrationReplayEntry("2026-03-06T12:00:00Z", 1, "trade:a", "ce:a")}}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute rebuild: %v", err)
	}
	decision, err := replayengine.ReplayResumeFromCheckpoint(manifest, *output.Checkpoint)
	if err != nil {
		t.Fatalf("resume from checkpoint: %v", err)
	}
	if decision.ResumeFromEventID != "ce:a" || decision.LogicalPartition == "" {
		t.Fatalf("unexpected resume decision: %+v", decision)
	}
}

func TestBackfillOverlapConflictHandling(t *testing.T) {
	reader := &integrationManifestReader{records: []ingestion.RawPartitionManifestRecord{{
		SchemaVersion:    "v1",
		LogicalPartition: ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
		StorageState:     ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/trades", EntryCount: 1, ContinuityChecksum: "sha256:abc",
	}}}
	builder, err := replayengine.NewManifestBuilder(reader, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	active, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "run-1", Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"coinbase"}, StreamFamilies: []string{"trades"}, WindowStart: "2026-03-06T00:00:00Z", WindowEnd: "2026-03-06T23:59:59Z"}, RuntimeMode: contracts.ReplayRuntimeModeRebuild, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build active manifest: %v", err)
	}
	candidate, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "run-2", Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"coinbase"}, StreamFamilies: []string{"trades"}, WindowStart: "2026-03-06T12:00:00Z", WindowEnd: "2026-03-06T18:00:00Z"}, RuntimeMode: contracts.ReplayRuntimeModeRebuild, ApplyIntent: true, ApprovalContext: &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build candidate manifest: %v", err)
	}
	key, conflicted, err := replayengine.ReplayRequestsConflict(candidate, []contracts.ReplayRunManifest{active})
	if err != nil {
		t.Fatalf("detect conflict: %v", err)
	}
	if !conflicted || key == "" {
		t.Fatalf("expected overlap conflict, got conflicted=%v key=%q", conflicted, key)
	}
}

func TestBackfillAuditTrailCompleteness(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion:    "v1",
		LogicalPartition: ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
		StorageState:     ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/trades", EntryCount: 1, ContinuityChecksum: "sha256:abc",
	}}
	builder, _ := replayengine.NewManifestBuilder(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	manifest, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "run-123", Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"coinbase"}, StreamFamilies: []string{"trades"}, WindowStart: "2026-03-06T00:00:00Z", WindowEnd: "2026-03-06T23:59:59Z"}, RuntimeMode: contracts.ReplayRuntimeModeRebuild, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	engine, err := replayengine.NewEngine(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, integrationEntryLoader{entries: []ingestion.RawAppendEntry{integrationReplayEntry("2026-03-06T12:00:00Z", 1, "trade:a", "ce:a")}}, integrationArtifactWriter{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute rebuild: %v", err)
	}
	if output.AuditTrail.Request.RequestID == "" || output.AuditTrail.Execution.RunID == "" || len(output.AuditTrail.Checkpoints) != 1 || output.AuditTrail.Outcome.TerminalStatus != "isolated-rebuild" {
		t.Fatalf("incomplete audit trail: %+v", output.AuditTrail)
	}
}

func TestBackfillApplyGateNegativePaths(t *testing.T) {
	records := []ingestion.RawPartitionManifestRecord{{
		SchemaVersion:    "v1",
		LogicalPartition: ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
		StorageState:     ingestion.RawStorageStateHot, Location: "hot://raw/BTC-USD/2026-03-06/trades", EntryCount: 1, ContinuityChecksum: "sha256:abc",
	}}
	builder, _ := replayengine.NewManifestBuilder(&integrationManifestReader{records: records}, integrationSnapshotLoader{snapshot: replayengine.ConfigSnapshot{Ref: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, OrderingPolicyID: "event-time-sequence-canonical-id.v1"}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	manifest, err := builder.Build(contracts.ReplayRunRequest{SchemaVersion: "v1", RunID: "run-apply", Scope: contracts.ReplayScope{Symbol: "BTC-USD", Venues: []string{"coinbase"}, StreamFamilies: []string{"trades"}, WindowStart: "2026-03-06T00:00:00Z", WindowEnd: "2026-03-06T12:00:00Z"}, RuntimeMode: contracts.ReplayRuntimeModeRebuild, ApplyIntent: true, ApprovalContext: &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}, ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"}, ContractVersions: []contracts.ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}}, Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"}})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	outcomes := make(map[string]contracts.ReplayOutcomeAuditRecord)
	first, err := replayengine.ReplayApplyGate(manifest, []contracts.ReplayArtifactRef{{Kind: "rebuild-output", Namespace: "runs/run-apply/rebuild", Digest: "sha256:artifact"}}, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("first apply gate: %v", err)
	}
	second, err := replayengine.ReplayApplyGate(manifest, []contracts.ReplayArtifactRef{{Kind: "rebuild-output", Namespace: "runs/run-apply/rebuild", Digest: "sha256:artifact"}}, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("second apply gate: %v", err)
	}
	if !reflect.DeepEqual(first, second) || !second.CrossedApplyGate {
		t.Fatalf("unexpected apply gate outcome: first=%+v second=%+v", first, second)
	}
}

type integrationManifestReader struct {
	records []ingestion.RawPartitionManifestRecord
}

func (r *integrationManifestReader) ResolveRawPartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error) {
	return append([]ingestion.RawPartitionManifestRecord(nil), r.records...), nil
}

type integrationSnapshotLoader struct{ snapshot replayengine.ConfigSnapshot }

func (l integrationSnapshotLoader) LoadConfigSnapshot(ref contracts.ReplaySnapshotRef) (replayengine.ConfigSnapshot, error) {
	return l.snapshot, nil
}

type integrationEntryLoader struct{ entries []ingestion.RawAppendEntry }

func (l integrationEntryLoader) LoadRawEntries(partitions []contracts.ReplayPartitionRef) ([]ingestion.RawAppendEntry, error) {
	return append([]ingestion.RawAppendEntry(nil), l.entries...), nil
}

type integrationArtifactWriter struct{}

func (integrationArtifactWriter) WriteArtifact(runID string, mode contracts.ReplayRuntimeMode, kind string, payload any) (contracts.ReplayArtifactRef, error) {
	digest, err := contracts.ReplayValueDigest(payload)
	if err != nil {
		return contracts.ReplayArtifactRef{}, err
	}
	return contracts.ReplayArtifactRef{Kind: kind, Namespace: "runs/" + runID + "/" + string(mode), Digest: digest}, nil
}

func integrationReplayEntry(bucketTimestamp string, sequence int64, sourceID string, canonicalID string) ingestion.RawAppendEntry {
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
		BucketTimestampSource:  ingestion.RawBucketTimestampSourceExchange,
		NormalizerService:      "normalizer",
		ConnectionRef:          "conn-1",
		SessionRef:             "session-1",
		BuildVersion:           "test",
		DuplicateAudit:         ingestion.RawDuplicateAudit{IdentityKey: sourceID, Occurrence: 1},
		PartitionKey: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.VenueCoinbase,
			StreamFamily: "trades",
		},
	}
}

func integrationSourceIDs(entries []ingestion.RawAppendEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.VenueMessageID)
	}
	return ids
}
