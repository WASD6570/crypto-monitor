# Child Plan Seeds

## `binance-runtime-config-profile-parity`

- Outcome: make the checked-in Binance `local`, `dev`, and `prod` ingestion profiles explicitly valid for the sustained Spot plus USD-M runtime while keeping them prod-like and identical in behavior
- Primary repo area: `configs/*`, `libs/go/ingestion`
- Dependencies: archived `binance-market-state-live-reader-cutover`, `binance-runtime-status-endpoint-and-ops-handoff`, and `binance-usdm-output-application-and-replay-proof`
- Validation shape: targeted Go tests that load each profile, assert Binance runtime invariants, and prove the fixed symbol set plus required polling fields remain intact
- Why it stands alone: this settles the profile contract before command startup logic, docs, or compose wiring depend on it

## `binance-rollout-compose-and-ops-handoff`

- Outcome: align `docker-compose.yml`, repo startup docs, and operator runbooks with the settled prod-like startup posture so all checked-in paths behave the same for now and override guardrails stay explicit
- Primary repo area: `docker-compose.yml`, `README.md`, `docs/runbooks`, `services/market-state-api/README.md`
- Dependencies: `binance-runtime-config-profile-parity`
- Validation shape: `docker compose config`, focused startup smoke, and runbook verification of warm-up plus runtime-status guidance under the same prod-like config posture everywhere
- Why it stands alone: it is the rollout and operator handoff slice that explains the single supported startup posture after config parity has removed runtime differences between checked-in paths

## `binance-live-reload-dev-workflow`

- Outcome: add an optional developer-only workflow that provides Vite HMR for `apps/web` and Go auto-restart for `market-state-api` while preserving the exact same live Binance connectivity, same-origin `/api` boundary, and prod-like runtime inputs as the default stack
- Primary repo area: `docker-compose.dev.yml`, `apps/web`, `services/market-state-api`, `scripts/dev`, `README.md`
- Dependencies: archived `binance-runtime-config-profile-parity` and `binance-rollout-compose-and-ops-handoff`
- Validation shape: overlaid Compose config proof, same-origin `/dashboard` plus `/api/runtime-status` smoke through the dev server, and a watcher-restart proof that does not switch to mocks or fixture-backed runtime responses
- Why it stands alone: it improves local iteration speed without reopening rollout semantics, and its dev-only wiring can be reviewed and validated independently of the default prod-like Compose path
