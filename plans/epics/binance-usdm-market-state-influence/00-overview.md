# Binance USD-M Market State Influence

## Epic Summary

Settle the bounded, deterministic role of already-landed Binance USD-M context in current-state and regime decisions without reopening the stable `/api/market-state/*` surface or moving live semantics outside Go-owned services.

## In Scope

- define one deterministic per-symbol USD-M influence signal for `BTC-USD` and `ETH-USD` from the existing funding, mark/index, liquidation, and open-interest context surfaces
- decide whether that signal remains auxiliary, acts as a bounded degrade/cap input, or applies a narrowly justified semantic modifier before any consumer-facing rollout
- apply the settled signal inside `services/feature-engine` and `services/regime-engine` while keeping the current market-state API stable by default
- add replay, fixture, and focused API proof so the same pinned Spot plus USD-M inputs always produce the same current-state and regime outputs

## Out Of Scope

- new USD-M acquisition work, new derivative venues, or private futures endpoints
- operator-health surface design; that remains in `binance-runtime-health-and-operator-observability`
- broad browser/dashboard redesign or direct browser-to-Binance logic
- environment rollout defaults or long-run soak validation

## Target Repo Areas

- `services/feature-engine`
- `services/regime-engine`
- `services/market-state-api`
- `services/venue-binance` only if a narrow additive influence seam is required
- `tests/integration`
- `tests/replay`

## Validation Shape

- targeted Go tests for deterministic USD-M influence evaluation and no-context fallback behavior
- replay fixtures and repeated-input checks proving the same pinned Spot plus USD-M inputs yield the same current-state and regime outputs
- focused API verification that `/api/market-state/global` and `/api/market-state/:symbol` stay backward-compatible unless a later child proves a small additive provenance seam is necessary

## Current Repository State

- `plans/epics/binance-usdm-context-sensors/` already settled the available USD-M sensor inventory and provenance expectations for funding, mark/index, liquidation, and open-interest inputs
- `plans/completed/binance-spot-runtime-read-model-owner/` and `plans/completed/binance-market-state-live-reader-cutover/` already keep the live market-state path on sustained Go-owned Spot runtime inputs and the existing `/api/market-state/*` contract
- `plans/epics/binance-runtime-health-and-operator-observability/` is the parallel Wave 2 operator-visibility slice and should not be bundled into this semantics epic
- the remaining gap is product and algorithm clarity: USD-M context exists, but the repo has not yet settled whether it should modify current-state and regime outputs or stay machine-visible background context only

## Major Constraints

- keep Go as the live runtime path; Python remains offline-only
- preserve `BTC-USD` and `ETH-USD` as the only tracked symbols in this epic
- preserve `/api/market-state/global` and `/api/market-state/:symbol` unless a later child proves one small additive metadata seam is required for operator honesty
- keep replay determinism explicit; any semantic change must be provable with pinned fixtures and repeated-input validation
- prefer the smallest backward-compatible USD-M influence posture first; do not assume broad positive weighting without child-level justification
