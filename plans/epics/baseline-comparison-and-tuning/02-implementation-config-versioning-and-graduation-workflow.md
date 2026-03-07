# Implementation: Config Versioning And Graduation Workflow

## Module Requirements And Scope

- Define immutable config snapshots for tuning candidates and active alert policy.
- Specify the promotion path from draft candidate to active config using replay-backed evidence.
- Keep graduation lightweight, auditable, and reversible.
- Separate optional offline candidate generation from live config authority.

## Target Repo Areas

- `configs/*`
- `services/alert-engine`
- `services/replay-engine`
- `services/outcome-engine`
- `services/simulation-api`
- `libs/go`
- `schemas/json/replay`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Snapshot Structure

- Every tuning candidate should produce a full immutable snapshot, not a partial diff applied to current state.
- Snapshot contents should include:
  - alert setup thresholds and enablement
  - market-state gating policy references
  - dedupe, cooldown, and clustering policy
  - baseline parameter bundles
  - outcome-evaluation settings referenced by the run
  - simulation cost assumptions when the report includes net viability
  - metadata: parent version, created-at, author, reason, and environment scope
- Keep snapshot naming explicit and sortable, such as timestamp plus semantic version label.

### 2. Candidate Lifecycle

- Recommended states: `draft`, `candidate`, `active`, `rejected`, `rolled-back`, and `superseded`.
- `draft`: may exist from manual edits or offline analysis, but has no promotion authority.
- `candidate`: replay-complete snapshot with required reports attached.
- `active`: the single environment-selected snapshot currently governing live alert decisions.
- `rejected`: failed promotion before activation.
- `rolled-back`: was active, later reverted due to failure thresholds.
- `superseded`: previously active but replaced by a newer snapshot without rollback semantics.

### 3. Required Evidence Before Promotion

- Candidate must pass pinned replay fixtures covering:
  - one clean trend or breakout day
  - one fragmented market day
  - one degraded-feed day
- Candidate must also pass a recent rolling window, default 14 days per tracked symbol, unless storage gaps force a shorter but explicitly declared window.
- Reports must include active-versus-candidate comparisons against the three naive baselines.
- Promotion should fail closed if any required window is missing, inconsistent, or unreproducible.

### 4. Safe Promotion Defaults

- Promote only if candidate improves or preserves all of the following on the same comparison window:
  - alert precision by regime
  - positive net viability rate when simulation data exists
  - fragmented-market false-positive rate
- Preserve means within neutral tolerance, default plus or minus 1 absolute point.
- Improvement should not rely on material alert-volume collapse alone; reports must show emitted volume alongside quality deltas.
- If a candidate helps one setup family but harms another beyond thresholds, keep the whole snapshot in `rejected` unless the implementation later supports family-scoped activation explicitly.

### 5. Approval Rules Without Heavy Governance

- Automated gates are mandatory and should be the first approver: replay determinism, report completeness, comparable-window coverage, and version-pin integrity.
- Human approval default is one person from either:
  - a repo maintainer responsible for alerting logic, or
  - the designated first-user operator consuming the alerts
- Human approval should confirm only three things: report reviewed, rollback target exists, and environment scope is correct.
- Do not require committees, tickets, or dual approval unless later production policy explicitly adds them.

### 6. Separation From Offline Research

- Python or notebook analysis may suggest candidate values or annotate a report.
- Promotion authority comes only from committed config snapshots and Go-owned replay outputs.
- If offline research generates a candidate, it must still be re-encoded as a versioned config snapshot before replay and approval.

### 7. Roll-Forward And Rollback Preparation

- Every newly active snapshot should preserve a direct pointer to the immediately previous known-good active snapshot.
- Promotion should atomically update the environment's active pointer and rollback pointer together.
- A rollback should never reconstruct old values manually; it should switch the active pointer back to the preserved prior snapshot.

## Unit And Integration Test Expectations

- unit tests for snapshot manifest validation and state transitions
- unit tests for promotion gate evaluation against threshold tolerances
- integration tests for activating one snapshot while preserving prior rollback target
- replay tests proving candidate reports remain reproducible from stored snapshot references
- negative tests for missing required windows, missing report artifacts, and mismatched version references

## Summary

This module defines the conservative config authority model: immutable snapshots, explicit lifecycle states, replay-backed evidence, and one-human lightweight approval. The next module should assume candidate and active snapshots are stable before defining reports, failure thresholds, and rollback behavior.
