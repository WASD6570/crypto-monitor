# Testing Plan: Replay Run Manifests And Ordering

## Testing Goal

Prove that replay runs can be assembled from preserved snapshots, executed with one deterministic ordering model, and reported through versioned output contracts without drifting into backfill orchestration or live-side effects.

## Report Output

- Write the execution report to `plans/completed/replay-run-manifests-and-ordering/testing-report.md`.

## Preconditions

- Raw partition manifest continuity from `plans/completed/raw-event-log-boundary/` remains available.
- Replay fixtures cover out-of-order, duplicate, late, and timestamp-degraded event windows.
- Any config snapshot used by replay is checked into a deterministic test fixture or generated inside the test harness with stable digests.

## Smoke Matrix

### 1. Manifest Assembly And Snapshot Freeze

- Command: `go test ./libs/go/... ./services/replay-engine/... -run 'TestReplayRunManifestSchemaDecode|TestReplaySnapshotVersionGuard|TestReplayManifestBuilderFreezesResolvedSnapshots|TestReplayRunFailsOnMissingConfigSnapshot|TestReplayRunFailsOnManifestChecksumDrift'`
- Verify:
  - run manifest records raw partition refs, config snapshot refs, contract version refs, and build provenance
  - unsupported schema versions fail before runtime execution
  - missing or drifted snapshots fail with structured result status, not fallback defaults

### 2. Deterministic Ordering

- Command: `go test ./services/replay-engine/... -run 'TestReplayDeterministicDoubleRun|TestReplayStableOrderingWithMixedSequenceAvailability|TestReplayPreservesDegradedTimestampOrdering'`
- Verify:
  - two runs over identical inputs emit identical order, counters, and output digests
  - equal timestamps plus partial sequence coverage still resolve deterministically
  - persisted `exchangeTs` vs degraded `recvTs` selection is preserved, not recomputed

### 3. Runtime Mode Safety

- Command: `go test ./services/replay-engine/... -run 'TestReplayInspectModeDoesNotWriteArtifacts|TestReplayRebuildModeWritesIsolatedArtifacts|TestReplayCompareModeEmitsDeterministicSummary|TestReplayRejectsUnsupportedMode'`
- Verify:
  - `inspect` emits manifest/result metadata only
  - `rebuild` writes isolated run-scoped artifacts only
  - `compare` emits deterministic comparison metadata only
  - unsupported or future modes fail clearly without partial writes

### 4. Integration Path

- Command: `go test ./tests/integration -run 'TestReplayManifestUsesResolvedRawPartitionRefs|TestReplayOneSymbolOneDayDeterministicManifestExecution'`
- Verify:
  - replay consumes logical partition resolution from the raw manifest boundary instead of guessing storage paths
  - one-symbol, one-day replay executes from the frozen manifest and remains within expected local/dev limits under normal fixture load

## Negative Cases

- missing config snapshot ref
- contract version mismatch between manifest and decoder
- checksum drift on raw partition manifest refs
- compare request without a valid comparison target
- unsupported runtime mode request

## Side-Effect Expectations

- No live venue calls are required for any replay test.
- No external notifications, alerts, or publish/apply side effects are allowed in this slice.
- Rebuild artifacts, if produced during tests, stay namespaced to the run and are disposable.

## Recommended Execution Order

1. manifest and schema validation tests
2. deterministic ordering unit tests
3. runtime mode safety tests
4. integration replay manifest path tests

## Exit Criteria

- All planned commands pass.
- The testing report records manifest, ordering, and mode-safety outcomes with any deviations called out explicitly.
- Results prove deterministic replay identity and ordering without introducing checkpoint or apply workflow behavior.
