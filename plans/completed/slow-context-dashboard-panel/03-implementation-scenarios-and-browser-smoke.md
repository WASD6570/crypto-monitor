# Scenarios And Browser Smoke

## Module Requirements And Scope

Target repo areas:

- `apps/web/src/test/dashboardScenarioCatalog.ts`
- `apps/web/src/test/dashboardScenarioCatalog.test.ts`
- `apps/web/tests/e2e/dashboardScenarioHelpers.ts`
- `apps/web/tests/e2e/visibility-dashboard-core.spec.ts`
- `apps/web/src/features/dashboard-shell/components/DashboardShell.test.tsx`
- `apps/web/src/features/dashboard-state/dashboardStateMapper.test.ts`

This module locks the slow-context dashboard behavior into the same deterministic scenario system that already covers the core route.

In scope:

- extending existing dashboard scenario fixtures with slow-context states
- focused mapper and shell tests for advisory panel behavior
- desktop and mobile Playwright smoke for the new panel

Out of scope:

- a large Cartesian-product matrix across every current-state and slow-context permutation
- duplicating fixtures outside the shared scenario catalog

## Scenario Strategy

- Keep the current dashboard scenario names as the primary route states and layer slow-context overrides onto them where needed.
- Add one small helper seam that can inject slow-context row states into the existing symbol fixtures without changing the underlying realtime market-state scenario.
- Cover this bounded set of slow-context outcomes:
  - all rows fresh
  - delayed CME with otherwise healthy route data
  - stale last-known-good row with explicit stale note
  - partial availability where ETF is unavailable and CME rows remain readable
  - full unavailable panel while current-state route still succeeds

## Assertion Focus

- `Context only` badge is always visible when the dashboard shell is visible.
- row freshness and cadence labels are visible in both component tests and browser smoke.
- slow-context errors or unavailable rows never alter summary-card trust, route warnings, or focused realtime panel copy.
- mobile layout keeps `as of` labels and freshness badges readable for the focused symbol.

## Browser Smoke Guidance

- Extend the existing `visibility-dashboard-core.spec.ts` file instead of creating a disconnected spec unless the added coverage becomes too noisy.
- Keep one desktop smoke for healthy slow context and one desktop smoke for partial or unavailable fallback.
- Keep one mobile smoke proving the panel remains readable and does not push critical route warnings out of view.
- Reuse `dashboardScenarioHelpers.ts` so Playwright keeps mocking one coherent dashboard response family.

## Summary

This module proves the new advisory panel with the same deterministic scenario vocabulary already used for the dashboard route. It keeps the test surface bounded while still locking in the honesty rules that matter for slow context.
