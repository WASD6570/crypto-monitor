# Alerting And Evaluation Feature Map

## 1. `alert-generation-and-hygiene`

- Goal: implement setups A/B/C with severity, dedupe, cooldown, clustering, and full trace payloads.
- Primary repo areas: `services/alert-engine`, `schemas/json/alerts`, `configs/*`, `tests/integration`
- Why it stands alone: this is the first point where the system starts asking for user attention.

## 2. `tactical-risk-state-and-permissioning`

- Goal: compute `NORMAL/DE-RISK/STOP` and combine it with market state to determine what alerts may escalate.
- Primary repo areas: `services/risk-engine`, `services/alert-engine`, `schemas/json/outcomes`, `docs/runbooks`
- Why it stands alone: permissioning and review logic should not be hidden inside setup code.

## 3. `alert-delivery-and-routing`

- Goal: deliver alerts through the UI, Telegram, and optional webhook with severity and state preserved.
- Primary repo areas: `services/alert-engine`, `apps/web`, `libs/ts`, `configs/*`
- Why it stands alone: delivery is a product surface, not just a transport detail.

## 4. `outcome-evaluation`

- Goal: compute target hits, invalidations, MAE/MFE, time-to-hit, timeout, and net viability records for every alert.
- Primary repo areas: `services/outcome-engine`, `schemas/json/outcomes`, `tests/integration`
- Why it stands alone: objective review is central to the product promise.

## 5. `simulated-execution`

- Goal: provide saved what-if execution with latency, slippage, fees, leverage, and confidence labels.
- Primary repo areas: `services/simulation-api`, `schemas/json/simulation`, `apps/web`, `tests/integration`
- Why it stands alone: simulation is related to outcomes but should be testable independently.

## 6. `operator-feedback-and-notes`

- Goal: let the user save, dismiss, rate, and annotate alerts without mutating live logic directly.
- Primary repo areas: `apps/web`, `services/outcome-engine`, `schemas/json/outcomes`
- Why it stands alone: this creates review memory and future tuning context.

## 7. `replay-and-analytics-ui`

- Goal: display alert replay, outcome details, regime slices, and saved simulation runs.
- Primary repo areas: `apps/web`, `libs/ts`, `tests/e2e`
- Why it stands alone: review UX depends on backend truth but is a separate operator workflow.

## 8. `baseline-comparison-and-tuning`

- Goal: compare production alerts against naive baselines and manage config-versioned tuning and rollback.
- Primary repo areas: `services/outcome-engine`, `apps/research`, `libs/python`, `tests/parity`, `docs/runbooks`
- Why it stands alone: experimentation should be explicit, auditable, and isolated from live alert emission.
