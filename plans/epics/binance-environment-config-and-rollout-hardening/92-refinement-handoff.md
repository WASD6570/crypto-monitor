# Refinement Handoff

## Next Recommended Action

- Run `feature-testing` for `plans/binance-live-reload-dev-workflow/`.
- Why next: the default prod-like Compose stack remains unchanged, and the dev-only live-reload child is now implemented and validated enough for final archive testing.

## Active Child Plan

- `plans/binance-live-reload-dev-workflow/`

## Archived Child Evidence

- `plans/completed/binance-runtime-config-profile-parity/`
- `plans/completed/binance-rollout-compose-and-ops-handoff/`

## Child Queue

- none beyond the active child plan above

## Safe Parallel Planning And Execution

- `binance-long-run-runtime-hardening` can still move through `program-refining` in parallel if planning capacity exists; it does not depend on the developer workflow landing first

## Prerequisites To Carry Forward

- keep `BTC-USD` and `ETH-USD` fixed throughout this epic
- keep `/healthz` process-health only and preserve `/api/runtime-status` as the additive operator route
- preserve `GET /api/market-state/global` and `GET /api/market-state/:symbol` by default; any rollout-facing metadata must stay additive and optional
- keep the conservative USD-M application posture unchanged and preserve the loud failure rule for partial Spot-only override environments
- keep checked-in config paths prod-like and identical in runtime behavior for now
- keep Go as the live runtime path; Python remains offline-only
- keep the default `docker-compose.yml` startup posture unchanged; developer iteration speedups must live in isolated dev-only wiring
- keep the developer workflow on the real market path: no frontend runtime mocks, no fixture-backed API fallback, and no browser-side Binance access

## Assumptions And Blockers

- assumption: the checked-in `configs/*/ingestion.v1.json` files remain the canonical runtime-profile homes for this epic rather than introducing a second config source
- assumption: compose should keep proving the same-origin Go API boundary while docs and runbooks describe the same prod-like startup posture everywhere
- blocker for later children: none; this follow-up child is bounded and optional, and Wave 4 can still move into `program-refining` separately
