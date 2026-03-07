# Implementation Module 2: Side-Effect Safety And Recovery Matrix

## Requirements And Scope

- Prove the completed replay/backfill control flow remains safe when exercised end to end with real retention-aware replay inputs.
- Cover `inspect`, `rebuild`, `compare`, resume-after-failure, overlap handling, and explicitly gated `apply` behavior.
- Keep the module bounded to integrated validation, sink instrumentation, and narrow evidence hooks; do not redesign request bounds, checkpoint identity, audit structures, or runtime modes.
- Show that late-event repair candidates and historical correction paths remain auditable instead of silently mutating prior outputs.

## Target Repo Areas

- `services/replay-engine`
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`
- `docs/runbooks`

## Implementation Notes

### Safety Matrix

- Exercise at least these integrated paths for one deterministic symbol/day scope:
  - `inspect` against hot partitions
  - `rebuild` against cold or restored partitions
  - `compare` with late-event correction evidence
  - resumed `rebuild` after injected failure
  - rejected overlapping `apply` request
  - successful gated `apply` with idempotent replay-aware sink keys
- Keep all external sinks instrumented doubles or local harnesses; no live endpoints or credentials.

### Evidence To Capture

- request and run identifiers
- checkpoint lineage before and after resume
- late, duplicate, degraded, and processed counters
- compare or diff artifact references for late-event repair cases
- sink emission counters by mode
- apply approval context and idempotency token reuse outcome

### Recovery Assertions

- Resume continues from the last materialized checkpoint using the pinned scope, snapshots, and build provenance.
- Resume with drifted config, contract set, build, or mode remains rejected by the existing rules.
- Overlapping rebuild/apply requests resolve deterministically with audit evidence instead of racing.

### Side-Effect Assertions

- `inspect`, `rebuild`, and `compare` produce zero external sink emissions.
- rejected `apply` still records audit evidence and emits zero side effects.
- repeated `apply` with the same approval/idempotency context returns the same stored outcome rather than duplicating promotion or notifications.

## Unit And Integration Test Expectations

- `go test ./tests/integration -run 'TestReplaySafetyMatrixAcrossModes|TestBackfillResumeAfterRetentionRestore|TestReplayApplyGateRejectsWithoutApproval|TestReplayApplyIsIdempotentAcrossRetries|TestReplayOverlapHandlingRemainsDeterministic'`
- `go test ./tests/replay/... -run 'TestReplayCompareCapturesLateEventRepairCandidates|TestReplayModesDoNotEmitSideEffectsByDefault'`
- `go test ./services/replay-engine/... -run 'TestReplayResumeKeepsPinnedSnapshotsAfterFailure|TestReplayRejectedApplyStillWritesAuditEvidence'`

## Contract / Fixture / Replay Impacts

- Prefer existing request, checkpoint, audit, and outcome contracts from the completed backfill slice.
- Add fixture coverage only for integrated late-event repair, injected failure, and sink instrumentation scenarios.
- If a new evidence field is required, keep it additive, Go-owned, and strictly for deterministic validation or runbook readability.

## Summary

This module closes the safety story for the replay foundation. The key result is an integrated recovery matrix that proves no-default-side-effect behavior, deterministic resume semantics, and auditable gated apply flows across the completed replay slices.
