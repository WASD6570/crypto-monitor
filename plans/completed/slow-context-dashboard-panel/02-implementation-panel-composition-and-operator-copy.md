# Panel Composition And Operator Copy

## Module Requirements And Scope

Target repo areas:

- `apps/web/src/features/dashboard-shell/components/DashboardShell.tsx`
- `apps/web/src/pages/dashboard/DashboardPage.tsx`
- `apps/web/src/styles.css`

This module defines how the dashboard shell renders slow context once the focused-symbol response already includes the completed service-owned advisory block.

In scope:

- rendering one dedicated slow-context panel in the dashboard shell
- panel header, `Context only` badge, helper copy, and row composition
- isolated loading, partial, stale, delayed, and unavailable states inside the panel
- desktop and mobile layout rules for the panel and row stack

Out of scope:

- backend changes beyond the already-bounded response seam
- modifying existing summary-strip or route-warning hierarchy
- new charts, drill-down interactions, or analytics affordances

## Placement Decision

- Keep the current dashboard structure intact and add the slow-context panel immediately after the focused-symbol detail region.
- This preserves the established operator order:
  1. route trust and summary cards
  2. realtime focused panels
  3. advisory slow context
- Do not add a second route parameter or a separate fetch-driven subpage for slow context.

## Rendering Rules

- Always reserve panel space once the dashboard shell is renderable; do not let slow-context loading cause layout jumps.
- Show a persistent `Context only` badge in the panel header.
- Use one panel-level framing sentence, then row-level details for each metric.
- For each row, show:
  - metric label
  - latest value and unit when available
  - freshness badge from the service response
  - cadence label
  - `as of` timestamp
  - previous value or revision note only when the response includes it
  - isolated fallback or unavailable copy when the row is not readable
- If all rows are unavailable, keep the panel visible with explicit unavailable framing instead of collapsing it.

## Operator Messaging Rules

- Panel framing stays stable and short: `These indicators update on a slower schedule than market-state feeds.`
- Keep the advisory rule unmistakable: `Use to explain backdrop, not to gate the live state.`
- Prefer service-supplied row messages for fresh, delayed, stale, and unavailable copy so the UI does not fork message vocabulary.
- Never reuse slow-context status as route-level warning text or summary-card callouts.

## Styling Guidance

- Keep the panel visually distinct but quieter than the realtime warning hierarchy.
- Use a neutral tone family and compact badge system so `Context only` never competes with `TRADEABLE/WATCH/NO-OPERATE` emphasis.
- On desktop, render rows as a tight card stack or compact grid inside the panel.
- On mobile, collapse to one column and keep badge, cadence, and `as of` text visible without truncation that hides status.
- Avoid decorative motion; a lightweight load-in is acceptable, but status changes must remain immediately legible.

## Component Test Expectations

- healthy panel renders all rows plus the persistent `Context only` badge
- delayed or stale rows keep the realtime panels unchanged and show advisory copy only inside the slow-context panel
- partial panel keeps available rows readable while unavailable rows stay explicit
- fully unavailable panel still renders header, framing text, and isolated fallback copy

## Summary

This module adds one honest dashboard surface for slow context without disturbing the route's existing trust hierarchy. The panel is always present, clearly secondary, and explicit about slower cadence so operators cannot confuse advisory backdrop with realtime gating.
