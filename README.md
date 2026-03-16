# Crypto Market Copilot

Monorepo scaffold for a multi-language crypto market system.

This repository starts with structure only. It does not implement business logic, schemas, migrations, or concrete service internals yet.

## Stack Ownership

- Go owns live, realtime, production-path services in `services/`.
- TypeScript owns the SPA in `apps/web`.
- Python owns research, ML, and offline analysis in `apps/research`.
- Python is intentionally optional for live operation.

## Repository Layout

```text
.
├── apps/
│   ├── research/
│   └── web/
├── configs/
│   ├── dev/
│   ├── local/
│   └── prod/
├── docs/
│   ├── architecture/
│   ├── decisions/
│   ├── runbooks/
│   └── specs/
├── libs/
│   ├── go/
│   ├── python/
│   └── ts/
├── initiatives/
├── plans/
├── schemas/
│   ├── json/
│   └── sql/
├── scripts/
│   ├── backfill/
│   ├── dev/
│   └── export/
├── services/
├── tests/
│   ├── e2e/
│   ├── fixtures/
│   ├── integration/
│   ├── parity/
│   └── replay/
├── .env.example
├── Makefile
└── docker-compose.yml
```

## Intent By Area

- `apps/web`: Vite + React + TypeScript SPA.
- `apps/research`: Python workspace for notebooks, experiments, offline jobs, and ML research.
- `services`: focused Go service boundaries for ingestion, normalization, features, alerts, replay, risk, simulation, and backfills.
- `libs`: shared code split by language.
- `initiatives`: initiative-level planning artifacts for work that is too large for one feature plan.
- `plans`: implementation-ready feature plans.
- `schemas`: canonical home for shared contracts and future database definitions.
- `docs`: human-readable architecture notes, specs, decisions, and runbooks.
- `configs`: environment-specific configuration homes.
- `scripts`: developer, backfill, and export helpers.
- `tests`: fixtures and higher-level test organization.

## Schema Homes

The shared contract area is intentionally explicit:

- `schemas/json/events`
- `schemas/json/features`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `schemas/json/replay`
- `schemas/json/simulation`

## Current Status

- The repo is scaffolded for gradual growth.
- Service and app internals are intentionally minimal.
- Program docs live under `docs/specs/`; initiative plans live under `initiatives/`; feature plans live under `plans/`.
- Later implementation work can fill in concrete code without redesigning the top-level layout.

## Compose Startup

- Run `pnpm run prod:like` from the repo root for the default prod-like stack.
- Run `docker compose up --build` from the repo root.
- The checked-in Compose stack always starts `market-state-api` with the prod-like config path at `/app/configs/prod/ingestion.v1.json`; there is no separate local/dev/prod startup posture in Compose.
- Open `http://127.0.0.1:4173/dashboard`.
- The current Compose stack runs `web` plus a Go-owned `market-state-api` service, and `apps/web` stays a same-origin consumer of `/api/*`; the browser does not talk to Binance directly.
- Use `GET /api/runtime-status` as the operator runtime-health route for fixed `BTC-USD` and `ETH-USD` status; `GET /healthz` stays process health only.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain consumer read routes and may briefly be unavailable while `/api/runtime-status` still reports `readiness=NOT_READY` during warm-up.
- Run `make compose-smoke` for the repeatable rollout proof, and use `docs/runbooks/binance-compose-rollout.md` for the full startup and handoff sequence.

## Developer Live Reload

- Keep `docker compose up --build` as the default prod-like reference path.
- Run `pnpm dev` from the repo root for the isolated developer workflow.
- For an isolated dev-only workflow with Vite HMR and Go auto-restart, run `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build`.
- The dev overlay still uses the real Go-owned live market path and the same checked-in prod-like config profile; it does not enable frontend mocks, fixture-backed runtime reads, or browser-side Binance access.
- Run `make compose-dev-smoke` to verify the dev overlay serves `/dashboard` through Vite, proxies same-origin `/api/*`, and restarts `market-state-api` after a watched-file touch.
