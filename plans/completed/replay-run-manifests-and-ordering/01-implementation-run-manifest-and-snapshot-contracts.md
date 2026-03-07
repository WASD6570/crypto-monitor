# Implementation Module 1: Run Manifest And Snapshot Contracts

## Module Requirements

- Define implementation-ready replay contracts for request scope, run manifest, snapshot references, and run result metadata.
- Preserve storage-engine neutrality by referencing logical raw partitions and artifact digests rather than vendor paths or migrations.
- Make every replay run self-describing enough that a later agent can audit the exact inputs without consulting mutable runtime state.
- Keep this module bounded away from checkpoint orchestration, operator approval workflows, and publish-side state transitions.

## Target Repo Areas

- `schemas/json/replay`
- `libs/go`
- `services/replay-engine`
- `configs/*`
- `tests/replay`
- `tests/integration`

## Scope

- extend the replay schema family beyond the current seed/window/result baseline to cover run-manifest and comparison metadata
- define the snapshot categories a replay run must load and record
- define the minimal Go interfaces and structs that load, validate, and surface those references inside `services/replay-engine`
- define failure semantics for missing or drifted snapshots before runtime execution begins

## Contract Additions To Plan

### Replay Run Manifest

Add a versioned manifest contract in `schemas/json/replay` that records at minimum:

- `schemaVersion`
- `runId`
- `requestRef` or deterministic request digest
- replay scope: symbol, venue set, stream families, UTC window
- requested runtime mode: `inspect`, `rebuild`, or `compare`
- resolved partition manifest refs and continuity checksums
- config snapshot ref or digest
- contract/schema version refs used by the run
- build provenance: git SHA, release ID, feature-flag digest
- initiator metadata: actor, reason code, requested-at timestamp
- manifest status: `planned`, `running`, `completed`, `failed`

Keep `apply` reserved as a future mode seam; do not require this slice to wire operator promotion or external side-effect controls.

### Replay Run Result

Plan a result contract revision or companion schema that records:

- `runId`
- terminal status and failure category
- input counters: partitions, events, duplicates, late events, degraded timestamp events
- output artifact refs and output digest(s)
- comparison summary ref when mode is `compare`
- start and finish timestamps
- manifest digest so result verification can prove which manifest produced the outputs

### Comparison Summary Contract

Plan a separate compare artifact schema for deterministic review instead of overloading the run result:

- run identity and compared target identity
- changed artifact counts by family
- first mismatch location or digest summary
- unchanged count and drift classification
- explicit note that compare output is audit-only and does not imply promotion

## Snapshot Categories To Preserve Exactly

### Raw Input Snapshot

- logical partition refs from `raw-partition-manifest.v1.schema.json`
- continuity checksums and entry counts
- first and last canonical event IDs for the scoped partitions

### Runtime Config Snapshot

- timestamp plausibility window
- late-event watermark thresholds
- ordering precedence policy identifier
- runtime mode and artifact namespace rules
- feature flags that change deterministic replay behavior

### Contract Snapshot

- replay schema versions
- any downstream output schema versions emitted by replay
- optional family manifest digest from `schemas/json/replay/family.v1.json`

### Build Snapshot

- git SHA or release ID
- build timestamp or artifact version
- service name and module version for `services/replay-engine`

## Go Module Notes

- Put schema-decoding and version guards in `libs/go` so every replay consumer uses one contract interpretation path.
- Keep `services/replay-engine` focused on loading snapshot references, validating immutability expectations, and passing a frozen run manifest into execution.
- Prefer small boring interfaces such as `ManifestResolver`, `SnapshotLoader`, and `ResultWriter` rather than storage-specific adapters embedded in business logic.

## Failure Rules

- Fail before execution when any required snapshot ref is absent, unsupported, or digest-mismatched.
- Treat missing config snapshot data as a hard failure, not a fallback to current defaults.
- Treat schema version mismatch as a hard failure surfaced in the structured run result.
- Keep partial artifacts clearly namespaced to the failed run and never treat them as current truth.

## Unit Test Expectations

- schema validation test for the new replay run manifest and compare artifact schemas
- Go decode/version-guard test for unsupported schema versions in `libs/go`
- manifest assembly test that records partition refs, config snapshot refs, and build provenance deterministically
- missing snapshot test that fails before ordered execution starts
- digest drift test that rejects a manifest when stored snapshot metadata no longer matches the resolved source

## Validation Commands

- `go test ./libs/go/... -run 'TestReplayRunManifestSchemaDecode|TestReplaySnapshotVersionGuard|TestReplayManifestDigestValidation'`
- `go test ./services/replay-engine/... -run 'TestReplayManifestBuilderFreezesResolvedSnapshots|TestReplayRunFailsOnMissingConfigSnapshot|TestReplayRunFailsOnManifestChecksumDrift'`
- `go test ./tests/integration -run 'TestReplayManifestUsesResolvedRawPartitionRefs'`

## Summary

This module gives replay one frozen description of what a run is and what immutable inputs shaped it. The next module should assume those snapshot refs already exist and focus only on deterministic event ordering and runtime-mode behavior.
