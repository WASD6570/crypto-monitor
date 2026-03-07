# UI State And Performance Constraints

## Module Requirements

- Target repo area: `apps/web`
- Primary responsibility: define SPA state boundaries, fetch behavior, disclosure rules, and performance limits for dense review surfaces
- Required inputs: recent alert list queries, selected alert detail queries, lazy replay-window queries, aggregate analytics queries, and service freshness metadata
- Validation focus: responsive UI smoke coverage, loading-state behavior, request-cancellation behavior, and performance-budget adherence under mocked payloads

## Scope

- client state segmentation for alert list state, selected alert detail state, replay-window state, and analytics slice state
- query-key and caching expectations that preserve freshness labels without stale cross-panel leakage
- route and component boundaries for code splitting if the review area grows heavy
- progressive disclosure rules for dense data on desktop and mobile
- performance constraints for the Vite SPA, especially around charting, table density, and replay evidence rendering

## Out Of Scope

- introducing SSR, Next.js, or server-rendered review pages
- adding heavyweight analytics libraries without proven need
- background client workers unless evidence shows presentational formatting is too expensive
- bespoke global state architecture unrelated to this review flow

## State Model

- Keep recent alert list state independent from selected detail state so list refreshes do not wipe the active review panel.
- Keep replay-window state lazy and local to the replay panel or route segment because it is optional and may be the heaviest payload.
- Keep analytics state separate from detail state because freshness and retry behavior differ.
- Persist only lightweight UI preferences locally, such as collapsed sections or selected filters, never canonical review results.

## Freshness And Retry Semantics

- Each panel owns its own freshness indicator using service-provided timestamps.
- A newer list payload must not silently overwrite an older but still active detail payload without a visible refresh transition.
- Retry buttons should be scoped per panel: list, detail, replay, analytics.
- If a request fails after previous success, keep the last successful payload visible with an error banner and age marker when safe.

## Progressive Disclosure Rules

- Default open sections: alert summary, why-fired explanation, outcome summary, replay coverage status.
- Default closed sections: delivery details, feedback history, raw identifiers, extended event evidence, low-priority aggregates.
- On mobile, move secondary panels into accordions or drawers while keeping core trust fields in the first viewport.
- Avoid modal-heavy flows for core review because comparison between summary and detail should remain easy.

## Performance Constraints

- Respect the operating-default query targets:
  - recent alert drill-down under 5 seconds
  - 24h review analytics under 10 seconds
- Avoid loading replay evidence until the user opens the replay panel.
- Prefer summary strips, compact tables, and bounded windows over always-on chart canvases.
- Route-split or component-split heavier analytics/replay bundles if bundle growth threatens first paint for the main web app.
- Virtualize only if measured row counts justify it; do not add virtualization complexity for short recent-alert lists.
- Memoize expensive presentational transforms, but never move business logic into the memoized client layer.
- Keep tap targets at least 44px and preserve no-horizontal-scroll access to primary trust fields on mobile.

## Degraded And Empty State Rules

- Empty list, stale detail, unavailable replay, and stale analytics must each have distinct copy and badges so the user knows what is missing.
- Do not collapse degraded sections away; show them with clear trust-reduction labels.
- If analytics lag detail by many minutes, show independent ages and avoid blended page-level freshness claims.
- If replay coverage is degraded because of retention or late events, keep the panel available with partial evidence markers instead of blocking access.

## Accessibility And Interaction Notes

- Keyboard access must reach filter controls, alert rows, section toggles, and replay expansion points.
- Severity, regime, and degraded state should never rely on color alone.
- Preserve stable focus when selecting alerts or opening disclosure panels.
- Screen-reader labels should expose stale, degraded, partial, and unavailable statuses explicitly.

## Unit And Component Test Expectations

- list refresh does not clear selected alert detail unexpectedly
- opening replay panel triggers lazy fetch once per query key and surfaces loading/error states cleanly
- stale detail and stale analytics banners can coexist with different timestamps
- mobile viewport keeps summary and action controls visible without broken layout
- code-split review route still renders shell and loading states without blocking the rest of the SPA

## Summary

This module defines how the review UI behaves as a Vite SPA under real latency and dense-data constraints. It keeps panel state isolated, disclosure conservative, and heavy replay or analytics work lazy enough to preserve first paint and trustworthy refresh behavior.
