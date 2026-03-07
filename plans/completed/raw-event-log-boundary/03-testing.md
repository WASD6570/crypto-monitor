# Testing Plan: Raw Event Log Boundary

## Goals

- prove append-only raw writes after normalization
- prove persisted timestamp provenance drives deterministic routing
- prove duplicate and late events remain audit-visible
- prove hot/cold manifest continuity stays replay-safe

## Expected Report Artifact

- `plans/completed/raw-event-log-boundary/testing-report.md`

## Recommended Validation Commands

```bash
go test ./libs/go/ingestion -run 'TestRawWriteBoundaryAppendsImmutableEntries|TestRawWriteBoundaryPersistsTimestampProvenance|TestRawPartitionRoutingUsesPersistedBucketDecision'
go test ./services/normalizer/... -run 'TestNormalizerRawWriteBoundary'
go test ./services/replay-engine/... ./tests/integration -run 'TestRawManifestContinuityAcrossTierTransition|TestReplayPartitionLookupDoesNotGuessStoragePaths|TestRawWriteBoundaryRecordsDuplicateAuditFacts'
```

## Smoke Matrix

### 1. Happy-Path Append

- Input: one normalized canonical event with plausible `exchangeTs`
- Verify:
  - one raw append entry is written
  - `exchangeTs` and `recvTs` are both persisted
  - selected bucket timestamp source is `exchangeTs`
  - partition key resolves to the expected UTC day, symbol, and venue

### 2. Timestamp-Degraded Append

- Input: one normalized canonical event with invalid or missing `exchangeTs`
- Verify:
  - raw append still succeeds
  - persisted bucket timestamp source falls back to `recvTs`
  - timestamp degradation reason is stored explicitly
  - routing uses the persisted fallback decision, not a recomputed choice

### 3. Duplicate Arrival Audit Visibility

- Input: the same source event arrives twice with the same identity precedence fields
- Verify:
  - both ingest attempts remain audit-visible under append-only rules or duplicate linkage metadata
  - duplicate handling does not mutate the first append entry
  - both events resolve to the same logical partition

### 4. Late Event Persistence

- Input: an event that arrives after the live watermark defined by operating defaults
- Verify:
  - raw append succeeds
  - late-event marker is persisted
  - no downstream correction or replay apply behavior is triggered by this feature

### 5. Hot-To-Cold Manifest Continuity

- Input: one logical partition moved from hot to cold retention in a test harness
- Verify:
  - manifest lookup still resolves the same logical partition
  - event count and continuity markers match before and after transition
  - replay-facing lookup does not need storage-path guessing

## Determinism And Safety Checks

- Repeat each routing test twice with the same fixtures and confirm identical partition keys.
- Repeat manifest lookup after simulated tier transition and confirm identical logical partition resolution.
- Confirm no test depends on wall-clock time or live network calls.
- Confirm all live-path coverage remains in Go.

## Fixture Guidance

- Reuse canonical fixture conventions already established in `tests/fixtures/events/...`.
- Add or extend fixtures for:
  - duplicate source event
  - timestamp-degraded event
  - late out-of-order event
  - one partition continuity case spanning hot and cold manifests

## Exit Criteria

- All recommended validation commands pass.
- The generated testing report at `plans/completed/raw-event-log-boundary/testing-report.md` records commands, fixture inputs, pass/fail results, and any deviations.
- Another agent can run the same commands and reach the same partition and manifest outcomes.
