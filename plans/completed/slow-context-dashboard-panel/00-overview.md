# Slow Context Dashboard Panel

## Ordered Implementation Plan

1. Extend the dashboard symbol contract and frontend adapter seam so the focused-symbol read includes one service-owned slow-context block without adding client freshness logic.
2. Add one dedicated advisory slow-context panel to the dashboard shell that keeps realtime panels first, shows `Context only` messaging, and renders CME plus ETF rows with isolated fallback states.
3. Extend the shared dashboard scenario catalog, unit coverage, and desktop/mobile browser smoke so fresh, delayed, stale, partial, and unavailable slow-context cases stay deterministic.

## Problem Statement

`plans/completed/slow-context-query-surface-and-freshness/` now fixes the service-owned slow-context vocabulary, freshness thresholds, and explicit unavailable behavior, but `apps/web` still has no way to surface that advisory context. Without this child feature, operators cannot see the already-normalized CME and ETF backdrop inside the dashboard, and later UI work would risk re-deriving freshness, hiding unavailable states, or mixing slow context into realtime trust cues.

This slice is the bounded UI follow-on only: consume the completed service seam, present it honestly, and prove the dashboard stays readable when slow context is delayed, stale, partial, or unavailable.

## Bounded Scope

- extend the focused-symbol dashboard contract consumption in `apps/web` for the service-owned slow-context block
- add one dedicated slow-context presentational model and panel in the dashboard shell
- render CME volume, CME open interest, and ETF daily flow rows with service-supplied freshness, cadence, timestamps, revision visibility, and advisory copy
- keep slow-context failures isolated to the panel while the existing dashboard route and realtime panels keep working
- add deterministic fixtures and browser smoke for fresh, delayed, stale, partial, and unavailable slow-context cases

## Out Of Scope

- new slow-source polling, normalization, freshness, or threshold logic in the client
- changing `TRADEABLE`, `WATCH`, `NO-OPERATE`, route trust, or summary-card semantics based on slow context
- new historical drill-downs, charts, saved views, or analytics workflows
- widening this slice into replay, schema rollout, or multi-endpoint orchestration work
- Python in the live runtime path

## Requirements

- Reuse `plans/completed/slow-context-query-surface-and-freshness/` as the authoritative source for slow-context availability, freshness, message keys, cadence metadata, timestamps, and revision semantics.
- Keep `apps/web` read-only and presentational: no client-side recomputation of freshness windows, delayed/stale thresholds, or operator-safe copy.
- Keep the main dashboard reading order intact: realtime state remains primary, and the slow-context panel remains explicitly advisory.
- Show a persistent `Context only` indicator and service-supplied helper text whenever the panel renders.
- Preserve row-level isolation: one missing or unavailable slow metric must not blank the whole panel if other metrics remain readable.
- Preserve timestamp honesty by showing service-owned `as of` and cadence labels directly and surfacing correction or revision notes only when the response includes them.
- Keep slow-context status out of the route-level `trustState`, summary-strip warnings, and market-state badges unless a future feature explicitly changes that rule.
- Keep the route non-blocking: initial load, symbol switching, retry, and stale last-known-good behavior must remain governed by the existing dashboard current-state flow.

## Target Repo Areas

- `apps/web/src/api/dashboard`
- `apps/web/src/features/dashboard-state`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`
- `apps/web/src/test`
- `apps/web/tests/e2e`

## Module Breakdown

### 1. Slow-Context Contract Consumption And View Model

- Extend the symbol-scoped dashboard contract types and decoders so the focused-symbol read can carry one nested slow-context block sourced from services.
- Map that block into a dedicated shell-facing panel model with panel-level helper copy plus row-level values, freshness badges, cadence labels, timestamps, and unavailable/error notes.
- Keep slow-context state separate from existing summary and section trust derivation so advisory data cannot silently downgrade realtime state.

### 2. Dashboard Panel Composition And Styling

- Render one dedicated slow-context panel in the dashboard shell after the existing realtime detail region, keeping the panel visible but visually secondary.
- Use a bespoke row layout rather than squeezing slow context into generic key/value metrics so partial availability, revision notes, and per-row freshness remain obvious.
- Keep desktop density high and mobile layout single-column, with timestamps and freshness badges still readable without hover.

### 3. Scenario Catalog, Unit Coverage, And Browser Smoke

- Extend the existing dashboard scenario helpers instead of inventing a second mock system.
- Add focused unit coverage for decoder and mapper behavior plus shell rendering assertions for advisory copy and row isolation.
- Extend desktop and mobile browser smoke around the same scenario catalog so the route still behaves honestly when slow context changes status.

## Acceptance Criteria

- Another agent can implement the slow-context dashboard slice without reopening the parent epic or re-deciding service freshness semantics.
- The plan names exact frontend repo areas for contract decoding, view-model mapping, panel composition, styles, fixtures, and browser smoke.
- The plan keeps slow context advisory-only and leaves summary-card, route-warning, and market-state semantics unchanged.
- The validation shape proves fresh, delayed, stale, partial, and unavailable slow-context cases on both desktop and mobile.

## ASCII Flow

```text
service-owned current-state symbol response
  + nested slow-context block
    - cme volume
    - cme open interest
    - etf daily flow
    - availability / freshness / cadence / timestamps / messages
                |
                v
apps/web dashboard decoder + mapper
  - no client freshness math
  - dedicated slow-context panel model
                |
                v
dashboard shell
  - realtime panels stay first
  - slow-context panel stays advisory
  - partial or unavailable rows stay isolated
                |
                v
desktop + mobile smoke matrix
  - fresh / delayed / stale / partial / unavailable
```

## Live-Path Boundary

- Services remain the source of truth for slow-context availability, freshness, timestamps, revision visibility, and advisory message content.
- This feature is a frontend consumer of that completed service seam; it does not move logic out of Go or introduce Python into the runtime path.
- If the symbol endpoint still needs a thin adapter update to expose the completed slow-context block, keep that change minimal and vocabulary-preserving rather than inventing a second frontend-owned contract.
