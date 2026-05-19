# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Feature model and service tests | `go test ./libs/go/features ./services/feature-engine` | Prove deterministic trade-flow aggregation, duplicate suppression, timestamp fallback accounting, and service wrapper behavior | New Spot trade-flow tests pass for BTC and ETH with stable output ordering |
| Binance runtime read-model tests | `go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner.*TradeFlow|TestBinanceSpotRuntimeOwner.*Publishes|TestBinanceSpotRuntimeOwner.*Reconnect'` | Prove accepted websocket trade frames are recorded without changing current-state publishability semantics | Runtime owner tests pass for trade-flow buckets, duplicates, degraded timestamps, reconnect, and repeated snapshots |
| Binance trade integration tests | `go test ./tests/integration -run 'TestIngestionBinanceSpotTrade|TestIngestionBinanceCurrentState|TestBinanceSpotSupervisor'` | Prove supervisor/parser/normalizer handoff remains compatible with trade-flow input recording | Existing trade handoff and current-state integration tests pass with new trade-flow assertions |
| Replay determinism | `go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism|TestReplayBinanceSpotTradeFlow'` | Prove accepted raw trade inputs rebuild identical trade-flow feature output across repeated runs | Replay output or digest remains stable and duplicate/timestamp-degraded trade evidence is deterministic |
| Full Go regression | `go test ./...` | Catch cross-package regressions from internal type/interface changes | Full Go suite passes |
| Contract and fixture validation | `make contracts-validate && make fixtures-validate` | Ensure any additive schemas/fixtures remain valid and existing families are not broken | Contract manifests and fixture corpus validate successfully |
| Replay smoke | `CONTRACT_FIXTURES=1 make replay-smoke` | Prove fixture-backed replay smoke remains deterministic after trade-flow changes | Replay smoke passes |
| Whitespace validation | `git diff --check` | Ensure plan/code/archive updates are patch-clean | No whitespace errors |

## Verification Checklist

- Trade-flow feature inputs are produced only for `BTC-USD` and `ETH-USD` Binance Spot trades.
- Duplicate trade IDs/source record IDs do not double-count any aggregate metric.
- Timestamp-degraded trades remain accepted only through the existing canonical timestamp policy and are counted as degraded/fallback evidence.
- Buy/sell notional, net aggressor notional, VWAP, first/last price, and price-change bps are deterministic for the same input order and also stable when inputs are replayed in canonical order.
- Degraded feed, stale, reconnect, and warm-up states are visible in internal feature output or absence semantics.
- `/api/market-state/global`, `/api/market-state/:symbol`, `/api/runtime-status`, and `/healthz` response shapes are unchanged in this child.
- No Python, browser-side Binance access, private Binance endpoint, or fixture-backed live runtime is introduced.

## Optional Validation

- Public Binance live boundary, only if explicitly enabled in a network-capable environment: `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive`.
- Compose smoke, only where Docker is available and runtime wiring materially changes startup behavior: `make compose-smoke`.

## Reporting

- Write feature-testing results to `plans/binance-spot-trade-flow-feature-inputs/testing-report.md` while this feature remains active.
- After implementation and validation finish, archive the full directory at `plans/completed/binance-spot-trade-flow-feature-inputs/` and update `plans/STATE.md`, the Binance market-intelligence epic handoff, and the Binance integration initiative handoff.
