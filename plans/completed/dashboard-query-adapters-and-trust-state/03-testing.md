# Testing Plan

## Testing Goal

Prove that the dashboard query adapter layer in `apps/web` consumes service-owned current-state payloads read-only, preserves trust metadata through normalization, and keeps the existing dashboard shell honest when surfaces are healthy, stale, degraded, partial, or unavailable.

## Output Artifact

- Record implementation-time results in `plans/completed/dashboard-query-adapters-and-trust-state/testing-report.md`.

## Required Validation Commands

### 1. Unit And Component Coverage

```bash
pnpm --dir apps/web test -- --runInBand
```

Purpose:

- validate decoder behavior, normalization, page-level fallback mapping, and shell trust rendering with mocked responses

### 2. Dashboard Build Smoke

```bash
pnpm --dir apps/web build
```

Purpose:

- confirm the Vite SPA builds cleanly after replacing fixture-default wiring with adapter-backed state

### 3. Desktop Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/dashboard-query-adapters-and-trust-state.spec.ts --project=chromium
```

Purpose:

- verify the adapter-backed dashboard route renders, symbol switching preserves trust cues, and partial/unavailable states stay explicit on desktop

### 4. Mobile Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/dashboard-query-adapters-and-trust-state.spec.ts --project=mobile-chrome
```

Purpose:

- verify the same trust and fallback behavior remains visible in the stacked mobile layout without expanding into later negative-state polish work

## Mocking Strategy

- Use deterministic mocked service responses at the adapter boundary rather than shell-only fixtures for the main implementation tests.
- Keep payloads aligned with the logical surfaces from `plans/completed/market-state-current-query-contracts/`:
  - dashboard snapshot
  - symbol detail
  - derivatives context
  - feed health and regime
- Preserve at least one payload with timestamp fallback or degraded-source markers so tests prove the UI surfaces service-owned trust notes.
- Keep browser smoke mocks lightweight and route-local; this feature does not need deep history or analytics fixtures.

## Smoke Matrix

### A. Healthy Initial Snapshot

Verify:

- `/dashboard` renders the rail and both BTC/ETH summaries from mocked adapter responses
- the focused symbol shows shell slots populated from normalized `ready` state
- service timestamps, freshness labels, and config/version metadata remain visible

### B. Symbol Cache Reuse

Verify:

- switching from BTC to ETH keeps a populated detail shell when ETH data is already cached
- the background refresh does not remove global degraded or stale notes
- the inactive symbol summary remains visible throughout the switch

### C. Degraded Trust Propagation

Verify:

- service-provided degraded reasons appear in the status rail or relevant section
- timestamp fallback markers remain visible without changing the service-supplied state label
- the UI never upgrades a degraded payload to healthy-looking presentation

### D. Stale After Refresh Failure

Verify:

- after one successful load, a failed refresh moves the affected surface to `stale`
- last-known-good content remains visible with explicit age and reduced-trust wording
- prolonged or severe staleness transitions to `unavailable` for the affected surface

### E. Partial And Unavailable Surfaces

Verify:

- missing derivatives context does not block overview or summary rendering
- missing feed-health/regime data elevates a trust warning while leaving safe summary data visible
- malformed or incomplete critical snapshot fields result in `unavailable`, not silent defaults

## Negative Cases

- malformed critical payload accepted as `ready`
- service timestamp fallback note dropped during normalization
- stale data rendered with healthy styling after refresh failure
- one unavailable lower surface incorrectly blanks the whole dashboard after initial success
- client-generated state labels or reason strings disagree with the service payload

## Implementation-Time Review Checklist

- confirm all market-state labels and regime/degradation reason families come from service payloads
- confirm adapter tests inject deterministic time so stale transitions are stable across runs
- confirm no new dependency adds unnecessary bundle weight for basic polling, caching, or decoding
- confirm out-of-scope seams remain deferred: no history route, no analytics panels, no detail-panel redesign

## Exit Criteria

- unit/component and browser smoke commands pass
- trust states map deterministically from mocked service responses
- the dashboard remains readable when one surface degrades or disappears
- implementation evidence is written to `plans/completed/dashboard-query-adapters-and-trust-state/testing-report.md`
