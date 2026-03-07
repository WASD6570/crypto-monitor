# Testing Report

- Feature: `backfill-checkpoints-and-audit-trail`
- Scope: bounded replay request normalization, checkpoint resume guards, overlap conflict handling, audit records, and apply-gate seams

## Commands

1. `/usr/local/go/bin/go test ./libs/go/... ./services/replay-engine/... ./tests/integration/... -run 'Test(ReplayRequestValidation|ReplayCheckpointResumeGuard|ReplayConflictKeyNormalization|ReplayAuditRecordSchema|ReplayApplyGateIdempotencyKey|ReplayOutcomeRecord|ReplayPromotionToken|ReplayRequestRejectsUnboundedScope|ReplayRequestNormalizesEquivalentScopes|ReplayResumeUsesLastMaterializedCheckpoint|ReplayResumeRejectsConfigSnapshotDrift|ReplayResumeRejectsBuildDrift|ReplayResumeRejectsModeDrift|ReplayAllowsConcurrentInspectAndCompare|ReplayRejectsOverlappingRebuildRequests|ReplayRejectsOverlappingApplyRequests|ReplayAuditRecordsCaptureRequestAndOutcome|ReplayAuditRecordsCaptureCheckpointLineage|ReplayRejectedApplyStillWritesAuditRecord|ReplayApplyModeRequiresApprovalContext|ReplayApplyGateIsIdempotent|ReplayDryRunNeverPublishesSideEffects|ReplayRebuildStopsAtIsolatedArtifacts|BackfillResumeAfterFailure|BackfillOverlapConflictHandling|BackfillAuditTrailCompleteness|BackfillApplyGateNegativePaths)'`
   - Result: passed

## Notes

- Request normalization now produces stable `requestId`, `scopeKey`, and `conflictKey` values before manifest creation.
- Resume coverage validates config, contract-set, build, and mode drift before continuing from the last materialized event boundary.
- Audit coverage verifies request, execution, checkpoint, and outcome facts for success and rejected-apply paths.
- Apply-gate coverage verifies approval is required and duplicate promotion attempts return the same stored outcome.
