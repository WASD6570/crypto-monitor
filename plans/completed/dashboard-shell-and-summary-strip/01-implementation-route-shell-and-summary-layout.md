# Implementation Module 1: Route Shell And Summary Layout

## Scope

Implement the current-state dashboard route shell, global trust rail, BTC/ETH summary strip, and reserved detail layout so the operator has one stable reading order before live data integration exists.

## Target Repo Areas

- `apps/web/src/app`
- `apps/web/src/pages/dashboard`
- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/components`
- optional shared styling tokens in `apps/web/src/utils` only if the shell needs a small display helper

## Module Requirements

- Register exactly one dashboard route for the visibility shell.
- Render the shell immediately with fixture-backed content or placeholders; avoid a blank route.
- Place the status rail before the summary strip and the summary strip before all lower panel slots.
- Keep BTC and ETH visible together in the summary strip on desktop and mobile.
- Add visible active-symbol styling without removing the peer symbol summary.
- Reserve four lower shell slots for later child features: `overview`, `microstructure`, `derivatives`, and `health`.
- Keep trust labels and timestamps near the top of the rail and summary cards.
- Preserve mobile-safe density with 44px minimum tap targets, one-column controls, and no hover-only meaning.

## Implementation Decisions To Lock

- Put route registration in `apps/web/src/app`; keep the route entry component in `apps/web/src/pages/dashboard`.
- Keep the feature shell self-contained under `apps/web/src/features/dashboard-shell` so later dashboard modules can attach to named slots instead of scattering layout logic.
- Use shell slots or section wrapper components for future panel insertion rather than temporary ad hoc markup.
- Keep summary cards simple and service-trusting: they render fixture-provided state, reasons, age, and trust markers only.
- Do not introduce charting or dense table dependencies in this module.

## Recommended Component Breakdown

- `DashboardPage`: route entry and top-level route metadata
- `DashboardShell`: page composition and responsive layout container
- `DashboardStatusRail`: global `asOf`, oldest-age, config version, and degraded summary presentation
- `DashboardSummaryStrip`: two-card group with desktop row and mobile stack variants
- `DashboardSummaryCard`: state chip, reason list, freshness label, WORLD/USA comparison label, and active/focus interaction
- `DashboardSectionSlot`: titled placeholder surface for the four future detail regions

## Layout Rules

### Status Rail

- Must show dashboard `asOf`, oldest critical panel age, config/ruleset version, and a global degraded note when fixtures say one exists.
- Must stay compact enough to fit above the fold on laptop and mobile widths.
- Must expose exact timestamps in markup where tests and assistive tech can read them.

### Summary Strip

- Desktop: render BTC and ETH as sibling cards in one row when width allows.
- Mobile: render the same cards in a stacked selector with full-width buttons or card actions.
- Each card must show:
  - symbol label
  - service-supplied state label
  - short service-supplied reason stack
  - freshness/trust badge
  - last-updated timestamp
  - WORLD vs USA comparison summary when fixture data provides it

### Reserved Detail Area

- Render the focused symbol name directly above the slot grid so later child features inherit the same page orientation.
- Use predictable section titles now: `Overview`, `Microstructure`, `Derivatives Context`, `Feed Health And Regime`.
- Placeholder copy should say the section is planned or waiting on the next feature, not that the market signal is absent.

## Responsive And Accessibility Requirements

- Summary-card interactions must be keyboard reachable and screen-reader named with symbol, state, and freshness.
- Active/focused styles must work without color alone.
- Mobile section controls must remain visible near the top of the detail area and must not force horizontal scrolling.
- Long degraded or reason text must wrap instead of clipping.

## Unit Test Expectations

- route render shows the status rail, both summary cards, and all four named section slots
- summary cards preserve peer visibility when one symbol is active
- mobile render keeps summary controls stacked and tap-safe
- rail and card trust labels remain visible for degraded and stale fixtures

## Acceptance Criteria

- Another agent can implement the route shell without guessing at component boundaries.
- The layout fixes the operator reading order for desktop and mobile before detail modules exist.
- The summary strip preserves simultaneous BTC/ETH visibility.
- Trust metadata is visible in the shell and not deferred to future polish work.

## Summary

This module fixes the page chrome for the dashboard: one route, one status rail, one BTC/ETH summary strip, and one stable set of lower slots. Later work should fill the named slots and replace placeholder content without moving the rail or summary strip.
