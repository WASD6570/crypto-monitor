# Testing Report: market-state-api-compose-integration

## Environment

- Target: local Compose stack with `web` + `market-state-api`
- Date/time: 2026-03-09
- Commit/branch: `main`

## Smoke Matrix

| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Go API tests | `"/usr/local/go/bin/go" test ./services/market-state-api/... ./cmd/market-state-api` | API handler/provider and entrypoint compile cleanly | Passed | PASS |
| Web build | `pnpm --dir apps/web build` | web bundle builds without frontend runtime mock middleware | Passed | PASS |
| Compose config | `docker compose config` | `web` and `market-state-api` render with proxy-ready wiring | Passed | PASS |
| Compose startup | `docker compose up --build -d` | both services start and API becomes healthy | Passed | PASS |
| Proxied global API | `curl http://127.0.0.1:4173/api/market-state/global` | Go-owned JSON reaches the browser origin via `/api` | Returned `market_state_current_global_v1` JSON | PASS |
| Proxied symbol API | `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | symbol payload includes slow-context and regime sections | Returned symbol JSON with `slowContext` and `regime` | PASS |
| Browser smoke | `PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 ... playwright test apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts` | dashboard loads from the Go API without Playwright route mocks | Passed after fixing omitted slow-context fields for unavailable contexts | PASS |
| Web state tests | `pnpm --dir apps/web test -- --run dashboardDecoders` and `pnpm --dir apps/web test -- --run dashboardStateMapper` | decoder and trust-state mapping still pass after Go API cutover | Passed | PASS |

## Execution Evidence

### Go API tests

- Command/Request: `"/usr/local/go/bin/go" test ./services/market-state-api/... ./cmd/market-state-api`
- Expected: new API boundary and entrypoint compile with passing tests
- Actual: passed
- Verdict: PASS

### Web build

- Command/Request: `pnpm --dir apps/web build`
- Expected: web bundle builds after removing frontend runtime API mocks
- Actual: passed
- Verdict: PASS

### Compose startup and proxy smoke

- Command/Request: `docker compose up --build -d`
- Expected: `market-state-api` becomes healthy and `web` serves `/dashboard`
- Actual: `docker compose ps` showed `market-state-api` healthy and `web` up on port `4173`
- Verdict: PASS

### HTTP smoke

- Command/Request: `curl http://127.0.0.1:4173/api/market-state/global`
- Expected: Go-owned current-state payload through same-origin `/api`
- Actual: returned global JSON with BTC/ETH symbol summaries from `market-state-api`
- Verdict: PASS

- Command/Request: `curl http://127.0.0.1:4173/api/market-state/BTC-USD`
- Expected: Go-owned symbol payload through same-origin `/api`
- Actual: returned symbol JSON with `slowContext`, `composite`, `buckets`, `regime`, and `recentContext`
- Verdict: PASS

### Browser smoke

- Command/Request: `PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 LD_LIBRARY_PATH=... pnpm --dir apps/web exec playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium`
- Expected: dashboard renders via the real Go API without `page.route` interception
- Actual: passed after making unavailable slow-context fields omit empty `sourceFamily` and `thresholdBasis`
- Verdict: PASS

### Web state tests

- Command/Request: `pnpm --dir apps/web test -- --run dashboardDecoders`
- Expected: decoder tests continue passing after API cutover
- Actual: passed
- Verdict: PASS

- Command/Request: `pnpm --dir apps/web test -- --run dashboardStateMapper`
- Expected: trust-state mapper tests continue passing after API cutover
- Actual: passed
- Verdict: PASS

## Side-Effect Verification

### Compose runtime

- Evidence: `docker compose ps`
- Expected state: `market-state-api` healthy and `web` serving port `4173`
- Actual state: both services were running; `market-state-api` reported healthy and `web` was bound to host port `4173`
- Verdict: PASS

## Blockers / Risks

- The API uses deterministic Go-owned local state, not live ingestion-backed state yet.

## Follow-Up Fixes

- `docker-compose up` in attached mode initially failed because `depends_on.condition: service_healthy` waited on a healthcheck with a `30s` interval, while Compose gave up before the first probe ran.
- Updated `docker-compose.yml` to use a fast API healthcheck (`interval: 2s`, `timeout: 2s`, `retries: 10`, `start_period: 1s`).
- Reproduced the fix with attached startup via `timeout 20s docker compose up`, which now reaches `market-state-api` healthy and starts `web` before the timeout stops the stack.

## Next Actions

1. Run the new browser smoke with the known local Playwright library path when desired.
2. Replace the deterministic provider behind `services/market-state-api` with live current-state assembly once the Binance/live read path is implemented.
