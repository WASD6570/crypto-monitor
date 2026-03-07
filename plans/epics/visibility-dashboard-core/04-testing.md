# Testing Plan

## Testing Goal

Validate that the visibility dashboard renders trusted current-state information for BTC and ETH, remains honest about freshness and degraded conditions, and stays usable on desktop and mobile without moving market logic into the client.

## Output Artifact

- Record implementation-time results in `plans/epics/visibility-dashboard-core/testing-report.md`.

## Required Validation Commands

These commands are the minimum expected validation entry points for the eventual implementation. If tool names differ slightly once `apps/web` is fully scaffolded, the implementing agent should preserve the intent and update this plan only if necessary.

### 1. Unit And Component Coverage

Command:

```bash
pnpm --dir apps/web test -- --runInBand
```

Purpose:

- validate component rendering, state adapters, panel state transitions, and accessibility-focused unit behavior

### 2. Dashboard Build Smoke

Command:

```bash
pnpm --dir apps/web build
```

Purpose:

- confirm the Vite SPA builds cleanly with dashboard route code, lazy boundaries, and client contract adapters

### 3. Desktop End-To-End Smoke

Command:

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=chromium
```

Purpose:

- verify first-load flow, BTC/ETH switching, and core panel visibility on a desktop viewport

### 4. Mobile End-To-End Smoke

Command:

```bash
pnpm --dir apps/web exec playwright test tests/e2e/visibility-dashboard-core.spec.ts --project=mobile-chrome
```

Purpose:

- verify stacked layout, section access, visible trust warnings, and tap-target-safe symbol switching on mobile

## Fixture And Mocking Strategy

- use deterministic fixture-backed responses derived from the visibility initiative contracts
- include at least one healthy snapshot and one degraded snapshot for each symbol
- include a partial-data fixture where one panel is unavailable but the rest of the dashboard remains usable
- include a timestamp-fallback fixture where services mark an event or block as degraded due to `recvTs` fallback semantics

## Smoke Matrix

### A. Healthy Current-State Render

Verify:

- both symbol overview summaries render on first load
- BTC default focus shows overview, microstructure, derivatives, and feed/regime panels
- timestamps, freshness labels, and config/version metadata are visible
- no panel invents missing logic client-side

### B. Symbol Switching

Verify:

- switching from BTC to ETH updates detail panels
- the non-focused symbol summary remains visible
- switching preserves the current degraded or stale banners when they apply globally

### C. Feed Degradation

Verify:

- a degraded venue row is visible with venue name and reason
- symbol tradeability remains whatever the service returns; the client only presents it
- the dashboard shows trust reduction language such as weakened confirmation or degraded health

### D. Stale State

Verify:

- when refresh fails after a successful load, last known data remains visible temporarily
- the affected panel or page enters `stale` with explicit age and retry affordance if applicable
- severe staleness eventually transitions to `unavailable` based on implementation thresholds rather than silently persisting old state forever

### E. Partial Data

Verify:

- missing derivatives context does not block the overview or health panels
- missing feed-health data elevates a trust warning because core interpretability is impaired
- partial snapshot completeness is visibly labeled

### F. Accessibility

Verify:

- keyboard navigation reaches symbol controls and mobile section controls in order
- regime and health states are understandable without color alone
- screen-reader labels include state and freshness wording for primary summary chips
- contrast remains acceptable for healthy, degraded, stale, and unavailable tokens

## Negative Cases

### Freshness Negative Cases

- stale payload age shown incorrectly or omitted
- timestamp-degraded payload shown as normal without trust note
- mixed panel ages hidden by a single misleading global timestamp

### Degraded-State Negative Cases

- venue degradation shown only in logs or console output
- UI recomputes state severity from raw metrics and disagrees with the service payload
- missing venue health displayed as healthy by default

### Accessibility Negative Cases

- mobile tabs or accordions hide active warnings from screen readers
- keyboard focus is lost after symbol switching
- chips rely on color alone to distinguish `WATCH` and degraded health

## Implementation-Time Review Checklist

- confirm the dashboard uses service-supplied labels and reasons for regime and degradation
- confirm current-state surfaces remain within the under-2-second target in local mocked or dev conditions where practical
- confirm bundle additions are justified and heavy charting is deferred unless proven necessary
- confirm the page stays usable when one logical surface fails independently

## Exit Criteria

- desktop and mobile smoke checks pass
- degraded, stale, and unavailable states render explicitly
- fixture-backed current-state scenarios match service outputs without client re-derivation
- accessibility checks cover summary state, warnings, and navigation across dense layouts
