# Dashboard Detail Panels And Symbol Switching

## Ordered Implementation Plan

1. Add focused-panel view models and shell-facing data contracts in `apps/web` that reuse the completed dashboard adapter and trust-state seam.
2. Replace the focused-symbol placeholders with overview and microstructure panels that present service-owned current-state details for BTC and ETH without client-side market logic.
3. Add the feed-health/regime panel, an explicit derivatives unavailable panel, and instant-feeling symbol switching that reuses cached detail state while revalidation happens in the background.
4. Validate unit, build, and desktop browser smoke coverage for panel composition and symbol switching; record results in `plans/completed/dashboard-detail-panels-and-symbol-switching/testing-report.md`.

## Problem Statement

`plans/completed/dashboard-shell-and-summary-strip/` established the dashboard route, and `plans/completed/dashboard-query-adapters-and-trust-state/` established the read-only data and trust boundary. The route is still missing the main operator read path inside the focused-symbol area: overview, microstructure, derivatives, and health/regime need real panel content, and BTC/ETH switching needs to feel immediate without hiding peer summaries or inventing a new client state layer.

## Bounded Scope

- focused-symbol overview, microstructure, derivatives, and health/regime panels inside the existing dashboard shell
- panel-level view-model mapping from the already-decoded current-state contracts
- instant-feeling BTC/ETH switching that reuses cached symbol data and keeps the inactive summary visible
- desktop and mobile-safe detail composition that reuses the existing route and section navigation seams
- deterministic mocked-response and fixture-backed tests for panel rendering and symbol-switch behavior

## Out Of Scope

- backend, schema, or service contract changes
- client-side derivation of tradeability, fragmentation, divergence, derivatives meaning, or regime rules
- history, audit, replay, analytics, or slow-context UI
- negative-state polish, keyboard/a11y hardening, or full mobile smoke expansion beyond what is needed to implement the panels
- alerting, saved views, personalization, or new global app architecture

## Requirements

- Keep the work bounded to `dashboard-detail-panels-and-symbol-switching`; later child plans own negative-state polish and the full fixture smoke matrix.
- Stay entirely inside `apps/web` and preserve the Vite SPA React + TypeScript setup.
- Reuse the existing query adapter and dashboard-state mapping seam from `plans/completed/dashboard-query-adapters-and-trust-state/`; do not add a parallel client state system.
- Keep the dashboard read-only: all state labels, reason families, availability, freshness, and version metadata come from service-owned payloads.
- Preserve both BTC and ETH summary cards during symbol switching; only the focused detail region changes.
- Make symbol switching feel immediate by showing cached last-known-good panel content when available, then reconciling with background refresh state.
- Keep the derivatives panel honest: if the current-state contract still lacks derivatives detail, render an explicit unavailable panel rather than placeholder optimism or client inference.
- Preserve the current route-state contract: `symbol` remains the focused-symbol source of truth and `section` remains the mobile-safe section selector.
- Keep the panel layout dense but readable on desktop and single-column practical on mobile, with trust cues visible without hover.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-state`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`
- `apps/web/tests/e2e`

## Module Breakdown

### 1. Focused Panel View Models And Shell Contract

- Extend the shell-facing model so the focused-symbol region receives panel-specific presentational data instead of generic placeholder notes.
- Keep mapping logic in `apps/web/src/features/dashboard-state` so panel components receive already-reduced display values, trust states, refresh notes, and reason stacks.
- Reuse service payload fields already present in the current-state contracts for overview, microstructure, and health/regime; keep derivatives as an explicit unavailable seam until upstream data exists.

### 2. Overview And Microstructure Panel Composition

- Replace the generic overview and microstructure slots with focused symbol panels that surface effective state, symbol/global context, WORLD vs USA composite details, bucket timing, and recent-context summaries.
- Preserve the existing detail shell structure and section navigation so the page shape remains stable for later child features.

### 3. Derivatives, Health, And Symbol Switching Behavior

- Implement the health/regime panel from the existing global and symbol current-state surfaces.
- Render a dedicated derivatives panel with explicit unavailable messaging tied to the focused symbol and the existing contract gap.
- Keep symbol switching responsive by showing the next symbol's cached panels immediately when present, retaining summary-strip visibility, and surfacing any background refresh as trust-reducing UI rather than a blank panel reset.

## Design Details

### Focused Panel Data Shape

- Prefer one shell-facing `focusedPanels` object keyed by `overview`, `microstructure`, `derivatives`, and `health` instead of overloading the existing placeholder `sections` record.
- Each panel view model should carry:
  - title and optional eyebrow copy
  - trust state chip
  - one primary summary sentence
  - small metric rows or callouts sourced from service payloads
  - ordered reason list
  - optional refresh or fallback note when last-known-good data is retained
- Keep formatting-only helpers in the client, such as compact percentages, ages, and timestamp labels.

### Panel Content Expectations

#### Overview

- show focused symbol effective state
- show symbol state and global ceiling side by side
- show WORLD and USA composite availability, latest bucket timestamps, and any degraded or unavailable markers already present in the payload
- keep reason text concise and service-derived

#### Microstructure

- show the latest trusted bucket family used for the detail view
- show closed-window timing and missing-bucket counts
- show recent-context completeness for 30s, 2m, and 5m families without building charting or analytics views

#### Derivatives

- keep the slot visually complete, but state plainly that the current-state contract does not yet provide derivatives context
- tie the message to the focused symbol so switching still updates the card framing even when content remains unavailable

#### Health And Regime

- show global ceiling state, focused symbol summary availability, and key degraded reasons
- keep feed-health and trust notes close to freshness/version metadata so the operator can judge whether the panel is still safe to read

### Symbol Switching Rules

- route `symbol` remains authoritative for focus
- preserve the current `section` value across symbol switches unless the URL is invalid, in which case normalize back to `overview`
- summary strip remains visible and interactive at all times
- if the next symbol already has cached data, render it immediately and show any in-flight revalidation as a secondary note or loading accent instead of clearing the panel body
- if the next symbol lacks any safe data, fall back to the existing explicit loading or unavailable states from the trust model

## Acceptance Criteria

- Another agent can implement the focused-symbol detail panels without reopening the parent epic.
- The plan names exact repo areas for shell components, panel view models, page integration, styles, and tests.
- The plan preserves the read-only service trust boundary and does not require backend or schema changes.
- The plan keeps derivatives honestly unavailable until the contract exists, while still treating the panel as part of the focused-symbol composition.
- The validation shape proves panel composition, cached symbol switching, and explicit trust handling on the current dashboard route.

## ASCII Flow

```text
service-owned current-state payloads
  - global current state
  - BTC symbol current state
  - ETH symbol current state
                 |
                 v
existing adapter + normalization seam
                 |
                 v
focused-panel mapper
  - overview panel data
  - microstructure panel data
  - derivatives unavailable panel data
  - health/regime panel data
                 |
                 v
dashboard shell detail region
  - summary strip stays visible
  - focused panel content swaps by route symbol
  - section nav keeps mobile-safe reading order
                 |
                 v
operator reads one symbol quickly
without the UI recomputing market logic
```

## Live-Path Boundary

- Services remain the source of truth for effective state, reason families, freshness, availability, and version context.
- This feature is frontend-only and stops at read-only presentation, formatting, and route-level interaction behavior.
- Python remains out of scope and out of the runtime path.
