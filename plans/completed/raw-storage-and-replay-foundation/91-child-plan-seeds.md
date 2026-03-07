# Child Plan Seeds: Raw Storage And Replay Foundation

## `raw-event-log-boundary` (completed)

- Outcome: canonical events persist append-only after `services/normalizer` with stable identity, provenance, partition routing, and replay-safe hot/cold manifest rules.
- Primary repo areas: `services/normalizer`, `services/*` raw-storage boundary, `libs/go`, `schemas/json/replay`, `configs/*`, `tests/integration`
- Depends on: `plans/completed/canonical-contracts-and-fixtures/`, `plans/completed/market-ingestion-and-feed-health/`
- Validation shape: targeted Go tests for append-only writes, partition routing, timestamp provenance persistence, and retention-manifest continuity
- Why it stands alone: it locks the immutable write boundary that every later replay and correction path depends on.
- Archive: `plans/completed/raw-event-log-boundary/`

## `replay-run-manifests-and-ordering` (completed)

- Outcome: replay runs load preserved raw partitions plus config/contract/build snapshots, apply one deterministic ordering model, and emit inspect/rebuild/compare manifests without live dependencies.
- Primary repo areas: `services/*` replay boundary, `libs/go`, `schemas/json/replay`, `tests/replay`, `tests/integration`
- Depends on: `plans/completed/raw-event-log-boundary/`
- Validation shape: double-run determinism tests, ordering tests for equal timestamps and mixed sequence availability, and missing-snapshot failure tests
- Why it stands alone: it turns immutable raw history into an auditable replay runtime without yet taking on operator recovery orchestration.
- Archive: `plans/completed/replay-run-manifests-and-ordering/`

## `backfill-checkpoints-and-audit-trail` (completed)

- Outcome: replay/backfill requests are bounded, resumable, idempotent, and fully auditable, with explicit apply gates and no default external side effects.
- Primary repo areas: `services/*` replay-control or backfill boundary, `libs/go`, `schemas/json/replay`, `tests/integration`, `docs/runbooks`
- Depends on: `plans/completed/replay-run-manifests-and-ordering/`
- Validation shape: checkpoint resume tests, overlapping-request conflict tests, audit-record assertions, and apply-gate negative tests
- Why it stands alone: operational recovery rules are a separate risk surface from raw persistence and deterministic readback.
- Archive: `plans/completed/backfill-checkpoints-and-audit-trail/`

## `replay-retention-and-safety-validation`

- Outcome: the integrated raw-storage/replay/backfill slice has a high-signal smoke matrix for hot-to-cold continuity, late-event repair behavior, and side-effect safety.
- Primary repo areas: `tests/integration`, `tests/replay`, `configs/*`, `docs/runbooks`
- Depends on: `plans/completed/raw-event-log-boundary/`, `plans/completed/replay-run-manifests-and-ordering/`, `plans/completed/backfill-checkpoints-and-audit-trail/`
- Validation shape: deterministic replay smoke, retention continuity smoke, backfill resume smoke, and inspect/rebuild/apply side-effect assertions
- Why it stands alone: it closes the epic with integrated proof rather than hiding cross-slice verification inside earlier implementation plans.
