# Information Architecture

## Module Goal

Define the operator-facing structure for the visibility dashboard so later implementation agents can build a trustworthy, dense, mobile-safe screen in `apps/web` without guessing about page layout or reading order.

## Target Repo Areas

- `apps/web/src/app`
- `apps/web/src/pages`
- `apps/web/src/features`
- `apps/web/src/components`
- optional shared types/helpers in `libs/ts` only if a frontend-only helper becomes necessary

## Requirements

- Present BTC and ETH state without requiring route changes for the first scan.
- Support a default desktop layout and a mobile-safe stacked layout with the same information hierarchy.
- Make the symbol overview, microstructure, derivatives context, and feed health/regime views all available from the first screen.
- Preserve service-owned trust signals, including freshness, degraded reasons, source timestamps, and config version.
- Do not hide critical state behind hover-only interactions.
- Keep tap targets at least 44px on mobile and keep filter/control surfaces single-column unless a different pattern clearly improves readability.

## Proposed Screen Structure

### Route Shape

- Primary route: one dashboard route for current-state monitoring.
- Default symbol focus: BTC on first load, with fast switching to ETH.
- URL state: symbol focus, selected panel tab on mobile, and optional compact mode preferences can live in query params so reloads are stable.

### Layout Model

#### Desktop

- top status rail
- symbol overview strip with BTC and ETH side by side
- primary content split into two vertical columns:
  - left: symbol overview detail and microstructure
  - right: derivatives context and feed health/regime
- sticky condensed status bar allowed if it improves long-session monitoring without obscuring content

#### Mobile

- top status rail remains visible at the top of the scroll
- symbol overview cards become a stacked two-card selector
- focused symbol detail collapses into ordered sections or tabs:
  - overview
  - microstructure
  - derivatives
  - feed/regime
- all critical warnings remain inline above section content

## Panel Definitions

### 1. Global Status Rail

Purpose: orient the operator before any symbol-specific interpretation.

Must include:

- dashboard last updated time
- oldest displayed payload age across required panels
- global regime ceiling if available
- explicit degraded or stale indicator when any critical upstream dependency is impaired
- active config version or state-version label when the service exposes it

### 2. Symbol Overview View

Purpose: answer the first-user question quickly for each symbol.

Must include for BTC and ETH:

- current `TRADEABLE`, `WATCH`, or `NO-OPERATE` state
- short reason stack supplied by the service, not inferred by the UI
- WORLD vs USA directional alignment or fragmentation signal
- market-quality summary and most recent update time
- key freshness badge showing whether the symbol state is fresh, stale, or degraded

Desktop behavior:

- show both symbols at once
- selecting a symbol updates the lower detail panels without losing the peer symbol summary

Mobile behavior:

- stacked cards with strong state labels and key reason bullets
- the active symbol card expands or anchors the detail section below

### 3. Microstructure View

Purpose: show how healthy or unstable the current tape is without requiring raw venue-book inspection.

Should display service-computed or service-derived fields only, such as:

- spread quality
- imbalance or pressure indicator
- short-horizon momentum or chop classification
- venue agreement vs dislocation markers
- local freshness and degraded markers tied to the source surface

Design rules:

- prefer small-multiple metric rows and compact sparkline-ready placeholders over oversized charts
- if charting is used, it must support a no-chart fallback card for low-bandwidth or degraded cases
- keep numeric displays short and comparable across BTC and ETH

### 4. Derivatives Context View

Purpose: expose offshore risk posture and perp context without making the UI invent derivatives logic.

Should display service-owned context such as:

- perp premium or basis context
- funding-related state if supplied by services
- open-interest or leverage stress proxy if already computed upstream
- WORLD spot vs perp alignment or mismatch summary
- timestamp cadence that makes slower derivatives updates obvious when applicable

Design rules:

- clearly label offshore/perp context as context, not a standalone trade signal
- visually distinguish missing data from neutral state

### 5. Feed Health And Regime View

Purpose: show whether the operator can trust the current surface.

Must include:

- per-venue health rows for Binance, Bybit, Coinbase, and Kraken when relevant to the focused symbol
- freshness state, reconnect or resync warnings, and degraded reasons supplied by services
- regime explanation area that ties the symbol state to service-owned market-quality and health outputs
- explicit note when a critical venue downgrade is suppressing tradeability

Design rules:

- unhealthy inputs should be impossible to miss but should not visually drown out the rest of the screen
- use consistent severity tokens for healthy, degraded, stale, and unavailable

## Density And Mobile-Safety Rules

- default to compact cards, short labels, and limited decimal noise
- preserve text alternatives for color-coded state chips
- avoid four-column metric grids on mobile; switch to stacked two-column or single-column groups
- pin critical state labels and timestamps near the top of each panel so the user does not scroll for trust context
- allow long degraded-reason text to wrap without breaking the layout

## Loading And Empty-State Rules

- first load shows skeleton or placeholder structure for the global rail and all major panels
- symbol overview placeholders should preserve final card dimensions to reduce layout shift
- a partially loaded screen is acceptable only if missing panels are clearly labeled `loading`, `stale`, or `unavailable`
- empty state for unsupported or absent derivatives context must say `not available from current service surface`, not imply a market-neutral reading

## Accessibility Requirements

- all severity and regime colors require text labels and accessible contrast
- keyboard navigation must reach symbol switching, panel tabbing, and degraded-detail disclosure
- aria labels or equivalent accessible names must describe freshness/state chips in plain language
- timestamps should expose machine-readable values for screen-reader and test usage where practical

## Unit And Component Test Expectations

- symbol overview cards render state, reasons, and freshness consistently for BTC and ETH fixtures
- mobile section switching keeps warnings visible and focus order predictable
- degraded venue rows remain readable with long reason strings
- missing derivatives data renders as unavailable, not as zero or normal

## Summary

This module fixes the dashboard structure before implementation begins: one current-state route, one stable reading order, and four core views that remain dense on desktop while collapsing safely on mobile. Later agents should not introduce new panels or reorder trust-critical sections unless the service contracts change materially.
