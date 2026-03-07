# Testing

## Testing Goals

- prove bounded replay/backfill requests reject ambiguous scope and normalize equivalent requests deterministically
- prove failed correction runs resume only from durable checkpoints pinned to identical scope and snapshot inputs
- prove overlapping correction requests and apply-capable overlaps are rejected or serialized deterministically
- prove every request path leaves structured audit facts and keeps external side effects disabled by default

## Output Artifact

- Write the implementation-phase test report to `plans/completed/backfill-checkpoints-and-audit-trail/testing-report.md`.

## Test Matrix

### 1. Request Bounds And Scope Normalization

- Purpose: verify the control surface accepts only explicit bounded requests and produces one canonical scope identity.
- Fixtures:
  - single-symbol single-day request for `BTC-USD`
  - equivalent requests with reordered venue or stream-family input fields
  - missing end time, missing reason, and wildcard scope negatives
- Validation commands:
  - `go test ./libs/go/... -run 'Test(ReplayRequestValidation|ReplayConflictKeyNormalization)'`
  - `go test ./services/replay-engine/... -run 'TestReplayRequestRejectsUnboundedScope|TestReplayRequestNormalizesEquivalentScopes'`
- Verify:
  - unbounded or ambiguous requests fail before replay execution starts
  - logically equivalent requests share the same normalized scope and conflict key
  - bounded window defaults remain policy-driven and explicit in the request record

### 2. Checkpoint Resume And Drift Rejection

- Purpose: prove checkpoint recovery is idempotent for identical inputs and refuses drifted resumes.
- Fixtures:
  - failed run with durable checkpoint after one partition step
  - retry with identical config snapshot, contract set, and build provenance
  - retry with changed config snapshot digest
  - retry with changed build provenance or output mode
- Validation commands:
  - `go test ./libs/go/... -run 'TestReplayCheckpointResumeGuard'`
  - `go test ./services/replay-engine/... -run 'TestReplayResumeUsesLastMaterializedCheckpoint|TestReplayResumeRejectsConfigSnapshotDrift|TestReplayResumeRejectsBuildDrift|TestReplayResumeRejectsModeDrift'`
  - `go test ./tests/integration/... -run 'TestBackfillResumeAfterFailure'`
- Verify:
  - resume starts from the last fully materialized event boundary
  - repeated resume attempts do not duplicate isolated outputs
  - snapshot or build drift forces a new request instead of silent continuation

### 3. Overlap Conflict Handling

- Purpose: verify deterministic behavior when requests cover the same or overlapping correction scope.
- Fixtures:
  - concurrent inspect and compare requests for the same scope
  - overlapping rebuild requests for the same symbol/day
  - overlapping apply-capable requests with one narrower and one wider window
- Validation commands:
  - `go test ./services/replay-engine/... -run 'TestReplayAllowsConcurrentInspectAndCompare|TestReplayRejectsOverlappingRebuildRequests|TestReplayRejectsOverlappingApplyRequests'`
  - `go test ./tests/integration/... -run 'TestBackfillOverlapConflictHandling'`
- Verify:
  - non-materializing modes can coexist when policy allows
  - materializing conflicts produce one deterministic rejection or queue outcome
  - overlap decisions are based on normalized scope and not raw request formatting

### 4. Audit Trail Completeness

- Purpose: verify request, execution, checkpoint, and outcome facts are emitted for both success and failure paths.
- Fixtures:
  - inspect request that completes successfully
  - rebuild request that fails and resumes
  - apply-capable request rejected for missing approval context
- Validation commands:
  - `go test ./libs/go/... -run 'Test(ReplayAuditRecordSchema|ReplayOutcomeRecord)'`
  - `go test ./services/replay-engine/... -run 'TestReplayAuditRecordsCaptureRequestAndOutcome|TestReplayAuditRecordsCaptureCheckpointLineage|TestReplayRejectedApplyStillWritesAuditRecord'`
  - `go test ./tests/integration/... -run 'TestBackfillAuditTrailCompleteness'`
- Verify:
  - each path writes machine-readable request, execution, checkpoint, and outcome records as applicable
  - rejected actions remain visible in audit history
  - outcome records link manifests, diffs, and replacement references when promotion is attempted

### 5. Apply Gates And Side-Effect Safety

- Purpose: prove correction work stops at isolated outputs unless an explicit apply path is authorized.
- Fixtures:
  - dry-run inspect request
  - isolated rebuild request
  - apply-capable request with missing approval context
  - duplicate approved apply attempt using the same promotion token
- Validation commands:
  - `go test ./libs/go/... -run 'Test(ReplayApplyGateIdempotencyKey|ReplayPromotionToken)'`
  - `go test ./services/replay-engine/... -run 'TestReplayApplyModeRequiresApprovalContext|TestReplayApplyGateIsIdempotent|TestReplayDryRunNeverPublishesSideEffects|TestReplayRebuildStopsAtIsolatedArtifacts'`
  - `go test ./tests/integration/... -run 'TestBackfillApplyGateNegativePaths'`
- Verify:
  - inspect and rebuild never cross the apply gate by default
  - missing approval context blocks apply-mode requests before promotion
  - duplicate approved apply attempts are idempotent or deterministically rejected
  - no alert, webhook, or notification side effects occur in any default path

## Required Negative Cases

- wildcard or open-ended backfill scope
- resume attempt with changed config snapshot digest
- resume attempt with changed build provenance or mode
- overlapping apply-capable requests on the same symbol/day window
- denied apply request that still requires an audit record
- failed run whose checkpoint exists but whose isolated outputs already partially materialized

## Downstream Validation Boundary

- Hot/cold retention continuity, broader replay safety smoke matrices, and final end-to-end correction validation now live in `plans/completed/replay-retention-and-safety-validation/` within the completed epic `plans/completed/raw-storage-and-replay-foundation/`.
- This testing plan should include only the minimum downstream assertion needed here: checkpoint and audit records must carry logical partition/manifest references that later retention validation can consume.

## Exit Criteria For Implementation

- targeted Go unit, service, replay, and integration commands pass
- checkpoint resume is proven idempotent for identical pinned inputs
- audit records exist for success, failure, rejection, and promotion-attempt paths
- apply-capable paths remain denied by default without explicit approval context
- the resulting test report is written to `plans/completed/backfill-checkpoints-and-audit-trail/testing-report.md`
