# Implementation Step 3: Derivatives, Health, And Symbol Switching Behavior

## Requirements And Scope

- Finish the focused-symbol detail region by implementing the feed-health/regime panel and the explicit derivatives panel.
- Preserve instant-feeling symbol switching using the existing route-state and adapter cache behavior.
- Keep the feature frontend-only; if richer derivatives content requires a backend contract, leave that as an explicit future seam.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/features/dashboard-state`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`
- `apps/web/tests/e2e`

## Implementation Notes

- Health/regime panel should render service-owned trust context such as:
  - global ceiling state
  - focused symbol availability summary
  - degraded or missing-input reasons
  - freshness or last-success notes that help the operator decide whether the current detail view is safe to trust
- Derivatives panel should render as a first-class panel with focused-symbol framing, but keep its body explicitly unavailable until the contract exists.
- Use the existing route-state hook for `symbol` and `section`; do not add a second focused-symbol store.
- Preserve the current behavior where switching symbols can reuse already-fetched state. The UI should:
  - swap to the cached panel content immediately when available
  - keep the status rail and summary strip stable during the swap
  - show any in-flight refresh as a secondary note or trust-reduced panel state, not as a full shell reset
- Keep invalid URL query values normalized through the existing route-state helpers instead of adding custom panel-level guards.

## Test Expectations

- Add or extend browser smoke coverage in `apps/web/tests/e2e/visibility-dashboard-core.spec.ts` for:
  - healthy BTC to ETH switching with populated focused panels
  - degraded health/regime notes remaining visible after symbol changes
  - derivatives panel staying explicit and honest instead of disappearing
  - URL `symbol` persistence across switches and route reloads
- Keep unit or component tests focused on trust-state and panel rendering logic; reserve deeper mobile and accessibility coverage for `dashboard-negative-state-mobile-a11y`.

## Summary

This step completes the main operator read path: all four detail panels render intentionally, symbol switching stays responsive, and the dashboard remains honest about the missing derivatives contract.
