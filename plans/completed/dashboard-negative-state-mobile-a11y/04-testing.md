# Testing Plan

## Testing Goal

Prove that the dashboard communicates stale, degraded, unavailable, and partial-data states honestly across desktop and mobile, while symbol switching and section navigation remain keyboard-safe and screen-reader-readable on the current route.

## Output Artifact

- Record implementation-time results in `plans/dashboard-negative-state-mobile-a11y/testing-report.md`.

## Required Validation Commands

### 1. Unit And Component Coverage

```bash
pnpm --dir apps/web test -- --runInBand
```

Purpose:

- validate negative-state mapping, warning rendering, keyboard interaction behavior, and accessibility-oriented component semantics

### 2. Dashboard Build Smoke

```bash
pnpm --dir apps/web build
```

Purpose:

- confirm the Vite SPA still builds cleanly after accessibility and mobile-state presentation changes

### 3. Desktop Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium
```

Purpose:

- verify negative-state warnings and keyboard-safe interaction stay intact on desktop

### 4. Mobile Browser Smoke

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome
```

Purpose:

- verify stacked mobile layout, visible warning hierarchy, and route-backed symbol/section behavior on a mobile viewport

## Mocking Strategy

- Reuse the existing dashboard mocked-response seam in `apps/web`.
- Keep coverage focused on current-state responses already supported by the dashboard:
  - global current state
  - BTC symbol current state
  - ETH symbol current state
- Include at least one degraded and one partial/unavailable scenario so the feature proves honest warning behavior under trust reduction.

## Smoke Matrix

### A. Desktop Degraded State Honesty

Verify:

- route-level trust warning remains visible when one symbol or panel is degraded
- summary-card warning text stays visible next to the affected symbol
- panel-level fallback or degraded notes remain close to the affected content

### B. Keyboard Symbol And Section Navigation

Verify:

- keyboard interaction can move through symbol and section controls in a predictable order
- current selection remains visible and semantically exposed
- switching symbol or section keeps route query params aligned with the visible state

### C. Mobile Warning Visibility

Verify:

- stacked mobile layout keeps trust warnings visible without relying on hover
- active symbol and active section remain obvious when panels are degraded or unavailable
- retry and navigation controls remain reachable and readable on mobile

### D. Screen-Reader-Oriented Trust Cues

Verify:

- trust chips, warning text, and current selection expose readable text equivalents
- focused symbol and active section are understandable without visual context alone

## Negative Cases

- degraded or stale panels relying on color alone without readable warning text
- mobile layout pushing the first critical warning below unrelated neutral content
- keyboard focus becoming ambiguous or invisible on symbol/section controls
- route query state drifting from the visible focused symbol or active section
- unavailable surfaces hiding warnings or implying healthy defaults

## Implementation-Time Review Checklist

- confirm no new market-state or derivatives logic was added client-side
- confirm warning hierarchy stays consistent across rail, summaries, and panels
- confirm mobile changes do not collapse dense operator content into unclear cards
- confirm no heavy dependency was added for accessibility or focus handling
- confirm the later `dashboard-fixture-smoke-matrix` work remains deferred

## Execution Notes

- If Playwright system libraries are still unavailable in the environment, reuse the established local-library extraction and `LD_LIBRARY_PATH` workaround from prior dashboard testing artifacts.

## Exit Criteria

- unit/component, build, desktop, and mobile validation commands pass
- negative-state warnings remain explicit and readable across route scopes
- symbol and section controls stay keyboard-safe and route-consistent
- implementation evidence is written to `plans/dashboard-negative-state-mobile-a11y/testing-report.md`
