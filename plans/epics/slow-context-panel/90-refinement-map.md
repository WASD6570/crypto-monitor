# Refinement Map: Slow Context Panel

## Already Covered By Completed Work

- `plans/completed/dashboard-shell-and-summary-strip/` established the dashboard reading order and kept slower context explicitly out of the core route shell.
- `plans/completed/dashboard-detail-panels-and-symbol-switching/` finished the focused-symbol dashboard panels and confirmed slow context remains a later UI seam rather than a hidden add-on to existing panels.
- `plans/completed/dashboard-fixture-smoke-matrix/` closed the core dashboard honesty matrix, so this epic should extend the route with advisory context rather than reopen realtime warning semantics.
- `plans/completed/market-state-current-query-contracts/` and `plans/completed/market-state-history-and-audit-reads/` already define the authoritative current-state and audit boundaries the slow-context work must not destabilize.
- `plans/completed/raw-storage-and-replay-foundation/` already provides the append-safe and deterministic baseline that later slow-context persistence must respect if it enters replay-visible paths.

## What Still Remains

- one service-owned normalized slow-context record and query seam with deterministic freshness classification
- one non-blocking dashboard integration path that keeps slow context visually separate from realtime state
- bounded implementation order and validation guidance so later feature planning does not mix storage semantics and UI copy into one oversized slice

## Refinement Decisions

- Keep the epic split into service-first then UI-last child features; the dashboard panel depends on service-owned cadence and freshness semantics.
- Preserve the product rule that slow context is advisory only in MVP; no child feature should modify realtime regime, tradeability, or feed-health gating.
- Avoid provider lock-in during refinement; name source families and ingestion responsibilities, not vendor commitments.
- Keep any future shared contract work optional and narrow. Do not invent concrete schemas from this epic alone unless the later feature plan proves a shared response seam is required.

## Refinement Waves

### Wave 1

- `slow-context-source-boundaries` (completed; archived under `plans/completed/slow-context-source-boundaries/`)
- Why first: later planning now inherits one clear ingestion responsibility model, publish-window vocabulary, and idempotent polling rule before storage or UI work is bounded.

### Wave 2

- `slow-context-query-surface-and-freshness` (next active child plan)
- Why next: query and persistence planning now builds directly on the completed source boundary and should stabilize before any dashboard copy or layout work is planned.

### Wave 3

- `slow-context-dashboard-panel`
- Why last: the web surface should consume the service-owned slow-context seam after cadence labels, freshness states, and failure isolation rules are already fixed.

## Parallelism Guidance

- Do not plan Wave 2 in parallel with Wave 1; source publication semantics and correction handling still shape the normalized record and query response.
- Do not plan Wave 3 in parallel with Wave 2; the dashboard panel copy and fallback behavior should follow the final service-owned freshness and availability vocabulary.

## Notes For Later Planning

- Keep `apps/web` read-only and presentational; slow-context labels, freshness, and operator-safe messaging must come from service outputs.
- Prefer one obvious Go-owned slow-context seam over scattering polling or freshness logic across multiple existing services.
- Treat partial slow-context availability as normal and non-fatal. CME and ETF context may diverge in freshness without weakening realtime market-state delivery.
