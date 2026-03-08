# Child Plan Seeds: Slow Context Panel

## `slow-context-source-boundaries` (completed)

- Outcome: Go-owned slow-source ingestion boundaries are defined for CME volume, CME open interest, and ETF daily flow with scheduled polling, publish-state classification, idempotent re-poll handling, and operator-visible source health that stays separate from realtime feed health.
- Primary repo areas: `services/*`, `tests/fixtures`, optional config under `configs/*`
- Depends on: `plans/completed/market-ingestion-and-feed-health/`, `plans/completed/raw-storage-and-replay-foundation/`, and the slow-context epic guidance in `plans/epics/slow-context-panel/`
- Validation shape: targeted Go tests for source parsing, repeated-poll idempotency, delayed-publication classification, and correction handling
- Why it stands alone: it fixes the acquisition and source-health boundary before storage shape, query assembly, or dashboard copy can drift around provider assumptions
- Archive: `plans/completed/slow-context-source-boundaries/`

## `slow-context-query-surface-and-freshness`

- Outcome: Services define one normalized slow-context record/query seam with explicit `fresh`, `delayed`, `stale`, and `unavailable` classification, append-safe history expectations, and non-blocking current-state integration guidance.
- Primary repo areas: `services/*`, optional `schemas/json/features`, `tests/fixtures`, `tests/integration`, optional `tests/replay`
- Depends on: `plans/completed/slow-context-source-boundaries/`, `plans/completed/market-state-current-query-contracts/`, and `plans/completed/market-state-history-and-audit-reads/`
- Validation shape: targeted Go tests for pinned-clock freshness thresholds, explicit unavailable payloads, and current-state query isolation when slow-context lookup fails
- Why it stands alone: it stabilizes the service-owned cadence and availability semantics that every later UI or consumer must trust

## `slow-context-dashboard-panel`

- Outcome: `apps/web` gains one advisory slow-context panel that renders service-supplied CME and ETF context, cadence labels, freshness badges, and isolated fallback states without changing symbol-state semantics or blocking the dashboard route.
- Primary repo areas: `apps/web/src/features`, `apps/web/src/pages/dashboard`, `apps/web/src/styles.css`, `apps/web/tests/e2e`
- Depends on: `slow-context-query-surface-and-freshness`, `plans/completed/dashboard-detail-panels-and-symbol-switching/`, and `plans/completed/dashboard-fixture-smoke-matrix/`
- Validation shape: targeted web tests plus desktop/mobile browser smoke proving `Context only` messaging, as-of/cadence visibility, and non-blocking fallback behavior for stale, delayed, partial, and unavailable slow context
- Why it stands alone: it keeps UI integration bounded to presentation and operator messaging after the service-owned semantics are already fixed
