# Testing: Binance Live Current State Query Assembly

## Validation Matrix

| Case | Flow | Expected Result |
|---|---|---|
| Live source seam | `services/market-state-api` live source constructor -> provider methods | source dependencies validate cleanly and unsupported-symbol behavior stays unchanged |
| Healthy Spot assembly | pinned Binance Spot top-of-book plus healthy depth posture -> symbol/global query assembly | `BTC-USD` and `ETH-USD` responses preserve the existing contract and produce stable `world` plus explicit `usa` unavailability |
| Degraded Spot posture | timestamp fallback, stale feed-health, or unsynchronized depth input -> query assembly | current-state sections degrade explicitly without changing response shape |
| Deterministic repeated run | same accepted Binance Spot inputs -> repeated symbol/global query assembly | symbol and global responses remain identical across runs |
| Consumer seam preservation | assembled current-state output -> existing history/audit and bucket references | reserved history seam and bucket refs remain populated for consumers |

## Commands

### Market State API Boundary

- `"/usr/local/go/bin/go" test ./services/market-state-api -run 'Test(Deterministic|Live|Handler|CurrentState).*' -v`

Expected coverage:
- live source seam construction
- symbol/global response assembly behavior
- unsupported symbol and provider failure paths

### Integration Proof

- `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(MarketStateCurrent|IngestionBinance.*CurrentState).*' -v`

Expected coverage:
- Binance-backed current-state symbol/global responses
- contract stability and degradation visibility
- reserved history seam and bucket reference preservation

### Determinism Proof

- `"/usr/local/go/bin/go" test ./tests/replay -run 'Test(Replay.*MarketState|MarketStateCurrentReplay).*' -v`

Expected coverage:
- repeated-run determinism for pinned current-state inputs
- stable config and algorithm version propagation

## Inputs / Env

- Default validation should require no secrets.
- This feature should not require live Binance websocket or REST access.
- Use pinned Binance fixtures or package-local deterministic helpers for accepted inputs.

## Verification Checklist

- `services/market-state-api` no longer requires package-local deterministic bundle building to define the future live source seam
- symbol and global current-state outputs preserve the existing contract shape for `BTC-USD` and `ETH-USD`
- `usa` remains explicit as unavailable or partial when no live USA contributor exists
- feed-health, timestamp, and depth-recovery degradation remain machine-readable in assembled output
- repeated runs over the same accepted inputs produce identical outputs

## Testing Report Output

- While active: `plans/binance-live-current-state-query-assembly/testing-report.md`
- After implementation and validation complete: move the full directory to `plans/completed/binance-live-current-state-query-assembly/`
