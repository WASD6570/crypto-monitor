# Slow Context Dashboard Panel Testing

## Goal

Verify that the dashboard renders service-owned slow CME and ETF context as advisory-only UI, with explicit cadence and freshness labeling, while leaving the core current-state route behavior unchanged.

Expected report output path while this feature is active: `plans/slow-context-dashboard-panel/testing-report.md`

## Validation Commands

```bash
pnpm --filter web test -- --run DashboardShell
pnpm --filter web test -- --run dashboardStateMapper
pnpm --filter web test -- --run dashboardScenarioCatalog
pnpm --filter web exec vite build
pnpm --filter web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium --grep "slow context|Context only"
pnpm --filter web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome --grep "slow context|Context only"
```

## Smoke Matrix

### 1. Healthy Slow Context

- Purpose: prove the dashboard shows all slow-context rows when the service seam is fully readable.
- Inputs:
  - healthy dashboard scenario with fresh CME volume, CME open interest, and ETF flow rows
- Verify:
  - `Context only` badge is visible
  - each row shows value, freshness, cadence, and `as of` metadata
  - summary cards and route warnings remain unchanged from the healthy realtime scenario

### 2. Delayed And Stale Advisory Rows

- Purpose: prove slower-cadence warnings stay isolated to the advisory panel.
- Inputs:
  - delayed CME row inside tolerated age
  - stale row beyond the service-owned threshold with last-known-good value still visible
- Verify:
  - delayed or stale copy appears only inside the slow-context panel
  - row-level freshness badges stay readable
  - realtime overview, microstructure, derivatives, and health panels do not change semantics

### 3. Partial Availability

- Purpose: prove one unreadable slow metric does not collapse neighboring readable rows.
- Inputs:
  - CME rows available
  - ETF row explicitly unavailable
- Verify:
  - panel still renders all rows
  - unavailable row keeps explicit fallback copy
  - available rows retain their values and timestamps

### 4. Fully Unavailable Slow Context

- Purpose: prove the dashboard route still loads when no trusted slow-context data is readable.
- Inputs:
  - symbol response with explicit unavailable slow-context block
- Verify:
  - panel header and advisory framing remain visible
  - panel fallback copy is explicit
  - summary cards, route warning hierarchy, and retry behavior remain intact

### 5. Mobile Readability

- Purpose: prove the advisory panel remains legible on the existing mobile dashboard route.
- Inputs:
  - one healthy or partial slow-context scenario on the mobile Playwright project
- Verify:
  - freshness badge, cadence label, and `as of` timestamp remain visible without hover
  - panel stays single-column and does not hide existing route warnings or focused-symbol context

## Required Negative Cases

- delayed CME publication with healthy realtime state
- stale slow-context row rendered as last-known-good with explicit stale note
- missing ETF row while CME remains readable
- fully unavailable slow-context block while current-state route still succeeds
- malformed slow-context contract rejected by decoder tests

## Exit Criteria

- slow-context rows render with service-owned cadence, freshness, and timestamp semantics only
- `Context only` framing remains persistent and unambiguous
- partial and unavailable slow-context states stay isolated to the advisory panel
- desktop and mobile smoke keep the dashboard route readable without changing realtime trust semantics
- results are written to `plans/slow-context-dashboard-panel/testing-report.md` during implementation and archived with the plan after completion
