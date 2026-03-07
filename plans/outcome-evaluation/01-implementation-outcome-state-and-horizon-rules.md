# Implementation: Outcome State And Horizon Rules

## Module Requirements And Scope

- Build the deterministic state machine that decides each horizon result from trusted alert inputs and trusted post-alert market data.
- Cover ordering for `TARGET_HIT`, `INVALIDATED`, and `TIMEOUT` across 30s, 2m, and 5m horizons.
- Produce shared path metrics needed by review and later simulation: MAE, MFE, time-to-hit, time-to-invalidation, and decisive timestamp.
- Attribute each evaluated horizon to regime context without making regime logic a client responsibility.

## Target Repo Areas

- `services/outcome-engine`
- `libs/go`
- `schemas/json/outcomes`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Inputs

The evaluator should require a fully emitted alert payload with, at minimum:

- `alertId`
- `symbol`
- `setupFamily` or baseline identifier
- `direction` such as long or short
- alert-open event timestamp and source timestamp metadata
- explicit target and invalidation thresholds carried from alert generation
- evaluation price-source identifier
- config and algorithm version references

It should also require trusted post-alert market observations with stable ordering metadata and regime snapshots for the same symbol and horizon window.

## Outcome Model

Each horizon should be evaluated independently and stored as its own closed result within a shared alert outcome record.

Recommended horizon result vocabulary:

- `TARGET_HIT`
- `INVALIDATED`
- `TIMEOUT`
- `UNDECIDED_DATA_GAP` only when the required market path is not trustworthy enough to close the horizon deterministically

Recommended lifecycle states:

- `PENDING`: alert exists but no horizon has closed yet
- `PARTIAL`: at least one horizon closed, longer horizons still open
- `COMPLETE`: all configured horizons closed or were marked with a terminal gap outcome

## Ordered Evaluation Rules

### 1. Open Boundary

- Start each horizon from the first eligible ordered event at or after the alert-open timestamp.
- Do not backfill pre-alert observations into the outcome window.
- Carry whether the open timestamp came from `exchangeTs` or degraded `recvTs` so replay and audit stay explicit.

### 2. Decisive Condition Order

For each ordered event in the horizon window:

1. Update path metrics for MAE and MFE.
2. Check invalidation condition.
3. Check target condition.
4. If neither fired and the horizon has ended, mark `TIMEOUT`.

This conservative order prevents optimistic classification when one ordered observation crosses both boundaries or when the underlying data granularity cannot prove target happened first.

### 3. Tie And Same-Event Policy

- If target and invalidation are both first observed on the same event, classify the horizon as `INVALIDATED`.
- If multiple targets exist in later alert versions, outcome evaluation for this slice should use the minimum required target only; later multi-target ladder logic belongs in execution or strategy-specific planning.
- If the input alert lacks either target or invalidation, reject evaluation and record an explicit validation error instead of inventing one.

### 4. Timeout Policy

- `TIMEOUT` is terminal only after the evaluator has processed all eligible ordered observations through the horizon end.
- Timeout keeps MAE and MFE metrics for the full window.
- Timeout does not imply profitability; it only means neither decisive threshold was reached first.

## Horizon Handling

- Evaluate exactly these default horizons: `30s`, `2m`, `5m`.
- Use one shared open event and one shared ordered path, then derive independent horizon windows.
- Permit different results across horizons. Example: `30s=TIMEOUT`, `2m=TARGET_HIT`, `5m=INVALIDATED` is valid.
- Record per-horizon:
  - `horizon`
  - `openedAt`
  - `closedAt`
  - `result`
  - `decisiveLevelType`
  - `timeToHitMs` when target hit
  - `timeToInvalidationMs` when invalidated
  - `timeToCloseMs` for every terminal result

## MAE / MFE And Time Metrics

- MAE should measure worst adverse excursion from alert-open to horizon close using the alert direction.
- MFE should measure best favorable excursion over the same window.
- Store both raw price delta and normalized basis-point forms when the input price basis supports it.
- `timeToHitMs` and `timeToInvalidationMs` should be null unless the corresponding event happened.
- If a horizon closes due to data gap, keep excursion metrics only for the trusted observed subwindow and mark coverage explicitly.

## Regime Attribution

Each horizon result should carry enough regime context for review and baseline slicing:

- `openRegime`: symbol regime at alert open
- `closeRegime`: symbol regime at horizon close
- `globalRegimeCeilingAtOpen`
- `fragmentedAtOpen`
- `degradedAtOpen`
- `experiencedFragmentationDuringWindow`
- `experiencedDegradationDuringWindow`

Safe default: do not invent fine-grained causal labels such as `fakeout` or `trend failure` inside this slice. Keep attribution structural and replayable.

## Determinism Notes

- Order observations using the replay foundation rules: primary event time, then stable sequence/order fields, then stable event ID.
- Do not depend on map iteration order, wall-clock time, or current mutable config.
- Pin the exact config snapshot that defined horizons, tie policy, and evaluation source.

## Unit And Fixture Expectations

- fixture for target-first path
- fixture for invalidation-first path
- fixture for same-event tie resolved to invalidation
- fixture for timeout with non-zero MAE and MFE
- fixture for horizon disagreement across 30s, 2m, and 5m
- fixture for degraded timestamp fallback that still remains deterministic
- fixture for missing threshold inputs rejected before evaluation

## Summary

This module delivers the pure outcome state machine: ordered horizon closure, excursion metrics, and regime attribution. It should stop at objective market-path answers so later simulation can reuse the same record without inheriting execution assumptions.
