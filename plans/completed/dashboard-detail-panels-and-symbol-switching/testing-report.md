# Testing Report

- Feature: `dashboard-detail-panels-and-symbol-switching`
- Scope: focused detail-panel view models, dashboard shell panel rendering, BTC/ETH symbol switching, and desktop browser smoke coverage.

## Commands

- `pnpm --dir apps/web test -- --runInBand`
- `pnpm --dir apps/web build`
- `LD_LIBRARY_PATH="/home/personal/code/alerts/.playwright-deps/extracted/usr/lib/x86_64-linux-gnu:/home/personal/code/alerts/.playwright-deps/extracted/lib/x86_64-linux-gnu" pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium`

## Results

- `pnpm --dir apps/web test -- --runInBand` passed with 13/13 tests.
- `pnpm --dir apps/web build` passed and produced the Vite production bundle.
- Desktop Playwright smoke for `tests/e2e/visibility-dashboard-core.spec.ts` passed after supplying locally extracted Chromium runtime libraries through `LD_LIBRARY_PATH`.

## Notes

- The focused detail region now renders overview, microstructure, derivatives-gap, and health/regime panels from the existing service-owned current-state seam without adding client-side market logic.
- Unit and component coverage now checks focused-panel mapping, stale fallback notes, shell rendering, and symbol-switch panel updates.
- Browser smoke now verifies focused panel composition, route-backed symbol persistence, and degraded panel honesty using mocked current-state responses.
- This environment still lacks Chromium runtime libraries such as `libnspr4.so` by default; validation used temporary local Ubuntu package extracts under `/home/personal/code/alerts/.playwright-deps/`, then removed those temporary artifacts after the smoke run completed.
