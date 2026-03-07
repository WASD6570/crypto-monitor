# Dashboard Shell And Summary Strip

## Ordered Implementation Plan

1. Add one fixture-backed current-state dashboard route shell in `apps/web` with the global trust rail, BTC/ETH summary strip, and reserved detail-panel slots.
2. Add local fixture-backed state, symbol-focus URL state, and mobile section navigation without depending on unstable upstream query surfaces.
3. Add targeted route, interaction, and responsive smoke coverage and record results in `plans/completed/dashboard-shell-and-summary-strip/testing-report.md`.

## Problem Statement

The visibility initiative already defines the dashboard reading order, trust boundary, and service-owned market-state expectations, but `apps/web` still lacks the first operator-facing screen.

This child feature creates the smallest implementation-ready slice: a Vite SPA dashboard route shell that renders a global trust rail, BTC/ETH summary strip, and stable symbol-focus behavior from deterministic fixtures so the UI can be built before live query surfaces are finalized.

## Bounded Scope

- one dashboard route entry for current-state monitoring in `apps/web`
- one top status rail that exposes service-owned freshness, degradation, and config/version placeholders from fixtures
- one BTC/ETH summary strip with clear active symbol focus and service-supplied reason text
- one desktop/mobile-safe reading order with reserved slots for later detail, derivatives, and feed-health panels
- local fixture-backed state and URL navigation for symbol focus and mobile section selection
- test seams that let later child features swap fixtures for real query adapters without rewriting the shell

## Out Of Scope

- live API integration, polling, streaming, or client contract decoding beyond fixture-backed interfaces
- client-side computation of tradeability, fragmentation, regime, feed quality, or derivatives meaning
- detailed panel implementation for overview, microstructure, derivatives, or feed-health content beyond shell placeholders
- slow context, alert workflows, or deep historical views
- SSR, Next.js, server components, or Python runtime dependencies

## Requirements

- Keep the work bounded to `dashboard-shell-and-summary-strip`; later child plans own query adapters, detailed panels, and negative-state depth.
- Keep `apps/web` as a React + TypeScript + Vite SPA.
- Preserve the service-trusting UI boundary: fixtures should mimic service-owned state, timestamps, degraded markers, and reason text rather than asking the client to derive them.
- Make the shell fixture-backed first so implementation is not blocked on `world-usa-composites-and-market-state` query-surface stabilization.
- Keep the route dense enough for monitoring while remaining mobile-safe: 44px minimum tap targets, single-column controls on mobile, and inline trust warnings near the top of every section.
- Keep BTC and ETH visible together in the summary strip even when one symbol is focused for the lower detail area.
- Keep freshness honest: show explicit `loading`, `ready`, `stale`, `degraded`, or `unavailable` labels from fixture state, and never imply neutral market conditions when data is missing.
- Keep global trust cues visible without requiring hover or hidden disclosure patterns.

## Target Repo Areas

- `apps/web/src/app`
- `apps/web/src/pages`
- `apps/web/src/features/dashboard-shell`
- `apps/web/src/components`
- `apps/web/src/hooks`
- `apps/web/src/state`
- optional fixture helpers under `apps/web/src/utils`
- optional browser smoke coverage under `apps/web/tests` or `apps/web/tests/e2e` once the SPA scaffold exists

## Module Breakdown

### 1. Route Shell And Summary Layout

- Register the dashboard route and page composition in `apps/web/src/app` and `apps/web/src/pages`.
- Create a feature-local shell in `apps/web/src/features/dashboard-shell` with:
  - global status rail
  - BTC/ETH summary strip
  - focused-symbol header
  - reserved detail-region slots for later child features
- Define the desktop and mobile reading order now so later feature work fills stable slots instead of reshaping the page.

### 2. Fixture-Backed State And Navigation

- Add feature-local fixture data, state selectors, and navigation helpers in `apps/web/src/features/dashboard-shell`, `apps/web/src/hooks`, and `apps/web/src/state` only where needed.
- Drive symbol focus, mobile section selection, and panel trust labels from fixture-owned state.
- Preserve a clean seam so the later query-adapter child feature can replace the fixture source behind the same shell-facing view model.

## Design Details

### Recommended Folder Shape

- `apps/web/src/app`: router registration and app-level layout seam
- `apps/web/src/pages/dashboard`: route entry component for the current-state screen
- `apps/web/src/features/dashboard-shell/components`: `DashboardShell`, `DashboardStatusRail`, `SummaryStrip`, `SummaryCard`, and shell placeholders
- `apps/web/src/features/dashboard-shell/fixtures`: healthy, degraded, stale, and partial-data dashboard fixtures
- `apps/web/src/features/dashboard-shell/model`: view-model mapper types that stay presentational only
- `apps/web/src/hooks`: query-param helpers only if they are reusable outside this feature

### Route And URL Defaults

- preferred route: one dedicated dashboard path such as `/dashboard`
- default focus: `BTC-USD`
- query params:
  - `symbol=BTC-USD|ETH-USD`
  - `section=overview|microstructure|derivatives|health` for mobile-safe section recall
- invalid query values fall back to the default fixture-backed state instead of breaking the route

### Shell Layout Rules

#### Desktop

- full-width status rail first
- two-card summary strip second, showing BTC and ETH together
- focused-symbol detail header third
- two-column placeholder grid below:
  - left slot: overview then microstructure
  - right slot: derivatives then feed health/regime

#### Mobile

- status rail first and always visible at the top of the route flow
- stacked summary cards with large tap targets and visible active-state styling
- compact section switcher or anchored section buttons below the active summary card
- ordered single-column panel slots with repeated trust note above the active section when needed

### Trust-State Constraints

- The shell displays service-shaped labels, timestamps, and reasons from fixtures.
- The shell may format ages and compact numbers for presentation, but it must not derive `TRADEABLE`, `WATCH`, `NO-OPERATE`, fragmentation, or regime outcomes.
- Global and per-card trust labels should use the same bounded vocabulary: `loading`, `ready`, `stale`, `degraded`, `unavailable`.
- Timestamp-fallback or degraded-source notes from fixture metadata must remain visible in compact text near the relevant rail or card.

### Placeholder Strategy For Later Child Features

- Reserve named shell slots for `overview`, `microstructure`, `derivatives`, and `health` now.
- Each slot should support placeholder shell states so the route is testable before real detail content exists.
- Keep slot props shallow and presentational so later modules can mount richer content without changing page structure.

## Acceptance Criteria

- Another agent can implement the dashboard route shell without rereading the parent epic.
- The plan names concrete repo areas under `apps/web` for route registration, page composition, feature components, fixture state, and tests.
- The plan keeps service-side market logic authoritative and limits the client to layout, formatting, and navigation.
- The plan is fixture-backed first and explicitly defers live query-surface integration to the next child feature.
- The layout and test expectations cover desktop density, mobile safety, and visible trust-state handling.

## ASCII Flow

```text
fixture-backed dashboard payloads
        |
        v
feature-local fixture adapter
        |
        +----> global rail view model
        |
        +----> BTC summary card view model
        |
        +----> ETH summary card view model
        |
        +----> focused symbol + section state
        v
apps/web dashboard route shell
  - status rail
  - summary strip
  - focused header
  - reserved panel slots
        |
        v
later child features replace fixture source with query adapters
without changing route shape or trust boundary
```
