# Testing Plan

## Testing Goal

Prove the dashboard-core route has one deterministic, reusable smoke matrix for healthy, degraded, stale, partial, and unavailable states across mapper, component, desktop-browser, and mobile-browser coverage.

## Output Artifact

- Record implementation-time results in `plans/completed/dashboard-fixture-smoke-matrix/testing-report.md` in the archived feature directory.

## Required Validation Commands

### 1. Unit And Component Coverage

```bash
pnpm --dir apps/web test -- --runInBand
```

Purpose:

- validate deterministic scenario helpers, mapper reduction, shell warning hierarchy, and route-backed selection semantics

### 2. Dashboard Build Smoke

```bash
pnpm --dir apps/web build
```

Purpose:

- confirm the Vite SPA still builds cleanly after smoke-matrix helper and test changes

### 3. Desktop Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium
```

Purpose:

- verify the shared scenario matrix holds in the composed desktop route

### 4. Mobile Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome
```

Purpose:

- verify the shared scenario matrix holds in the stacked mobile route

## Scenario Matrix

### A. Healthy Baseline

Verify:

- route trust remains calm and readable
- both symbol summaries stay visible
- no scenario helper injects warning copy that is not justified by the payload

### B. Degraded Timestamp-Trust Reduction

Verify:

- route warning, focused-symbol warning, and affected panel warning remain visible
- timestamp fallback or degraded-trust copy stays text-readable
- unaffected peer surfaces remain readable

### C. Stale Last-Known-Good Refresh Failure

Verify:

- stale trust labels remain explicit
- last-known-good copy stays attached to the affected summary or panel
- retry and route controls remain usable

### D. Partial And Unavailable Surface Fallback

Verify:

- partial input scenarios keep comparison and panel fallback honest
- unavailable surfaces stay explicit without blanking the whole route
- safe neighboring summaries or panels remain visible

### E. Route Persistence Across Desktop And Mobile

Verify:

- `symbol` and `section` query params remain aligned with visible state
- reload preserves focused symbol and active section in supported scenarios
- mobile layout keeps warnings and both summary cards visible

## Negative Cases

- stale or degraded scenarios silently collapsing back to healthy-looking copy
- Playwright and Vitest using different scenario semantics for the same named state
- unavailable surfaces disappearing instead of rendering explicit fallback messaging
- route query state drifting from focused symbol or active section after scenario setup or reload
- browser smoke relying on ad hoc inline payload edits instead of the shared scenario seam

## Implementation-Time Review Checklist

- confirm scenario names map to service-owned payload conditions rather than invented UI states
- confirm no production dashboard logic changes are required to support the matrix
- confirm unit/component/browser layers reuse the same scenario helpers where practical
- confirm desktop and mobile coverage remain bounded and maintainable
- confirm the final dashboard-core archive path is ready once implementation completes

## Execution Notes

- If Playwright system libraries remain unavailable in the environment, reuse the local-library extraction plus `LD_LIBRARY_PATH` workaround already documented in `plans/completed/dashboard-negative-state-mobile-a11y/testing-report.md`.

## Exit Criteria

- unit/component, build, desktop, and mobile validation commands pass
- the scenario catalog is deterministic and reused across the intended test layers
- route honesty, warning hierarchy, and route-backed selection are proven for healthy, degraded, stale, partial, and unavailable paths
- implementation evidence is written to `plans/completed/dashboard-fixture-smoke-matrix/testing-report.md`
