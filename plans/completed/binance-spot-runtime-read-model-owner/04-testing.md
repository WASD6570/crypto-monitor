# Testing

## Validation Matrix

### 1. Sustained owner startup and warm-up

- Goal: prove the command-owned runtime starts cleanly and returns no publishable observation until accepted Spot data exists.
- Command: `go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner|TestNewProviderWithOptions'`
- Verify:
  - provider construction succeeds with the local Binance runtime config
  - initial snapshots are empty or partial rather than silently fabricating data
  - supported symbol ordering remains deterministic

### 2. Runtime progression and degradation

- Goal: prove accepted runtime inputs update the read model and preserve explicit degraded state across reconnect or resync paths.
- Command: `go test ./cmd/market-state-api ./tests/integration -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState'`
- Verify:
  - accepted Spot progression publishes `BTC-USD` and `ETH-USD` observations with best bid and ask fields
  - reconnect, sequence-gap, or stale paths remain machine-readable in `FeedHealth` and `DepthStatus`
  - previously accepted observations remain readable while the runtime is degraded

### 3. Deterministic repeated-input proof

- Goal: prove the sustained owner produces stable observation ordering and content for identical accepted inputs.
- Command: `go test ./tests/replay -run 'TestReplayBinanceMarketStateDeterminism'`
- Verify:
  - identical accepted input sequences yield identical read-model snapshots
  - timestamps and degradation reasons remain stable for the same scripted sequence
  - no per-request polling fallback path remains in the implementation under test

### 4. Full targeted regression pass

- Goal: prove the runtime owner does not break existing current-state assembly behavior.
- Command: `go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay`
- Verify:
  - live current-state provider tests still pass without route or schema changes
  - integration and replay coverage remain green with the sustained owner in place

## Notes For Testing Agent

- Prefer deterministic fixture-backed runtime scripts over live Binance network access for this slice.
- If a new scripted runtime fixture is added, archive it under the existing Binance fixture tree and note the path in `testing-report.md`.
- Record whether the final implementation keeps background goroutines bounded on shutdown.
- Write results to `plans/binance-spot-runtime-read-model-owner/testing-report.md` while the feature is active, then move the full directory to `plans/completed/binance-spot-runtime-read-model-owner/` after implementation and validation finish.
