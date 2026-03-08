# Implementation Step 3: Accessibility Checks And Mobile Smoke

## Requirements And Scope

- Prove the negative-state and mobile-safe interaction work with focused browser and component validation.
- Keep test scope bounded to accessibility-critical behavior and the current route, leaving the final broad smoke matrix for the later child plan.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/tests/e2e`
- optional shared test helpers under `apps/web/src/test`

## Implementation Notes

- Add or extend tests for:
  - keyboard navigation across symbol and section controls
  - visible focus state on interactive dashboard controls
  - degraded, stale, and unavailable warning visibility in mobile layout
  - screen-reader-readable trust and selection cues where component testing can assert them directly
- Prefer deterministic mocked current-state responses over ad hoc fixture-only rendering for browser coverage.
- Use mobile Playwright coverage for the stacked layout and desktop/unit tests for more granular semantics where appropriate.

## Browser Validation Shape

- Extend `tests/e2e/visibility-dashboard-core.spec.ts` or nearby files with a mobile-chrome scenario that verifies:
  - warning text is visible in the stacked layout
  - symbol switching remains reachable and obvious on mobile
  - active section and focused symbol remain understandable after route reload

## Summary

This step closes the feature with proof that negative-state honesty and accessibility-critical interaction behavior hold under the current dashboard route, without broadening into the full end-state smoke matrix.
