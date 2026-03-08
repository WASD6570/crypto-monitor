# Program Handoff

## Initiative Queue

1. `initiatives/crypto-market-copilot-visibility-foundation/`
2. `initiatives/crypto-market-copilot-alerting-and-evaluation/`

## Why This Order

- Initiative 1 builds the trusted market-state substrate.
- Initiative 2 depends on that substrate for every alert, outcome, and review surface.
- Starting with alerts before trusted state would create noisy signals and brittle debugging loops.

## Handoff Seed: Initiative 1

### `initiatives/crypto-market-copilot-visibility-foundation/`

- Problem statement: the first user needs a trustworthy market screen that explains WORLD vs USA state, feed integrity, and no-operate conditions without monitoring raw venue streams manually.
- In scope: contracts, fixtures, ingestion, normalization, replay, composite features, regime state, dashboard surfaces, and slow context where non-blocking.
- Out of scope: alert setup logic, push notifications, outcome evaluation, simulated execution, tuning workflow.
- Validation shape: deterministic replay, stable dashboard state, feed degradation visibility, and current-state explainability within 60 seconds.

## Handoff Seed: Initiative 2

### `initiatives/crypto-market-copilot-alerting-and-evaluation/`

- Problem statement: once market state is trustworthy, the first user needs bounded alerts, realistic outcome verification, and a durable review loop that proves whether attention should have been spent.
- In scope: alert logic, permissions, risk states, delivery surfaces, outcomes, simulation, feedback, baselines, and tuning workflow.
- Out of scope: live order submission, AI-led live ranking, exchange credential flows.
- Validation shape: baseline comparisons, replay-safe alert/outcome behavior, saved simulation records, and delivery plus review surfaces.

## First Recommended Next Step

Completed initiative-1 slices already archived:

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`
- `plans/completed/raw-event-log-boundary/`
- `plans/completed/dashboard-shell-and-summary-strip/`
- `plans/completed/replay-run-manifests-and-ordering/`
- `plans/completed/world-usa-composite-snapshots/`
- `plans/completed/backfill-checkpoints-and-audit-trail/`
- `plans/completed/market-quality-and-divergence-buckets/`
- `plans/completed/replay-retention-and-safety-validation/`
- `plans/completed/raw-storage-and-replay-foundation/`
- `plans/completed/symbol-and-global-regime-state/`
- `plans/completed/market-state-current-query-contracts/`
- `plans/completed/dashboard-query-adapters-and-trust-state/`
- `plans/completed/market-state-history-and-audit-reads/`
- `plans/completed/dashboard-detail-panels-and-symbol-switching/`
- `plans/completed/dashboard-negative-state-mobile-a11y/`
- `plans/completed/dashboard-fixture-smoke-matrix/`

The visibility dashboard core child queue is complete. Run `program-refining` for the remaining initiative-1 epic before any further implementation:

- `plans/epics/slow-context-panel/`

Source refinement handoff:

- `plans/epics/slow-context-panel/00-overview.md`

Current completion note:

- `plans/completed/dashboard-fixture-smoke-matrix/` now holds the archived implementation and validation evidence for the final `visibility-dashboard-core` child feature.

Do not start initiative 2 feature planning until initiative 1 has at least:

- stable contract boundaries
- deterministic replay
- composite features and regime outputs
- feed health and degradation states exposed to downstream consumers
