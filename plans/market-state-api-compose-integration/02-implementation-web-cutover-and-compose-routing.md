# Implementation: Web Cutover And Compose Routing

## Requirements And Scope

- Remove frontend-owned runtime API mocking.
- Keep `apps/web` consuming the same `/api/market-state/*` contract family through its existing decoder and trust-state layers.
- Route browser requests to the Go API in both local dev and Docker Compose.

## Target Repo Areas

- `apps/web/src/api`
- `apps/web/vite.config.ts`
- `apps/web/tests/e2e`
- `apps/web/Dockerfile`
- `docker-compose.yml`

## Implementation Notes

- Delete or retire `apps/web/mock-api/dashboardMockApi.ts` and any runtime env flags that enabled frontend-owned API routes.
- Keep the dashboard client request paths relative (`/api/...`) so the browser remains agnostic to backend location.
- For local non-compose dev, add a Vite proxy to a configurable Go API origin rather than reintroducing frontend mock handlers.
- For Compose, serve the built SPA behind a web runtime that can reverse-proxy `/api` to `market-state-api` so browser traffic stays same-origin.
- If a multi-stage static web image is simpler than `vite preview`, prefer that production-like runtime for Compose.
- Preserve existing decoder/state tests where they validate presentation logic only; add at least one integration smoke that proves a real Go API response reaches the dashboard without Playwright route fulfillment.

## Testing Expectations

- `pnpm --dir apps/web build`
- targeted dashboard decoder / state tests still pass
- one browser or integration smoke runs without frontend-owned API route mocks

## Summary

This step makes the frontend a true client again: it reads the Go API over `/api`, keeps its decoder/trust-state boundary, and stops owning runtime response simulation.
