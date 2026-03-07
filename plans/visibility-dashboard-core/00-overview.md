# Visibility Dashboard Core

## Ordered Implementation Plan

1. Define the dashboard information architecture and operator reading order for BTC and ETH current-state monitoring.
2. Define the dashboard query surfaces, freshness semantics, and UI trust boundaries against service-owned market logic.
3. Implement the Vite SPA delivery shell, responsive dense layouts, and degraded-state interaction model.
4. Add fixture-backed rendering, loading, stale-state, and accessibility coverage.
5. Run targeted web smoke checks against fixture and mocked service responses before handing off to feature implementation.

## Feature Role In The Visibility Initiative

`visibility-dashboard-core` is the first operator-facing slice that turns canonical events, composite features, market state, and feed health into a screen the first user can trust in under 60 seconds.

It sits after:

- `canonical-contracts-and-fixtures`
- `market-ingestion-and-feed-health`
- `raw-storage-and-replay-foundation`
- `world-usa-composites-and-market-state`

It must not invent new market interpretation inside `apps/web`. The dashboard exists to present service-owned truth clearly, expose degradation honestly, and keep the user oriented during fast-changing market conditions.

## First-User Workflow

1. Open the web app and land on a current-state dashboard that renders usable content within the dashboard query targets from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`.
2. Scan BTC and ETH summary rows/cards first to answer: `TRADEABLE`, `WATCH`, or `NO-OPERATE`.
3. Expand or focus into one symbol to understand why the state exists now across:
   - symbol overview
   - microstructure
   - derivatives context
   - feed health and regime visibility
4. Confirm whether degraded feeds, stale panels, or fragmented conditions reduce trust before interpreting any price move.
5. Leave the screen open as a monitoring surface that remains readable on desktop and mobile without hiding critical state.

## Scope

- Vite SPA dashboard plan for `apps/web`
- current-state dashboard shell for BTC and ETH
- symbol overview view
- microstructure view
- derivatives context view
- feed health and regime visibility view
- dense but mobile-safe layout rules
- query-surface plan for current state, health, and recent context
- loading, stale-state, and degraded-feed behavior
- dashboard interaction model for the first operator workflow
- fixture-backed and mocked-response testing strategy

## Out Of Scope

- alert setup, alert review, or outcome workflow
- new market logic, thresholds, or scoring formulas in the client
- slow context panel work for CME and ETF data beyond reserving integration seams
- SSR, Next.js, or server-component patterns
- speculative charting systems or heavy analytics not needed for current-state trust
- implementation of backend services or concrete API schema files

## Key Constraints

- `apps/web` remains a React + TypeScript + Vite SPA.
- The UI consumes service outputs and does not rederive fragmentation, tradeability, regime, or feed-quality decisions.
- Service-owned timestamps, degraded markers, freshness indicators, and config versioning remain visible in the UI.
- Mobile safety is mandatory, but the screen must stay dense enough for real market monitoring.
- Slower institutional context stays visibly separate and must not block this core dashboard slice.
- Python stays offline-only and never becomes a runtime dependency for dashboard behavior.

## Assumptions

- `world-usa-composites-and-market-state` will provide service-owned current-state outputs for BTC and ETH before this feature is implemented.
- The initial operator workflow needs current-state and recent-window context, not deep historical exploration.
- The web app can rely on a bounded set of service query surfaces, with transport choice left open between HTTP polling plus optional streaming updates.

## Design Overview

### Reading Order

The screen should follow one stable reading sequence so the operator does not hunt for meaning:

1. global timestamp and freshness banner
2. symbol overview strip for BTC and ETH
3. focused symbol detail stack
4. feed health and degraded-reason visibility
5. recent context timestamps and config provenance

### Trust Model

- services compute market state, fragmentation, regime, feed degradation, and derived metrics
- the UI formats, arranges, color-codes, and explains those outputs
- any UI-only calculation must be presentational only, such as sorting, local layout grouping, or converting numbers to compact display strings

### Response And Freshness Defaults

Use the operating defaults as implementation guardrails:

- current-state panels target under 2 seconds
- if a request exceeds the target, the UI shows loading or stale-state affordances rather than silently freezing
- stale content stays visible with explicit age and degraded trust messaging until a newer payload arrives or the screen enters a failed state
- the operator should always see whether a panel is fresh, stale, degraded, or unavailable

## ASCII Flow

```text
venue feeds + replay inputs
          |
          v
services/venue-* -> services/normalizer -> canonical events
                                          |
                                          v
                        services/feature-engine + services/regime-engine
                                          |
                     +--------------------+----------------------+
                     |                    |                      |
                     v                    v                      v
             symbol current state   derivatives context    feed health state
                     |                    |                      |
                     +---------- query surfaces / streams -------+
                                          |
                                          v
                                 apps/web Vite SPA
                                          |
         +----------------------+---------+-----------+------------------+
         |                      |                     |                  |
         v                      v                     v                  v
  symbol overview         microstructure       derivatives view   feed/regime view
         |                      |                     |                  |
         +----------------------+---------------------+------------------+
                                          |
                                          v
                         operator answers "what changed" and "can I trust it"
```

## Ordered Delivery Notes

- Build the IA first so later agents know the exact page and panel structure.
- Lock the query surface contract expectations before selecting client state patterns.
- Implement summary and detail surfaces before adding optional richer interactions.
- Treat degraded and stale states as first-class UX, not polish work.
- Keep a reserved seam for `slow-context-panel`, but do not let it block the core dashboard path.
