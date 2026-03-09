# Implementation: Mock Dashboard API

## Requirements And Scope

- Add one explicit server-side mock API seam for `/api/market-state/global` and `/api/market-state/:symbol`.
- Keep it opt-in through environment configuration.
- Reuse the existing dashboard scenario fixtures so the mock responses stay aligned with current UI expectations.
- Do not change client-side market logic ownership; the API response stays the source of truth for the current screen.

## Target Repo Areas

- `apps/web/vite.config.ts`
- `apps/web/src/test/dashboardScenarioCatalog.ts`
- `apps/web/src/features/dashboard-state/dashboardStateFixtures.ts`

## Implementation Notes

- Add a small Vite middleware plugin that can run in both dev and preview servers.
- Read a scenario selector from env, defaulting to `healthy` when mock mode is enabled.
- Return JSON payloads with proper status codes and a narrow route surface.
- Keep unsupported symbols explicit with `404` JSON instead of silent fallback.

## Unit / Smoke Expectations

- `pnpm build` still succeeds for `apps/web`.
- When mock mode is enabled, the dashboard can fetch same-origin market-state payloads without Playwright route interception.

## Summary

This step makes the current dashboard bootable with deterministic API data while keeping the future swap to real service endpoints explicit and low-risk.
