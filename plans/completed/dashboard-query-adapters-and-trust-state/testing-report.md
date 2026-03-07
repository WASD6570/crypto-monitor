# Testing Report

- Feature: `dashboard-query-adapters-and-trust-state`
- Scope: `apps/web` adapter decoding, normalized trust mapping, retry/stale fallback behavior, and browser smoke coverage.

## Commands

- `pnpm --dir apps/web test -- --runInBand`
- `pnpm --dir apps/web build`
- `LD_LIBRARY_PATH="/home/personal/code/alerts/.playwright-deps/extracted/usr/lib/x86_64-linux-gnu:/home/personal/code/alerts/.playwright-deps/extracted/lib/x86_64-linux-gnu" pnpm --dir apps/web exec playwright test tests/e2e/dashboard-query-adapters-and-trust-state.spec.ts --project=chromium`
- `LD_LIBRARY_PATH="/home/personal/code/alerts/.playwright-deps/extracted/usr/lib/x86_64-linux-gnu:/home/personal/code/alerts/.playwright-deps/extracted/lib/x86_64-linux-gnu" pnpm --dir apps/web exec playwright test tests/e2e/dashboard-query-adapters-and-trust-state.spec.ts --project=mobile-chrome`

## Results

- `pnpm --dir apps/web test -- --runInBand` passed with 11/11 tests.
- `pnpm --dir apps/web build` passed and produced the Vite production bundle.
- Desktop Playwright smoke passed after supplying locally extracted Chromium runtime libraries through `LD_LIBRARY_PATH`.
- Mobile Playwright smoke passed with the same local library override.

## Notes

- Unit coverage now exercises decoder validation, partial-surface mapping, adapter-backed page rendering, and stale last-known-good fallback after a failed refresh.
- Browser smoke now routes mocked service-owned current-state payloads through the real fetch adapters instead of fixture-default page wiring.
- Playwright needed temporary local Ubuntu package extracts under `/home/personal/code/alerts/.playwright-deps/` because the session image does not provide `libnspr4.so`, `libnss3.so`, and related Chromium dependencies by default; the extracted runtime directory was removed after validation.
