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

Run `feature-planning` for the first slice inside initiative 1:

- `canonical-contracts-and-fixtures`

Do not start initiative 2 feature planning until initiative 1 has at least:

- stable contract boundaries
- deterministic replay
- composite features and regime outputs
- feed health and degradation states exposed to downstream consumers
