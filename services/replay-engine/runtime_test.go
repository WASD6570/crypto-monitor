package replayengine

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestReplayManifestBuilderFreezesResolvedSnapshots(t *testing.T) {
	reader := &stubManifestReader{records: []ingestion.RawPartitionManifestRecord{
		manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111"),
		manifestRecord("coinbase", "books", "hot://raw/coinbase/books", "sha256:222"),
	}}
	builder, err := NewManifestBuilder(reader, stubSnapshotLoader{snapshot: ConfigSnapshot{
		Ref:              contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		OrderingPolicyID: orderingPolicyID,
	}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}

	manifest, err := builder.Build(replayRequest(contracts.ReplayRuntimeModeInspect))
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.ConfigSnapshot.ID != "cfg-1" {
		t.Fatalf("config snapshot id = %q, want cfg-1", manifest.ConfigSnapshot.ID)
	}
	if manifest.Build.GitSHA != "deadbeef" {
		t.Fatalf("git sha = %q, want deadbeef", manifest.Build.GitSHA)
	}
	if len(manifest.RawPartitions) != 2 {
		t.Fatalf("raw partitions = %d, want 2", len(manifest.RawPartitions))
	}
	if manifest.RawPartitions[0].LogicalPartition != "2026-03-06/BTC-USD/coinbase/books" {
		t.Fatalf("first partition = %q, want sorted logical partition", manifest.RawPartitions[0].LogicalPartition)
	}
}

func TestReplayRequestRejectsUnboundedScope(t *testing.T) {
	builder, err := NewManifestBuilder(&stubManifestReader{records: []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}}, stubSnapshotLoader{snapshot: ConfigSnapshot{
		Ref:              contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		OrderingPolicyID: orderingPolicyID,
	}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	request := replayRequest(contracts.ReplayRuntimeModeInspect)
	request.Scope.Venues = []string{"*"}
	if _, err := builder.Build(request); err == nil {
		t.Fatal("expected wildcard scope rejection")
	}
}

func TestReplayRequestNormalizesEquivalentScopes(t *testing.T) {
	builder, err := NewManifestBuilder(&stubManifestReader{records: []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "books", "hot://raw/coinbase/books", "sha256:111"), manifestRecord("kraken", "trades", "hot://raw/kraken/trades", "sha256:222")}}, stubSnapshotLoader{snapshot: ConfigSnapshot{
		Ref:              contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		OrderingPolicyID: orderingPolicyID,
	}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new manifest builder: %v", err)
	}
	left := replayRequest(contracts.ReplayRuntimeModeRebuild)
	left.Scope.Venues = []string{"kraken", "coinbase"}
	left.Scope.StreamFamilies = []string{"trades", "books"}
	right := replayRequest(contracts.ReplayRuntimeModeRebuild)
	right.Scope.Venues = []string{"coinbase", "kraken", "coinbase"}
	right.Scope.StreamFamilies = []string{"books", "trades"}

	leftManifest, err := builder.Build(left)
	if err != nil {
		t.Fatalf("build left manifest: %v", err)
	}
	rightManifest, err := builder.Build(right)
	if err != nil {
		t.Fatalf("build right manifest: %v", err)
	}
	if leftManifest.ScopeKey != rightManifest.ScopeKey || leftManifest.ConflictKey != rightManifest.ConflictKey || leftManifest.RequestID != rightManifest.RequestID {
		t.Fatalf("normalized manifest drift: left=%q/%q/%q right=%q/%q/%q", leftManifest.RequestID, leftManifest.ScopeKey, leftManifest.ConflictKey, rightManifest.RequestID, rightManifest.ScopeKey, rightManifest.ConflictKey)
	}
}

func TestReplayRunFailsOnMissingConfigSnapshot(t *testing.T) {
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, nil, failingSnapshotLoader{err: fmt.Errorf("config snapshot missing")}, &recordingArtifactWriter{})
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})

	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected missing config snapshot error")
	}
	if output.Result.FailureCategory != "missing-snapshot" {
		t.Fatalf("failure category = %q, want missing-snapshot", output.Result.FailureCategory)
	}
}

func TestReplayRunFailsOnManifestChecksumDrift(t *testing.T) {
	record := manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{record})
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:drift")}, nil, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})

	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected manifest checksum drift error")
	}
	if output.Result.FailureCategory != "snapshot-drift" {
		t.Fatalf("failure category = %q, want snapshot-drift", output.Result.FailureCategory)
	}
}

func TestReplayRetentionUsesPreservedSnapshots(t *testing.T) {
	record := manifestRecordWithState("coinbase", "trades", "restore://raw/BTC-USD/2026-03-06/trades", "sha256:111", ingestion.RawStorageStateTransition)
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{record})
	if manifest.RawPartitions[0].StorageState != string(ingestion.RawStorageStateTransition) {
		t.Fatalf("storage state = %q, want transition", manifest.RawPartitions[0].StorageState)
	}
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{record}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})

	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute preserved snapshot replay: %v", err)
	}
	if output.Result.ManifestDigest == "" || output.AuditTrail.Execution.ConfigSnapshot.Digest != manifest.ConfigSnapshot.Digest {
		t.Fatalf("expected preserved snapshots in replay evidence: result=%+v audit=%+v", output.Result, output.AuditTrail.Execution)
	}
}

func TestReplayDeterministicDoubleRun(t *testing.T) {
	entries := []ingestion.RawAppendEntry{
		replayEntry("2026-03-06T12:00:00Z", 0, "trade:b", "ce:b", ingestion.RawBucketTimestampSourceExchange),
		replayEntry("2026-03-06T12:00:00Z", 2, "book:a", "ce:a", ingestion.RawBucketTimestampSourceExchange),
		replayEntry("2026-03-06T12:00:01Z", 0, "trade:c", "ce:c", ingestion.RawBucketTimestampSourceExchange),
	}
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, entries, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})

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
	if !reflect.DeepEqual(sourceIDs(first.OrderedEntries), sourceIDs(second.OrderedEntries)) {
		t.Fatalf("ordered entries differ: %v vs %v", sourceIDs(first.OrderedEntries), sourceIDs(second.OrderedEntries))
	}
	if !reflect.DeepEqual(first.Result.InputCounters, second.Result.InputCounters) {
		t.Fatalf("input counters differ: %+v vs %+v", first.Result.InputCounters, second.Result.InputCounters)
	}
}

func TestReplayStableOrderingWithMixedSequenceAvailability(t *testing.T) {
	ordered := sortReplayEntries([]ingestion.RawAppendEntry{
		replayEntry("2026-03-06T12:00:00Z", 0, "trade:z", "ce:z", ingestion.RawBucketTimestampSourceExchange),
		replayEntry("2026-03-06T12:00:00Z", 2, "book:b", "ce:b", ingestion.RawBucketTimestampSourceExchange),
		replayEntry("2026-03-06T12:00:00Z", 1, "book:a", "ce:a", ingestion.RawBucketTimestampSourceExchange),
		replayEntry("2026-03-06T12:00:00Z", 0, "trade:c", "ce:c", ingestion.RawBucketTimestampSourceExchange),
	})
	got := sourceIDs(ordered)
	want := []string{"book:a", "book:b", "trade:c", "trade:z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered ids = %v, want %v", got, want)
	}
}

func TestReplayPreservesDegradedTimestampOrdering(t *testing.T) {
	ordered := sortReplayEntries([]ingestion.RawAppendEntry{
		replayEntry("2026-03-07T00:00:00.100Z", 0, "trade:next-day", "ce:next-day", ingestion.RawBucketTimestampSourceRecv),
		replayEntry("2026-03-06T23:59:59.900Z", 0, "trade:degraded", "ce:degraded", ingestion.RawBucketTimestampSourceRecv),
	})
	got := sourceIDs(ordered)
	want := []string{"trade:degraded", "trade:next-day"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered ids = %v, want %v", got, want)
	}
}

func TestReplayInspectModeDoesNotWriteArtifacts(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	writer := recordingArtifactWriter{}
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &writer)

	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute inspect: %v", err)
	}
	if len(output.Result.ArtifactRefs) != 0 || len(writer.calls) != 0 {
		t.Fatalf("inspect mode wrote artifacts: result=%d writer=%d", len(output.Result.ArtifactRefs), len(writer.calls))
	}
}

func TestReplayRebuildModeWritesIsolatedArtifacts(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	writer := recordingArtifactWriter{}
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &writer)

	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute rebuild: %v", err)
	}
	if len(output.Result.ArtifactRefs) != 1 {
		t.Fatalf("artifact refs = %d, want 1", len(output.Result.ArtifactRefs))
	}
	if output.Result.ArtifactRefs[0].Namespace != "runs/run-123/rebuild" {
		t.Fatalf("artifact namespace = %q, want run-scoped rebuild namespace", output.Result.ArtifactRefs[0].Namespace)
	}
}

func TestReplayCompareModeEmitsDeterministicSummary(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeCompare
	writer := recordingArtifactWriter{}
	entry := replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{entry}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &writer)
	targetDigest, err := digestEntries([]ingestion.RawAppendEntry{entry})
	if err != nil {
		t.Fatalf("target digest: %v", err)
	}

	output, err := engine.Execute(manifest, &CompareTarget{ID: "baseline-1", Digest: targetDigest})
	if err != nil {
		t.Fatalf("execute compare: %v", err)
	}
	if output.CompareSummary == nil {
		t.Fatal("expected compare summary")
	}
	if output.CompareSummary.DriftClassification != "match" {
		t.Fatalf("drift classification = %q, want match", output.CompareSummary.DriftClassification)
	}
	if output.Result.ComparisonSummaryRef == nil {
		t.Fatal("expected comparison summary ref")
	}
}

func TestReplayRejectsUnsupportedMode(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeMode("apply")
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, nil, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})

	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected unsupported mode error")
	}
	if output.Result.FailureCategory != "invalid-manifest" {
		t.Fatalf("failure category = %q, want invalid-manifest", output.Result.FailureCategory)
	}
}

func TestReplayResumeUsesLastMaterializedCheckpoint(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	checkpoint := replayCheckpoint(manifest, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, "2026-03-07T00:00:00Z")
	decision, err := ReplayResumeFromCheckpoint(manifest, *checkpoint)
	if err != nil {
		t.Fatalf("resume from checkpoint: %v", err)
	}
	if decision.ResumeFromEventID != "ce:a" || decision.NextCheckpointSequence != 2 || decision.RetryCount != 1 {
		t.Fatalf("unexpected resume decision: %+v", decision)
	}
}

func TestReplayResumeRejectsConfigSnapshotDrift(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	checkpoint := replayCheckpoint(manifest, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, "2026-03-07T00:00:00Z")
	manifest.ConfigSnapshot.Digest = "sha256:drift"
	if _, err := ReplayResumeFromCheckpoint(manifest, *checkpoint); err == nil {
		t.Fatal("expected config snapshot drift rejection")
	}
}

func TestReplayResumeRejectsBuildDrift(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	checkpoint := replayCheckpoint(manifest, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, "2026-03-07T00:00:00Z")
	manifest.Build.GitSHA = "beaded"
	if _, err := ReplayResumeFromCheckpoint(manifest, *checkpoint); err == nil {
		t.Fatal("expected build drift rejection")
	}
}

func TestReplayResumeRejectsModeDrift(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	checkpoint := replayCheckpoint(manifest, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, "2026-03-07T00:00:00Z")
	manifest.RuntimeMode = contracts.ReplayRuntimeModeCompare
	if _, err := ReplayResumeFromCheckpoint(manifest, *checkpoint); err == nil {
		t.Fatal("expected mode drift rejection")
	}
}

func TestReplayResumeKeepsPinnedSnapshotsAfterFailure(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecordWithState("coinbase", "trades", "restore://raw/BTC-USD/2026-03-06/trades", "sha256:111", ingestion.RawStorageStateTransition)})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	checkpoint := replayCheckpoint(manifest, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, "2026-03-07T00:00:00Z")
	checkpoint.LastErrorClass = "injected-failure"
	decision, err := ReplayResumeFromCheckpoint(manifest, *checkpoint)
	if err != nil {
		t.Fatalf("resume after failure: %v", err)
	}
	if decision.ResumeFromEventID != "ce:a" || decision.NextCheckpointSequence != 2 || decision.RetryCount != 1 || decision.LogicalPartition != checkpoint.LogicalPartition {
		t.Fatalf("unexpected resume decision: %+v", decision)
	}
}

func TestReplayAllowsConcurrentInspectAndCompare(t *testing.T) {
	inspect := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	compare := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	compare.RuntimeMode = contracts.ReplayRuntimeModeCompare
	compare.RequestID = "sha256:compare"
	key, conflicted, err := ReplayRequestsConflict(compare, []contracts.ReplayRunManifest{inspect})
	if err != nil {
		t.Fatalf("detect conflict: %v", err)
	}
	if conflicted || key != "" {
		t.Fatalf("expected no conflict, got conflicted=%v key=%q", conflicted, key)
	}
}

func TestReplayRejectsOverlappingRebuildRequests(t *testing.T) {
	active := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	active.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	active.RequestID = "sha256:active"
	active.ConflictKey = "sha256:active-conflict"
	candidate := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	candidate.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	candidate.RequestID = "sha256:candidate"
	key, conflicted, err := ReplayRequestsConflict(candidate, []contracts.ReplayRunManifest{active})
	if err != nil {
		t.Fatalf("detect conflict: %v", err)
	}
	if !conflicted || key == "" {
		t.Fatalf("expected overlapping rebuild conflict, got conflicted=%v key=%q", conflicted, key)
	}
}

func TestReplayRejectsOverlappingApplyRequests(t *testing.T) {
	active := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	active.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	active.ApplyIntent = true
	active.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}
	active.RequestID = "sha256:active-apply"
	active.ConflictKey = "sha256:active-apply-conflict"
	candidate := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	candidate.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	candidate.ApplyIntent = true
	candidate.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-2", PromotionToken: "promo-2"}
	candidate.RequestID = "sha256:candidate-apply"
	key, conflicted, err := ReplayRequestsConflict(candidate, []contracts.ReplayRunManifest{active})
	if err != nil {
		t.Fatalf("detect conflict: %v", err)
	}
	if !conflicted || key == "" {
		t.Fatalf("expected overlapping apply conflict, got conflicted=%v key=%q", conflicted, key)
	}
}

func TestReplayAuditRecordsCaptureRequestAndOutcome(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute inspect: %v", err)
	}
	if output.AuditTrail.Request.RequestID == "" || output.AuditTrail.Outcome.TerminalStatus != "no-op" {
		t.Fatalf("unexpected audit trail: %+v", output.AuditTrail)
	}
	if err := contracts.ValidateReplayExecutionAuditRecord(output.AuditTrail.Execution); err != nil {
		t.Fatalf("validate execution audit: %v", err)
	}
}

func TestReplayAuditRecordsCaptureCheckpointLineage(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute rebuild: %v", err)
	}
	if output.Checkpoint == nil || len(output.AuditTrail.Checkpoints) != 1 {
		t.Fatalf("expected checkpoint lineage, got checkpoint=%v audit=%d", output.Checkpoint, len(output.AuditTrail.Checkpoints))
	}
	if err := contracts.ValidateReplayCheckpointAuditRecord(output.AuditTrail.Checkpoints[0]); err != nil {
		t.Fatalf("validate checkpoint audit: %v", err)
	}
}

func TestReplayRejectedApplyStillWritesAuditRecord(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	manifest.ApplyIntent = true
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, nil, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})
	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected approval-required error")
	}
	if output.AuditTrail.Outcome.TerminalStatus != "rejected-apply" || output.AuditTrail.Request.RequestID == "" {
		t.Fatalf("unexpected rejected apply audit trail: %+v", output.AuditTrail)
	}
}

func TestReplayRejectedApplyStillWritesAuditEvidence(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	manifest.ApplyIntent = true
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, nil, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})
	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected approval-required error")
	}
	if output.Result.FailureCategory != "approval-required" || output.AuditTrail.Outcome.TerminalStatus != "rejected-apply" {
		t.Fatalf("unexpected rejected apply output: %+v", output)
	}
	if err := contracts.ValidateReplayExecutionAuditRecord(output.AuditTrail.Execution); err != nil {
		t.Fatalf("validate execution audit: %v", err)
	}
}

func TestReplayApplyModeRequiresApprovalContext(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	manifest.ApplyIntent = true
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, nil, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &recordingArtifactWriter{})
	output, err := engine.Execute(manifest, nil)
	if err == nil {
		t.Fatal("expected approval-required error")
	}
	if output.Result.FailureCategory != "approval-required" {
		t.Fatalf("failure category = %q, want approval-required", output.Result.FailureCategory)
	}
}

func TestReplayApplyGateIsIdempotent(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	manifest.ApplyIntent = true
	manifest.ApprovalContext = &contracts.ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}
	outcomes := make(map[string]contracts.ReplayOutcomeAuditRecord)
	artifacts := []contracts.ReplayArtifactRef{{Kind: "rebuild-output", Namespace: "runs/run-123/rebuild", Digest: "sha256:artifact"}}
	first, err := ReplayApplyGate(manifest, artifacts, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("first apply gate: %v", err)
	}
	second, err := ReplayApplyGate(manifest, artifacts, "sha256:prior", outcomes, time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("second apply gate: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("idempotent apply outcome drift: %+v vs %+v", first, second)
	}
}

func TestReplayDryRunNeverPublishesSideEffects(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	writer := recordingArtifactWriter{}
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &writer)
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute inspect: %v", err)
	}
	if len(writer.calls) != 0 || output.AuditTrail.Outcome.CrossedApplyGate {
		t.Fatalf("inspect run published side effects: writer=%d crossedApplyGate=%v", len(writer.calls), output.AuditTrail.Outcome.CrossedApplyGate)
	}
}

func TestReplayRebuildStopsAtIsolatedArtifacts(t *testing.T) {
	manifest := builtManifest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")})
	manifest.RuntimeMode = contracts.ReplayRuntimeModeRebuild
	writer := recordingArtifactWriter{}
	engine := newEngineForTest(t, []ingestion.RawPartitionManifestRecord{manifestRecord("coinbase", "trades", "hot://raw/coinbase/trades", "sha256:111")}, []ingestion.RawAppendEntry{replayEntry("2026-03-06T12:00:00Z", 0, "trade:a", "ce:a", ingestion.RawBucketTimestampSourceExchange)}, stubSnapshotLoader{snapshot: ConfigSnapshot{Ref: manifest.ConfigSnapshot, OrderingPolicyID: manifest.OrderingPolicyID}}, &writer)
	output, err := engine.Execute(manifest, nil)
	if err != nil {
		t.Fatalf("execute rebuild: %v", err)
	}
	if output.AuditTrail.Outcome.TerminalStatus != "isolated-rebuild" || output.AuditTrail.Outcome.CrossedApplyGate {
		t.Fatalf("unexpected rebuild outcome: %+v", output.AuditTrail.Outcome)
	}
}

type stubSnapshotLoader struct{ snapshot ConfigSnapshot }

func (l stubSnapshotLoader) LoadConfigSnapshot(ref contracts.ReplaySnapshotRef) (ConfigSnapshot, error) {
	if l.snapshot.Ref.ID == "" {
		return ConfigSnapshot{}, fmt.Errorf("config snapshot missing")
	}
	return l.snapshot, nil
}

type failingSnapshotLoader struct{ err error }

func (l failingSnapshotLoader) LoadConfigSnapshot(ref contracts.ReplaySnapshotRef) (ConfigSnapshot, error) {
	return ConfigSnapshot{}, l.err
}

type stubEntryLoader struct{ entries []ingestion.RawAppendEntry }

func (l stubEntryLoader) LoadRawEntries(partitions []contracts.ReplayPartitionRef) ([]ingestion.RawAppendEntry, error) {
	return append([]ingestion.RawAppendEntry(nil), l.entries...), nil
}

type recordingArtifactWriter struct{ calls []contracts.ReplayArtifactRef }

func (w *recordingArtifactWriter) WriteArtifact(runID string, mode contracts.ReplayRuntimeMode, kind string, payload any) (contracts.ReplayArtifactRef, error) {
	digest, err := contracts.ReplayValueDigest(payload)
	if err != nil {
		return contracts.ReplayArtifactRef{}, err
	}
	ref := contracts.ReplayArtifactRef{Kind: kind, Namespace: fmt.Sprintf("runs/%s/%s", runID, mode), Digest: digest}
	w.calls = append(w.calls, ref)
	return ref, nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func newEngineForTest(t *testing.T, records []ingestion.RawPartitionManifestRecord, entries []ingestion.RawAppendEntry, snapshotLoader ConfigSnapshotLoader, writer *recordingArtifactWriter) *Engine {
	t.Helper()
	engine, err := NewEngine(&stubManifestReader{records: records}, snapshotLoader, stubEntryLoader{entries: entries}, writer)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.clock = fixedClock{now: time.Date(2026, time.March, 7, 1, 2, 3, 0, time.UTC)}
	return engine
}

func builtManifest(t *testing.T, records []ingestion.RawPartitionManifestRecord) contracts.ReplayRunManifest {
	t.Helper()
	builder, err := NewManifestBuilder(&stubManifestReader{records: records}, stubSnapshotLoader{snapshot: ConfigSnapshot{
		Ref:              contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		OrderingPolicyID: orderingPolicyID,
	}}, contracts.ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("new builder: %v", err)
	}
	manifest, err := builder.Build(replayRequest(contracts.ReplayRuntimeModeInspect))
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	return manifest
}

func replayRequest(mode contracts.ReplayRuntimeMode) contracts.ReplayRunRequest {
	return contracts.ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "run-123",
		Scope: contracts.ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{"coinbase"},
			StreamFamilies: []string{"books", "trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:    mode,
		ConfigSnapshot: contracts.ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []contracts.ReplaySnapshotRef{{
			SchemaVersion: "v1",
			Kind:          "schema",
			ID:            "replay-family",
			Version:       "v1",
			Digest:        "sha256:family",
		}},
		Initiator: contracts.ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}
}

func manifestRecord(venue string, streamFamily string, location string, checksum string) ingestion.RawPartitionManifestRecord {
	return manifestRecordWithState(venue, streamFamily, location, checksum, ingestion.RawStorageStateHot)
}

func manifestRecordWithState(venue string, streamFamily string, location string, checksum string, state ingestion.RawStorageState) ingestion.RawPartitionManifestRecord {
	return ingestion.RawPartitionManifestRecord{
		SchemaVersion: "v1",
		LogicalPartition: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.Venue(venue),
			StreamFamily: streamFamily,
		},
		StorageState:          state,
		Location:              location,
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            1,
		FirstCanonicalEventID: "ce:first",
		LastCanonicalEventID:  "ce:last",
		ContinuityChecksum:    checksum,
	}
}

func replayEntry(bucketTimestamp string, sequence int64, sourceID string, canonicalID string, source ingestion.RawBucketTimestampSource) ingestion.RawAppendEntry {
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
		PartitionKey: ingestion.RawPartitionKey{
			UTCDate:      "2026-03-06",
			Symbol:       "BTC-USD",
			Venue:        ingestion.VenueCoinbase,
			StreamFamily: "trades",
		},
	}
}

func sourceIDs(entries []ingestion.RawAppendEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.VenueMessageID)
	}
	return ids
}
