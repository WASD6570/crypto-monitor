# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Handler and provider route coverage | `go test ./services/market-state-api ./cmd/market-state-api` | Prove the new runtime-status route, optional provider seam, and existing routes work together | Targeted tests for `/api/runtime-status`, `/healthz`, and `/api/market-state/*` pass |
| Deterministic repeated-read proof | `go test ./services/market-state-api ./cmd/market-state-api -count=2` | Catch unstable symbol ordering or time-relative status drift in repeated accepted-input runs | Repeated runs stay green with stable runtime-status assertions |
| Race coverage for live route reads | `go test -race ./cmd/market-state-api` | Check concurrency safety for runtime snapshot reads through the new route | Command-level route and runtime snapshot tests pass under the race detector |
| Focused API regression proof | `go test ./tests/integration -run 'TestIngestionBinanceCurrentState|TestIngestionBinance.*Runtime'` | Confirm current-state behavior stays intact while runtime-status proof lands beside it | Existing Binance integration coverage stays green and any new focused status smoke passes |

## Verification Checklist

- `GET /api/runtime-status` returns only `BTC-USD` and `ETH-USD` in deterministic order.
- Warm-up remains distinguishable from degraded runtime through explicit readiness.
- Shared feed-health states and reasons remain unchanged from the existing runbooks.
- `/healthz` still reports process health only.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain backward-compatible.
- No new Binance polling or websocket ownership path is introduced for the route.

## Reporting

- Historical results now live in `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/testing-report.md`.
- The full validated feature history lives in `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/`.
