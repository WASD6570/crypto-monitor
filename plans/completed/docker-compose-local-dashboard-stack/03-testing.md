# Testing: Docker Compose Local Dashboard Stack

## Smoke Matrix

| Case | Command / Flow | Expected Evidence |
|---|---|---|
| Compose config | `docker compose config` | Compose file renders successfully |
| Container startup | `docker compose up --build -d` | Web container starts and stays running |
| Dashboard HTML | `curl http://127.0.0.1:4173/dashboard` | HTML contains the app shell |
| Mock global API | `curl http://127.0.0.1:4173/api/market-state/global` | JSON payload contains dashboard global state |
| Mock symbol API | `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | JSON payload contains symbol state |
| Cleanup | `docker compose down` | Stack stops cleanly |

## Required Commands

- `pnpm --dir apps/web build`
- `docker compose config`
- `docker compose up --build -d`
- `curl http://127.0.0.1:4173/dashboard`
- `curl http://127.0.0.1:4173/api/market-state/global`
- `curl http://127.0.0.1:4173/api/market-state/BTC-USD`
- `docker compose down`

## Verification Checklist

- The compose stack is clearly documented as fixture-backed.
- The dashboard can boot without external API infrastructure.
- The API routes answer with deterministic data aligned with the current UI contract.
- No Python or extra infra dependency is introduced.

## Report Path

- Write validation results to `plans/docker-compose-local-dashboard-stack/testing-report.md` while the feature is active.
- Move the report with the rest of the directory when archiving to `plans/completed/docker-compose-local-dashboard-stack/`.
