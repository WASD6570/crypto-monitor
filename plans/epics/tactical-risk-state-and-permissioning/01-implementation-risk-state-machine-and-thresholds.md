# Implementation: Risk State Machine And Thresholds

## Module Requirements And Scope

- Define deterministic tactical risk transitions for `NORMAL`, `DE-RISK`, and `STOP`.
- Consume trusted loss-accounting inputs plus market-state outputs.
- Specify config-owned soft daily, hard daily, and weekly stop thresholds.
- Define effective permission ceilings without implementing delivery or execution logic.

## Target Repo Areas

- `services/*` risk or alert-control evaluator
- `libs/go`
- `configs/*`
- `schemas/json/alerts` if state or ceiling payloads are shared outward
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Inputs

- global market-state ceiling and reason set
- symbol market-state ceiling and reason set
- trusted realized-loss or outcome aggregates for daily and weekly windows
- config snapshot containing threshold values, reset windows, and permission ceilings
- prior persisted tactical state and any active review-required marker

## State Machine Rules

### Transition Priority

Evaluate in this order so replay stays deterministic:

1. validate config and input freshness
2. apply weekly stop rules
3. apply hard daily stop rules
4. apply soft daily de-risk rules
5. preserve stricter prior review-required restrictions when reset conditions are not satisfied
6. combine with market-state ceilings to compute the effective permission result

### Base Transitions

- `NORMAL -> DE-RISK` when the soft daily threshold is met or exceeded.
- `NORMAL -> STOP` when the hard daily threshold or weekly stop threshold is met or exceeded.
- `DE-RISK -> STOP` when the hard daily threshold or weekly stop threshold is met or exceeded.
- `DE-RISK -> NORMAL` only after loss inputs recover below configured hysteresis or reset rules and no review-required block remains.
- `STOP -> DE-RISK` or `STOP -> NORMAL` only after the configured reset boundary is reached and required human review has been completed.

### Hysteresis And Safe Defaults

- Use explicit hysteresis or separate exit thresholds so borderline losses do not flap state repeatedly.
- If exit thresholds are absent, default to requiring the next reset window rather than same-session automatic recovery.
- If multiple thresholds conflict, choose the stricter resulting state.
- If trusted loss input is unavailable or stale, do not silently downgrade from `STOP` or `DE-RISK`; hold the stricter prior state and log the data-quality reason.

## Threshold Handling

### Daily Limits

- Soft daily limit is the first conservative guardrail and maps to `DE-RISK`.
- Hard daily limit maps to `STOP` and should require review before any later re-escalation.
- Daily windows should be UTC and config-defined so replay can pin exact boundaries.

### Weekly Stop

- Weekly stop is a separate threshold family and must dominate daily recovery.
- Once hit, the effective tactical state remains `STOP` until the next configured weekly boundary and review clearance.
- Weekly stop should be recorded distinctly from hard daily stop so operators can understand the longer lockout reason.

### `NO-OPERATE` Interaction

- `NO-OPERATE` is an external market-state ceiling, not a tactical risk transition trigger by default.
- While global or symbol market state is `NO-OPERATE`, alert consumers receive informational-only permission regardless of tactical state.
- Tactical state history should still advance independently from loss inputs so replay can explain both dimensions separately.

## Permission Ceiling Model

- Tactical state should output a machine-readable ceiling, not just a label.
- Safe default ceiling mapping:
  - `NORMAL`: inherit full market-state ceiling
  - `DE-RISK`: allow only conservative config-approved setups/severities
  - `STOP`: informational-only
- Keep setup-level and severity ceilings config-owned and versioned; do not hardcode business policy into the client.

## Determinism And Replay Notes

- Pin evaluation to ordered inputs, UTC windows, config version, and algorithm version.
- Distinguish event time from processing time in records so late inputs are auditable.
- Replays must produce the same transition sequence and effective permission ceilings for the same fixture set.

## Unit Test Expectations

- threshold crossing at exact equality boundaries
- hysteresis or reset-window behavior around soft and hard limits
- weekly stop dominance over daily recovery
- stale or missing trusted loss inputs holding stricter state
- `NO-OPERATE` informational ceiling without mutating tactical state history
- deterministic replay of identical input streams

## Summary

This module defines the tactical risk evaluator as a deterministic, config-owned state machine that sits on top of market-state outputs. It must fail closed, preserve strict ordering, and emit an effective permission ceiling that downstream alert components can trust without inferring hidden logic.
