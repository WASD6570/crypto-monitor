# UI Delivery And Operator Workflow

## Module Goal

Translate the dashboard architecture and query surfaces into an implementation plan for a trustworthy operator workflow inside the Vite SPA.

## Target Repo Areas

- `apps/web/src/app`
- `apps/web/src/pages`
- `apps/web/src/features`
- `apps/web/src/state`
- `apps/web/src/hooks`
- `apps/web/src/components`
- `tests/e2e`

## Requirements

- Deliver a first-load experience that gets the operator to a trustworthy current-state read in under 60 seconds.
- Make BTC and ETH switching instant-feeling once initial data is present.
- Keep critical trust information visible during loading, errors, staleness, and degraded-feed conditions.
- Preserve a mobile-safe experience without downgrading the screen into a sparse consumer dashboard.
- Keep state management simple and local to this feature unless a shared app-level pattern is clearly needed.

## Ordered Implementation Steps

1. Create the route shell and page composition for the dashboard.
2. Add the top status rail and symbol overview strip.
3. Add focused symbol detail sections for overview and microstructure.
4. Add derivatives context and feed health/regime sections.
5. Add loading, stale-state, degraded, and unavailable overlays or banners.
6. Add responsive mobile collapse behavior and keyboard/screen-reader affordances.
7. Add fixture-backed stories/tests and end-to-end smoke coverage.

## Operator Workflow Plan

### Entry

- The operator lands directly on the dashboard route.
- The page renders a shell immediately.
- The first complete screen should answer, at minimum:
  - current symbol states
  - freshness of the view
  - whether any critical feed degradation is active

### Primary Read Path

The default path should require no customization:

1. read the status rail
2. compare BTC and ETH state cards
3. focus the more interesting symbol
4. inspect microstructure and derivatives context
5. confirm feed health and regime explanation before acting on the view

### Ongoing Monitoring

- automatic refresh or streaming updates should update timestamps and affected panels without full-page reflow
- notable degraded-state changes should be visually elevated but not trigger modal interruptions
- the operator should be able to leave the screen open for long sessions without memory-heavy or animation-heavy behavior

## UI State Model

At planning level, each panel should support the same bounded state machine:

- `loading`
- `ready`
- `stale`
- `degraded`
- `unavailable`

Guidance:

- `stale` means last known good data is shown with trust reduction
- `degraded` means the data is current enough to show but upstream conditions reduce confidence
- `unavailable` means the panel cannot provide a safe reading and should stop pretending otherwise

## Interaction Rules

### Symbol Switching

- switching between BTC and ETH should reuse cached detail data when still fresh enough
- the inactive symbol summary remains visible so cross-symbol comparison is never lost
- on mobile, symbol switching must not reset scroll unexpectedly or hide current warnings

### Panel Expansion And Collapse

- desktop can keep all four core views visible together
- mobile can use accordions or tabs, but the feed/regime warnings must remain pinned or repeated near the active section
- avoid deep nesting or drill-down patterns that obscure the first-user workflow

### Timestamps And Provenance

- each panel should show its own timestamp context
- the global rail should show the oldest critical data age so the operator sees if one panel is lagging the rest
- config version or ruleset provenance should be visible in a compact footer or rail detail

## Degraded And Negative-State UX

### Degraded Feed Messaging

- show venue names and affected symbol scope when available
- explain effect on trust in plain terms, such as `Coinbase stale; USA confirmation weakened`
- degrade optimism, not visibility: keep prior good data visible if safe, but mark it clearly

### Partial Data

- if derivatives context is missing while overview is healthy, the screen remains usable and says the derivatives panel is unavailable
- if feed health is missing, the dashboard should elevate a trust warning because core interpretability is impaired
- if symbol current state is missing for one symbol, keep the other symbol usable and isolate the failure explicitly

### Severe Failure

- if the snapshot surface fails completely, render a full-page unavailable state with retry control and last-success timestamp if known
- avoid generic `something went wrong` messaging; include the failed surface and whether other local cached data exists

## Mobile-Safe Dense Delivery Rules

- prioritize compact typography and short labels over oversized chart chrome
- keep key chips and warning labels within the initial viewport of each section
- use fixed-height cards sparingly; let warning content expand naturally
- if horizontal tables are needed for venue health, provide a stacked mobile alternative rather than forcing horizontal scrolling for critical values

## Accessibility And Usability Cases

- keyboard users can switch symbols and move through sections in logical order
- screen readers hear state, freshness, and degradation before dense metric lists
- color-blind users can distinguish healthy/degraded/stale states via text, icon, and placement, not color alone
- reduced-motion users should not depend on animation to detect state changes

## Implementation Risks And Guardrails

- avoid chart-library adoption before a fixture-backed no-chart layout proves sufficient
- avoid centralizing all remote state if local feature-scoped hooks and caches keep the design simpler
- avoid bespoke client normalization layers that duplicate service contracts
- reserve optional seams for later alert drill-down and slow context integration without prebuilding them now

## End-To-End Test Expectations

- first load renders both symbol summaries and at least one focused symbol detail using fixture or mocked service data
- symbol switching updates detail sections without losing trust warnings
- stale and degraded states remain visible after refresh failures
- mobile viewport keeps critical warnings and state chips accessible

## Summary

This module turns the data plan into an operator workflow: open the route, read the global trust cues, compare BTC and ETH, inspect one symbol in depth, and keep monitoring even when upstream conditions degrade. The implementation should stay compact, explicit, and service-trusting rather than clever.
