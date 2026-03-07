# Implementation: Net Viability And Cost Model Boundaries

## Module Requirements And Scope

- Add conservative net-viability fields to outcome records so review can distinguish raw directional correctness from edge after simple costs.
- Keep this slice explicitly below simulated execution: no fill simulation, leverage logic, venue routing, or position management.
- Define clean seams so `simulated-execution` can replace or augment these coarse fields later without changing pure outcome logic.

## Target Repo Areas

- `services/outcome-engine`
- `libs/go`
- `schemas/json/outcomes`
- `configs/*`
- `tests/fixtures`
- `tests/integration`

## Boundary Rules

### Pure Outcome Evaluation Must Own

- gross favorable and adverse movement from the trusted price path
- decisive outcome by horizon
- coarse time-based measures such as time-to-hit and time-to-close
- conservative net-viability fields computed from pinned default costs

### Simulated Execution Must Own Later

- entry and exit fill selection
- latency modeled against actual book or trade path
- fee tier, maker/taker assumptions, funding, borrow, or leverage rules
- slippage walk logic and venue-specific execution constraints
- realized trade PnL semantics

## Net Viability Fields

Each horizon result should carry a minimal viability block such as:

- `grossMoveBps`
- `grossMovePrice`
- `assumedEntryDelayMs`
- `assumedExitDelayMs`
- `assumedFeeBps`
- `assumedSlippageBps`
- `assumedRoundTripCostBps`
- `netMoveBps`
- `netViabilityStatus` with `POSITIVE`, `NEGATIVE`, or `UNKNOWN`
- `costModelVersion`
- `costCoverage` describing whether the value is coarse-only or simulation-backed

Safe default: for this slice, `costCoverage=COARSE_EVALUATION`.

## Safe Default Cost Policy

- Keep cost assumptions in versioned config, not code constants.
- Use one conservative default round-trip fee and one conservative default slippage assumption per instrument family.
- Represent latency as delayed entry and delayed decisive exit timestamps when the required price observations exist; otherwise use the next trusted observation after the configured delay.
- If the delayed observation is missing or a horizon closes under a trusted data gap, set `netViabilityStatus=UNKNOWN`.
- Never claim a profitable simulated trade from a coarse field alone; only claim that the alert remained positive or negative under the pinned conservative assumptions.

## Relationship To Baselines

- Baseline alerts should use the same viability field names and cost-model versioning as production alerts.
- Comparisons against `naive-breakout`, `naive-vwap-reversion`, and `single-venue-trigger` should use the same horizon, symbol, and regime slices.
- Store baseline identity separately from `setupFamily` so production-vs-baseline comparisons do not rely on name parsing.

## Replay, Config, And Versioning

- Pin `costModelVersion` alongside `configVersion` and `algorithmVersion` for every evaluated horizon.
- Replays must use the preserved cost-model snapshot that was active for the original run unless the replay is explicitly a counterfactual comparison.
- Counterfactual cost comparisons should append new evaluation artifacts rather than mutate historical viability fields.

## Negative Boundaries To Preserve

- Do not infer fills from candles or path extrema and call them execution.
- Do not add portfolio sizing, Kelly logic, risk budgeting, or stop-moving rules.
- Do not assume venue-specific fee tiers unless those are part of a later execution-specific config.
- Do not depend on Python research code to compute live viability fields.

## Unit And Fixture Expectations

- positive gross outcome that becomes `NEGATIVE` after conservative costs
- positive gross outcome that remains `POSITIVE` after conservative costs
- timeout case with non-zero MFE but `NEGATIVE` net viability
- missing delayed observation leading to `UNKNOWN` viability
- replay with pinned old cost model matching original viability fields
- baseline and production alerts producing directly comparable viability blocks

## Summary

This module adds an intentionally limited cost boundary to the outcome record. It helps Initiative 2 answer whether attention was still economically plausible, while leaving full trade realism for `simulated-execution`.
