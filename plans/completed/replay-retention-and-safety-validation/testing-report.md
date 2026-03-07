# Testing Report: Replay Retention And Safety Validation

## Outcome

- Added retention-aware replay coverage for hot, cold, and explicit restore partition states while preserving logical partition identity, ordered canonical IDs, and late/degraded counters.
- Added integrated safety matrix coverage for `inspect`, `rebuild`, `compare`, rejected `apply`, resume from restored checkpoints, and idempotent approved `apply` flows.
- Added Go-visible storage-state evidence on replay partition refs so deterministic retention continuity assertions do not rely on log scraping.

## Commands

1. `/usr/local/go/bin/go test ./services/replay-engine/... -run 'TestReplayScopeResolutionDoesNotGuessTierPaths|TestReplayRetentionUsesPreservedSnapshots|TestReplayResumeKeepsPinnedSnapshotsAfterFailure|TestReplayRejectedApplyStillWritesAuditEvidence'`
   - Result: passed
2. `/usr/local/go/bin/go test ./tests/replay/... -run 'TestReplayDeterminismAcrossRetentionTiers|TestReplayRetentionPreservesLateAndDegradedEvidence|TestReplayCompareCapturesLateEventRepairCandidates|TestReplayModesDoNotEmitSideEffectsByDefault'`
   - Result: passed
3. `/usr/local/go/bin/go test ./tests/integration -run 'TestReplayRetentionContinuityAcrossHotAndCold|TestReplayColdRestoreIsExplicitAndDeterministic|TestReplaySafetyMatrixAcrossModes|TestBackfillResumeAfterRetentionRestore|TestReplayApplyGateRejectsWithoutApproval|TestReplayApplyIsIdempotentAcrossRetries|TestReplayOverlapHandlingRemainsDeterministic'`
   - Result: passed

## Notes

- Explicit restore proof uses the additive replay partition `storageState` field with `transition` plus a restore-scoped location reference; no storage-vendor behavior was introduced.
- Retention continuity assertions stay storage-engine-neutral by comparing manifest continuity, counters, digests, and ordered event identities rather than concrete storage layouts.
- Side-effect safety remains bounded to apply-gate and audit evidence; non-apply modes remain isolated and never cross the apply gate by default.
