# Testing Plan: Dashboard Shell And Summary Strip

## Goals

- prove the dashboard route shell renders a stable operator reading order from fixtures
- prove BTC/ETH summary visibility and symbol-focus navigation work on desktop and mobile
- prove trust-state labels remain visible without client-side market-logic recomputation
- prove the fixture-backed shell stays implementation-ready while upstream query surfaces are still changing

## Expected Report Artifact

- `plans/completed/dashboard-shell-and-summary-strip/testing-report.md`

## Recommended Validation Commands

```bash
pnpm --dir apps/web test -- --runInBand
pnpm --dir apps/web build
pnpm --dir apps/web exec playwright test tests/e2e/dashboard-shell-and-summary-strip.spec.ts --project=chromium
pnpm --dir apps/web exec playwright test tests/e2e/dashboard-shell-and-summary-strip.spec.ts --project=mobile-chrome
```

## Smoke Matrix

### 1. Healthy Shell Render

- Input: healthy fixture-backed snapshot
- Verify:
  - the dashboard route renders the status rail first
  - BTC and ETH summary cards both render on first paint
  - BTC is focused by default
  - all four lower shell slots render with predictable titles
  - config version and freshness context are visible

### 2. Symbol Switching

- Input: healthy fixture-backed snapshot
- Verify:
  - activating ETH updates the focused-symbol header and query param
  - BTC summary remains visible while ETH is focused
  - keyboard and pointer interactions both work

### 3. Degraded Trust Messaging

- Input: degraded fixture where a service-owned venue or timestamp note weakens trust
- Verify:
  - the status rail shows a global degraded note
  - the affected summary card shows degraded or stale wording from fixture data
  - the UI does not recompute or relabel market state beyond what the fixture provides

### 4. Partial Shell Availability

- Input: partial fixture where one lower section is unavailable
- Verify:
  - the route still renders the shell and summary strip
  - the unavailable slot is labeled explicitly
  - the page does not imply the missing section is neutral or healthy

### 5. Mobile Layout And Navigation

- Input: healthy and degraded fixtures on a mobile viewport
- Verify:
  - summary cards stack vertically with tap-safe controls
  - section switching updates the `section` query param
  - trust labels remain near the top of the active mobile view
  - no critical content requires horizontal scrolling

## Negative Cases

- invalid `symbol` query param causes a safe fallback to `BTC-USD`
- invalid `section` query param causes a safe fallback to `overview`
- a stale or degraded fixture never hides the last-known timestamp
- missing lower-section data does not collapse the whole route when top-level shell data exists
- summary cards never infer a new state label from missing or degraded metadata

## Determinism And Safety Checks

- keep fixtures static and repeatable; do not generate runtime timestamps in tests
- confirm tests assert service-supplied labels and reasons rather than derived client labels
- confirm the route builds without charting or heavyweight dashboard dependencies
- confirm mobile and desktop checks use the same route shape and trust vocabulary

## Exit Criteria

- All recommended validation commands pass.
- The generated testing report at `plans/completed/dashboard-shell-and-summary-strip/testing-report.md` records commands, fixture inputs, pass/fail results, and any deviations.
- Another agent can rerun the same commands and reach the same shell, navigation, and trust-state outcomes.
