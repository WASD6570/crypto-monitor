# Implementation: Runtime Owner And Read Model

## Requirements And Scope

- Add one sustained runtime owner that satisfies `marketstateapi.SpotCurrentStateReader` through a process-owned read model instead of per-request polling.
- Limit the feature to Spot `BTC-USD` and `ETH-USD`.
- Reuse existing Binance runtime components for websocket supervision, depth bootstrap, and depth recovery.
- Keep the runtime owner explicit about startup, shutdown, readiness, and symbol ordering.

## Target Repo Areas

- `cmd/market-state-api/live_provider.go`
- `cmd/market-state-api/*.go` for new runtime owner or read-model files
- `services/venue-binance/*.go` only for the smallest helper exports needed to compose the existing runtime surfaces cleanly

## Implementation Notes

- Replace `binanceSpotSnapshotReader` with a sustained owner such as `binanceSpotRuntimeOwner` plus an internal read-model store.
- Keep provider construction in `newProviderWithOptions(...)`, but move symbol state ownership out of the per-request `Snapshot(...)` path.
- Model per-symbol runtime state explicitly: binding, latest accepted observation, last accepted timestamps, current feed-health inputs, current depth recovery status, and whether the symbol is publishable yet.
- Keep snapshot reads concurrency-safe and deterministic by returning observations in fixed symbol order.
- Add explicit lifecycle methods such as `Start(context.Context)`, `Stop(context.Context)`, and `Snapshot(context.Context, now)` so the later cutover slice can reuse the boundary without guessing hidden goroutine behavior.
- If existing venue code lacks one small orchestration helper for accepted Spot frames or recovery decisions, add the helper in `services/venue-binance` rather than duplicating protocol logic in the command.

## Unit Test Expectations

- Startup returns no observations until the owner has accepted enough runtime data to publish one.
- Accepted top-of-book or synchronized depth progression updates the correct symbol and leaves the other symbol stable.
- Snapshot reads preserve `BTC-USD`, then `ETH-USD` ordering across repeated calls.
- Snapshot reads surface the last accepted observation together with current degraded or stale status instead of silently dropping already-known symbols.
- Shutdown cancels background runtime work cleanly and prevents goroutine leaks in tests.

## Summary

This module creates the process-owned state source the command is currently missing. The provider still reads a `SpotCurrentStateReader`, but the reader becomes a sustained runtime owner with explicit lifecycle and deterministic snapshot semantics instead of a bounded REST polling helper.
