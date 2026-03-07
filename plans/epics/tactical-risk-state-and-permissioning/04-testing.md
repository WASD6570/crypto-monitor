# Tactical Risk State And Permissioning Testing

## Test Goals

- verify deterministic tactical state transitions and effective permission ceilings
- verify append-only persistence and structured decision logging
- verify review-required authz boundaries and operator control behavior
- verify replayability, negative cases, and `NO-OPERATE` dominance

## Expected Output Artifact

- `plans/epics/tactical-risk-state-and-permissioning/testing-report.md`

## Validation Commands

- `go test ./... -run TacticalRiskState`
- `go test ./... -run TacticalRiskPermissioning`
- `go test ./... -run TacticalRiskTransitionLog`
- `go test ./... -run TacticalRiskReplay`
- `go test ./... -run TacticalRiskAuthz`
- `npm --prefix apps/web test -- tactical-risk-review`

Implementation may rename packages, but the final test targets should preserve this split: state machine, permissioning, logging, replay, authz, and UI review smoke coverage.

## Smoke Matrix

### 1. Soft Daily Breach

- Fixture: realized loss crosses soft daily threshold exactly once while market state remains `TRADEABLE`.
- Verify: tactical state becomes `DE-RISK`, effective ceiling tightens per config, decision log records threshold values and reason code.
- Negative case: repeated evaluation with identical inputs does not append duplicate transition rows.

### 2. Hard Daily Breach

- Fixture: realized loss crosses hard daily threshold.
- Verify: tactical state becomes `STOP`, effective ceiling becomes informational-only, review-required item is created.
- Negative case: stale loss input must not clear `STOP` on a subsequent retry.

### 3. Weekly Stop Dominance

- Fixture: weekly threshold breached, then later daily loss recovers below soft threshold.
- Verify: tactical state remains `STOP` until weekly reset boundary plus approved review clearance.
- Negative case: manual restore attempt before weekly reset is rejected and logged.

### 4. `NO-OPERATE` Interaction

- Fixture: tactical state is `NORMAL` or `DE-RISK`, but symbol or global market state becomes `NO-OPERATE`.
- Verify: effective permission ceiling becomes informational-only without mutating tactical transition history solely because of `NO-OPERATE`.
- Negative case: review approval while `NO-OPERATE` remains active does not elevate the effective ceiling.

### 5. Invalid Or Missing Threshold Config

- Fixture: config snapshot missing exit threshold, has inverted hard/soft thresholds, or is absent for the evaluation window.
- Verify: evaluator fails closed to the stricter state, emits structured config error reason codes, and does not fabricate permissive ceilings.
- Negative case: client-supplied fallback thresholds are ignored.

### 6. Replay Determinism

- Fixture: same ordered input stream, config version, and prior-state seed replayed twice.
- Verify: identical transition sequence, decision-log reason codes, and effective permission ceilings across both runs.
- Negative case: out-of-order replay input is marked as such and does not silently rewrite prior live audit records.

### 7. Review Surface Authz

- Fixture: authorized and unauthorized users call review endpoints and use the review UI.
- Verify: authorized actions append audit records; unauthorized actions return rejection with no state mutation.
- Negative case: duplicate acknowledgement or approval calls remain idempotent.

## Structured Log Verification Checklist

- every evaluation has an evaluation id
- no-transition evaluations still emit decision logs
- transitions carry previous state, new state, thresholds, reason codes, config version, market-state references, and actor metadata
- manual review actions link back to the reviewed transition ids
- replay-produced artifacts include replay correlation ids

## Persistence Verification Checklist

- transition records are append-only
- duplicate retries do not create multiple equivalent transitions
- current state queries match the latest valid persisted record plus active review status
- historical queries preserve event-time ordering and correction markers

## Handoff Notes For Feature Testing

- Prefer deterministic fixtures over ad hoc time-based tests.
- Pin UTC window boundaries in fixtures for daily and weekly calculations.
- Include at least one fixture where market state is permissive but tactical state is restrictive, and one where the reverse is true.
- Record any implementation-specific command renames in `plans/epics/tactical-risk-state-and-permissioning/testing-report.md` after feature implementation.
