# Replay And Analytics UI

## Ordered Implementation Plan

1. Define the alert review information architecture, reading order, and service-owned query contract expectations for recent alert detail and replay inspection.
2. Plan the alert detail and replay window module so a user can explain why an alert fired, what evidence was present, and what replay coverage is available without client-side business logic.
3. Plan the analytics slices and regime review module so the user can review outcome quality, simulation-summary seams, and best/avoid patterns from service-owned aggregates.
4. Plan the SPA state model, freshness behavior, and performance constraints so dense review views remain trustworthy on desktop and mobile.
5. Validate rendering, stale/degraded states, and negative-path handling with fixture-backed UI and contract smoke coverage.

## Role In Initiative 2

- This is slice 7 of `crypto-market-copilot-alerting-and-evaluation`.
- It is the main review surface that turns Initiative 2 alert, outcome, and later simulation records into a user-visible trust loop.
- Its role is not to create new trading logic; it makes service-owned alert review data understandable fast enough to support the program's time-to-trust target.
- Success means a recent alert can be opened and explained end to end in under 60 seconds using only product surfaces.

## Problem Statement

The product can only earn trust if the user can inspect a recent alert and quickly answer two questions: why did it fire, and did it work. That review must come from service-owned alert, outcome, replay, and analytics records rather than ad hoc client interpretation. The UI needs one dense but readable Vite SPA workflow that supports recent alert detail, replay-window inspection, regime-aware slices, and best/avoid summaries without becoming a heavy chart wall.

## First-User Workflow

1. Open the alert review area from the web UI and land on a recent alert stream with severity, setup family, regime, freshness, and review-status context.
2. Select one recent alert and immediately see the service-owned explanation bundle: trigger time, validation horizon context, market-state gate, degraded markers, delivery summary, and outcome summary.
3. Open the replay window to inspect the bounded before/after event window, available evidence, and any degraded or incomplete coverage markers.
4. Review outcome and simulation-summary seams to understand market truth first, then optional execution-truth context if a saved simulation exists.
5. Move to analytics slices to answer which regimes, setups, and condition bundles are performing best, which should be avoided, and whether freshness or coverage limits reduce confidence.

## In Scope

- Vite SPA planning for recent alert stream detail and review navigation in `apps/web`
- alert detail surface for service-owned alert payload, gating reasons, timestamps, provenance, delivery summary, operator feedback summary, and outcome summary
- replay-window review surface for bounded pre/post alert evidence and explicit degraded coverage handling
- simulation-summary seam that renders stored simulation status and net-viability summary when available without recomputing fills in the client
- regime-aware analytics slices for setup family, severity, horizon, market state, fragmentation, and degradation cohorts
- best-condition and avoid-condition summaries rendered from service-owned aggregate review data
- safe loading, empty, stale, degraded, and partial-data defaults
- mobile-safe progressive disclosure for dense review data
- fixture-backed testing strategy and performance guardrails for the Vite SPA

## Out Of Scope

- alert generation, tactical risk logic, delivery transport behavior, outcome computation, or simulation execution logic
- client-side recomputation of alert quality, regime state, outcomes, simulation PnL, or best/avoid condition ranking
- SSR, Next.js, server components, or backend implementation details outside defining query-contract expectations for the UI
- speculative freeform charting work, notebook-style analysis, or research-only dashboards
- config mutation, baseline comparison workflow, or historical tuning reports beyond reserved seams

## Requirements

- `apps/web` remains a React + TypeScript + Vite SPA.
- The UI consumes service-owned review data and does not rederive business logic for setup validity, regime state, outcome ordering, simulation results, or analytics ranking.
- Every review surface shows freshness, coverage, and degradation honestly using service-provided timestamps and markers.
- Recent alert drill-down should target the under-5-second query goal from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`; wider historical analytics should target under 10 seconds or fall back to precomputed aggregates.
- Replay and analytics views must keep progressive disclosure so the user can answer the core review questions before opening richer evidence panels.
- The simulation surface is a seam: if no saved simulation exists, the UI shows explicit absence rather than inferring viability from outcome fields.
- Empty states, stale analytics, and degraded replay coverage must default conservative: surface what is known, what is missing, and what should not be trusted.
- Python remains offline-only and never becomes a runtime dependency for the live review UI.

## Assumptions

- Prior slices provide stable alert records, outcome records, operator feedback records, and at least one service-owned review query surface for recent alert detail.
- Saved simulation records may lag the initial alert/outcome path, so the UI must treat simulation as optional and separately timestamped.
- Replay inspection is bounded to a review window around one alert, not a general-purpose historical charting product.
- Analytics slices are served as precomputed or service-aggregated review data when query latency would exceed SPA targets.

## Target Repo Areas

- `apps/web`
- service-owned review APIs or gateway surfaces already backing alerts, outcomes, replay, and simulations
- `tests/fixtures`
- `tests/integration`
- `tests/replay` for replay-window determinism smoke inputs when needed

## Trust Boundary

- services compute and persist alert reasons, state gates, outcomes, simulation summaries, replay coverage, and analytics aggregates
- the UI arranges, filters, and formats those records for operator review
- any client-only computation stays presentational only, such as sorting, grouping, compact-number formatting, or viewport-specific layout choices

## Safe Defaults

### Empty States

- If no recent alerts exist, show the review shell with a clear `No recent alerts yet` state, a time-range note, and the last successful refresh timestamp if available.
- If analytics slices return no rows for the selected filters, show `No review data for this slice` instead of zero-performance claims.

### Stale Analytics

- Keep the last successful aggregate visible with explicit age, stale badge, and a note that rankings may lag newer alerts.
- Do not mix stale analytics timestamps with fresher alert detail timestamps without labeling each section independently.

### Degraded Replay Coverage

- If the replay window is partial because of retention, late events, or feed degradation, show the partial evidence with exact missing-coverage markers.
- If no replay evidence is available, keep the alert detail usable and label replay as unavailable rather than blocking the page.

### Progressive Disclosure

- Default view shows the smallest answer set first: alert summary, why-fired reasons, outcome summary, simulation availability, and replay coverage status.
- Richer evidence such as event tables, regime-path detail, delivery history, and operator notes open in expandable panels, tabs, or drawers only when requested.

## ASCII Flow

```text
alert-engine + delivery + outcome + feedback + simulation storage
                            |
                            v
                 service-owned review query surfaces
                            |
          +-----------------+------------------+
          |                                    |
          v                                    v
   recent alert detail                  review analytics slices
   - alert payload                       - regime cohorts
   - market/risk gates                   - setup/horizon cohorts
   - outcome summary                     - best/avoid summaries
   - replay coverage                     - freshness + sample size
   - simulation seam
          |                                    |
          +-----------------+------------------+
                            |
                            v
                      apps/web Vite SPA
                            |
        +-------------------+--------------------+
        |                   |                    |
        v                   v                    v
   alert stream        alert detail         analytics review
        |                   |                    |
        |                   v                    |
        |            replay window panel         |
        +-------------------+--------------------+
                            |
                            v
            user answers "why did it fire" and "did it work"
```

## Ordered Delivery Notes

- Lock the review reading order before choosing component boundaries so the page answers the time-to-trust questions first.
- Define query-contract expectations before local state patterns so the SPA does not start deriving missing business logic.
- Treat replay coverage and stale analytics as first-class UX, not follow-up polish.
- Keep simulation and baseline comparison as explicit seams so later slices can attach without redesigning the review flow.
