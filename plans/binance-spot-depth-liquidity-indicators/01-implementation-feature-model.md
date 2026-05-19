# Implementation: Feature Model And Service Wrapper

## Requirements And Scope

- Target repo areas: `libs/go/features`, `services/feature-engine`, and focused unit tests in the same packages.
- Add an internal Spot depth-liquidity evaluator that consumes already-synchronized book state, not native websocket frames directly.
- Keep the evaluator independent from `services/venue-binance` so replay and service tests can feed deterministic book inputs without a live runtime.
- Support only `BTC-USD`/`BTCUSDT` and `ETH-USD`/`ETHUSDT` in this child.
- Do not add public JSON schemas unless implementation explicitly persists or exposes standalone depth-liquidity snapshots.

## Proposed Types

- `SpotDepthLiquidityConfig`
- `SpotDepthLiquidityLevel`
- `SpotDepthLiquidityInput`
- `SpotDepthLiquiditySnapshot`
- Optional `SpotDepthLiquidityResult` if the evaluator returns validation or assignment metadata.

`SpotDepthLiquidityInput` should include:

- symbol, venue, market type, source symbol, quote currency
- sequence or source record ID for deterministic provenance
- bid and ask levels with price/size, already sorted best-to-worst by side
- best bid/ask from the same synchronized depth state or the latest trusted `bookTicker`
- exchange timestamp, receive timestamp, timestamp status
- feed-health state and degradation reasons
- depth recovery state fields needed to cap quality, especially synchronized, sequence gap, refresh due, and stale posture

`SpotDepthLiquiditySnapshot` should include:

- schema version, config version, algorithm version
- symbol, source symbol, quote currency, venue, market type
- exchange/receive/canonical timestamps and timestamp status
- sequence or source record ID
- spread bps
- best bid size, best ask size, best bid notional, best ask notional
- bid depth notional, ask depth notional, minimum-side depth notional
- depth imbalance ratio and depth pressure ratio
- buy and sell slippage proxy bps for configured quote notionals, plus flags when the visible book cannot fill the configured notional
- liquidity score and cap reasons
- feed-health state and degradation reasons

## Algorithm Notes

- Validate all prices and sizes are positive, bid prices are non-increasing, ask prices are non-decreasing, and best bid is strictly below best ask.
- Calculate `mid = (bestBid + bestAsk) / 2` and `spreadBps = (bestAsk - bestBid) / mid * 10000`.
- Calculate side notional as `sum(price * size)` across configured visible levels, using the bounded depth cache supplied by the runtime.
- Calculate `imbalanceRatio = (bidNotional - askNotional) / (bidNotional + askNotional)` when both sides have positive notional.
- Calculate `depthPressureRatio` as the signed imbalance after applying any one-sided or insufficient-depth caps.
- Calculate slippage proxy by walking visible ask levels for a configured buy quote notional and visible bid levels for a configured sell quote notional; if visible depth cannot fill the notional, mark the side unavailable and cap liquidity.
- Compute `LiquidityScore` as a deterministic bounded score from spread quality, minimum-side depth quality, imbalance quality, slippage quality, and feed/timestamp caps.
- Use `FeedHealthState` and reasons to cap or exclude scores consistently with current composite behavior: stale and sequence-gap input should not become an eligible contributor; degraded input may remain eligible with a lower score.

## Initial Config Shape

- `SchemaVersion`: `v1`
- `ConfigVersion`: `feature-engine.binance-spot-depth-liquidity.v1`
- `AlgorithmVersion`: `binance-spot-depth-liquidity.v1`
- `MarketType`: `spot`
- `Symbols`: `BTC-USD`, `ETH-USD`
- `SourceSymbols`: `BTCUSDT`, `ETHUSDT`
- `MaxLevels`: `5`
- Per-symbol target depth and slippage quote notionals should be explicit in config so tests can pin BTC/ETH behavior without hidden constants.

## Service Wrapper

- Add `WithSpotDepthLiquidityConfig` to `services/feature-engine`.
- Add a method such as `EvaluateSpotDepthLiquidity(input features.SpotDepthLiquidityInput) (features.SpotDepthLiquiditySnapshot, error)`.
- Keep the wrapper thin, mirroring the `SpotTradeFlow` service pattern.

## Unit Test Expectations

- Happy BTC and ETH books produce stable spread, notional, imbalance, slippage proxy, and score values.
- Wide spread lowers the score.
- Low or one-sided visible depth lowers or caps the score.
- Stale or sequence-gap feed health produces unusable or severely capped output and explicit reasons.
- Timestamp fallback increments/degrades the timestamp posture without losing deterministic output.
- Unsupported symbols, source-symbol mismatches, crossed books, unsorted levels, zero/negative prices or sizes, missing timestamps, and unknown feed-health states are rejected.
- Service wrapper tests prove config injection and output parity with the underlying evaluator.

## Summary For Next Agent

- Build the deterministic feature model first and keep it independent from Binance runtime internals.
- Use explicit config/version values and tests for every scoring cap before wiring current-state behavior.
