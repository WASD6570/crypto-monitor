# Market State API

- Owns Go-served read endpoints for dashboard current-state consumption.
- Serves `GET /healthz`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Defaults to a live Binance Spot-backed provider for local and command-line runtime use.
- Warm-up is explicit: symbol and global payloads may be partial or unavailable until live Spot observations are fetched.
- Keeps the first live cutover Spot-driven; `usa` remains explicit rather than synthesized.
- Surfaces degradation in payloads instead of failing `/healthz` for market-data freshness.
- Supports `MARKET_STATE_API_CONFIG_PATH`, `MARKET_STATE_API_BINANCE_BASE_URL`, and `MARKET_STATE_API_SPOT_POLL_INTERVAL` for local runtime overrides.
- Preserves the service-owned current-state trust boundary so `apps/web` remains read-only and presentational.
