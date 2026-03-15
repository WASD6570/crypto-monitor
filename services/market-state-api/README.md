# Market State API

- Owns Go-served read endpoints for dashboard current-state consumption.
- Serves `GET /healthz`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Defaults to a live Binance Spot-backed provider for local and command-line runtime use.
- The dashboard consumes these routes through the same-origin `/api/market-state/*` path; it does not talk to Binance directly.
- Warm-up is explicit: symbol and global payloads may be partial or unavailable until the sustained Spot runtime publishes readable observations.
- Local users may briefly see `Current State Unavailable` in `apps/web` and can use `Retry current state` to re-read the API after warm-up.
- Keeps the first live cutover Spot-driven; `usa` remains explicit rather than synthesized.
- Surfaces degradation in payloads instead of failing `/healthz`; `/healthz` remains process health only and does not encode market-data freshness.
- Supports `MARKET_STATE_API_CONFIG_PATH`, `MARKET_STATE_API_BINANCE_BASE_URL`, and `MARKET_STATE_API_BINANCE_WS_URL` for local runtime overrides.
- Preserves the service-owned current-state trust boundary so `apps/web` remains read-only and presentational.
