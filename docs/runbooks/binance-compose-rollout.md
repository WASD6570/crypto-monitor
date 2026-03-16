# Binance Compose Rollout

## Startup Contract

- The checked-in Compose stack always starts `market-state-api` with `MARKET_STATE_API_CONFIG_PATH=/app/configs/prod/ingestion.v1.json`.
- Compose does not switch runtime behavior between `local`, `dev`, and `prod`; the checked-in runtime posture is intentionally prod-like everywhere.
- `web` remains the public entry point on `http://127.0.0.1:4173`, and it proxies same-origin `/api/*` requests to `market-state-api` on the internal Compose network.
- `GET /healthz` stays process health only.
- `GET /api/runtime-status` is the bounded operator runtime-health route for fixed `BTC-USD` and `ETH-USD`.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain consumer read routes.

## Fast Proof

- Preferred command: `make compose-smoke`
- The smoke helper renders Compose, starts the stack, checks `/api/runtime-status` and `/api/market-state/global` through the same-origin web path, checks `/healthz` from inside `market-state-api`, and tears the stack down on success or failure.

## Manual Rollout Check

1. Start the stack: `docker compose up --build -d`
2. Confirm the services are up: `docker compose ps`
3. Query runtime health through the web entry point: `curl -fsS http://127.0.0.1:4173/api/runtime-status`
4. Check current-state reachability through the same-origin web path: `curl -sS -o /tmp/market-state-global.json -w '%{http_code}\n' http://127.0.0.1:4173/api/market-state/global`
5. Confirm process health from inside the Go container: `docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz`

## How To Read The Results

- `readiness=NOT_READY` on `/api/runtime-status` is expected during warm-up.
- A temporary current-state miss during warm-up is acceptable if `/api/runtime-status` is reachable and still reports `NOT_READY`.
- Once runtime status reaches `readiness=READY`, treat `DEGRADED` or `STALE` feed health as an operator issue, not a Compose-startup success.
- A healthy `/healthz` response does not prove market-data freshness; it only proves the process is serving.
- If you override Spot URLs for testing, set the paired USD-M override URLs too or startup will fail fast.

## Escalation Path

- For feed-health vocabulary and `GET /api/runtime-status` field interpretation, use `docs/runbooks/ingestion-feed-health-ops.md`.
- For active degraded or stale runtime investigation, use `docs/runbooks/degraded-feed-investigation.md`.

## Clean Shutdown

- Stop the stack with `docker compose down --remove-orphans`.
