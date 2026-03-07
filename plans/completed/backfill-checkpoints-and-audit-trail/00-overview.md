# Backfill Checkpoints And Audit Trail

## Ordered Implementation Plan

1. Implement bounded replay/backfill request validation, checkpoint identity, resume semantics, and deterministic conflict handling in Go.
2. Implement structured audit records, apply-gate seams, and isolated correction promotion rules without introducing default external side effects.
3. Validate checkpoint resume, overlapping-request rejection, audit completeness, and apply-gate negative paths with targeted Go and integration coverage.

## Problem Statement

The replay foundation now preserves immutable raw history, run identity, and deterministic ordering, but the platform still lacks the operational safety layer that makes replay and backfill usable for real recovery work. Operators and services need one bounded, resumable, and auditable path for correction requests that can recover from failure without duplicating derived outputs or silently re-emitting external effects.

## Bounded Scope

- replay/backfill request bounds by symbol, venue, stream family, time window, mode, and reason
- durable checkpoint identity and resume validation for failed or restarted runs
- deterministic overlap/conflict handling for requests that target the same correction scope
- structured audit records for request, execution, checkpoint lineage, and outcome facts
- explicit apply-gate seams between isolated rebuild outputs and any later promoted correction path
- focused Go-side validation for resume, conflict, audit, and apply-gate behavior

## Out Of Scope

- retention smoke matrices, hot/cold continuity proof, or the final integrated replay-retention safety slice beyond naming downstream validation hooks
- storage-engine selection, concrete schemas, tables, or migrations
- UI/operator console work, dashboard controls, or web replay UX
- downstream feature, regime, alert, or outcome business logic
- final external publish workflows beyond the apply-gate seam required to keep them off by default
- any Python dependency in the live replay or backfill path

## Requirements

- Build directly on `plans/completed/replay-run-manifests-and-ordering/` and preserve its run identity, preserved snapshots, and runtime mode assumptions.
- Reuse canonical event, timestamp, and degraded-marker semantics from completed prerequisite plans.
- Keep all request validation, checkpointing, and audit behavior in Go service or shared Go helper code.
- Keep the slice storage-engine-neutral by defining logical request, checkpoint, and audit records instead of concrete database layouts.
- Preserve operating defaults from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`:
  - `exchangeTs` remains primary when plausible
  - `recvTs` remains stored and auditable
  - late events are corrected through replay/backfill, not hidden mutation
  - replay/backfill remains feasible for one symbol and one day within the local/dev expectation
- Default every replay/backfill request to inspect or isolated rebuild behavior unless an explicit apply intent and approval context are present.
- Resume must reject mismatched scope, config snapshot, contract set, or build provenance rather than continuing with drifted inputs.

## Target Repo Areas

- `services/replay-engine`
- `services/*` replay-control or operator-recovery boundary added during implementation if needed
- `libs/go`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`
- `docs/runbooks`

## Module Breakdown

### 1. Request Bounds And Checkpoints

- Own request scope validation, run/request identity linkage, checkpoint record shape, resume eligibility, and overlapping-scope conflict rules.
- Keep request handling bounded to deterministic Go runtime behavior and explicit validation failures.

### 2. Audit Records And Apply Gates

- Own structured audit families, checkpoint lineage recording, apply-approval seams, and isolated-to-promoted correction transitions.
- Keep external side effects out of scope except for the idempotent gate that later publish work must pass through.

## Acceptance Criteria

- Another agent can implement bounded replay/backfill control flow without reopening the parent epic.
- Repo areas for Go logic, shared helpers, replay contracts, config, tests, and runbooks are explicit.
- Validation commands cover checkpoint resume, overlap rejection, audit record completeness, and apply-gate denial paths.
- The plan stays bounded to request controls, checkpoints, audit facts, and apply seams while leaving final retention smoke work for `replay-retention-and-safety-validation`.

## ASCII Flow

```text
operator or service recovery request
          |
          v
bounded request validator
  - symbol / venue / stream family
  - utc start / end
  - mode + reason + initiator
  - config / contract / build refs
          |
          +----> reject unbounded or conflicting scope
          |
          v
replay run + checkpoint controller
  - stable request id
  - stable run id linkage
  - manifest position
  - last materialized event id
  - retry lineage
          |
          v
isolated replay/backfill execution
  - inspect / rebuild / compare by default
  - resume only when pinned inputs match
          |
          +----> structured audit records
          |       - request facts
          |       - execution facts
          |       - checkpoint lineage
          |       - outcome facts
          |
          +----> apply gate seam
                  - explicit approval context
                  - idempotent promotion token
                  - no default notifications/webhooks
```

## Live-Path Boundary

- This feature stops at Go-owned replay/backfill request control, checkpointing, audit records, and apply-gate seams.
- Later replay-retention validation may exercise these seams end to end, but it must not redefine request bounds, checkpoint identity, or audit semantics established here.
