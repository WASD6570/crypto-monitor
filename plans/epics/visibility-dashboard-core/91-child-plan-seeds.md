# Child Plan Seeds: Visibility Dashboard Core

## `dashboard-shell-and-summary-strip` (completed)

- Outcome: the Vite SPA gains one current-state dashboard route with a global status rail, BTC/ETH summary strip, stable symbol-focus URL state, and dense desktop/mobile reading order using fixture-backed presentational data.
- Primary repo areas: `apps/web/src/app`, `apps/web/src/pages`, `apps/web/src/features`, `apps/web/src/components`
- Depends on: `plans/completed/canonical-contracts-and-fixtures/`, `plans/completed/market-ingestion-and-feed-health/`, epic IA docs in `plans/epics/visibility-dashboard-core/`
- Validation shape: `pnpm --dir apps/web test -- --runInBand` for route and summary-card rendering plus `pnpm --dir apps/web build`
- Why it stands alone: it fixes the operator reading order and responsive shell before live query integration decisions are finalized.
- Archive: `plans/completed/dashboard-shell-and-summary-strip/`

## `dashboard-query-adapters-and-trust-state` (completed)

- Outcome: the SPA gets a bounded client data layer for dashboard snapshot, symbol detail, derivatives context, and feed-health/regime surfaces, including `loading`, `ready`, `stale`, `degraded`, and `unavailable` panel states with service-owned timestamps and completeness markers.
- Primary repo areas: `apps/web/src/api`, `apps/web/src/features`, `apps/web/src/hooks`, `apps/web/src/state`, optional `libs/ts`
- Depends on: `plans/completed/dashboard-shell-and-summary-strip/`, `plans/completed/market-state-current-query-contracts/`
- Validation shape: `pnpm --dir apps/web test -- --runInBand` for adapter decoding, partial snapshot handling, stale propagation, and trust metadata rendering
- Why it stands alone: it locks the client trust boundary and state machine separately from panel layout work.
- Archive: `plans/completed/dashboard-query-adapters-and-trust-state/`

## `dashboard-detail-panels-and-symbol-switching` (completed)

- Outcome: focused symbol overview, microstructure, derivatives, and feed health/regime panels render for BTC and ETH, with instant-feeling symbol switching that reuses fresh cached detail data without hiding peer summaries.
- Primary repo areas: `apps/web/src/features`, `apps/web/src/components`, `apps/web/src/hooks`, `apps/web/src/pages`
- Depends on: `plans/completed/dashboard-shell-and-summary-strip/`, `plans/completed/dashboard-query-adapters-and-trust-state/`
- Validation shape: `pnpm --dir apps/web test -- --runInBand` for panel composition and symbol-switch behavior plus `pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium`
- Why it stands alone: it implements the main operator read path after shell and data boundaries are already fixed.
- Archive: `plans/completed/dashboard-detail-panels-and-symbol-switching/`

## `dashboard-negative-state-mobile-a11y` (completed)

- Outcome: the dashboard exposes stale, degraded, unavailable, and partial-data states consistently across desktop and mobile, with keyboard-safe symbol switching, visible warnings, and screen-reader-readable trust labels.
- Primary repo areas: `apps/web/src/components`, `apps/web/src/features`, `apps/web/src/pages`, `tests/e2e`
- Depends on: `plans/completed/dashboard-shell-and-summary-strip/`, `plans/completed/dashboard-query-adapters-and-trust-state/`, `plans/completed/dashboard-detail-panels-and-symbol-switching/`
- Validation shape: `pnpm --dir apps/web test -- --runInBand` for accessibility-focused component behavior plus `pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome`
- Why it stands alone: negative-state honesty and mobile-safe density are a separate risk surface from basic panel rendering.
- Archive: `plans/completed/dashboard-negative-state-mobile-a11y/`

## `dashboard-fixture-smoke-matrix` (completed)

- Outcome: deterministic healthy, degraded, stale, and partial-data fixtures or mocks back the dashboard unit and Playwright smoke matrix, including timestamp-fallback trust notes and per-panel availability cases.
- Primary repo areas: `tests/fixtures`, `apps/web/src/api`, `apps/web/src/features`, `tests/e2e`
- Depends on: `plans/completed/dashboard-query-adapters-and-trust-state/`, `plans/completed/dashboard-detail-panels-and-symbol-switching/`, `plans/completed/dashboard-negative-state-mobile-a11y/`, shared fixture conventions from `plans/completed/canonical-contracts-and-fixtures/`
- Validation shape: `pnpm --dir apps/web test -- --runInBand`, `pnpm --dir apps/web build`, `pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium`, and `pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome`
- Why it stands alone: it closes the epic with implementation-proof coverage instead of scattering cross-panel verification through earlier plans.
- Archive: `plans/completed/dashboard-fixture-smoke-matrix/`
