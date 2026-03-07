# Refinement Map: World USA Composites And Market State

## Current Status

This epic is still too broad for direct `feature-planning` or `feature-implementing`.

Completed prerequisite coverage already exists in:

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`
- `plans/completed/raw-event-log-boundary/`
- `plans/completed/dashboard-shell-and-summary-strip/`

Those completed slices already provide:

- shared symbol, venue, timestamp, degraded-state, and versioning vocabulary for service outputs and UI consumers
- live canonical events plus feed-health signals for Binance, Bybit, Coinbase, and Kraken
- the append-only raw event boundary and partition-discovery seam later replay-backed audit reads will depend on
- a fixture-backed dashboard shell that is waiting on service-owned market-state query surfaces rather than inventing client logic

Completed child feature coverage now also exists in:

- `plans/completed/world-usa-composite-snapshots/`
- `plans/completed/market-quality-and-divergence-buckets/`
- `plans/completed/symbol-and-global-regime-state/`
- `plans/completed/market-state-current-query-contracts/`

## Upstream Dependency Status

- `plans/completed/raw-storage-and-replay-foundation/` now provides stable replay identity, ordering, backfill safety, and retention continuity evidence for later historical market-state reads.
- Replay-backed market-state history is no longer blocked on replay/storage stabilization, but it should still wait for current-state contract sequencing inside this epic.
- `plans/epics/visibility-dashboard-core/` is already refined and waiting on this epic to define service-owned read models for dashboard query adapters.

## What This Epic Still Needs

- versioned current-state consumer contracts the dashboard and later alert/risk services can read without recomputing market logic
- replay-aware historical and audit read seams that preserve bucket, config, and algorithm version context once replay/storage stabilization is ready


## What Should Not Be Re-Done

- canonical contract families, timestamp semantics, or deterministic fixture conventions already established in `plans/completed/canonical-contracts-and-fixtures/`
- venue ingestion, normalization, and degraded-feed detection already established in `plans/completed/market-ingestion-and-feed-health/`
- raw append-only event persistence already established in `plans/completed/raw-event-log-boundary/`
- fixture-only route shell work already established in `plans/completed/dashboard-shell-and-summary-strip/`
- any client-side recomputation of composite weights, divergence, market quality, or `TRADEABLE/WATCH/NO-OPERATE`

## Refinement Waves

### Wave 1

- `world-usa-composite-snapshots` (completed; archived under `plans/completed/world-usa-composite-snapshots/`)
- Why first: every later bucket, regime, and consumer surface depended on a trusted service-owned composite and contributor-provenance boundary.

### Wave 2

- `market-quality-and-divergence-buckets` (completed; archived under `plans/completed/market-quality-and-divergence-buckets/`)
- Why next: with composite snapshots complete, bucketed feature families could build on stable snapshot seams without reopening weighting policy.

### Wave 3

- `symbol-and-global-regime-state` (completed; archived under `plans/completed/symbol-and-global-regime-state/`)
- Why here: regime logic depended on stabilized composite-derived buckets and remained a distinct conservative trust gate rather than being buried inside feature assembly.

### Wave 4

- `market-state-current-query-contracts` (completed; archived under `plans/completed/market-state-current-query-contracts/`)
- Why here: once composite, bucket, and regime outputs were explicit, the repo could define a read-only current-state contract for dashboards and future service consumers without exposing raw internals.

### Wave 5

- `market-state-history-and-audit-reads` (completed; archived under `plans/completed/market-state-history-and-audit-reads/`)
- Why last: this slice should remain after current-state contracts so historical retrieval reuses the same service-owned read model rather than inventing a parallel contract.

## Plan-Now Vs Dependency-Gated

Safe to plan now:

- none; all child plans in this epic are implemented and archived

No additional dependency-gated child plans remain inside this epic.

## Notes For Future Planning

- Keep all market-state logic in Go services or shared Go helpers; the web app renders service-owned outputs only.
- Keep quote normalization, weighting, degradation penalties, and regime thresholds config-versioned and replay-pinned.
- Prefer child plans that end with targeted Go validation plus deterministic fixture or replay coverage, especially for timestamp fallback, degraded venues, and fragmentation transitions.
- Use current-state contracts to unblock `dashboard-query-adapters-and-trust-state`, but keep replay-backed history and audit reads explicitly separate so UI work does not inherit unstable replay assumptions.
