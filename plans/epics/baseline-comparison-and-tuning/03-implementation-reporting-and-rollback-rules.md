# Implementation: Reporting And Rollback Rules

## Module Requirements And Scope

- Define the outputs that tuning reports must produce for operators and future review surfaces.
- Specify failure thresholds and rollback rules for active configs.
- Keep reporting deterministic, append-only, and tied to snapshot versions.
- Ensure rollback decisions stay simple enough for MVP operations.

## Target Repo Areas

- `services/replay-engine`
- `services/alert-engine`
- `services/outcome-engine`
- `services/simulation-api`
- `libs/go`
- `configs/*`
- `schemas/json/replay`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Required Reporting Outputs

- Every candidate comparison run should emit a report bundle with at least:
  - snapshot metadata and parent active version
  - pinned replay windows used
  - rolling window coverage and any gaps
  - production-versus-baseline metrics by setup family, symbol, horizon, and regime
  - alert volume, precision, false-positive counts, and positive net viability rate
  - fragmented-market slices and degraded-feed slices
  - promotion recommendation: `PROMOTE`, `HOLD`, `REJECT`, or `ROLL_BACK`
- Prefer machine-readable primary output plus a concise human-readable summary derived from the same data.

### 2. Recommended Report Artifacts

- Machine-readable artifact under a stable replay/report path, such as a versioned JSON report keyed by snapshot version and replay manifest.
- Human-readable markdown or text summary for handoff and review.
- Aggregate deltas should always include absolute counts beside percentages so low-sample improvements are obvious.
- Report bundles should be immutable and referenced from the candidate manifest.

### 3. Baseline Comparison Rules

- Always report production against each required baseline separately before any blended score.
- Minimum comparisons:
  - production setup `A` vs `naive-breakout`
  - production setup `B` vs `naive-vwap-reversion`
  - production multi-venue logic vs `single-venue-trigger`
- If a report cannot compute one required baseline comparison, promotion should default to `HOLD` or `REJECT`, not inferred success.

### 4. Failure Threshold Defaults

- Reject a candidate pre-promotion if, on the same comparable window, it:
  - loses more than 3 absolute precision points in any required setup family or regime slice
  - loses more than 5 relative percent of positive net viability where simulation coverage exists
  - raises fragmented-market false positives by more than 2 absolute points or by more than 10 relative percent, whichever is larger
  - materially reduces sample size through over-suppression without an explained quality gain
- Roll back an active snapshot if the rolling live-data window crosses the same thresholds relative to the active snapshot's promotion report or to the previous known-good active snapshot.

### 5. Rollback Rules

- Default rollback trigger sources:
  - scheduled rolling-window replay against the active snapshot
  - operator-observed degradation confirmed by the same replay command path
  - repeated deterministic failures in required report generation for the active snapshot
- Default rollback action:
  - set the previous known-good snapshot back to `active`
  - mark the failing snapshot as `rolled-back`
  - attach rollback reason and report artifact references
- Do not hot-edit the failing snapshot in place.

### 6. Approval And Audit Trail

- Promotion or rollback decisions should record who approved the action, which report artifact justified it, and which snapshot became active.
- Keep the audit trail minimal but sufficient: approver identity, timestamp, environment, old active version, new active version, and reason code.
- Operator notes may supplement a report but cannot replace the deterministic report bundle.

### 7. Query And Review Expectations

- Reports should remain queryable by snapshot version, symbol, setup family, baseline family, regime slice, and replay manifest.
- Future UI or analytics views should read stored report outputs, not recompute promotion logic client-side.
- Keep report schema boring and explicit so future slices can surface it without hidden business logic.

## Unit And Integration Test Expectations

- unit tests for recommendation classification and failure-threshold evaluation
- unit tests for rollback pointer selection and audit-log creation
- integration tests for report generation completeness across all three baselines
- replay tests for append-only report history and identical regenerated report bundles
- negative tests for missing baseline sections, low-sample windows, and stale parent-version pointers

## Summary

This module defines how tuning evidence is packaged and how operations respond when a config helps or harms performance. It keeps promotion and rollback deterministic, append-only, and lightweight enough for MVP use without hiding decisions behind informal judgment.
