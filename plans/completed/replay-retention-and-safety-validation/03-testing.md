# Testing Plan: Replay Retention And Safety Validation

Expected output artifact: `plans/completed/replay-retention-and-safety-validation/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Hot-to-cold continuity | Resolve and replay the same symbol/day scope from hot, cold, and restored partition states | Counts, ordered event IDs, late/degraded counters, and replay digests stay identical; manifest resolution never guesses tier paths | Integration output plus manifest/digest assertions |
| Replay determinism after retention movement | Run the same retention-aware replay twice with identical snapshots | Run manifests, result digests, and ordered output references match exactly | Replay smoke output plus checksum evidence |
| Late-event continuity | Replay a scope containing persisted late events before and after tier movement | Late-event markers remain visible and compare/repair candidates stay stable | Replay compare output |
| Resume after failure | Fail a rebuild after checkpoint creation, then resume using the same pinned inputs | Resume continues from the checkpoint boundary without duplicate materialization or drifted inputs | Checkpoint lineage output |
| Side-effect-safe mode matrix | Run `inspect`, `rebuild`, `compare`, rejected `apply`, and idempotent approved `apply` against instrumented sinks | Non-apply modes emit zero side effects; rejected apply emits zero side effects; repeated approved apply is idempotent | Sink counters plus audit evidence |
| Overlap/conflict handling | Submit overlapping rebuild/apply requests for the same scope during recovery smoke | Requests are rejected or serialized deterministically with audit evidence | Integration output |
| Missing snapshot / manifest gap negative | Remove a required snapshot ref or create a manifest continuity gap | Replay fails before materialization with a clear validation error | Error-path test output |

## Required Commands

The implementing agent should provide these exact commands or repo-standard direct equivalents:

- `/usr/local/go/bin/go test ./services/replay-engine/... -run 'TestReplayScopeResolutionDoesNotGuessTierPaths|TestReplayRetentionUsesPreservedSnapshots|TestReplayResumeKeepsPinnedSnapshotsAfterFailure|TestReplayRejectedApplyStillWritesAuditEvidence'`
- `/usr/local/go/bin/go test ./tests/replay/... -run 'TestReplayDeterminismAcrossRetentionTiers|TestReplayRetentionPreservesLateAndDegradedEvidence|TestReplayCompareCapturesLateEventRepairCandidates|TestReplayModesDoNotEmitSideEffectsByDefault'`
- `/usr/local/go/bin/go test ./tests/integration -run 'TestReplayRetentionContinuityAcrossHotAndCold|TestReplayColdRestoreIsExplicitAndDeterministic|TestReplaySafetyMatrixAcrossModes|TestBackfillResumeAfterRetentionRestore|TestReplayApplyGateRejectsWithoutApproval|TestReplayApplyIsIdempotentAcrossRetries|TestReplayOverlapHandlingRemainsDeterministic'`

If package paths differ during implementation, replace them with equally explicit commands that still isolate retention continuity, deterministic replay, resume safety, and apply-gate behavior.

## Verification Checklist

- Replay resolves hot, cold, and restored partitions through logical references, not tier-specific guessing.
- Hot/cold transitions preserve canonical event IDs, ordering inputs, timestamp provenance, and degraded markers.
- Repeated retention-aware replay runs produce identical manifests, counts, and digests for the same inputs.
- Late events remain persisted, visible, and repairable through replay rather than hidden mutation.
- Resume behavior stays idempotent and rejects snapshot/build/mode drift.
- `inspect`, `rebuild`, and `compare` emit no external side effects.
- `apply` requires explicit approval context and remains idempotent across retries.
- Python remains optional and offline-only.

## Negative Cases

- missing cold-tier manifest reference for a historical partition
- missing config, contract, or build snapshot for a replay resume
- tier transition that changes ordered event IDs or event counts
- apply attempt without approval context or replay-aware idempotency key
- repeated apply attempt that tries to materialize a second outcome
- overlap request that bypasses deterministic conflict handling

## Handoff Notes

- Use deterministic local fixtures and instrumented sinks only; no exchange credentials or live endpoints.
- Keep testing evidence focused on integrated proof across completed replay slices rather than broad new functionality.
- The feature is done only when another agent can run the commands above and produce `plans/completed/replay-retention-and-safety-validation/testing-report.md` with continuity, resume, and side-effect evidence.
