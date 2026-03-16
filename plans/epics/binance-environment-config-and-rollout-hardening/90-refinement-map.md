# Refinement Map

## Already Done

- `plans/completed/binance-market-state-live-reader-cutover/` settled the sustained Spot-backed API cutover and the current warm-up posture
- `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/` settled `/api/runtime-status` as the additive operator route while keeping `/healthz` process-only
- `plans/completed/binance-usdm-output-application-and-replay-proof/` settled the bounded USD-M runtime inputs, additive provenance posture, and the requirement to avoid silent Spot/USD-M miswiring
- `configs/local/ingestion.v1.json`, `configs/dev/ingestion.v1.json`, and `configs/prod/ingestion.v1.json` already provide the checked-in profile shape this epic should keep prod-like rather than diversify
- `plans/completed/binance-runtime-config-profile-parity/` and `plans/completed/binance-rollout-compose-and-ops-handoff/` already completed the checked-in startup contract and operator rollout handoff for the default Compose path

## Remaining Work

- add an optional developer workflow that gives `apps/web` hot-module reload against the real Go API boundary instead of a built Nginx bundle
- add Go auto-restart for the local developer path without changing the live market wiring, fixed symbols, config path expectations, or override guardrails
- document and validate how the developer overlay differs from the default Compose stack only in iteration speed, not in market connectivity or runtime semantics

## Overlap And Non-Goals

- do not reopen runtime-health surface design; `/api/runtime-status` and `/healthz` semantics are already settled
- do not reopen USD-M current-state policy; carry forward the conservative Wave 2 posture unchanged
- do not bundle long-run reconnect or failure-path hardening into this epic; keep that for `binance-long-run-runtime-hardening`
- do not introduce infrastructure-specific deployment templates, secret management, or non-Binance environment policy here
- do not let a developer-only workflow replace or dilute the default prod-like Compose startup path

## Refinement Waves

### Archived Wave 1

- `binance-runtime-config-profile-parity`
- Why first: the checked-in environment profiles are the base contract for every later rollout step, and the current Binance `dev` and `prod` profiles were not yet aligned with the live open-interest runtime requirements

### Archived Wave 2

- `binance-rollout-compose-and-ops-handoff`
- Why later: docs and compose proof needed to describe the now-settled prod-like profile posture everywhere instead of inventing separate environment behavior

### Follow-On Wave

- `binance-live-reload-dev-workflow`
- Why now: the prod-like default stack is already settled, so the repo can safely add a faster local iteration path as long as it keeps the exact same live market boundary and remains clearly optional

### Direct Post-Implementation Checks

- verify all three checked-in profiles remain parseable through `LoadEnvironmentConfig(...)`, preserve the fixed Binance symbol set, and stay behaviorally identical
- verify `/healthz` stays process-health only while `/api/runtime-status` remains the runtime-health investigation route during rollout verification
- verify local compose and runbook steps do not imply browser-side Binance access or silent fallback to mismatched Spot versus USD-M endpoints
- verify the developer overlay serves `apps/web` through Vite HMR, proxies `/api/*` to Go, and shows Go watcher restarts without introducing mocks or fixture-backed runtime reads
