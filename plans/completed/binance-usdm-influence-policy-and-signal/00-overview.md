# Binance USD-M Influence Policy And Signal

## Ordered Implementation Plan

1. Define one deterministic internal USD-M influence contract for `BTC-USD` and `ETH-USD`, including explicit `no-context` and degraded-context posture, without changing current-state or regime outputs yet.
2. Add the smallest Go-owned USD-M input seam needed to assemble funding, mark/index, liquidation, and open-interest context into one stable per-symbol evaluator input.
3. Implement the bounded evaluator in `services/feature-engine`, keeping the first child limited to auxiliary or degrade-cap posture rather than consumer-facing output changes.
4. Add replay and focused regression proof that identical pinned Spot plus USD-M inputs yield identical signal outputs and that `/api/market-state/*` stays unchanged in this slice.
5. Record validation evidence in `plans/binance-usdm-influence-policy-and-signal/testing-report.md`, then move the full directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to settling the internal policy and evaluator seam for Binance USD-M influence.
- Do not change `/api/market-state/global`, `/api/market-state/:symbol`, or `/healthz` in this child.
- Keep the symbol scope fixed to `BTC-USD` and `ETH-USD`.
- Keep Go as the live runtime path; Python remains offline-only.
- Reuse the already-landed USD-M context surfaces for funding, mark/index, liquidation, and open interest instead of reopening acquisition work.
- Make no-context, stale, and degraded USD-M behavior explicit and deterministic.
- Prefer the smallest backward-compatible posture first: auxiliary or bounded degrade-cap semantics only; no positive weighting or broad bullish/bearish scoring in this child.
- Keep replay determinism explicit with pinned fixtures and repeated-input validation.

## Design Notes

### Policy boundary

- Treat this child as the contract-settling slice for an internal signal, not the consumer-facing application slice.
- Keep the evaluator output internal and machine-readable so the follow-on child can apply it to current-state and regime assembly without guessing at semantics.
- Preserve current Spot-only outputs as the default external behavior until the second child lands.

### Signal posture

- Default to an explicit posture model that can represent healthy auxiliary context, bounded degrade-cap pressure, `NO_CONTEXT`, and degraded-context conditions.
- Keep the signal per-symbol and deterministic for the same accepted USD-M input bundle.
- Carry reason codes and trigger metrics so later application work can remain auditable without reverse-engineering evaluator internals.

### Input seam

- Reuse the settled USD-M sensor inventory from the completed websocket and open-interest work.
- Prefer a narrow Go-owned snapshot or evaluator-input seam that gathers the latest accepted funding, mark/index, liquidation, and open-interest context for each tracked symbol.
- Keep websocket-derived and REST-derived freshness/degradation semantics explicit instead of collapsing them into a generic derivatives health flag.

### Replay and compatibility posture

- Extend replay proof to include the new internal signal while keeping the current API response stable in this child.
- Treat unchanged current-state output as a required regression check, not an incidental side effect.
- Defer any additive provenance metadata until the follow-on application child proves it is needed.

## ASCII Flow

```text
existing Binance USD-M context inputs
  - funding
  - mark/index
  - liquidation
  - open interest
            |
            v
internal evaluator input seam
  - per-symbol latest accepted context
  - freshness / degraded posture
  - deterministic ordering
            |
            v
services/feature-engine
  USD-M influence evaluator
  - auxiliary / degrade-cap policy only
  - explicit no-context behavior
            |
            +--> replay proof
            +--> later output-application child

external surfaces remain unchanged in this child
  - GET /api/market-state/global
  - GET /api/market-state/:symbol
  - GET /healthz
```

## Archive Intent

- Keep this feature active under `plans/binance-usdm-influence-policy-and-signal/` while implementation and validation are in progress.
- When complete, move the full directory and `testing-report.md` to `plans/completed/binance-usdm-influence-policy-and-signal/`.
