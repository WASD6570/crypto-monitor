# Replay Retention And Safety Validation

## Ordered Implementation Plan

1. Add integrated hot/cold retention and replay continuity smoke coverage on top of the completed raw log, replay manifest, and checkpoint slices.
2. Add side-effect safety, recovery, and late-event validation coverage that exercises inspect, rebuild, compare, resume, and gated apply paths without reopening core runtime behavior.
3. Document deterministic validation commands and evidence expectations in `plans/completed/replay-retention-and-safety-validation/testing-report.md`.

## Problem Statement

The replay foundation already has append-only raw storage, deterministic replay manifests, and bounded checkpointed backfill controls, but the epic is not done until those slices are proven together under retention movement, recovery pressure, and side-effect safety rules.

This feature closes the epic with integrated validation for hot/cold continuity, deterministic replay smoke behavior, late-event repair visibility, and no-default-side-effect recovery flows.

## Bounded Scope

- integrated replay validation across `plans/completed/raw-event-log-boundary/`, `plans/completed/replay-run-manifests-and-ordering/`, and `plans/completed/backfill-checkpoints-and-audit-trail/`
- hot-to-cold retention continuity proof for replay-visible partitions and manifests
- deterministic replay smoke coverage for one-symbol, one-day scopes using preserved snapshots and existing ordering rules
- side-effect safety assertions for `inspect`, `rebuild`, `compare`, and explicitly gated `apply` paths
- resume and failure-recovery validation that proves checkpoints, audit lineage, and apply guards still behave under integrated runs
- minimal Go-side validation hooks, harness wiring, and runbook notes needed to exercise the existing seams

## Out Of Scope

- redefining raw write envelopes, partition identity rules, replay ordering, request bounds, checkpoint identity, or audit schema semantics already completed
- new business logic for market state, alerts, outcomes, or simulation
- storage-engine-specific schemas, migrations, object-store layouts, or restore tooling
- UI/operator console work, web replay UX, or dashboard changes
- Python in the live runtime path

## Requirements

- Build directly on the completed plans for `raw-event-log-boundary`, `replay-run-manifests-and-ordering`, and `backfill-checkpoints-and-audit-trail`.
- Preserve `exchangeTs` vs `recvTs` timestamp selection semantics and degraded markers from the completed prerequisite work.
- Keep the feature storage-engine-neutral: validate logical partition references, restore seams, and manifests without baking in vendor-specific storage assumptions.
- Keep all live-safe validation hooks, smoke harnesses, and recovery assertions in Go service/shared-helper paths.
- Do not reopen replay runtime or request/control implementation except for narrow hooks needed to expose deterministic test seams and evidence.
- Prove the operating defaults from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md` remain intact:
  - raw canonical events stay replayable across 30-day hot and 365-day compressed cold retention defaults
  - one-symbol, one-day replay remains runnable within the local/dev expectation
  - late events are persisted and corrected through replay/backfill, not silent mutation
  - replay/backfill defaults stay side-effect-safe unless explicit apply context is present

## Target Repo Areas

- `services/replay-engine`
- `services/normalizer` only if a fixture or validation hook must expose raw-write evidence already persisted there
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`
- `docs/runbooks`

## Module Breakdown

### 1. Retention And Continuity Smokes

- Prove replay callers can resolve the same logical scope before and after hot-to-cold retention movement.
- Keep implementation focused on fixtures, harnesses, and manifest/restore assertions rather than new storage logic.

### 2. Side-Effect Safety And Recovery Matrix

- Prove integrated replay and backfill flows remain deterministic, resumable, auditable, and side-effect-safe across inspect, rebuild, compare, and gated apply modes.
- Limit service changes to validation seams that expose evidence, counters, or sink instrumentation.

## Design Details

### Validation Philosophy

- This slice is proof-oriented, not architecture-expanding.
- Prefer high-signal integration and replay smoke tests over broad new abstractions.
- Reuse completed contracts and fixtures first; add only the minimum extra fixture coverage needed for hot/cold continuity, late-event repair, and apply-gate assertions.

### Retention Continuity Expectations

- The same requested symbol/day scope should resolve through logical partition references, not tier-specific guesses.
- Tier movement must preserve event counts, ordering inputs, canonical event IDs, timestamp provenance, and degraded markers.
- If a cold partition requires staged restore, replay evidence should show the restore path is explicit and deterministic rather than hidden.

### Side-Effect Safety Expectations

- `inspect`, `rebuild`, and `compare` must not emit alerts, notifications, or webhook-like side effects.
- `apply` may only proceed with explicit approval/apply context and replay-aware idempotency protection already defined by the completed backfill slice.
- Recovery validation should show rejected or repeated apply attempts are auditable and deterministic.

### Live vs Research Boundary

- Go owns replay retention validation, harness wiring, and safety assertions for the live platform path.
- Python may remain optional for offline analysis later, but this feature cannot require Python to execute retention, replay, or recovery smoke coverage.

## Acceptance Criteria

- Another agent can implement this validation slice without reopening the parent epic or redesigning earlier replay/runtime modules.
- The plan names explicit repo areas for Go validation hooks, fixtures, integration coverage, and runbook evidence.
- Validation commands are deterministic, repo-local, and specific enough for another agent to run without extra interpretation.
- Out-of-scope boundaries clearly preserve the completed raw, replay-ordering, and checkpoint/audit slices.

## ASCII Flow

```text
raw hot partition manifest ------+
                                 |
raw cold partition manifest -----+--> replay scope resolver
                                        - same logical symbol/day scope
                                        - preserved config/contract/build refs
                                        - explicit restore evidence when needed
                                                 |
                                                 v
                                  deterministic replay smoke harness
                                    - inspect / rebuild / compare / apply
                                    - late-event and degraded fixtures
                                    - checkpoint resume injection
                                                 |
                    +----------------------------+-----------------------------+
                    |                            |                             |
                    v                            v                             v
         continuity assertions         audit + checkpoint assertions   sink safety assertions
         - counts and digests          - resume lineage                - no default side effects
         - ordering unchanged          - conflict handling             - gated apply only
         - event IDs preserved         - deterministic outcomes        - idempotent promotion
```
