# Refinement Map: Visibility Dashboard Core

## Current Status

This epic is still too broad for direct `feature-planning` or `feature-implementing`.

Completed prerequisite coverage already exists in:

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`

Those completed slices already provide:

- shared BTC/ETH symbol vocabulary, timestamp semantics, degraded markers, and config/version metadata expectations
- feed-health outputs and venue degradation semantics that the dashboard must display rather than recompute
- deterministic fixture conventions that later dashboard mocks can reuse

Completed child feature coverage now also exists in:

- `plans/completed/dashboard-shell-and-summary-strip/`

## Upstream Dependency Status

- `plans/completed/market-state-current-query-contracts/` now provides the service-owned current-state consumer contract seam this epic needed for dashboard query wiring.
- `plans/completed/raw-storage-and-replay-foundation/` now provides replay-backed auditability history, but it did not block read-only SPA shell planning or fixture-backed rendering work.

## What This Epic Still Needs

- a read-only client query boundary that preserves service-owned freshness, degradation, completeness, and provenance metadata
- focused symbol detail panels for overview, microstructure, derivatives context, and feed health/regime
- explicit stale, degraded, unavailable, and partial-data UX that never implies neutral market state
- mobile-safe dense delivery and accessibility behavior for symbol switching and panel access
- deterministic fixture-backed and mocked-response smoke coverage for desktop and mobile

## What Should Not Be Re-Done

- shared contract semantics already established in `plans/completed/canonical-contracts-and-fixtures/`
- ingestion and venue-health derivation already established in `plans/completed/market-ingestion-and-feed-health/`
- service-owned market-state, fragmentation, and regime computation that belongs in `plans/epics/world-usa-composites-and-market-state/`
- any backend service implementation, concrete schema authoring, or client-side market-logic re-derivation

## Refinement Waves

### Wave 1

- `dashboard-shell-and-summary-strip` (completed; archived under `plans/completed/dashboard-shell-and-summary-strip/`)
- Why first: the route shell, global status rail, BTC/ETH summary strip, and symbol-focus behavior were fixture-safe and could be built without waiting for final transport choices.

### Wave 2

- `dashboard-query-adapters-and-trust-state` (completed; archived under `plans/completed/dashboard-query-adapters-and-trust-state/`)
- Why next: with the shell complete and current-state contracts now finished, the next missing UI capability is the client adapter boundary.

### Wave 3

- `dashboard-detail-panels-and-symbol-switching` (completed; archived under `plans/completed/dashboard-detail-panels-and-symbol-switching/`)
- `dashboard-negative-state-mobile-a11y` (completed; archived under `plans/completed/dashboard-negative-state-mobile-a11y/`)
- Why here: panel composition and negative-state/mobile behavior depend on the shell and query-state model, but they can be planned as separate implementation slices once those boundaries are stable.

### Wave 4

- `dashboard-fixture-smoke-matrix` (completed; archived under `plans/completed/dashboard-fixture-smoke-matrix/`)
- Why last: integrated fixture, build, and Playwright coverage validates the composed dashboard after the UI slices above are defined.

## Notes For Future Planning

- Keep `apps/web` as a React + TypeScript + Vite SPA; do not introduce SSR, Next.js, or server-owned rendering assumptions.
- Keep the UI read-only and service-trusting: format values, arrange panels, and expose trust metadata, but do not compute tradeability, fragmentation, or regime locally.
- Treat slow institutional context as a later seam only; do not let `slow-context-panel` expand this epic.
- Prefer child plans that each end with direct `apps/web` validation commands and mocked or fixture-backed proof.
