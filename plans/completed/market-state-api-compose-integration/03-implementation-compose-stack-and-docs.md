# Implementation: Compose Stack And Docs

## Requirements And Scope

- Run both `web` and `market-state-api` from the repo-root Compose file.
- Make startup and shutdown reproducible for another developer.
- Keep the docs honest about deterministic Go-owned local state versus future live exchange connectivity.

## Target Repo Areas

- `docker-compose.yml`
- `README.md`
- `apps/web/README.md`
- `services/README.md`
- `services/market-state-api/README.md`

## Implementation Notes

- Expose the web app on the current user-facing port and the Go API on an internal service port; optional host API exposure is acceptable if it simplifies debugging.
- Add Compose healthchecks or readiness ordering if startup races appear likely.
- Document exactly how to run the stack, what endpoints exist, and that the current API data is deterministic/local until the live ingestion slice lands.
- Make the next replacement seam explicit: the Go provider inside `market-state-api` is what later live Binance-backed state should swap out.

## Testing Expectations

- `docker compose config`
- `docker compose up --build -d`
- `curl http://127.0.0.1:4173/dashboard`
- `curl http://127.0.0.1:4173/api/market-state/global`
- `curl http://127.0.0.1:4173/api/market-state/BTC-USD`
- `docker compose down`

## Summary

This step turns the temporary one-container web demo into a real two-service local stack with a clear frontend/backend boundary and an honest on-ramp to live data.
