# Implementation 02: Costs, Latency, And Confidence Labels

## Module Requirements And Scope

- Define conservative cost-model defaults for fees, slippage, and latency.
- Specify how leverage options, including `x5`, affect gross/net result reporting for perps.
- Create stable confidence labels that explain when assumptions remain strong, when fallbacks were used, and when simulation should be refused.
- Keep the model deterministic and versioned so replay and baseline comparison stay credible.

## Target Repo Areas

- `services/simulation-api`
- `libs/go`
- `schemas/json/simulation`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Cost-Model Requirements

- Fee presets must be mode-aware: separate default tables for spot and perps, with explicit entry and exit fee assumptions.
- MVP default is taker-on-entry and taker-on-exit for all modes unless a specific saved preset opts into maker logic later.
- Slippage must be computed from either L2 walk data or a deterministic fallback impact curve tied to venue, mode, and notional bucket.
- Latency must include at least alert-delivery-to-operator reaction delay plus execution-start delay; zero-latency runs are out of scope for default presets.
- Net result reporting must separate gross move, fees, slippage, latency impact, and final net estimate so the user can see what erased edge.

## Latency Plan

- Define a small set of pinned presets such as `fast`, `default`, and `conservative`, each with explicit milliseconds and source-of-truth labels.
- Measure latency from alert emission processing time, not signal nomination time.
- If clock health or timestamp provenance is degraded beyond safe thresholds from operating defaults, downgrade confidence or refuse if event ordering can no longer be trusted.
- Store both configured latency and observed data-timestamp health on the run so later review can distinguish market lag from simulation policy.

## Slippage Plan

- Preferred slippage method: venue-specific L2 depth walk for the configured review notional.
- Fallback slippage method: deterministic impact curve from historical depth bands or top-of-book spread buckets when L2 is partially degraded but not critically unsafe.
- Refusal condition: no trusted way to approximate impact without inventing liquidity.
- Every run must store `slippageMethod` and `slippageConfidenceReason`.

## Fee Defaults

- Spot defaults should assume standard taker spot fees with room for venue override by config version.
- Perp defaults should assume standard taker perp fees and may include funding only if deterministic funding snapshots are available for the horizon; otherwise funding is omitted and confidence degrades if omission matters materially.
- If fee data is missing for a requested venue/mode, refuse rather than silently applying another venue's economics.

## Leverage And Net Result Rules

- Leverage multiplies exposure for `perp-long` and `perp-short`, but reported output must include both raw price move and leverage-adjusted PnL estimate.
- Support `x1`, `x2`, `x3`, and `x5`; treat `x5` as allowed but higher-risk and always display more conservative confidence messaging when fills are based on fallbacks.
- If liquidation-distance logic cannot be evaluated deterministically, do not fake liquidation probabilities; instead record a guardrail field such as `liquidationRiskUnknown=true` and lower confidence.
- Spot runs always report leverage as `x1`.

## Confidence Label Framework

- Suggested labels: `HIGH_CONFIDENCE`, `NORMAL_CONFIDENCE`, `LOW_CONFIDENCE`, `REFUSED`.
- `HIGH_CONFIDENCE`: trusted venue selected, healthy timestamps, healthy L2 on entry and exit, complete fee inputs, no material substitutions.
- `NORMAL_CONFIDENCE`: small degradations allowed, such as healthy top-of-book fallback for one leg or venue substitution among trusted venues.
- `LOW_CONFIDENCE`: deterministic fallback run allowed, but one or more material assumptions reduce realism, such as missing L2, omitted funding, or degraded timestamp source still within audit tolerance.
- `REFUSED`: assumptions are too weak to produce a trustworthy result, such as critically degraded L2 with no safe fallback, missing fee table, unsupported leverage, or untrusted ordering.
- Confidence must be derived from stable reason codes, not free-form text alone.

## Baseline And Review Compatibility

- Production alerts and baseline alerts should use the same simulation presets so net-viability comparisons remain honest.
- Confidence labels and reason codes should be queryable by setup family, regime, mode, leverage, and venue so review surfaces can distinguish weak assumptions from weak signals.

## Unit-Test Expectations

- fee-calculation tests for spot vs perp defaults and refusal on missing fee table
- latency-preset tests that show net-entry time shifts and preserved timestamp provenance
- slippage tests for L2 walk, quote fallback, and refusal when both are unsafe
- confidence-label tests that cover `HIGH_CONFIDENCE`, `NORMAL_CONFIDENCE`, `LOW_CONFIDENCE`, and `REFUSED`
- leverage tests for `perp-long` and `perp-short` at `x1`, `x2`, `x3`, and `x5`
- replay tests proving identical net outputs and labels for identical fixtures

## Summary

This module defines the conservative economics of simulation: pinned latency presets, deterministic fee and slippage methods, supported leverage up to `x5`, and stable confidence labels that explain exactly when assumptions remain strong, when they weaken, and when the system must refuse the run.
