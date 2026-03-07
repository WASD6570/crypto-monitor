# Raw Event Log Boundary Testing Report

## Commands

1. `go test ./libs/go/ingestion -run 'TestRawWriteBoundaryAppendsImmutableEntries|TestRawWriteBoundaryPersistsTimestampProvenance|TestRawWriteBoundaryRecordsDuplicateAuditFacts|TestRawWriteBoundaryRejectsContractMismatch|TestRawPartitionRoutingUsesPersistedBucketDecision|TestRawPartitionRoutingIsStableForDuplicateInputs'`
   - Result: PASS
   - Coverage notes: append-only writes, timestamp provenance, duplicate audit visibility, contract mismatch rejection, persisted bucket routing, duplicate partition stability.

2. `go test ./services/normalizer/... -run 'TestNormalizerRawWriteBoundary'`
   - Result: PASS
   - Coverage notes: `services/normalizer` writes one raw append entry on the successful normalization path and persists ingest provenance.

3. `go test ./services/replay-engine/... ./tests/integration -run 'TestRawManifestContinuityAcrossTierTransition|TestReplayPartitionLookupDoesNotGuessStoragePaths'`
   - Result: PASS
   - Coverage notes: replay lookup consumes manifest resolution directly and hot/cold manifest continuity preserves the same logical partition identity and continuity markers.

## Fixture And Scenario Notes

- Used deterministic in-test canonical events plus existing Coinbase trade fixture coverage in the normalizer boundary test.
- Covered a degraded timestamp case with `recvTs` fallback routing across a UTC day boundary.
- Covered duplicate trade identity by repeating the same venue message ID and verifying append-only duplicate audit metadata.
- Covered manifest continuity by resolving the same logical partition before and after a simulated hot-to-cold transition.

## Deviations And Assumptions

- Kept the replay-facing boundary storage-neutral by adding a minimal Go manifest resolver and JSON schema instead of introducing storage-vendor paths or a full replay runtime.
- Used a conservative default late marker threshold of `2s` for the raw boundary until a later replay/state slice defines per-window lateness policy wiring.
- Partition routing includes `streamFamily` only for families that materially change replay or retention access cost in the current code path (`order-book`, `feed-health`, and derivative sensor families); trades and top-of-book stay keyed by day, symbol, and venue.
