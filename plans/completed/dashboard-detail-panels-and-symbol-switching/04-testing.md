# Testing Plan

## Testing Goal

Prove that the dashboard detail panels render from the existing service-owned current-state seam, BTC/ETH switching keeps the focused region populated without hiding peer summaries, and unavailable or degraded panel states remain explicit instead of falling back to guessed client logic.

## Output Artifact

- Record implementation-time results in `plans/completed/dashboard-detail-panels-and-symbol-switching/testing-report.md`.

## Required Validation Commands

### 1. Unit And Component Coverage

```bash
pnpm --dir apps/web test -- --runInBand
```

Purpose:

- validate focused-panel mapping, panel rendering, and symbol-switch behavior with deterministic mocked responses or injected fixtures

### 2. Dashboard Build Smoke

```bash
pnpm --dir apps/web build
```

Purpose:

- confirm the Vite SPA still builds cleanly after replacing shell placeholders with real detail panels

### 3. Desktop Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium
```

Purpose:

- verify focused panels render on the dashboard route, symbol switching feels immediate on desktop, and trust cues remain visible during degraded or unavailable states

## Mocking Strategy

- Reuse the existing dashboard mocked-response seam in `apps/web` rather than introducing backend dependencies.
- Keep the browser smoke focused on the already-supported current-state contracts:
  - global current state
  - BTC symbol current state
  - ETH symbol current state
- Include at least one degraded or partial payload so the health and microstructure panels prove they keep service-owned trust notes visible.
- Keep the derivatives panel mocked through the existing contract gap: the test should prove honest unavailable rendering, not synthesize derivatives data.

## Smoke Matrix

### A. Healthy Focused Panel Render

Verify:

- `/dashboard` renders overview, microstructure, derivatives, and health/regime panels for the default focused symbol
- BTC and ETH summary cards remain visible while the focused panel region renders BTC by default
- status rail freshness and version metadata remain visible

### B. Cached Symbol Switching

Verify:

- switching from BTC to ETH updates the focused heading and panel content immediately when ETH data is already cached
- the summary strip stays visible during the switch
- background refresh does not blank the panel bodies after initial success

### C. Degraded Microstructure Or Health

Verify:

- missing-input or degraded reasons appear inside the relevant panel
- trust chips stay aligned with the existing `loading` / `ready` / `stale` / `degraded` / `unavailable` vocabulary
- the UI never upgrades a degraded payload to healthy-looking copy

### D. Honest Derivatives Gap

Verify:

- the derivatives panel renders for both BTC and ETH with focused-symbol framing
- the panel states plainly that no derivatives context is present in the current contract
- unavailable derivatives content does not blank overview, microstructure, or health/regime panels

### E. Route Persistence

Verify:

- changing symbols updates the `symbol` query param
- reloading the route preserves the focused symbol
- invalid `symbol` or `section` query values normalize back to the default route state

## Negative Cases

- focused panel content inferred from missing contract fields
- symbol switch clears already-safe cached content before revalidation completes
- degraded health or microstructure notes disappear when another panel is healthy
- derivatives panel silently vanishes because data is unavailable
- panel rendering changes the existing route-state contract or hides the summary strip

## Implementation-Time Review Checklist

- confirm no new client-side market-state or derivatives logic was introduced
- confirm panel metrics are formatted from service-owned fields rather than recomputed summaries
- confirm the desktop layout remains dense and mobile behavior remains single-column practical
- confirm no new dependency is added for basic panel rendering or symbol switching
- confirm later work remains deferred: mobile a11y polish and the full fixture smoke matrix stay in their own child plans

## Execution Notes

- If Playwright system libraries are still unavailable in the environment, reuse the existing local-library extraction and `LD_LIBRARY_PATH` workaround already documented in prior dashboard testing artifacts.

## Exit Criteria

- unit/component, build, and desktop browser smoke commands pass
- focused panels render deterministically from mocked current-state payloads
- symbol switching keeps the focused region readable after initial success
- derivatives remain honestly unavailable until upstream contracts change
- implementation evidence is written to `plans/completed/dashboard-detail-panels-and-symbol-switching/testing-report.md`
