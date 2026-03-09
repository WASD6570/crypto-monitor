# Docker Compose Local Dashboard Stack

## Ordered Implementation Plan

1. Add a server-owned mock market-state API seam for the current dashboard route so the web app can start with meaningful data instead of empty network failures.
2. Add a containerized web runtime and root `docker-compose.yml` entry that serves the built SPA and the mock API together.
3. Document and validate the local Docker startup flow so future agents can run the stack without guessing.

## Problem Statement

The repository has a usable dashboard SPA, but there is no local container startup path and no live backend service entrypoint yet. Before starting real Binance connectivity, the project needs one reproducible Docker Compose path that boots the current app surface with deterministic data.

## Requirements

- Keep the current live/runtime boundary honest; do not invent a fake Go market-data service or claim live exchange connectivity.
- Start the existing `apps/web` dashboard via Docker Compose from the repo root.
- Serve stable market-state responses for the current dashboard route so `/dashboard` is useful immediately after startup.
- Reuse existing fixture/scenario data where practical instead of duplicating ad hoc JSON.
- Keep the setup light: no new database, queue, or Python runtime.
- Preserve a clean seam for replacing the mock API with real service-owned endpoints later.

## Out Of Scope

- Real Binance websocket connectivity.
- New Go service entrypoints or multi-service orchestration.
- Alerting, risk, replay, or contract redesign.
- Production deployment hardening.

## Design Notes

- The smallest truthful stack is one containerized web runtime with same-origin mock API routes.
- Vite should serve the current dashboard bundle and answer `/api/market-state/*` only when an explicit mock env flag is enabled.
- Docker Compose should remain repo-root owned and focus on the current runnable surface.
- Documentation should call out that this stack is fixture-backed and intended for local smoke/dev use until real service endpoints exist.

## Target Repo Areas

- `apps/web`
- `docker-compose.yml`
- `README.md`
- `apps/web/README.md`

## ASCII Flow

```text
docker compose up --build
          |
          v
  apps/web container
          |
          +--> built Vite SPA at /dashboard
          |
          +--> mock /api/market-state/global
          |
          +--> mock /api/market-state/:symbol
```

## Archive Intent

- Keep this feature active under `plans/docker-compose-local-dashboard-stack/` while implementation and validation are in progress.
- When implementation and validation finish, move the full directory and `testing-report.md` to `plans/completed/docker-compose-local-dashboard-stack/`.
