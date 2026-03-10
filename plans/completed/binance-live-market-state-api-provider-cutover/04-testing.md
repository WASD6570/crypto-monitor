# Testing: Binance Live Market State API Provider Cutover

## Validation Matrix

| Area | Command / Flow | What It Proves |
|---|---|---|
| Command and provider wiring | `"/usr/local/go/bin/go" test ./services/market-state-api ./cmd/market-state-api -v` | handler/provider tests still pass and the command entrypoint builds the live-backed provider cleanly |
| Live current-state regression | `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceCurrentState.*' -v` | previous Spot-first current-state semantics remain stable after command cutover |
| Web contract guard | `pnpm --dir apps/web test -- --run dashboardDecoders` | current Go payload shape still satisfies the existing dashboard decoder boundary |
| Compose render | `docker compose config` | local stack renders with the updated `market-state-api` runtime wiring |
| Compose startup | `docker compose up --build -d` | `web` and `market-state-api` start with the live-backed command path |
| Same-origin API smoke | `curl http://127.0.0.1:4173/api/market-state/global` and `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | browser origin still proxies Go-owned JSON for global and symbol routes |
| Browser smoke | `PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 pnpm --dir apps/web exec playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium` | dashboard loads against the real Go API path without route mocks and tolerates live variability |

## Verification Checklist

- `GET /healthz` returns `200` while the process is up.
- `GET /api/market-state/global` returns the existing `market_state_current_global_v1` shape with `global`, `symbols`, and `provenance`.
- `GET /api/market-state/BTC-USD` returns the existing symbol response shape with `slowContext`, `composite`, `buckets`, `regime`, `recentContext`, and `provenance`.
- API responses may be partial or unavailable during warm-up, but they must remain schema-stable and machine-readable.
- Unsupported symbols still return the existing not-found behavior.
- Browser smoke passes without deterministic state text assertions.

## Notes

- This feature owns Compose, API, and browser verification directly; do not create a separate smoke-only child plan.
- If live external connectivity is unavailable in the execution environment, record that blocker in `testing-report.md` and still run the non-networked Go and web contract checks.
- Write the active validation report to `plans/binance-live-market-state-api-provider-cutover/testing-report.md`.
- After implementation and validation complete, move the entire directory to `plans/completed/binance-live-market-state-api-provider-cutover/`.
