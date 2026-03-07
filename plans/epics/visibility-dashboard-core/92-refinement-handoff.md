# Refinement Handoff: Visibility Dashboard Core

## Recommended Next Child Feature

- `dashboard-negative-state-mobile-a11y`

## What Just Finished

- `dashboard-shell-and-summary-strip` is implemented, validated, and archived under `plans/completed/dashboard-shell-and-summary-strip/`.
- `plans/completed/market-state-current-query-contracts/` now provides the service-owned current-state consumer seam this dashboard reads.
- `dashboard-query-adapters-and-trust-state` is implemented, validated, and archived under `plans/completed/dashboard-query-adapters-and-trust-state/`.
- `dashboard-detail-panels-and-symbol-switching` is implemented, validated, and archived under `plans/completed/dashboard-detail-panels-and-symbol-switching/`.

## Why This Is Next

- The dashboard now has a complete focused-symbol read path built on the completed cache and trust seam.
- Negative-state polish and mobile accessibility are the next bounded risks now that detail composition exists.
- Full smoke-matrix work remains downstream once negative-state behavior is explicit.

## Recommended `feature-planning` Order

1. `dashboard-negative-state-mobile-a11y`
2. `dashboard-fixture-smoke-matrix`

## Parallelism Guidance

- `dashboard-negative-state-mobile-a11y` is now unblocked by the completed detail-panel implementation seam.
- `dashboard-fixture-smoke-matrix` should remain last because it validates the composed route, adapters, and negative-state behavior together.

## Blockers And Dependency Notes

- Main blocker cleared: `plans/completed/dashboard-detail-panels-and-symbol-switching/` now provides the panel composition seam later child plans depend on.
- Non-blocker: `plans/completed/raw-storage-and-replay-foundation/` now provides replay-backed auditability history, but it does not block shell planning or fixture-backed UI work.
- Preserve completed prerequisites from `plans/completed/canonical-contracts-and-fixtures/`, `plans/completed/market-ingestion-and-feed-health/`, and `plans/completed/dashboard-shell-and-summary-strip/`; later child plans should reuse their timestamp, degraded-state, venue-health vocabulary, and route-shell layout seams.

## Assumptions To Preserve

- `apps/web` stays a Vite SPA with React + TypeScript only.
- The dashboard remains read-only and does not compute tradeability, fragmentation, derivatives logic, or regime state.
- Slow context from CME or ETF sources stays out of this epic except for future integration seams.
- The next planner should keep validation centered on `apps/web` tests, desktop/mobile smoke, and deterministic fixture or mocked-response behavior.
