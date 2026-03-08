# Implementation Step 3: Browser Smoke Matrix And Reporting

## Requirements And Scope

- Extend desktop and mobile Playwright coverage around the shared scenarios.
- Finish the dashboard-core epic with one maintainable browser matrix and recorded validation evidence.

## Target Repo Areas

- `apps/web/tests/e2e`
- `apps/web/playwright.config.ts`
- `plans/completed/dashboard-fixture-smoke-matrix/testing-report.md`

## Implementation Notes

- Reuse the shared scenario and mocking helpers from Step 1 instead of repeating route stubs per spec.
- Keep the browser matrix bounded to the current dashboard route and its key operator journeys.
- Cover route reload behavior, warning visibility, active selection cues, and panel fallback honesty on both desktop and mobile.
- If a helper split makes the browser spec easier to scan, prefer that over one oversized spec file.
- Record implementation-time results and any environment-specific Playwright notes in the testing report.

## Browser Test Expectations

- Desktop: healthy baseline, degraded or partial warning visibility, keyboard or route-backed selection persistence.
- Mobile: stacked summary visibility, route warning visibility, active section clarity, and reload-safe route state.
- Negative case proof: unavailable or stale surfaces remain explicit and do not imply healthy defaults.

## Summary

This step closes the feature with final browser-level proof that the dashboard-core route stays deterministic, honest, and maintainable across its supported scenario matrix.
