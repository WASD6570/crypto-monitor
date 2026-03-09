# Market State API Compose Integration

## Ordered Implementation Plan

1. Add a Go-owned `market-state-api` service boundary that serves the existing dashboard current-state endpoints and owns any temporary deterministic fixture data.
2. Cut `apps/web` over to the Go API entirely, remove frontend-owned runtime mock routes, and preserve the existing read-only decoder/trust-state boundary.
3. Rework the local Docker Compose stack to run both `web` and `market-state-api` with same-origin `/api` delivery and clear local run documentation.
4. Validate Go API correctness, web integration, and Compose/browser smoke; record evidence in `plans/market-state-api-compose-integration/testing-report.md`.

## Problem Statement

The dashboard in `apps/web` already consumes service-shaped current-state contracts, but the current local startup path still keeps runtime API mocking inside the frontend container. That breaks the intended trust boundary: the web should be a read-only consumer of Go-owned current-state responses, and Docker Compose should exercise that boundary directly before any live Binance connectivity work begins.

## Bounded Scope

- one new Go read service boundary for dashboard current-state delivery
- temporary deterministic state ownership on the Go side only
- web client cutover to the Go API with no frontend-owned runtime mock routes
- Docker Compose integration for `web` + `market-state-api`
- targeted tests and browser smoke for the integrated stack

## Out Of Scope

- real Binance websocket connectivity or ingestion orchestration
- alerting, risk, replay-control APIs, or authenticated endpoints
- replacing temporary Go fixture/state providers with live stores in the same slice
- broad refactors of dashboard component tests that do not block API cutover

## Requirements

- Keep the current-state source of truth in Go service-owned code.
- `apps/web` remains read-only and presentational; it must not derive tradeability, freshness, global ceiling, or slow-context semantics on its own.
- Any temporary mocks/fixtures used to make the stack runnable must live on the Go side, not in frontend runtime middleware.
- Preserve the existing current-state contract family and dashboard decoder boundary from `plans/completed/market-state-current-query-contracts/` and `plans/completed/dashboard-query-adapters-and-trust-state/`.
- Favor same-origin `/api` access in Compose so the browser hits one web origin while Go owns the API responses behind it.
- Keep Python out of the live/runtime path.

## Design Notes

- Add an explicit `services/market-state-api/` boundary rather than hiding HTTP handlers inside `feature-engine` or `regime-engine`.
- The API service should assemble responses through existing Go query builders where possible, even if the backing state source is still deterministic and local for now.
- Temporary deterministic dashboard state should be represented in Go as an API-owned provider seam, not frontend-only route stubs.
- The web app should continue to request `/api/market-state/global` and `/api/market-state/:symbol`, but the request path should terminate in Go through dev/proxy or Compose routing rather than frontend middleware.
- Compose should run two services and make `/dashboard` usable without pretending that live exchange connectivity already exists.

## Target Repo Areas

- `services/market-state-api`
- `services/feature-engine`
- `services/regime-engine`
- `services/slow-context`
- `apps/web/src/api`
- `apps/web/vite.config.ts`
- `apps/web/tests/e2e`
- `docker-compose.yml`
- `README.md`
- `apps/web/README.md`

## Dependency Context

- `plans/completed/market-state-current-query-contracts/` already defined the service-owned contract boundary for symbol/global current-state payloads.
- `plans/completed/dashboard-query-adapters-and-trust-state/` already moved `apps/web` to contract-shaped current-state reads and trust-state reduction.
- `services/feature-engine/service.go` already exposes `QueryCurrentState` and `QueryCurrentStateWithSlowContext`.
- `services/regime-engine/service.go` already exposes `QueryCurrentGlobalState`.
- The current `apps/web/mock-api/dashboardMockApi.ts` and root `docker-compose.yml` should be treated as temporary scaffolding to remove in this feature.

## Acceptance Criteria

- Another agent can implement the feature without reopening initiative-level planning.
- A Go-owned `market-state-api` serves the dashboard endpoints used by `apps/web`.
- No frontend runtime API mocks remain in the shipped web app or Compose path.
- `docker compose up --build` runs `web` and `market-state-api`, and `/dashboard` reads current-state data through that Go boundary.
- Validation commands are concrete and reproducible.

## ASCII Flow

```text
browser
  |
  v
web origin (/dashboard, /api/*)
  |
  +--> static dashboard assets
  |
  +--> reverse proxy /api/market-state/*
           |
           v
   services/market-state-api
     - global state handler
     - symbol state handler
     - health/readiness
     - temporary deterministic provider in Go
           |
           +--> feature-engine current-state assembly
           +--> regime-engine global-state assembly
           +--> slow-context query seam where present
```

## Live-Path Boundary

- This feature introduces only a read API and local stack wiring.
- It does not add live venue sockets, authenticated exchange flows, or client-owned market interpretation.
- The next live-data slice can replace the API's deterministic provider behind the same Go-owned endpoints.

## Archive Intent

- Keep this feature active under `plans/market-state-api-compose-integration/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/market-state-api-compose-integration/`.
