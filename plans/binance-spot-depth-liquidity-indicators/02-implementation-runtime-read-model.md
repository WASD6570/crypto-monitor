# Implementation: Runtime Read Model And Current-State Wiring

## Requirements And Scope

- Target repo areas: `services/venue-binance`, `cmd/market-state-api`, `services/market-state-api`, and focused tests under those packages.
- Extend existing Spot runtime/read-model seams only; do not create another websocket supervisor, depth bootstrap owner, depth recovery owner, or provider contract.
- Preserve public response schemas for `/api/market-state/global`, `/api/market-state/:symbol`, `/api/runtime-status`, and `/healthz`.
- Replace the fixed Binance current-state liquidity contribution with observed depth-liquidity score only after synchronized depth exists.

## Venue Parsing Notes

- Extend internal parsed depth/top-of-book types to carry size/level information needed by the runtime cache.
- Keep `ingestion.OrderBookMessage` and canonical `order-book-top` output unchanged unless implementation proves a direct blocker.
- For `ParseTopOfBookEvent`, retain `BestBidSize` and `BestAskSize` in `ParsedTopOfBook` or an adjacent internal struct.
- For `ParseOrderBookSnapshotWithSourceSymbol` and `ParseOrderBookDelta`, retain bid/ask levels as price/size strings or parsed internal levels while continuing to populate the existing best-price `OrderBookMessage`.
- Preserve existing parser validation and add only the minimal validation needed for sizes/levels.

## Compact Book Cache

- Add one bounded per-symbol compact depth cache owned by `spotRuntimeState` or a small helper local to `cmd/market-state-api`.
- Bootstrap snapshots initialize the cache from snapshot levels.
- Accepted synchronized deltas update levels deterministically: non-zero size upserts a level, zero size removes it, and the cache remains sorted best-to-worst per side.
- Resync, reconnect, bootstrap failure, sequence gap, and owner reset paths must clear or mark the cache unusable until a fresh synchronized snapshot/delta path is available.
- The cache should be capped at the configured visible depth level count, matching the runtime depth snapshot limit unless implementation makes a small explicit config-only change.

## Runtime Evaluation Flow

- On bootstrap success and refresh/recovery success, update the cache from the synchronized snapshot and aligned deltas before marking depth-liquidity usable.
- On each accepted synchronized depth delta, update the cache and evaluate a new `SpotDepthLiquiditySnapshot` with current depth health.
- On `bookTicker`, continue to record best bid/ask for price publication; when synchronized depth exists, reconcile the latest top-of-book price/size with the cache before evaluating or publishing current-state observation.
- Add internal fields to `marketstateapi.SpotCurrentStateObservation` as needed, such as `LiquidityScore` and `DepthLiquidity`.
- Update `spotContributorInput` to use the observed liquidity score instead of `100`.
- Update `validateSpotObservation` and tests so live observations require a positive observed liquidity score when publishable.

## Readiness And Degradation Rules

- Warm-up with no synchronized depth must not publish a fake liquidity score.
- Synchronized depth with healthy feed should produce an eligible positive score.
- Degraded feed may produce a positive but reduced score and carry reasons.
- Stale feed, sequence-gap state, cooldown/rate-limit blocked recovery, bootstrap failure, and reconnect-not-ready must not silently improve score; if the existing composite path excludes stale/sequence-gap contributors, keep that behavior intact.
- A depth-liquidity score must not make otherwise unpublishable top-of-book/depth state appear ready.

## Current-State Compatibility

- Public JSON field names stay unchanged.
- Existing current-state contributor `rawWeight`/`finalWeight`, market-quality caps, and regime output may change because they already represent derived quality.
- Existing `/api/runtime-status` stays the operator runtime-health route; detailed depth-liquidity metrics remain internal in this child.
- If implementation needs to expose the score in runtime status for debugging, stop and convert that into an explicit additive contract decision before coding.

## Unit And Runtime Test Expectations

- Runtime owner records deterministic depth-liquidity snapshots for BTC and ETH after bootstrap plus accepted deltas.
- Repeated snapshots are stable and sorted by supported symbol order.
- A wide-spread or low-depth scenario reduces `LiquidityScore` and the current-state contributor raw weight.
- Sequence-gap/resync, stale snapshot, reconnect, cooldown, and rate-limit paths clear or cap liquidity and preserve existing readiness semantics.
- Public market-state and runtime-status response shapes remain unchanged.
- Existing runtime tests for publishability, reconnect, and deterministic repeated input continue to pass with updated expected contributor weights.

## Summary For Next Agent

- Wire the feature model through the existing runtime/read-model path only after the evaluator is tested.
- The safest implementation is one compact per-symbol depth cache plus observed score on `SpotCurrentStateObservation`; avoid public schema changes in this child.
