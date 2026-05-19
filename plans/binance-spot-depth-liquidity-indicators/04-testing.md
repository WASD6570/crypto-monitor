# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Feature model and service tests | `go test ./libs/go/features ./services/feature-engine -run 'TestSpotDepthLiquidity|TestService.*SpotDepthLiquidity'` | Prove deterministic spread, notional, imbalance, slippage, scoring caps, validation, and service wrapper behavior | New Spot depth-liquidity tests pass for BTC and ETH with pinned score/metric output |
| Venue parser and depth runtime tests | `go test ./services/venue-binance -run 'TestParse(TopOfBook|OrderBook)|TestSpotDepth'` | Prove internal size/level parsing and existing bootstrap/recovery behavior remain compatible | Venue tests pass for level parsing, bootstrap, resync, stale, cooldown, and rate-limit paths |
| Binance runtime read-model tests | `go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner.*(DepthLiquidity|Publishes|Reconnect|Deterministic|RuntimeStatus)'` | Prove accepted websocket/snapshot depth inputs produce internal liquidity snapshots without changing publishability semantics | Runtime tests pass for happy BTC/ETH, wide spread/low depth, repeated snapshots, reconnect, and degraded paths |
| Market-state API compatibility | `go test ./services/market-state-api -run 'Test(LiveSpot|Current|RuntimeStatus|Handler)'` | Prove public response shapes remain stable while Binance contributor weight uses observed liquidity | API/provider tests pass with updated expected weights and unchanged runtime-status schema |
| Binance depth integration tests | `go test ./tests/integration -run 'TestIngestionBinanceSpotDepth|TestIngestionBinanceCurrentState|TestBinanceSpotSupervisor'` | Prove supervisor/parser/normalizer/depth-liquidity handoff remains compatible | Existing depth recovery tests and new liquidity assertions pass |
| Replay determinism | `go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism|TestReplayBinanceSpotDepthLiquidity'` | Prove accepted raw depth/top-of-book inputs rebuild identical liquidity output across repeated runs | Replay output or digest remains stable, including degraded and timestamp-fallback cases |
| Full Go regression | `go test ./...` | Catch cross-package regressions from internal type/interface and current-state value changes | Full Go suite passes |
| Contract and fixture validation | `make contracts-validate && make fixtures-validate` | Ensure existing schemas/manifests and fixture corpus remain valid | Contract manifests and fixture corpus validate successfully |
| Replay smoke | `CONTRACT_FIXTURES=1 make replay-smoke` | Prove fixture-backed replay smoke remains deterministic after liquidity changes | Replay smoke passes |
| Whitespace validation | `git diff --check` | Ensure plan/code/archive updates are patch-clean | No whitespace errors |

## Verification Checklist

- Depth-liquidity feature output is produced only for `BTC-USD` and `ETH-USD` Binance Spot inputs.
- The fixed Binance `LiquidityScore: 100` current-state contribution is replaced by an observed score derived from synchronized depth/top-of-book state.
- Spread bps, top-level sizes/notional, total side notional, imbalance, depth pressure, slippage proxy, and score are deterministic for the same input order.
- One-sided, low-depth, wide-spread, stale, sequence-gap, reconnect, cooldown, and rate-limit scenarios lower, cap, or exclude liquidity quality instead of improving it.
- Timestamp-degraded inputs remain accepted only through the existing canonical timestamp policy and carry fallback/degraded evidence.
- `/api/market-state/global`, `/api/market-state/:symbol`, `/api/runtime-status`, and `/healthz` response shapes remain unchanged in this child.
- No Python, browser-side Binance access, private Binance endpoint, or fixture-backed live runtime is introduced.

## Optional Validation

- Public Binance live boundary, only if explicitly enabled in a network-capable environment: `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive`.
- Compose smoke, only where Docker is available and runtime wiring materially changes startup behavior: `make compose-smoke`.

## Reporting

- Write feature-testing results to `plans/binance-spot-depth-liquidity-indicators/testing-report.md` while this feature remains active.
- After implementation and validation finish, archive the full directory at `plans/completed/binance-spot-depth-liquidity-indicators/` and update `plans/STATE.md`, the Binance market-intelligence epic handoff, and the Binance integration initiative handoff.
