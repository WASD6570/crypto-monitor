# Testing Report: docker-compose-local-dashboard-stack

## Environment

- Target: local Docker Compose dashboard stack
- Date/time: 2026-03-08
- Commit/branch: `main`

## Smoke Matrix

| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Compose config | `docker compose config` | Compose file renders successfully | Compose rendered one `web` service with port `4173` and mock API env vars | PASS |
| Web build | `pnpm --dir apps/web build` | Web bundle builds cleanly | Build passed after fixing one stale test-only malformed-value assignment | PASS |
| Container startup | `docker compose up --build -d` | Web container starts and stays running | `crypto-copilot-web-1` built and remained `Up` on port `4173` | PASS |
| Dashboard HTML | `curl http://127.0.0.1:4173/dashboard` | HTML contains the app shell | Returned `index.html` for the SPA with root element and built assets | PASS |
| Mock global API | `curl http://127.0.0.1:4173/api/market-state/global` | JSON payload contains dashboard global state | Returned `market_state_current_global_v1` JSON with BTC/ETH summaries | PASS |
| Mock symbol API | `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | JSON payload contains symbol state | Returned `market_state_current_response_v1` JSON with slow-context, composite, bucket, and regime fields | PASS |

## Execution Evidence

### Compose config

- Command/Request: `docker compose config`
- Expected: root compose file resolves without errors
- Actual: resolved build context `apps/web`, published port `4173`, and mock API env values
- Verdict: PASS

### Web build

- Command/Request: `pnpm --dir apps/web build`
- Expected: TypeScript and Vite build succeed
- Actual: initial run failed on `apps/web/src/api/dashboard/dashboardDecoders.test.ts` because the test assigned an invalid literal to a narrowed union; updated the malformed-value test to cast the field as `string`, then the build passed
- Verdict: PASS

### Compose startup

- Command/Request: `docker compose up --build -d`
- Expected: image builds and service stays running
- Actual: image `crypto-copilot-web` built successfully and `docker compose ps` showed `crypto-copilot-web-1` as `Up`
- Verdict: PASS

### HTTP smoke

- Command/Request: `curl http://127.0.0.1:4173/dashboard`
- Expected: HTML entry page for the dashboard SPA
- Actual: returned the built `index.html` document with `/assets/index-BRH-8ZHv.js`
- Verdict: PASS

### API smoke

- Command/Request: `curl http://127.0.0.1:4173/api/market-state/global`
- Expected: deterministic mock current-state global payload
- Actual: returned `market_state_current_global_v1` JSON with `WATCH` global state and BTC/ETH symbol summaries
- Verdict: PASS

### Symbol smoke

- Command/Request: `curl http://127.0.0.1:4173/api/market-state/BTC-USD`
- Expected: deterministic mock symbol payload
- Actual: returned `market_state_current_response_v1` JSON with `slowContext`, `composite`, `buckets`, `regime`, and `recentContext`
- Verdict: PASS

## Side-Effect Verification

### Compose runtime

- Evidence: `docker compose ps`
- Expected state: one local web service is running and bound to host port `4173`
- Actual state: `crypto-copilot-web-1` was running with `0.0.0.0:4173->4173/tcp`
- Verdict: PASS

## Blockers / Risks

- The Compose stack is intentionally fixture-backed and does not prove live exchange or Go service connectivity.
- The current image keeps dev/test dependencies because `vite preview` is used directly for the local stack; that is acceptable for local startup but not tuned for production deployment.

## Next Actions

1. Add a browser-level Compose smoke once a stable CLI/browser dependency path is desired for container validation.
2. Replace the mock API seam with real service-owned endpoints when the first live ingestion/query service entrypoint exists.
