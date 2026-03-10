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

## Local Docker Startup

- Run `docker compose up --build` from the repo root.
- Open `http://127.0.0.1:4173/dashboard`.
- The current Compose stack runs `web` plus a Go-owned `market-state-api` service.
- Dashboard `/api/market-state/*` responses now come from Go, not frontend runtime mocks.
- The default `market-state-api` command now fetches live Binance Spot snapshots for `BTC-USD` and `ETH-USD` behind the existing Go API boundary.
- `/healthz` reflects process readiness; market-data warm-up, unavailability, and degradation stay visible in the JSON payloads.
- The first live cutover remains Spot-driven, so `usa` stays explicit and may be unavailable or partial.
