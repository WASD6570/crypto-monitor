# Market State API

- Owns Go-served read endpoints for dashboard current-state consumption.
- Serves `GET /healthz`, `GET /api/runtime-status`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Runs a live Binance Spot-backed provider in the checked-in Compose stack and from direct command-line startup.
- The dashboard consumes these routes through the same-origin `/api/market-state/*` path; it does not talk to Binance directly.
- Warm-up is explicit: symbol and global payloads may be partial or unavailable until the sustained Spot runtime publishes readable observations.
- During startup, users may briefly see `Current State Unavailable` in `apps/web` and can use `Retry current state` to re-read the same-origin API after warm-up.
- `GET /api/runtime-status` is the bounded operator-facing runtime-health surface for `BTC-USD` and `ETH-USD`; it exposes `readiness`, canonical feed-health state/reasons, connection posture, reconnect counts, depth-recovery posture, and runtime timestamps.
- Use `readiness=NOT_READY` to identify warm-up separately from `readiness=READY` with degraded or stale runtime health.
- Keeps the first live cutover Spot-driven; `usa` remains explicit rather than synthesized.
- Surfaces degradation in payloads instead of failing `/healthz`; `/healthz` remains process health only and does not encode market-data freshness.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain consumer current-state reads, not the primary operator runtime-health contract.
- Supports `MARKET_STATE_API_CONFIG_PATH`, `MARKET_STATE_API_BINANCE_BASE_URL`, `MARKET_STATE_API_BINANCE_WS_URL`, `MARKET_STATE_API_BINANCE_USDM_BASE_URL`, and `MARKET_STATE_API_BINANCE_USDM_WS_URL` for explicit runtime overrides.
- If Spot base or websocket URLs are overridden, the paired USD-M base and websocket URLs must also be set or startup fails fast.
- The checked-in Compose posture pins `MARKET_STATE_API_CONFIG_PATH=/app/configs/prod/ingestion.v1.json`; see `docs/runbooks/binance-compose-rollout.md` for rollout verification.
- The optional dev overlay runs the same service through a Go watcher with `MARKET_STATE_API_CONFIG_PATH=/workspace/configs/prod/ingestion.v1.json`; it changes restart ergonomics only, not live market wiring or route ownership.
- Preserves the service-owned current-state trust boundary so `apps/web` remains read-only and presentational.
