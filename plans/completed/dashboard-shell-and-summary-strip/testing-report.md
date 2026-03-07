# Dashboard Shell And Summary Strip Testing Report

## Commands

1. `pnpm --dir apps/web test -- --runInBand`
   - Result: PASS
   - Coverage notes: route render, BTC/ETH summary visibility, symbol switching, safe query-param fallback, degraded trust messaging, and partial section availability.

2. `pnpm --dir apps/web build`
   - Result: PASS
   - Coverage notes: TypeScript compile and production Vite build for the new shell route and fixture-backed feature modules.

3. `pnpm --dir apps/web exec playwright test tests/e2e/dashboard-shell-and-summary-strip.spec.ts --project=chromium`
   - Result: PASS
   - Coverage notes: desktop browser route smoke, summary-card visibility, and URL-state updates passed after supplying temporary local Chromium runtime libraries through `LD_LIBRARY_PATH`.

4. `pnpm --dir apps/web exec playwright test tests/e2e/dashboard-shell-and-summary-strip.spec.ts --project=mobile-chrome`
   - Result: PASS
   - Coverage notes: mobile browser route smoke, stacked summary layout, and section-query updates passed after supplying temporary local Chromium runtime libraries through `LD_LIBRARY_PATH`.

## Fixture And Scenario Notes

- Healthy route rendering uses a deterministic fixture-backed dashboard snapshot with BTC focused by default.
- Unit coverage also exercises degraded, stale, and partial-data shell states without asking the UI to derive market logic.
- Query-param normalization confirms invalid `symbol` and `section` values fall back to `BTC-USD` and `overview`.

## Notes And Assumptions

- Browser smoke validation passed by temporarily downloading Ubuntu runtime libraries locally and exporting `LD_LIBRARY_PATH`; the repo does not keep those extracted system libraries checked in.
- `playwright install --with-deps chromium` still cannot be used here because it requires `sudo` for system-package installation.
- The route remains fixture-backed by design; live query adapters are deferred to `dashboard-query-adapters-and-trust-state`.
- Service-owned labels, timestamps, degraded notes, and reason text remain presentation inputs only.
