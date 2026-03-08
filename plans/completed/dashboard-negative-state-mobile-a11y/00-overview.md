# Dashboard Negative State Mobile A11y

## Ordered Implementation Plan

1. Add a bounded negative-state presentation layer in `apps/web` so stale, degraded, unavailable, and partial-data conditions render with consistent warning hierarchy, copy treatment, and screen-reader-readable trust labels.
2. Tighten mobile reading order and keyboard-safe symbol/section navigation so the focused-symbol route remains usable on smaller screens and without pointer input.
3. Add deterministic component and browser coverage for negative states, mobile layout behavior, and accessibility-critical interactions; record results in `plans/dashboard-negative-state-mobile-a11y/testing-report.md`.

## Problem Statement

`plans/completed/dashboard-detail-panels-and-symbol-switching/` gives the dashboard a complete focused-symbol read path, but the current route still treats negative states mostly as visual trust chips and inline text. The next bounded risk is operator honesty and usability when data is stale, degraded, partial, or unavailable, especially on mobile and for keyboard or assistive-technology users who need explicit warnings, clear focus movement, and stable section access without the UI hiding critical trust context.

## Bounded Scope

- negative-state presentation rules for the existing dashboard rail, summary cards, and focused panels
- mobile-safe section and trust warning behavior inside the current dashboard route
- keyboard-safe symbol switching and section navigation within the existing URL-state model
- screen-reader-readable trust labels, warnings, and focused-symbol context for the current dashboard shell
- deterministic tests for negative-state rendering and mobile-safe interaction behavior

## Out Of Scope

- backend, schema, or service contract changes
- new dashboard information architecture, new panel families, or deeper analytics views
- a comprehensive fixture matrix across every dashboard state permutation; that remains the later `dashboard-fixture-smoke-matrix` slice
- slow-context integration, alerting UX, or historical/audit UI
- client-side derivation of market state, derivatives logic, or trust semantics beyond formatting and accessibility treatment

## Requirements

- Keep the work bounded to `dashboard-negative-state-mobile-a11y`; later child plans own the full fixture/mocked-response smoke matrix.
- Stay inside `apps/web` and preserve the React + TypeScript + Vite SPA boundary.
- Reuse the completed dashboard detail-panel seam from `plans/completed/dashboard-detail-panels-and-symbol-switching/`; do not redesign the route or create a second interaction model.
- Keep the dashboard read-only and service-trusting: negative-state labels, reasons, availability, and freshness still come from service-owned payloads.
- Make degraded, stale, partial, and unavailable conditions visually distinct without making healthy content unreadable.
- Keep warnings visible near the affected card or panel and near the route-level trust rail; do not hide critical trust reductions behind hover-only or collapsed affordances.
- Preserve both symbol summary cards on mobile, with tap targets at least 44px and active-state styling that remains obvious under degraded or unavailable conditions.
- Ensure symbol switching and section navigation are keyboard-safe, use semantic button behavior, and expose the current selection to assistive tech.
- Ensure screen readers can understand current trust state, focused symbol, active section, and critical warning text without inferring from visual layout alone.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-state`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`
- `apps/web/tests/e2e`

## Module Breakdown

### 1. Negative-State Presentation And Trust Copy

- Add shell-facing negative-state display rules and small presentational metadata that let components render warning banners, trust summaries, and panel-level fallback notes consistently.
- Keep the mapping seam in `apps/web/src/features/dashboard-state` so component code remains presentational.

### 2. Mobile Layout And Keyboard Interaction Hardening

- Refine the current summary strip, detail shell, and section navigation so mobile reading order remains practical and keyboard navigation stays obvious.
- Preserve route query params for `symbol` and `section`; accessibility work should strengthen the existing seam, not replace it.

### 3. Accessibility And Browser Validation

- Add focused component and browser checks for warning visibility, keyboard symbol switching, mobile section access, and screen-reader-readable trust/selection cues.
- Keep the smoke matrix minimal and high-signal so it proves the negative-state contract without expanding into the full final fixture matrix.

## Design Details

### Negative-State UI Direction

- Treat healthy state as the calm baseline and trust-reduced states as layered warnings.
- Prefer a restrained operator-console look: compact warning bands, stronger border/label contrast, and repeated trust cues at route and panel scope.
- Keep motion minimal; negative-state emphasis should come from hierarchy, color, and copy, not distracting animation.

### Warning Hierarchy

- Route-level rail: summarize the most important current trust reduction and stale/unavailable condition.
- Summary card: show compact symbol-specific warning text when one symbol is degraded, stale, or unavailable.
- Focused panel: show local fallback or degraded notes close to the affected panel title and metrics.
- Mobile: do not require scrolling past neutral-looking content before a warning becomes visible.

### Accessibility Expectations

- Symbol buttons and section buttons should expose current selection with semantic pressed/current cues and visible focus styles.
- Trust chips and warning text should have text equivalents that do not rely on color alone.
- Important route-level or panel-level warnings should be readable to screen readers in context with the affected symbol or section.
- Retry and section-navigation controls should stay reachable in a logical tab order.

### Mobile Expectations

- Keep the summary strip stacked and readable without clipping reason text or warning notes.
- Keep the active section obvious even when several panels are degraded or unavailable.
- Prefer sticky-feeling or repeated trust context only if it can be done without obscuring content; otherwise keep the warning copy near the top of the detail shell.

## Acceptance Criteria

- Another agent can implement negative-state and accessibility hardening without reopening the parent epic.
- The plan names exact repo areas for trust-copy mapping, shell components, mobile styles, and tests.
- The plan preserves the service-trusting boundary and does not require backend changes.
- The validation shape proves degraded/stale/unavailable honesty plus keyboard/mobile-safe interaction on the current dashboard route.

## ASCII Flow

```text
service-owned trust + reason metadata
          |
          v
existing dashboard mapper
  + negative-state display metadata
          |
          v
dashboard shell
  - route trust rail warnings
  - symbol card warning cues
  - focused panel fallback notes
  - keyboard/mobile-safe navigation
          |
          v
operator can read trust reductions quickly
on desktop, mobile, and assistive-tech paths
```

## Live-Path Boundary

- Services remain the source of truth for trust state, degraded reasons, freshness, and availability.
- This feature is frontend-only and limited to presentation, accessibility semantics, and interaction hardening.
- Python remains out of scope and out of the runtime path.
