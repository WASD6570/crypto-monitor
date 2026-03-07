# Alerting And Evaluation Dependencies

## Suggested Order

1. `alert-generation-and-hygiene`
2. `tactical-risk-state-and-permissioning`
3. `alert-delivery-and-routing`
4. `outcome-evaluation`
5. `simulated-execution`
6. `operator-feedback-and-notes`
7. `replay-and-analytics-ui`
8. `baseline-comparison-and-tuning`

## Dependency Notes

### `alert-generation-and-hygiene`

- Depends on: initiative 1 state outputs, contracts, replay inputs
- Unlocks: delivery, outcomes, review
- Risk: high

### `tactical-risk-state-and-permissioning`

- Depends on: market state from initiative 1 plus alert payloads
- Unlocks: severity ceilings and future execution safety
- Risk: high

### `alert-delivery-and-routing`

- Depends on: alert payloads and permissioning
- Unlocks: actual user attention loop
- Risk: medium

### `outcome-evaluation`

- Depends on: alerts, replay/raw market data, config versions
- Unlocks: truth, analytics, and baseline comparison
- Risk: high

### `simulated-execution`

- Depends on: alert payloads, L2 data quality, cost-model config, outcomes surfaces
- Unlocks: realistic viability review
- Risk: high

### `operator-feedback-and-notes`

- Depends on: delivered alerts and stored outcomes
- Unlocks: human review memory and future tuning context
- Risk: medium

### `replay-and-analytics-ui`

- Depends on: outcomes, simulations, replay surfaces, feedback
- Unlocks: efficient review loop
- Risk: medium

### `baseline-comparison-and-tuning`

- Depends on: outcomes, replay, baselines, config versioning, optionally offline analysis support
- Unlocks: disciplined improvement and rollback
- Risk: medium-high

## Must-Have Inputs From Initiative 1

- state and feature contracts are versioned
- replay inputs are deterministic
- feed degradation signals are included in alert permissioning inputs
- market state and fragmentation state are queryable by symbol and globally
