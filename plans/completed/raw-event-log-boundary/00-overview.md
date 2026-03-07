# Raw Event Log Boundary

## Ordered Implementation Plan

1. Lock the append-only raw write boundary immediately after `services/normalizer`.
2. Define partition routing, hot/cold manifest continuity, and replay-facing partition resolution.
3. Add deterministic append-only, routing, and manifest continuity validation.

## Problem Statement

The visibility foundation already defines canonical event semantics and the live normalization handoff, but it still lacks an immutable raw event boundary that makes later replay and backfill trustworthy.

This child feature establishes the smallest missing capability: persist canonical events after normalization with stable identity, timestamp provenance, and replay-safe partition discovery.

## Bounded Scope

- append-only raw event write behavior after `services/normalizer`
- raw event identity precedence and duplicate audit handling
- persisted timestamp provenance for `exchangeTs` vs `recvTs` bucket decisions
- partition routing by UTC day, symbol, venue, and stream family when needed
- hot/cold manifest continuity so replay can resolve partitions without guessing
- deterministic validation for writes, routing, and retention-manifest continuity

## Out Of Scope

- storage-engine product selection, concrete tables, or migrations
- replay runtime execution, replay ordering, or replay apply/publish controls
- backfill checkpoint orchestration and operator audit workflows beyond raw-write audit facts
- downstream feature, regime, alert, or UI logic
- any Python dependency in the live path

## Requirements

- Build directly on `plans/completed/canonical-contracts-and-fixtures/` and `plans/completed/market-ingestion-and-feed-health/`.
- Treat `services/normalizer` as the only live write-side source for raw persistence.
- Preserve canonical event semantics already established for symbol, venue, market type, `exchangeTs`, `recvTs`, and degraded markers.
- Persist late and duplicate events as raw audit-visible facts; do not silently mutate or drop history.
- Keep the write boundary append-only. Corrections happen later through replay or rebuild flows, not in-place updates.
- Keep all live persistence logic in Go service or shared Go helper code.
- Keep the default retention assumptions visible in the design:
  - 30 days hot raw retention
  - 365 days compressed cold raw retention
- Make one-symbol, one-day replay partition lookup cheap enough for the later replay slice to stay within the program's local/dev expectation.

## Target Repo Areas

- `services/normalizer`
- `services/replay-engine`
- `libs/go/ingestion`
- `libs/go/contracts`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/integration`

## Module Breakdown

### 1. Raw Write Boundary

- Add a Go-owned raw log writer boundary behind `services/normalizer`.
- Define the persisted record envelope, append-only guarantees, and duplicate audit handling.
- Persist timestamp provenance and ingest provenance at write time so replay never infers them later.

### 2. Partitioning And Manifests

- Define deterministic partition keys and manifest records for hot/cold lookup.
- Keep partition naming and manifest fields storage-engine-neutral.
- Require hot/cold transitions to preserve raw event identity, ordering inputs, and replay discoverability.

## Design Details

### Raw Record Shape To Preserve

Every raw append entry should be planned as one immutable canonical record plus write metadata:

- canonical event payload or stable payload reference
- canonical event ID
- source identity fields used for duplicate detection
- `exchangeTs`, `recvTs`, and the selected bucket timestamp source
- timestamp degradation reason when fallback occurred
- venue, symbol, market type, and stream family
- ingest service identity, connection/session reference, and normalization build/version reference
- audit flags for duplicate, late, and degraded-feed-linked events

### Identity And Duplicate Policy

- Identity precedence for planning and tests:
  1. venue message ID
  2. venue sequence or update ID plus stream key
  3. deterministic canonical event ID
- Duplicate inputs stay visible as audit facts or counters.
- The raw log is the immutable ingest history, not the deduplicated downstream materialization.

### Partition And Manifest Policy

- Primary partition dimensions:
  - UTC day from the persisted bucket timestamp decision
  - canonical symbol
  - venue
- Add stream family only when the data family materially changes retention or replay access cost.
- Maintain a replay-readable manifest that maps requested scope to concrete hot or cold partition locations.
- Manifest continuity must survive hot-to-cold transitions without changing partition identity from the caller's perspective.

### Live vs Research Boundary

- Live writes, partition resolution, and retention-manifest logic stay in Go.
- Python may inspect manifests offline later, but this feature cannot require Python to write, resolve, or restore raw data.

## Acceptance Criteria

- Another agent can implement the raw writer without revisiting the parent epic.
- The plan names the concrete repo areas for write-boundary, manifest, config, and test work.
- Validation commands cover append-only writes, partition routing, timestamp provenance persistence, and hot/cold manifest continuity.
- The plan remains bounded to the raw event log boundary and does not absorb replay runtime or backfill orchestration.

## ASCII Flow

```text
venue adapters
    |
    v
services/normalizer
    |
    v
raw write boundary
  - canonical event payload
  - canonical event identity
  - exchangeTs / recvTs provenance
  - duplicate / late / degraded markers
    |
    +----> partition router
    |        - utc day
    |        - symbol
    |        - venue
    |        - optional stream family
    |
    +----> hot partition manifest (30d)
    |
    +----> cold partition manifest (365d)
    |
    v
later replay-engine lookup
  - resolve scope to partitions
  - read preserved raw inputs
```
