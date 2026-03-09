# Market State API

- Owns Go-served read endpoints for dashboard current-state consumption.
- Serves `GET /healthz`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Keeps temporary deterministic local state on the Go side until live ingestion-backed read models are wired in.
- Preserves the service-owned current-state trust boundary so `apps/web` remains read-only and presentational.
