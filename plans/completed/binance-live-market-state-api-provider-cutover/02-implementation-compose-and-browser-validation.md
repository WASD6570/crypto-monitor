# Implementation: Compose And Browser Validation

## Requirements And Scope

- Keep the same-origin `/api` delivery path from `web` to `market-state-api` unchanged.
- Update local Compose wiring only as needed to run the live-backed command path.
- Update browser smoke coverage so it verifies the real Go API boundary without asserting deterministic market-state values.
- Keep `apps/web` decoder and client contracts stable; do not add frontend mocks or venue-aware logic.

## Target Repo Areas

- `docker-compose.yml`
- `apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts`
- `apps/web/playwright.config.ts` only if the external-server flow truly needs a minimal adjustment

## Implementation Notes

- If the live command path needs env vars or config-path overrides, wire them into the `market-state-api` Compose service only.
- Preserve the quick healthcheck cadence from `plans/completed/market-state-api-compose-integration/`; do not regress `depends_on.condition: service_healthy` startup behavior.
- Update browser assertions to focus on durable outcomes:
  - `/dashboard` loads from the external server
  - summary cards for `BTC-USD` and `ETH-USD` render
  - focused symbol content updates when switching symbols
  - current-state sections render through same-origin `/api`
  - partial/unavailable/degraded live responses remain user-visible and non-crashing
- Remove expectations tied to deterministic regime text or fixed seeded slow-context values.

## Testing Expectations

- `docker compose config` stays clean after wiring changes.
- Compose startup brings up `market-state-api` and `web` with the live-backed command path.
- Direct `curl` checks against `http://127.0.0.1:4173/api/market-state/global` and `http://127.0.0.1:4173/api/market-state/BTC-USD` confirm the browser origin is still proxying Go-owned JSON.
- Playwright external-server smoke passes without `page.route(...)` mocks and without deterministic state assumptions.

## Summary

This module keeps the existing browser-to-Go trust boundary intact while swapping the runtime source underneath it. Compose remains the local proving ground, and browser smoke becomes robust to real live/current-state variability instead of deterministic fixture text.
