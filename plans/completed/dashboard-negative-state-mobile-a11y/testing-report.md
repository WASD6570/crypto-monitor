# Dashboard Negative State Mobile A11y Testing Report

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

- Unit and component coverage passed with 14 assertions across dashboard mapper and shell behavior.
- Production build passed with `tsc --noEmit` and Vite bundle output.
- Desktop Playwright passed the focused-panel smoke plus the new keyboard route-state test.
- Mobile Playwright passed the stacked warning-hierarchy and route-reload smoke.
- Playwright required temporary local Ubuntu shared-library extraction under `.playwright-deps/` to provide Chromium runtime libraries in this environment.

## Coverage Highlights

- Route rail, summary cards, focused-symbol shell, and panel cards now render explicit warning text for degraded, stale, partial, and unavailable states.
- Summary and section controls expose current selection semantics and preserve `symbol` / `section` query params across reloads.
- Mobile coverage confirms both summary cards remain visible while focused warning copy and section state stay readable.
