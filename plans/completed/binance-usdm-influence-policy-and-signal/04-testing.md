# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Signal contract and evaluator tests | `go test ./libs/go/features ./services/feature-engine ./services/venue-binance` | Prove the internal signal contract, input seam, and bounded evaluator logic work in Go-owned services | Targeted tests for no-context, degraded-context, and deterministic policy behavior pass |
| Deterministic replay proof | `go test ./tests/replay -run 'TestReplayBinance.*Determinism|TestReplayBinanceMarketStateDeterminism'` | Prove repeated pinned Spot plus USD-M inputs yield identical signal outputs and no replay drift | Replay tests stay green across repeated runs |
| Focused integration regression | `go test ./tests/integration -run 'TestIngestionBinanceCurrentState|TestIngestionBinanceUSDM'` | Confirm USD-M signal work does not change current API behavior in this child while preserving mixed-surface USD-M evidence | Existing Binance current-state and USD-M integration coverage stays green and any new regression assertions pass |
| Repeated-run stability | `go test ./libs/go/features ./services/feature-engine ./services/venue-binance ./tests/replay -count=2` | Catch unstable reason ordering or nondeterministic evaluator output | Repeated runs stay green with stable signal assertions |

## Verification Checklist

- The internal signal is emitted only for `BTC-USD` and `ETH-USD`.
- No-context, degraded-context, and fresh-context policy states remain explicit and deterministic.
- The first child does not add positive weighting or direct regime/current-state mutation.
- `/api/market-state/global` and `/api/market-state/:symbol` remain unchanged in this slice.
- Replay proof shows the same pinned Spot plus USD-M inputs always yield the same signal output.

## Reporting

- Write results to `plans/binance-usdm-influence-policy-and-signal/testing-report.md` while the feature remains active.
- When implementation and validation finish, move the full directory to `plans/completed/binance-usdm-influence-policy-and-signal/`.
