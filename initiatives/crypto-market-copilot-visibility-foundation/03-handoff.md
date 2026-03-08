# Visibility Foundation Handoff

## Completed Feature Plans

1. `plans/completed/canonical-contracts-and-fixtures/`
2. `plans/completed/market-ingestion-and-feed-health/`
3. `plans/completed/raw-event-log-boundary/`
4. `plans/completed/dashboard-shell-and-summary-strip/`
5. `plans/completed/replay-run-manifests-and-ordering/`
6. `plans/completed/world-usa-composite-snapshots/`
7. `plans/completed/backfill-checkpoints-and-audit-trail/`
8. `plans/completed/market-quality-and-divergence-buckets/`
9. `plans/completed/replay-retention-and-safety-validation/`
10. `plans/completed/raw-storage-and-replay-foundation/`
11. `plans/completed/symbol-and-global-regime-state/`
12. `plans/completed/market-state-current-query-contracts/`
13. `plans/completed/dashboard-query-adapters-and-trust-state/`
14. `plans/completed/market-state-history-and-audit-reads/`
15. `plans/completed/dashboard-detail-panels-and-symbol-switching/`
16. `plans/completed/dashboard-negative-state-mobile-a11y/`
17. `plans/completed/dashboard-fixture-smoke-matrix/`
18. `plans/completed/slow-context-source-boundaries/`
19. `plans/completed/slow-context-query-surface-and-freshness/`
20. `plans/completed/slow-context-dashboard-panel/`

## Epic Queue

All initiative-1 epics now have their child features implemented and archived.

- `world-usa-composites-and-market-state` now has all child features implemented, including `market-state-history-and-audit-reads`.
- `visibility-dashboard-core` now has all child features implemented, including `dashboard-fixture-smoke-matrix`.
- `slow-context-panel` now has all child features implemented, including `slow-context-dashboard-panel`.

## Planning Waves

### Wave 1

- `canonical-contracts-and-fixtures` (completed)
- Why now: every later slice depends on shared event, feature, replay, and market-state vocabulary.

### Wave 2

- `market-ingestion-and-feed-health` (completed)
- Why now: trusted live inputs and degradation signals must exist before replay, state, and dashboard consumers can be implemented safely.

### Wave 3

- `raw-storage-and-replay-foundation` (completed)
- `visibility-dashboard-core`
- Why parallel: replay depended on ingestion semantics and contracts; dashboard IA/query planning could proceed once the market-state contract boundary was understood, without redefining ingestion or replay semantics.

### Wave 4

- `world-usa-composites-and-market-state`
- Why later: this slice depends on contracts and ingestion semantics being stable; replay-backed history is now unblocked, but current-state consumer sequencing still depends on regime outputs.

### Wave 5

- `slow-context-panel`
- Why later: this is intentionally non-blocking and should not shape the core realtime state path.

## Epic Seeds

### `plans/completed/canonical-contracts-and-fixtures/`

- Problem statement: visibility cannot be trusted if events, features, and market-state payloads are ambiguous.
- In scope: event families, feature families, replay/state output contracts, canonical symbol naming, deterministic fixtures.
- Out of scope: storage implementation and UI work.
- Validation shape: schema checks and deterministic fixture conventions.

### `plans/completed/market-ingestion-and-feed-health/`

- Problem statement: the user needs continuous multi-venue market state, which requires resilient ingestion and explicit degradation handling.
- In scope: WS resilience, snapshot bootstrap, delta sequence integrity, feed freshness, reconnect and gap metrics.
- Out of scope: composites, alert logic, UI presentation beyond health payloads.
- Validation shape: resync smoke tests, staleness scenarios, normalized event fixtures.

### `plans/completed/raw-storage-and-replay-foundation/`

- Problem statement: state surfaces must be reproducible, auditable, and backfill-safe.
- In scope: append-only raw persistence, replay contracts, deterministic replay loop, audit hooks.
- Out of scope: alert outcomes, simulation, web replay UX.
- Validation shape: repeated identical replay runs with identical outputs.
- Refinement and completion artifacts: `plans/completed/raw-storage-and-replay-foundation/90-refinement-map.md`, `plans/completed/raw-storage-and-replay-foundation/91-child-plan-seeds.md`, `plans/completed/raw-storage-and-replay-foundation/92-refinement-handoff.md`

### `plans/epics/world-usa-composites-and-market-state/`

- Problem statement: the product needs a coherent cross-venue view and a clear market-state gate before it can ask the user to trust future alerts.
- In scope: composites, weighting, clamping, fragmentation metrics, market-quality metrics, 5m market state.
- Out of scope: setup A/B/C logic, notifications, outcomes.
- Validation shape: deterministic fixture runs and regime-transition checks.

### `plans/epics/visibility-dashboard-core/`

- Problem statement: state is only useful if the user can read it quickly.
- In scope: symbol overview, microstructure, derivatives view, feed health panel, mobile-safe dense UI.
- Out of scope: alert replay, saved simulations, deep analytics.
- Validation shape: UI smoke checks and contract fixture rendering.
- Refinement artifacts: `plans/epics/visibility-dashboard-core/90-refinement-map.md`, `plans/epics/visibility-dashboard-core/91-child-plan-seeds.md`, `plans/epics/visibility-dashboard-core/92-refinement-handoff.md`

### `plans/epics/slow-context-panel/`

- Problem statement: slower USA context should be visible without muddying realtime state.
- In scope: CME volume/OI and ETF flow ingestion surfaces, slower context cards or panels, separate timestamps and cadence indicators.
- Out of scope: using slow context as a hard realtime gating dependency by default.
- Validation shape: slow-feed fixture checks and panel rendering.
