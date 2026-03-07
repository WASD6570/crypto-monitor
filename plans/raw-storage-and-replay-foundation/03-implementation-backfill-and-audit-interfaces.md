# Implementation Module 3: Backfill And Audit Interfaces

## Scope

Define how operators and services request replay/backfill work, how resume points behave after failure, how audit trails are recorded, and how side effects stay safe during recovery flows.

## Target Repo Areas

- `services/*` backfill or replay-control boundaries added during implementation if needed
- `libs/go`
- `schemas/json/replay`
- `configs/*`
- `tests/integration`
- `tests/replay`

## Requirements

- Backfill requests must be explicit, bounded, and auditable.
- Resume behavior must be idempotent across crashes, retries, and operator restarts.
- Replay and backfill must not re-trigger user-facing side effects by default.
- Audit trails must explain who initiated a run, why it ran, what scope it covered, and what it changed.
- Preserve live/research boundaries by keeping operational backfills in Go.

## Backfill Interface Recommendations

Backfill and replay-control requests should require explicit scoping fields rather than implicit "fix everything" behavior:

- symbol scope
- venue scope
- stream family scope
- UTC start and end boundaries
- mode: inspect, rebuild, compare, or apply
- reason code and free-form operator note
- config snapshot reference or explicit config version
- side-effect policy reference

Prefer a CLI or service-internal job interface that can be scripted and audited over ad hoc manual database edits.

## Resume And Checkpoint Rules

- Each run should emit durable checkpoints that include:
  - replay/backfill run ID
  - scoped partition or manifest position
  - last successfully materialized event identity
  - config snapshot reference
  - output mode
  - retry count and last error class
- Restarting from a checkpoint must not duplicate derived outputs.
- If inputs or config snapshot change after a failed run, resume should be rejected and a new run should be required.
- Checkpoints should be structured records, not inferred from logs.

## Audit Trail Requirements

Plan for structured audit records covering:

- request metadata:
  - initiator
  - request source
  - reason
  - approval or apply gate context when relevant
- input metadata:
  - partition manifest reference
  - config snapshot reference
  - contract version reference
  - code/build provenance
- execution metadata:
  - start/end time
  - checkpoint lineage
  - counts of processed, late, duplicate, degraded, and rejected events
  - promotion/apply decision
- outcome metadata:
  - no-op, compare-only, isolated rebuild, or promoted correction
  - links to diff artifacts or output manifests

Audit records should remain available even after raw data ages from hot to cold tiers.

## Side-Effect Safety Controls

- Default every operational interface to no external side effects.
- Separate internal materialization from external publishing.
- Require explicit apply intent for any run that can update canonical derived outputs.
- Require replay-aware idempotency keys on sinks that could emit alerts, notifications, or webhooks.
- If a correction changes historical state, record both the prior materialized output reference and the replacement reference in the audit trail.

## Backfill Interaction With Late Events

- Late events discovered after live watermarks should create replay/backfill work items or operator-visible repair candidates.
- The plan should avoid silent in-place mutation of already-consumed history.
- When a late event falls into a closed bucket, the correction path should be replay-driven and auditable.
- If the scope is too wide for immediate replay, the system should preserve a pending-repair marker rather than pretend history is already corrected.

## Negative And Risk Cases To Cover

- overlapping backfill requests for the same scope
- resume attempt with a different config snapshot than the original run
- apply-mode request without approval or side-effect policy
- missing cold-storage partition during historical replay
- late event that would alter historical derived state after alerts were already emitted
- duplicate source events across reconnect windows

## Unit Test Expectations

- Request validation tests reject unbounded or ambiguous scopes.
- Checkpoint tests confirm failed runs can resume exactly once without duplicate materialization.
- Side-effect gate tests confirm dry-run and rebuild modes never emit external notifications.
- Audit tests confirm every run leaves a structured request, execution, and outcome trail.
- Conflict tests confirm overlapping apply-mode requests are serialized or rejected deterministically.

## Contract / Fixture / Replay Impacts

- Replay contracts should include request, checkpoint, and outcome manifest families or equivalent reserved space.
- Fixtures should include late-arriving repair cases and failed-run resume cases.
- Integration coverage should test correction flow behavior without relying on live external services.

## Summary

This module makes correction work operable instead of dangerous. The key result is a bounded, restartable, and fully auditable recovery path that protects users from hidden history rewrites or duplicate downstream effects.
