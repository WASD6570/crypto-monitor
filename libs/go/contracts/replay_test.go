package contracts

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReplayRunManifestSchemaDecode(t *testing.T) {
	manifest := ReplayRunManifest{
		SchemaVersion:    "v1",
		RunID:            "run-123",
		RequestID:        "sha256:reqid",
		RequestDigest:    "sha256:req",
		ScopeKey:         "symbol=BTC-USD|venues=coinbase|window=2026-03-06T00:00:00Z/2026-03-06T23:59:59Z",
		ConflictKey:      "sha256:conflict",
		RuntimeMode:      ReplayRuntimeModeInspect,
		Status:           ReplayManifestStatusPlanned,
		OrderingPolicyID: "event-time-v1",
		Scope: ReplayScope{
			Symbol:      "BTC-USD",
			Venues:      []string{"coinbase"},
			WindowStart: "2026-03-06T00:00:00Z",
			WindowEnd:   "2026-03-06T23:59:59Z",
		},
		RawPartitions: []ReplayPartitionRef{{
			SchemaVersion:         "v1",
			LogicalPartition:      "2026-03-06/BTC-USD/coinbase/trades",
			Location:              "hot://raw/BTC-USD/2026-03-06",
			EntryCount:            2,
			FirstCanonicalEventID: "ce:first",
			LastCanonicalEventID:  "ce:last",
			ContinuityChecksum:    "sha256:abc",
		}},
		ConfigSnapshot:   ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-run-manifest", Version: "v1", Digest: "sha256:schema"}},
		Build:            ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"},
		Initiator:        ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}

	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	var decoded ReplayRunManifest
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if err := ValidateReplayRunManifest(decoded); err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
}

func TestReplaySnapshotVersionGuard(t *testing.T) {
	err := ValidateReplaySnapshotRef(ReplaySnapshotRef{
		SchemaVersion: "v2",
		Kind:          "config",
		ID:            "cfg-1",
		Digest:        "sha256:cfg",
	})
	if err == nil {
		t.Fatal("expected unsupported snapshot schema version error")
	}
}

func TestReplayManifestDigestValidation(t *testing.T) {
	manifest := ReplayRunManifest{
		SchemaVersion:    "v1",
		RunID:            "run-123",
		RequestID:        "sha256:reqid",
		RequestDigest:    "sha256:req",
		ScopeKey:         "symbol=BTC-USD|venues=coinbase|window=2026-03-06T00:00:00Z/2026-03-06T23:59:59Z",
		ConflictKey:      "sha256:conflict",
		RuntimeMode:      ReplayRuntimeModeInspect,
		Status:           ReplayManifestStatusCompleted,
		OrderingPolicyID: "event-time-v1",
		Scope: ReplayScope{
			Symbol:      "BTC-USD",
			Venues:      []string{"coinbase"},
			WindowStart: "2026-03-06T00:00:00Z",
			WindowEnd:   "2026-03-06T23:59:59Z",
		},
		RawPartitions: []ReplayPartitionRef{{
			SchemaVersion:      "v1",
			LogicalPartition:   "2026-03-06/BTC-USD/coinbase/trades",
			Location:           "hot://raw/BTC-USD/2026-03-06",
			EntryCount:         2,
			ContinuityChecksum: "sha256:abc",
		}},
		ConfigSnapshot:   ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Version: "v1", Digest: "sha256:cfg"},
		ContractVersions: []ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-run-manifest", Version: "v1", Digest: "sha256:schema"}},
		Build:            ReplayBuildProvenance{Service: "replay-engine"},
		Initiator:        ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}

	digest, err := ReplayManifestDigest(manifest)
	if err != nil {
		t.Fatalf("manifest digest: %v", err)
	}
	if err := ReplayDigestMatches(digest, digest); err != nil {
		t.Fatalf("digest should match: %v", err)
	}
	if err := ReplayDigestMatches(digest, "sha256:different"); err == nil {
		t.Fatal("expected digest drift error")
	}
}

func TestReplayRequestValidation(t *testing.T) {
	request := replayRequestFixture(ReplayRuntimeModeInspect)
	request.Scope.Venues = []string{"coinbase", "*"}
	if err := ValidateReplayRunRequest(request); err != nil {
		t.Fatalf("base request should validate before normalization-specific checks: %v", err)
	}
	if _, _, err := NormalizeReplayScope(request.Scope); err == nil {
		t.Fatal("expected wildcard venue rejection")
	}

	request = replayRequestFixture(ReplayRuntimeModeInspect)
	request.Scope.WindowEnd = ""
	if err := ValidateReplayRunRequest(request); err == nil {
		t.Fatal("expected missing end window rejection")
	}
}

func TestReplayConflictKeyNormalization(t *testing.T) {
	left := replayRequestFixture(ReplayRuntimeModeRebuild)
	left.Scope.Venues = []string{"kraken", "coinbase"}
	left.Scope.StreamFamilies = []string{"trades", "books"}

	right := replayRequestFixture(ReplayRuntimeModeRebuild)
	right.Scope.Venues = []string{"coinbase", "kraken", "coinbase"}
	right.Scope.StreamFamilies = []string{"books", "trades"}

	leftID, leftScopeKey, leftConflictKey, err := ReplayRequestIdentity(left, ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("left identity: %v", err)
	}
	rightID, rightScopeKey, rightConflictKey, err := ReplayRequestIdentity(right, ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("right identity: %v", err)
	}
	if leftID != rightID || leftScopeKey != rightScopeKey || leftConflictKey != rightConflictKey {
		t.Fatalf("normalized identity drift: left=%q/%q/%q right=%q/%q/%q", leftID, leftScopeKey, leftConflictKey, rightID, rightScopeKey, rightConflictKey)
	}
}

func TestReplayCheckpointResumeGuard(t *testing.T) {
	contractSetDigest, err := ReplayContractSetDigest([]ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}})
	if err != nil {
		t.Fatalf("contract set digest: %v", err)
	}
	buildDigest, err := ReplayBuildDigest(ReplayBuildProvenance{Service: "replay-engine", GitSHA: "deadbeef"})
	if err != nil {
		t.Fatalf("build digest: %v", err)
	}
	checkpoint := ReplayCheckpoint{
		SchemaVersion:           "v1",
		RequestID:               "sha256:reqid",
		RunID:                   "run-123",
		ScopeKey:                "scope-key",
		LogicalPartition:        "2026-03-06/BTC-USD/coinbase/trades",
		LastMaterializedEventID: "ce:last",
		LastScannedEventID:      "ce:last",
		ConfigSnapshot:          ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Digest: "sha256:cfg"},
		ContractSetDigest:       contractSetDigest,
		BuildDigest:             buildDigest,
		OutputMode:              ReplayRuntimeModeRebuild,
		Sequence:                1,
		RetryCount:              0,
		RecordedAt:              "2026-03-07T00:00:00Z",
	}
	if err := ValidateReplayCheckpoint(checkpoint); err != nil {
		t.Fatalf("validate checkpoint: %v", err)
	}
	checkpoint.BuildDigest = "sha256:drift"
	if err := ValidateReplayCheckpoint(checkpoint); err != nil {
		if !strings.Contains(err.Error(), "incomplete") {
			t.Fatalf("unexpected checkpoint validation error: %v", err)
		}
	}
}

func TestReplayAuditRecordSchema(t *testing.T) {
	record := ReplayRequestAuditRecord{
		SchemaVersion: "v1",
		RequestID:     "sha256:reqid",
		RunID:         "run-123",
		Initiator:     "tester",
		ReasonCode:    "audit",
		ScopeKey:      "scope-key",
		RuntimeMode:   ReplayRuntimeModeInspect,
		RecordedAt:    "2026-03-07T00:00:00Z",
	}
	if err := ValidateReplayRequestAuditRecord(record); err != nil {
		t.Fatalf("validate request audit record: %v", err)
	}
}

func TestReplayApplyGateIdempotencyKey(t *testing.T) {
	left, err := ReplayApplyGateIdempotencyKey("sha256:reqid", "promo-1")
	if err != nil {
		t.Fatalf("left key: %v", err)
	}
	right, err := ReplayApplyGateIdempotencyKey("sha256:reqid", "promo-1")
	if err != nil {
		t.Fatalf("right key: %v", err)
	}
	if left != right {
		t.Fatalf("idempotency key mismatch: %q vs %q", left, right)
	}
}

func TestReplayOutcomeRecord(t *testing.T) {
	record := ReplayOutcomeAuditRecord{
		SchemaVersion:  "v1",
		RequestID:      "sha256:reqid",
		RunID:          "run-123",
		TerminalStatus: "isolated-rebuild",
		RecordedAt:     "2026-03-07T00:00:00Z",
	}
	if err := ValidateReplayOutcomeAuditRecord(record); err != nil {
		t.Fatalf("validate outcome record: %v", err)
	}
}

func TestReplayPromotionToken(t *testing.T) {
	if err := ValidateReplayApprovalContext(&ReplayApprovalContext{AuthorizationContext: "ops-admin", ApprovalRef: "apr-1", PromotionToken: "promo-1"}); err != nil {
		t.Fatalf("validate approval context: %v", err)
	}
}

func replayRequestFixture(mode ReplayRuntimeMode) ReplayRunRequest {
	return ReplayRunRequest{
		SchemaVersion: "v1",
		RunID:         "run-123",
		Scope: ReplayScope{
			Symbol:         "BTC-USD",
			Venues:         []string{"coinbase"},
			StreamFamilies: []string{"trades"},
			WindowStart:    "2026-03-06T00:00:00Z",
			WindowEnd:      "2026-03-06T23:59:59Z",
		},
		RuntimeMode:      mode,
		ConfigSnapshot:   ReplaySnapshotRef{SchemaVersion: "v1", Kind: "config", ID: "cfg-1", Digest: "sha256:cfg"},
		ContractVersions: []ReplaySnapshotRef{{SchemaVersion: "v1", Kind: "schema", ID: "replay-family", Version: "v1", Digest: "sha256:family"}},
		Initiator:        ReplayInitiatorMetadata{Actor: "tester", ReasonCode: "audit", RequestedAt: "2026-03-07T00:00:00Z"},
	}
}
