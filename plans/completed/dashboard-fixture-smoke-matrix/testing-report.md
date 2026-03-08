# Dashboard Fixture Smoke Matrix Testing Report

## Result

- Status: passed
- Date: 2026-03-08

## Validation Commands

```bash
pnpm --dir apps/web test -- --runInBand
pnpm --dir apps/web build
LD_LIBRARY_PATH="/home/personal/code/alerts/.playwright-deps/extracted/usr/lib/x86_64-linux-gnu:/home/personal/code/alerts/.playwright-deps/extracted/lib/x86_64-linux-gnu" pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium
LD_LIBRARY_PATH="/home/personal/code/alerts/.playwright-deps/extracted/usr/lib/x86_64-linux-gnu:/home/personal/code/alerts/.playwright-deps/extracted/lib/x86_64-linux-gnu" pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome
```

## Notes

- Vitest passed with the shared dashboard scenario catalog, mapper matrix, shell matrix, and decoder coverage.
- The Vite production build passed with `tsc --noEmit` and emitted the updated SPA bundle successfully.
- Desktop Playwright passed healthy, degraded, partial, unavailable, stale-retry, and keyboard route-state smoke checks.
- Mobile Playwright passed healthy, degraded, partial, unavailable, and route-warning reload smoke checks.
- Chromium still required temporary local Ubuntu package extracts under `.playwright-deps/`; validation used `LD_LIBRARY_PATH` and the temporary dependency directory was removed afterward.

## Coverage Highlights

- One named scenario seam now drives healthy, degraded, stale, partial, and unavailable dashboard states across Vitest and Playwright.
- Browser smoke now proves degraded timestamp-trust copy, unavailable focused-symbol fallback, stale last-known-good retry behavior, and existing desktop/mobile route persistence.
- The final dashboard-core matrix stays bounded to route honesty, warning hierarchy, focused selection, and explicit fallback behavior without inventing client-owned market logic.
