# Raw Storage And Replay Foundation

## Ordered Implementation Plan

1. Define append-only raw event persistence boundaries, storage responsibilities, and retention behavior
2. Define deterministic replay runtime inputs, ordering rules, metadata snapshots, and output contracts
3. Define backfill, resume, and audit interfaces with explicit side-effect safety controls
4. Add targeted replay, retention, and backfill validation coverage
5. Confirm the feature stays inside Go live paths with Python optional and offline-only

## Problem Statement

The visibility initiative needs a way to prove that the market state shown to the user can be reproduced later, audited when feeds degrade, and corrected safely when late or missing data arrives.

This feature establishes the raw append-only event history and replay/runtime rules that sit between ingestion and later state, feature, and alert layers.

## Role In The Visibility Initiative

- Turns live ingestion from a transient stream into an auditable source of truth
- Gives later market-state and dashboard slices a deterministic replay substrate instead of best-effort recovery
- Preserves operator trust by making degraded feeds, late events, and backfills explainable after the fact
- Protects future alerting and outcome evaluation from hidden mutations in historical state

## In Scope

- append-only persistence rules for canonical raw events and related ingest metadata
- partitioning and storage-layout recommendations for symbol, venue, day, and stream family access
- retention interaction for hot and cold raw storage under the operating defaults
- replay runtime contract, preserved metadata, ordering semantics, and output expectations
- deterministic handling of late, out-of-order, duplicate, and timestamp-degraded events
- backfill and resume behavior for replay-safe correction flows
- audit trails for ingest provenance, replay provenance, and operator-triggered recovery actions
- side-effect safety boundaries so replay and backfill do not silently re-emit live mutations

## Out Of Scope

- concrete business logic for market-state scoring, alert setup, or outcomes
- final database schema definitions, migrations, or storage-vendor-specific table design
- UI replay controls or dashboard implementation
- simulation workflows beyond the raw replay foundation
- making Python part of the live runtime path

## Requirements

- Consume canonical event and replay vocabulary from `plans/completed/canonical-contracts-and-fixtures/`.
- Assume canonical events already include `exchangeTs`, `recvTs`, symbol, venue, market type, quote context, and degradation markers.
- Preserve the operating defaults from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`:
  - `exchangeTs` remains primary event time when plausible
  - `recvTs` is always stored and remains the staleness and latency source of truth
  - late events are persisted, marked, and corrected via replay/backfill rather than hidden live mutation
  - raw canonical events keep 30 days hot and 365 days compressed cold retention by default
  - a single-symbol single-day replay should remain runnable on local/dev infrastructure within 10 minutes under normal conditions
- Keep all live persistence and replay execution in Go services or shared Go helpers.
- Keep Python optional for offline inspection, parity, or research-only replay validation.

## Ordered Slice Boundaries

### 1. Raw Event Persistence

- Persist canonical raw events after normalization and before downstream feature or state derivation.
- Keep storage append-only for event records; corrections happen by replaying from preserved inputs, not by mutating source records in place.
- Store ingest provenance needed for audit and deterministic reordering.

### 2. Replay Runtime And Determinism

- Replay reads raw persisted events plus preserved config and metadata snapshots.
- Replay produces deterministic downstream outputs for the same input set, code version, and config snapshot.
- Replay must distinguish dry-run inspection from materialized rebuilds.

### 3. Backfill And Audit Interfaces

- Backfill scopes must be explicit by symbol, venue, stream family, and time window.
- Resume points must be auditable and idempotent.
- Operator actions and automated recovery flows must leave traceable audit records.

## Design Notes

### Storage Boundary

- Treat raw storage as the immutable history of canonicalized venue facts plus ingest metadata.
- Do not store user-facing derived market state in the raw layer.
- Prefer one clear write boundary: `venue adapter -> normalizer -> raw persistence`.
- Downstream consumers should rebuild derived state from raw persisted events or derived snapshots, not from adapter-local buffers.

### Append-Only Storage Recommendations

- Use append-only event records with stable source identity where available:
  - source message ID when the venue provides one
  - sequence number or update ID when applicable
  - deterministic canonical event ID when source identifiers are partial
- Record duplicate detection metadata without deleting the original ingest attempt from the audit trail.
- Keep payload bytes or canonicalized raw payload references available long enough to explain normalization and timestamp decisions.

### Partitioning Strategy

- Prefer boring physical partition keys that support replay and retention first:
  - day in UTC based on persisted event-time bucket source
  - symbol
  - venue
  - stream family or market data type when materially different access patterns exist
- Avoid speculative partition depth beyond what one-day replay and retention jobs need.
- Keep a partition manifest or equivalent index so replay can resolve hot vs cold locations without guessing.
- Partition naming should make hot/cold migration transparent to replay callers.

### Retention Interaction

- Raw canonical events remain queryable in hot storage for 30 days by default.
- Older raw events move to compressed cold storage for 365 days by default, preserving replayability.
- Retention transitions must not change canonical ordering or strip replay-critical metadata.
- If cold restore is needed, it should stage data back into a replay-readable form without changing event identity.

### Replay Inputs And Outputs

- Required replay inputs:
  - raw persisted canonical events for the scoped window
  - feed-health and degradation records needed to reproduce downstream trust decisions
  - the exact config snapshot used for bucketing, lateness windows, and downstream deterministic calculations
  - the exact contract/schema versions for touched payload families
  - code/build provenance, such as git SHA or release identifier, for audit comparison
  - replay invocation metadata: requested scope, mode, initiator, reason, and run timestamp
- Replay outputs should be versioned and auditable, not implicit log side effects. At minimum plan for:
  - rebuilt derived state or feature artifacts for the requested scope
  - replay run manifest with counts, watermark behavior, degraded-event counts, and input snapshot references
  - diff or comparison artifacts when replay is used for audit or correction

### Determinism Rules

- Same inputs plus same config snapshot plus same code/build version must produce identical event ordering and derived outputs.
- Ordering precedence should be fixed and documented:
  1. canonical event-time source (`exchangeTs` when plausible, else degraded `recvTs` fallback)
  2. stable venue sequence or source ordering field when present
  3. stable canonical event ID as final tie-breaker
- Replay must not depend on wall-clock time, nondeterministic iteration order, or live network calls.
- Any config or schema lookup needed during replay must come from preserved snapshots, not mutable current defaults.

### Late Events, Backfill, And Resume

- Late events remain persisted even when they miss live watermarks.
- Live services may mark already-emitted outputs as needing replay correction, but must not silently rewrite operator-visible history.
- Backfill jobs should resume from durable checkpoints that include partition position, last stable event ID, and the config snapshot reference in use.
- Resume behavior must be idempotent: rerunning a failed backfill should not duplicate derived outputs or corrupt audit history.

### Side-Effect Safety

- Replay and backfill must default to dry-run or isolated materialization unless a caller explicitly selects a publish/apply mode.
- Replay should never re-send alerts, webhooks, or user notifications by default.
- Materialized correction flows should write to scoped rebuild outputs first, then promote atomically or via explicit operator action.
- Any downstream sink that can cause external side effects must require replay-aware idempotency keys or an apply gate.

### Live vs Research Boundary

- Go owns raw persistence, replay orchestration, and live-safe backfill paths.
- Python may read the same fixtures or replay manifests offline for analysis, but the live platform cannot depend on Python to persist, restore, or replay canonical data.

## Target Repo Areas

- `services/normalizer`
- `services/*` raw-storage or replay service boundaries added during implementation if needed
- `libs/go`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## ASCII Flow

```text
exchange ws/rest feeds
        |
        v
venue adapter -> normalizer
        |
        v
append-only raw persistence
  - canonical event record
  - recv/exchange timestamp decision
  - ingest provenance
  - feed-health / degraded markers
        |
        +----> hot storage (30d)
        |
        +----> compressed cold storage (365d)
        |
        v
replay selector
  - scope: symbol / venue / stream / time window
  - config snapshot
  - contract versions
  - code/build provenance
        |
        v
deterministic replay runtime
  - stable ordering
  - late-event handling
  - dry-run or apply mode
        |
        +----> rebuilt features/state artifacts
        +----> replay manifest + audit trail
        +----> correction diff / publish gate
```
