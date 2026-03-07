# Visibility Foundation Dependencies

## Suggested Order

1. `canonical-contracts-and-fixtures`
2. `market-ingestion-and-feed-health`
3. `raw-storage-and-replay-foundation`
4. `world-usa-composites-and-market-state`
5. `visibility-dashboard-core`
6. `slow-context-panel`

## Dependency Notes

### `canonical-contracts-and-fixtures`

- Depends on: none
- Unlocks: every later slice
- Risk: high

### `market-ingestion-and-feed-health`

- Depends on: `canonical-contracts-and-fixtures`
- Unlocks: raw persistence, replay, state computation
- Risk: high

### `raw-storage-and-replay-foundation`

- Depends on: `canonical-contracts-and-fixtures`, `market-ingestion-and-feed-health`
- Unlocks: deterministic audits and later alert replay
- Risk: high

### `world-usa-composites-and-market-state`

- Depends on: contracts, ingestion, replay
- Unlocks: dashboards, future alert gating, fragmentation analysis
- Risk: high

### `visibility-dashboard-core`

- Depends on: market state outputs and stable query surfaces
- Unlocks: the first real operator workflow
- Risk: medium

### `slow-context-panel`

- Depends on: dashboard shell and context contracts
- Unlocks: richer interpretation, not core correctness
- Risk: medium

## Initiative-2 Handoff Requirements

Initiative 2 should not start feature implementation until this initiative provides:

- stable contracts for features and regime outputs
- deterministic replay for raw data and state outputs
- visible feed health and degradation states
- queryable market-state surfaces that alerting can depend on
