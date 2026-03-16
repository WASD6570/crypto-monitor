# Binance Environment Config And Rollout Hardening

## Epic Summary

Keep the checked-in Binance runtime profiles prod-like everywhere and align rollout docs around one startup posture for `cmd/market-state-api` so operators do not drift into environment-specific behavior or silently change current-state, runtime-status, or replay semantics.

## In Scope

- keep checked-in Binance runtime profiles in `configs/local`, `configs/dev`, and `configs/prod` identical in runtime behavior, with prod-like defaults everywhere
- preserve the existing override guardrails in `cmd/market-state-api` without adding environment-selection behavior
- update compose and operator docs so startup, warm-up expectations, and `/healthz` versus `/api/runtime-status` usage stay explicit under one prod-like posture everywhere
- add focused validation that checked-in profiles parse cleanly, stay behaviorally identical, and startup still fails loudly on unsafe overrides
- add an isolated local developer workflow that provides Vite HMR and Go auto-restart while preserving the exact same live market connectivity, same-origin `/api` boundary, and prod-like runtime inputs

## Out Of Scope

- non-Binance deployment automation, secrets distribution, or infrastructure-specific rollout tooling
- changes to current-state or regime semantics beyond preserving the settled Wave 2 behavior
- long-run soak, reconnect torture, or final failure-path hardening; that remains in `binance-long-run-runtime-hardening`
- browser-side Binance access or any Python runtime dependency in the live path
- fixture-backed frontend runtime routes, mock market-state providers, or dev-only market semantics that diverge from the settled live stack

## Target Repo Areas

- `configs/*`
- `libs/go/ingestion`
- `cmd/market-state-api`
- `services/market-state-api/README.md`
- `docker-compose.yml`
- `docker-compose.dev.yml`
- `apps/web`
- `services/market-state-api/Dockerfile.dev`
- `scripts/dev`
- `docs/runbooks`

## Validation Shape

- targeted Go tests for config parsing and prod-like identical behavior across `local`, `dev`, and `prod`
- targeted `cmd/market-state-api` startup tests covering explicit override behavior and loud failure on partial or unsafe configuration
- compose proof plus runbook-driven startup verification that preserves `/healthz`, `/api/runtime-status`, and `/api/market-state/*` roles during warm-up
- dev-workflow proof that the optional overlay serves `/dashboard` through Vite HMR, proxies `/api/*` to the same Go boundary, and auto-restarts the Go runtime without switching to mocks or alternate market wiring

## Current Repository State

- Wave 1 and Wave 2 completed the sustained Spot runtime owner, additive runtime-status route, and conservative USD-M influence application, so rollout work can now assume the live runtime shape is settled enough to keep one prod-like posture everywhere
- checked-in ingestion profiles already exist for `local`, `dev`, and `prod`, and this epic now keeps them behaviorally identical instead of assigning different runtime personalities
- `cmd/market-state-api` still defaults to `configs/local/ingestion.v1.json` and supports endpoint override environment variables, but identical checked-in profiles now prevent local/dev/prod behavior drift while the override guardrails stay in place
- `plans/completed/binance-rollout-compose-and-ops-handoff/` already settled the prod-like Compose startup contract and operator handoff for the checked-in stack
- the remaining follow-up in this epic is an optional developer-only workflow that speeds local iteration without changing live Binance connectivity or the prod-like default stack

## Major Constraints

- keep `BTC-USD` and `ETH-USD` as the tracked symbols for this epic
- keep `/healthz` process-health only and preserve `/api/runtime-status` as the operator runtime-health route
- preserve `GET /api/market-state/global` and `GET /api/market-state/:symbol` by default; rollout hardening must not silently widen or repurpose those contracts
- keep replay and repeated-run current-state behavior deterministic for the same config inputs
- fail loudly on unsafe or partial runtime override combinations rather than silently mixing Spot and USD-M endpoints
- keep Go as the live runtime path; Python remains offline-only
