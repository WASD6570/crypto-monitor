# Binance Spot Depth Liquidity Indicators

## Ordered Implementation Plan

1. Define one internal Go-owned Spot depth-liquidity feature model for `BTC-USD` and `ETH-USD` that computes observed spread, top-level size, depth notional, imbalance, depth pressure, slippage proxy, and a bounded liquidity score from synchronized Binance Spot top-of-book plus depth state.
2. Extend the Binance Spot runtime read model to maintain compact synchronized depth levels from bootstrap snapshots and accepted deltas, evaluate depth-liquidity snapshots, and replace the fixed Binance `LiquidityScore: 100` contributor weight with the observed score while preserving the existing public response shape.
3. Add deterministic fixtures, integration coverage, and replay proof for happy depth, low-liquidity/wide-spread, sequence-gap/resync, stale/degraded, and repeated-input scenarios.
4. Run the validation matrix, write `plans/binance-spot-depth-liquidity-indicators/testing-report.md`, then move the full directory to `plans/completed/binance-spot-depth-liquidity-indicators/` after implementation and validation finish.

## Requirements

- Scope is limited to service-owned Spot depth liquidity indicators derived from public Binance Spot `bookTicker`, `depth@100ms`, and REST `/api/v3/depth` snapshot inputs already owned by the runtime.
- Keep tracked symbols fixed to `BTC-USD` and `ETH-USD`, with source symbols `BTCUSDT` and `ETHUSDT`.
- Reuse the sustained Spot runtime owner, completed depth bootstrap/recovery owners, runtime-status surface, and archived runtime-soak evidence; do not create another websocket lifecycle owner or another depth synchronization policy.
- Keep Go as the live runtime path; Python remains offline-only.
- Preserve event time versus processing time: bucket or snapshot by resolved canonical event time where available, use receive time only through the existing timestamp fallback policy, and carry fallback/degraded counts into feature output.
- Degraded feeds, stale snapshots, sequence gaps, cooldown/rate-limit blocks, reconnect loops, and warm-up states must lower or remove liquidity contribution instead of silently improving market quality.
- Replace the current fixed Binance current-state liquidity contribution with the observed internal score, but do not add public API/dashboard fields in this child.
- Public `/api/market-state/global`, `/api/market-state/:symbol`, `/api/runtime-status`, and `/healthz` response schemas must remain unchanged; existing numeric values may change where they already represent contributor weights or market quality.
- Add `schemas/json/features` only if implementation chooses to expose or persist standalone depth-liquidity feature output in this child; otherwise leave formal public indicator contracts to `binance-market-indicator-api-and-dashboard-readiness`.

## Design Notes

### Feature boundary

- The completed depth bootstrap and recovery slices already prove snapshot alignment, delta sequencing, resync, cooldown/rate-limit, snapshot freshness, and machine-readable health.
- This feature consumes those settled surfaces to derive liquidity quality; it must not reinterpret sequence-gap or recovery readiness rules.
- Current-state assembly already uses Binance Spot as a `ContributorInput`, but it hard-codes `LiquidityScore: 100`; implementation should make that score come from the depth-liquidity model.
- Public indicator surfacing and dashboard rendering are intentionally deferred until trade-flow, depth-liquidity, and USD-M derivatives indicators settle.

### Planned internal output

- Add a deterministic model in `libs/go/features`, likely named around `SpotDepthLiquidityInput` and `SpotDepthLiquiditySnapshot`.
- Required first metrics: spread bps, best bid/ask size, best bid/ask notional, total bid/ask notional across configured visible levels, minimum-side depth notional, depth imbalance ratio, depth pressure ratio, buy/sell slippage proxy bps for configured quote notionals, timestamp fallback count, feed-health state/reasons, and liquidity score.
- Keep config and algorithm versions explicit, for example `feature-engine.binance-spot-depth-liquidity.v1` and `binance-spot-depth-liquidity.v1`.
- Initial scoring should be deterministic and bounded in `[1,100]` for usable synchronized books, with severe caps for stale, sequence-gap, one-sided, invalid, or insufficient-depth books.
- The model should accept only supported Binance Spot symbols and should reject unsupported symbols, source-symbol mismatches, non-positive prices/sizes, crossed books, missing timestamps, and missing feed-health state.

### Runtime read boundary

- Extend only internal Go structs and interfaces as needed, such as `SpotCurrentStateObservation`, with a depth-liquidity score/snapshot used by current-state assembly.
- Keep the compact book cache bounded to the snapshot depth already requested by the runtime (`limit=5` today) unless implementation proves one small config-only depth limit change is required.
- Bootstrap snapshots initialize the compact book; synchronized deltas update levels deterministically; resync/reconnect resets or blocks the cache until fresh synchronized depth exists.
- `bookTicker` remains the price publication source for current-state, but the feature output must require synchronized depth posture before assigning an observed liquidity score.
- Snapshot reads must stay deterministic in `BTC-USD`, then `ETH-USD` order.

### Compatibility posture

- The public current-state schema remains byte-shape compatible; this child changes existing contributor weight/market-quality values, not field names or JSON structure.
- Existing runtime-status shape remains unchanged, with depth readiness still visible through `depthStatus` and `feedHealth`.
- Later API/dashboard readiness work can expose detailed depth-liquidity snapshots as additive fields after this internal feature is tested and archived.

## ASCII Flow

```text
Binance Spot runtime
  - bookTicker frame
  - depth@100ms frame
  - REST depth snapshot
          |
          v
services/venue-binance
  existing supervisor + bootstrap + recovery
  - synchronized snapshot/delta state
  - explicit depth health and resync posture
          |
          v
cmd/market-state-api runtime owner
  compact depth book cache
  - apply snapshot levels
  - apply accepted delta levels
  - reset on reconnect/resync/gap
          |
          v
libs/go/features + services/feature-engine
  Spot depth-liquidity evaluator
  - spread and depth notional
  - imbalance and pressure
  - slippage proxy
  - degradation-aware liquidity score
          |
          v
services/market-state-api current-state assembly
  existing public response shape
  - Binance contributor uses observed liquidity score
  - detailed indicator surfacing deferred
```

## Acceptance Criteria

- Accepted Binance Spot top-of-book and synchronized depth inputs for `BTC-USD` and `ETH-USD` produce deterministic internal depth-liquidity snapshots through Go-owned feature code.
- Binance current-state contributor liquidity no longer uses a fixed `100`; it uses the observed depth-liquidity score when depth is synchronized and usable.
- Sequence gaps, stale snapshots, resync/cooldown/rate-limit blocks, reconnect warm-up, one-sided books, and timestamp fallback are represented in feature output and cannot improve liquidity quality.
- Repeated fixture and replay runs produce identical score, metric, and ordering results.
- Existing public market-state and runtime-status response schemas remain unchanged in this child.
- The feature remains replayable from accepted raw public Binance Spot depth inputs and does not require Python, browser-side Binance access, private endpoints, or fixture-backed live runtime.

## Archive Intent

- Keep this feature active under `plans/binance-spot-depth-liquidity-indicators/` while implementation and validation are in progress.
- When implementation and feature-testing pass, move the full directory and `testing-report.md` to `plans/completed/binance-spot-depth-liquidity-indicators/`.
