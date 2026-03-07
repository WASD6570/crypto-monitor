# Implementation Module 2: Partitioning And Manifests

## Scope

Define how raw append entries are routed into replay-safe partitions and how hot/cold manifests expose those partitions deterministically to later replay work.

## Target Repo Areas

- `services/normalizer`
- `services/replay-engine`
- `libs/go/ingestion`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/integration`

## Module Requirements

- Route each raw append entry into a deterministic partition using persisted timestamp provenance rather than mutable runtime defaults.
- Keep partition keys simple enough for one-symbol, one-day replay and retention jobs.
- Define a manifest format or equivalent index that lets replay resolve partitions without scanning storage heuristically.
- Preserve partition discoverability and event identity across hot-to-cold transitions.
- Keep storage-medium details abstract; the plan should define behavior, not product-specific layout.

## Partition Strategy To Implement

- Partition on UTC day derived from the persisted bucket timestamp decision.
- Partition within the day by canonical symbol and venue.
- Include stream family only for materially different access patterns, such as trades vs book or derivative sensors, when mixing them would raise replay or retention cost.
- Avoid deeper speculative sharding in this slice.

## Manifest Responsibilities

- Record which logical partition exists for each day/symbol/venue/(optional stream family) scope.
- Record whether the partition is hot, cold, or in transition.
- Preserve enough metadata to verify continuity:
  - logical partition key
  - storage tier
  - retention window boundaries
  - partition checksum, count, or equivalent continuity signal
  - first and last canonical event IDs or equivalent range markers
  - build/config snapshot references only when they are required to interpret the manifest, not to replay the events themselves
- Keep manifest naming stable so replay callers do not care whether the backing data lives in hot or cold storage.

## Retention And Tiering Rules

- Hot retention default remains 30 days; cold retention default remains 365 days.
- Tier transitions must not reorder records, collapse duplicates, or drop timestamp provenance.
- Cold compression is acceptable only if replay can stream entries with identical record boundaries and identity semantics.
- Manifest updates for tier transitions must be atomic from the replay caller's perspective: a partition is always discoverable in exactly one valid location or one explicit transition state.

## Replay-Facing Boundary

- `services/replay-engine` should consume manifest lookups, not infer paths from storage conventions.
- Replay partition resolution should accept explicit scope: symbol, venue, stream family when needed, and time window.
- Manifest resolution must remain deterministic for repeated requests with the same scope and retention state.

## Delivery Checklist

- partition key definition and normalization rules
- manifest record shape for hot/cold continuity
- retention-transition state model and caller expectations
- replay-engine lookup contract for logical partition resolution

## Unit Test Expectations

- `TestRawPartitionRoutingUsesPersistedBucketDecision` verifies routing follows the stored timestamp decision, including degraded fallback cases.
- `TestRawPartitionRoutingIsStableForDuplicateInputs` verifies duplicates resolve to the same logical partition.
- `TestRawManifestContinuityAcrossTierTransition` verifies hot-to-cold movement preserves logical partition identity and continuity signals.
- `TestReplayPartitionLookupDoesNotGuessStoragePaths` verifies replay consumes manifests or equivalent indexes rather than implicit path math.

## Contract / Fixture / Replay Impacts

- `schemas/json/replay` may need a manifest or partition reference contract added or extended for later replay-engine consumption.
- `tests/fixtures` should include duplicate, late, and timestamp-degraded events that route into deterministic partitions.
- `tests/integration` should assert hot/cold manifest continuity without depending on vendor-specific storage behavior.

## Acceptance Criteria

- Another agent can implement partition routing and manifests without revisiting the parent epic.
- Partition keys are explicit, minimal, and replay-oriented.
- Manifest continuity preserves replay-safe lookup across retention transitions.
- The plan stays bounded to raw partition discovery and does not pull in replay ordering or backfill checkpoints.

## Summary

This module makes the raw log discoverable. It defines the logical partition model and the manifest continuity rules that let later replay code read immutable history by scope instead of guessing where older data lives.
