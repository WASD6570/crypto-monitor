# Refinement Map: Raw Storage And Replay Foundation

## Current Status

This epic is still too broad for direct `feature-planning` or `feature-implementing`.

Completed prerequisite coverage already exists in:

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`

Those completed slices already provide:

- canonical event and replay vocabulary roots under `schemas/json/...`
- deterministic fixture conventions and replay-sensitive timestamp semantics
- canonical ingestion outputs with `exchangeTs`, `recvTs`, degraded markers, and feed-health linkage
- explicit `services/normalizer` handoff boundary as the write-side source for raw persistence

Completed child feature coverage now also exists in:

- `plans/completed/raw-event-log-boundary/`
- `plans/completed/replay-run-manifests-and-ordering/`
- `plans/completed/backfill-checkpoints-and-audit-trail/`

## What This Epic Still Needs

- config, contract, and build snapshot preservation for replay
- explicit side-effect safety and apply gating for correction flows
- focused integration and replay validation for retention, determinism, and resume behavior

## What Should Not Be Re-Done

- canonical event field semantics already locked in `plans/completed/canonical-contracts-and-fixtures/`
- venue ingestion, normalization, timestamp fallback, and feed-health behavior already implemented in `plans/completed/market-ingestion-and-feed-health/`
- frontend replay UX, market-state logic, alert logic, and storage-vendor-specific schema/migration design remain out of scope here

## Refinement Waves

### Wave 1

- `raw-event-log-boundary` (completed; archived under `plans/completed/raw-event-log-boundary/`)
- Why first: replay and backfill could not be made concrete until the immutable raw write boundary and persisted provenance model were fixed.

### Wave 2

- `replay-run-manifests-and-ordering` (completed; archived under `plans/completed/replay-run-manifests-and-ordering/`)
- Why next: deterministic replay depended on the raw event log identity, partition resolution, manifest continuity, and preserved snapshots from the completed wave-1 child slice.

### Wave 3

- `backfill-checkpoints-and-audit-trail` (completed; archived under `plans/completed/backfill-checkpoints-and-audit-trail/`)
- Why later: bounded correction flows depended on the now-completed replay modes, run identity, and manifest semantics from wave 2.

### Wave 4

- `replay-retention-and-safety-validation`
- Why last: the smoke matrix should validate the integrated persistence, replay, and backfill flow now that the earlier child slices are stable.

## Notes For Future Planning

- Keep all live persistence, replay, and backfill logic in Go services or shared Go helpers.
- Keep storage-engine choices abstract at the planning stage; do not invent concrete migrations here.
- Keep child plans narrow enough that each one has a direct validation command and no hidden rollout sequence.
