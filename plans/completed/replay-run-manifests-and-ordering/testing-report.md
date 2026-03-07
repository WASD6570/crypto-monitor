# Testing Report: Replay Run Manifests And Ordering

## Outcome

- Passed manifest/schema validation for replay run manifests, snapshot version guards, manifest digest validation, manifest freeze behavior, missing snapshot failure, and checksum drift failure.
- Passed deterministic ordering coverage for repeated runs, mixed sequence availability, and degraded timestamp ordering.
- Passed runtime mode safety coverage for `inspect`, `rebuild`, `compare`, and unsupported mode rejection.
- Passed focused integration coverage for manifest partition resolution and one-symbol/one-day deterministic execution.

## Commands

1. `/usr/local/go/bin/go test ./libs/go/... -run 'TestReplayRunManifestSchemaDecode|TestReplaySnapshotVersionGuard|TestReplayManifestDigestValidation' && /usr/local/go/bin/go test ./services/replay-engine/... -run 'TestReplayManifestBuilderFreezesResolvedSnapshots|TestReplayRunFailsOnMissingConfigSnapshot|TestReplayRunFailsOnManifestChecksumDrift' && /usr/local/go/bin/go test ./tests/integration -run 'TestReplayManifestUsesResolvedRawPartitionRefs'`
   - Result: passed
2. `/usr/local/go/bin/go test ./services/replay-engine/... -run 'TestReplayDeterministicDoubleRun|TestReplayStableOrderingWithMixedSequenceAvailability|TestReplayPreservesDegradedTimestampOrdering' && /usr/local/go/bin/go test ./services/replay-engine/... -run 'TestReplayInspectModeDoesNotWriteArtifacts|TestReplayRebuildModeWritesIsolatedArtifacts|TestReplayCompareModeEmitsDeterministicSummary|TestReplayRejectsUnsupportedMode' && /usr/local/go/bin/go test ./tests/integration -run 'TestReplayOneSymbolOneDayDeterministicManifestExecution'`
   - Result: passed

## Notes

- Deterministic ordering uses persisted bucket timestamp first, then available venue sequence, then canonical event ID.
- For equal timestamps with mixed sequence coverage, sequenced events sort before non-sequenced events; this keeps one stable total order without recomputing timestamp decisions.
- `compare` remains audit-only and emits a deterministic compare summary artifact ref without any publish/apply side effects.
