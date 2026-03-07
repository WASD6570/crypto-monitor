# Visibility Foundation Handoff

## Feature Queue

1. `canonical-contracts-and-fixtures`
2. `market-ingestion-and-feed-health`
3. `raw-storage-and-replay-foundation`
4. `world-usa-composites-and-market-state`
5. `visibility-dashboard-core`
6. `slow-context-panel`

## Planning Waves

### Wave 1

- `canonical-contracts-and-fixtures`
- Why now: every later slice depends on shared event, feature, replay, and market-state vocabulary.

### Wave 2

- `market-ingestion-and-feed-health`
- Why now: trusted live inputs and degradation signals must exist before replay, state, and dashboard consumers can be implemented safely.

### Wave 3

- `raw-storage-and-replay-foundation`
- `visibility-dashboard-core`
- Why parallel: replay depends on ingestion semantics and contracts; dashboard IA/query planning can proceed once the market-state contract boundary is understood, without redefining ingestion or replay semantics.

### Wave 4

- `world-usa-composites-and-market-state`
- Why later: this slice depends on contracts, ingestion, and replay semantics being stable.

### Wave 5

- `slow-context-panel`
- Why later: this is intentionally non-blocking and should not shape the core realtime state path.

## Child Plan Seeds

### `plans/canonical-contracts-and-fixtures/`

- Problem statement: visibility cannot be trusted if events, features, and market-state payloads are ambiguous.
- In scope: event families, feature families, replay/state output contracts, canonical symbol naming, deterministic fixtures.
- Out of scope: storage implementation and UI work.
- Validation shape: schema checks and deterministic fixture conventions.

### `plans/market-ingestion-and-feed-health/`

- Problem statement: the user needs continuous multi-venue market state, which requires resilient ingestion and explicit degradation handling.
- In scope: WS resilience, snapshot bootstrap, delta sequence integrity, feed freshness, reconnect and gap metrics.
- Out of scope: composites, alert logic, UI presentation beyond health payloads.
- Validation shape: resync smoke tests, staleness scenarios, normalized event fixtures.

### `plans/raw-storage-and-replay-foundation/`

- Problem statement: state surfaces must be reproducible, auditable, and backfill-safe.
- In scope: append-only raw persistence, replay contracts, deterministic replay loop, audit hooks.
- Out of scope: alert outcomes, simulation, web replay UX.
- Validation shape: repeated identical replay runs with identical outputs.

### `plans/world-usa-composites-and-market-state/`

- Problem statement: the product needs a coherent cross-venue view and a clear market-state gate before it can ask the user to trust future alerts.
- In scope: composites, weighting, clamping, fragmentation metrics, market-quality metrics, 5m market state.
- Out of scope: setup A/B/C logic, notifications, outcomes.
- Validation shape: deterministic fixture runs and regime-transition checks.

### `plans/visibility-dashboard-core/`

- Problem statement: state is only useful if the user can read it quickly.
- In scope: symbol overview, microstructure, derivatives view, feed health panel, mobile-safe dense UI.
- Out of scope: alert replay, saved simulations, deep analytics.
- Validation shape: UI smoke checks and contract fixture rendering.

### `plans/slow-context-panel/`

- Problem statement: slower USA context should be visible without muddying realtime state.
- In scope: CME volume/OI and ETF flow ingestion surfaces, slower context cards or panels, separate timestamps and cadence indicators.
- Out of scope: using slow context as a hard realtime gating dependency by default.
- Validation shape: slow-feed fixture checks and panel rendering.
