# Implementation: Live Runtime Read Model

## Scope

- Target areas: `cmd/market-state-api/spot_runtime_owner.go`, `cmd/market-state-api/*_test.go`, `services/market-state-api/live_spot_provider.go`, and any minimal supporting types in `services/market-state-api`.
- Promote accepted Binance Spot trade frames from parse-only runtime handling into internal trade-flow feature observations.
- Keep public market-state handlers unchanged in this child.

## Runtime Requirements

- Reuse the existing `binanceSpotRuntimeOwner` and `spotRuntimeState`; do not add a second websocket connection, goroutine owner, or raw trade listener.
- In the `trade` branch of `handlePayload`, parse the frame, normalize the trade using existing ingestion timestamp policy and binding metadata, then record one feature observation on the matching symbol state.
- Parse numeric price and size once at the boundary, reject invalid values, and keep errors explicit enough for tests.
- Preserve source identity as `trade:<id>` through `ingestion.NormalizeTradeMessage` so duplicate accepted frames can be suppressed by the feature processor.
- Carry current connection/feed-health posture when recording observations; if depth health is unavailable during warm-up, use the supervisor/feed status conservatively rather than marking healthy by fiat.
- Snapshot reads must include trade-flow buckets in deterministic symbol order and must not require public Binance access in tests.

## Read-Model Options

Use the smallest internal contract that fits the existing provider shape:

- Add an internal `TradeFlow` section to `marketstateapi.SpotCurrentStateSnapshot`, or add a small sibling interface if that avoids forcing all existing test readers to care about trade flow.
- Keep `SpotCurrentStateObservation` for top-of-book/current-state price input; trade flow should not be smuggled into the fixed `LiquidityScore: 100` path in this child.
- `spotLiveCurrentStateSource.Bundle` may preserve and ignore trade-flow snapshots until the later indicator/API child consumes them, but tests must prove the internal data is available and stable.

## Warm-Up And Reconnect Behavior

- Accepted trades before depth synchronization may be recorded as trade-flow inputs, but they must not make price current-state publishable by themselves.
- During reconnect, retain already closed trade-flow buckets but do not mutate a closed bucket on duplicate or late input after the watermark.
- If a trade is accepted while feed health is degraded, the resulting bucket must carry degraded feed reasons so later alert-readiness work can cap trust.
- Stop/shutdown should be idempotent and must not lose already snapshotted closed buckets during repeated reads.

## Command Test Expectations

- Extend existing websocket test helpers to emit `trade` frames alongside `bookTicker` and depth frames.
- Prove a runtime owner records trade-flow buckets for `BTC-USD` and `ETH-USD` in deterministic order.
- Prove duplicate trade frames with the same Binance trade ID do not double-count.
- Prove timestamp-degraded trade frames increment fallback/degraded accounting without changing public current-state response shape.
- Prove reconnect or stale/feed-degraded posture is reflected in internal trade-flow bucket quality fields.
- Prove repeated `Snapshot` calls return equal trade-flow data for the same accepted input sequence.

## Summary For Next Agent

Wire trade-flow recording only after the feature model exists. Keep runtime ownership in `cmd/market-state-api`, keep public API output stable, and make deterministic snapshots the live-boundary proof.
