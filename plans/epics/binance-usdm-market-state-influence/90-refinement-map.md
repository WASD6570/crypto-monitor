# Refinement Map

## Already Done

- `plans/epics/binance-usdm-context-sensors/` settled the available USD-M context surfaces, provenance fields, and symbol policy this epic must consume
- `plans/completed/binance-spot-runtime-read-model-owner/` and `plans/completed/binance-market-state-live-reader-cutover/` already moved live current-state reads onto the sustained Go-owned Spot runtime while preserving `/api/market-state/*`
- `plans/epics/binance-runtime-health-and-operator-observability/` already owns the separate operator-visibility problem, including keeping `/healthz` process-only
- initiative and dependency docs already mark this slice as replay-sensitive and limited to `BTC-USD` and `ETH-USD`

## Remaining Work

- settle one deterministic USD-M influence policy and evaluator seam that downstream current-state and regime logic can consume
- apply that settled signal to current-state and regime outputs without breaking the existing market-state routes
- add or confirm minimal provenance exposure only if the new semantics would otherwise be operator-opaque
- expand replay and focused integration proof so semantic changes remain deterministic and backward-compatible

## Overlap And Non-Goals

- do not reopen USD-M sensor ingestion, schema-family ownership, or canonical symbol policy
- do not bundle runtime-health, `/healthz`, or broader operator-status work into this epic
- do not plan environment defaults, rollout sequencing, or long-run soak checks here
- do not require browser-side logic changes; keep `apps/web` on same-origin API consumption only

## Refinement Waves

### Wave 1

- `binance-usdm-influence-policy-and-signal`
- Why first: the repo needs one bounded deterministic policy seam before current-state, regime, or API consumers can safely commit to USD-M semantics

### Wave 2

- `binance-usdm-output-application-and-replay-proof`
- Why later: consumer-facing application and any additive provenance should follow the settled influence signal rather than inventing semantics during endpoint work

## Direct Post-Implementation Checks

- verify the same pinned Spot plus USD-M inputs produce stable current-state and regime outputs across repeated runs
- verify `BTC-USD` and `ETH-USD` responses remain backward-compatible on `/api/market-state/*`, with only small additive metadata if child planning proves it necessary
