# Testing: Market State API Compose Integration

## Smoke Matrix

| Case | Command / Flow | Expected Evidence |
|---|---|---|
| Go API unit tests | `"/usr/local/go/bin/go" test ./services/market-state-api/...` | handlers/provider pass with deterministic responses |
| Web build | `pnpm --dir apps/web build` | web bundle builds after API cutover |
| Compose config | `docker compose config` | `web` and `market-state-api` render successfully |
| Compose startup | `docker compose up --build -d` | both services are running |
| Dashboard API proxy | `curl http://127.0.0.1:4173/api/market-state/global` | JSON current-state payload comes from Go API path |
| Symbol API proxy | `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | symbol payload includes slow-context and regime fields |
| Browser smoke | `LD_LIBRARY_PATH=... pnpm --dir apps/web exec playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium` | dashboard loads through real Go API with no Playwright route fulfillment |
| Cleanup | `docker compose down` | stack stops cleanly |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/market-state-api/...`
- `pnpm --dir apps/web test -- --run dashboardDecoders`
- `pnpm --dir apps/web test -- --run dashboardStateMapper`
- `pnpm --dir apps/web build`
- `docker compose config`
- `docker compose up --build -d`
- `curl http://127.0.0.1:4173/api/market-state/global`
- `curl http://127.0.0.1:4173/api/market-state/BTC-USD`
- `curl http://127.0.0.1:4173/dashboard`
- `docker compose down`

## Verification Checklist

- No frontend runtime API mocks remain in the compose path.
- The browser reads service-owned current-state payloads through Go.
- The dashboard still honors service-owned trust/degraded/unavailable semantics.
- Docs clearly distinguish deterministic local API state from future live exchange connectivity.

## Report Path

- Write validation results to `plans/market-state-api-compose-integration/testing-report.md` while the feature is active.
- Move the report with the rest of the directory when archiving to `plans/completed/market-state-api-compose-integration/`.
