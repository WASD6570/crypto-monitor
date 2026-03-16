# Implementation: Compose Startup Contract

## Module Requirements And Scope

- Align `docker-compose.yml` with one explicit prod-like startup posture for `market-state-api`.
- Keep the same-origin `web` to `market-state-api` proxy path intact.
- Preserve `/healthz` as the compose healthcheck target and do not widen it into a runtime-freshness signal.
- Keep override handling explicit and narrow; do not add environment-selection variables, profile switches, or different compose behavior for `local`, `dev`, and `prod`.

## Target Repo Areas

- `docker-compose.yml`
- `services/market-state-api/Dockerfile` only if the compose contract proves a tiny container-path clarification is required

## Key Decisions

- Prefer one visible startup contract in compose, even if the underlying checked-in configs are already behaviorally identical.
- If an explicit config path is needed, pin exactly one checked-in prod-like profile inside the container rather than teaching compose to branch on environment labels.
- Keep `market-state-api` internal to the compose network and continue exposing the browser entry point through `web` on `:4173`.
- Leave Spot and USD-M override guardrails in application code; compose should not mask or bypass those failures.

## Implementation Notes

- Keep the compose diff minimal and operator-readable.
- Avoid adding secrets, `.env` indirection, or deployment-tooling assumptions.
- Preserve the current startup ordering where `web` waits for `market-state-api` process health before serving the same-origin dashboard path.

## Unit Test And Proof Expectations

- `docker compose config` renders successfully after the compose changes.
- The rendered compose output shows one startup posture for `market-state-api`, not a local/dev/prod selection branch.
- The compose smoke proof can reach `/api/runtime-status` and `/api/market-state/global` through the `web` service while `/healthz` stays internal to `market-state-api`.

## Summary

- This module settles the checked-in container startup contract so later docs can describe one prod-like runtime story without hedging around environment-specific behavior.
