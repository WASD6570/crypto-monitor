# Alerting And Evaluation Handoff

## Feature Queue

1. `alert-generation-and-hygiene`
2. `tactical-risk-state-and-permissioning`
3. `alert-delivery-and-routing`
4. `outcome-evaluation`
5. `simulated-execution`
6. `operator-feedback-and-notes`
7. `replay-and-analytics-ui`
8. `baseline-comparison-and-tuning`

## Planning Waves

### Wave 1

- `alert-generation-and-hygiene`
- `tactical-risk-state-and-permissioning`
- Why now: alert generation and permission ceilings define the core alert decision model for every later slice.

### Wave 2

- `alert-delivery-and-routing`
- `outcome-evaluation`
- Why parallel: both depend on stable alert payloads; delivery handles interruption flow while outcome evaluation handles review truth.

### Wave 3

- `simulated-execution`
- `operator-feedback-and-notes`
- Why parallel: both depend on alerts and outcomes, but they serve different review functions and do not redefine the same contract boundaries.

### Wave 4

- `replay-and-analytics-ui`
- `baseline-comparison-and-tuning`
- Why later: both depend on alert, outcome, and review data being stable enough to inspect and compare.

## Child Plan Seeds

### `plans/alert-generation-and-hygiene/`

- Problem statement: the product needs to notify the user only when conditions deserve attention, not every time price twitches.
- In scope: setups A/B/C, 30s trigger and 2m validator wiring, 5m market-state gating, dedupe, cooldown, clustering, severity.
- Out of scope: outcome metrics, simulation, UI analytics.
- Validation shape: replay-based alert scenarios, duplicate suppression checks, setup-level fixture tests.

### `plans/tactical-risk-state-and-permissioning/`

- Problem statement: alert severity must respect both market state and tactical risk state so bad days do not create more noise.
- In scope: `NORMAL/DE-RISK/STOP`, soft and hard drawdown rules, weekly stop, no-operate interaction, transition logging.
- Out of scope: real order flattening, live execution auth.
- Validation shape: deterministic state-transition tests and structured log checks.

### `plans/alert-delivery-and-routing/`

- Problem statement: an alert product without reliable delivery still forces the user to stare at dashboards.
- In scope: in-app stream, Telegram, optional webhook, severity/state-preserving payloads, delivery audit trail.
- Out of scope: email and Slack by default.
- Validation shape: delivery smoke checks and payload contract verification.

### `plans/outcome-evaluation/`

- Problem statement: every alert needs objective proof of what happened afterward.
- In scope: target/invalidation ordering, MAE/MFE, time-to-hit, timeout, regime-tagged outcome records, net viability fields.
- Out of scope: simulated fills and leverage assumptions beyond the outcome boundary.
- Validation shape: deterministic outcome fixtures and replay-driven scenario checks.

### `plans/simulated-execution/`

- Problem statement: directional correctness is not enough if costs and latency erase the edge.
- In scope: spot long and perp long/short simulation, latency defaults, fee defaults, slippage model, low-confidence labels, saved runs.
- Out of scope: real orders and exchange credentials.
- Validation shape: saved simulation payload checks, L2-walk vs fallback cases, cost-model scenario tests.

### `plans/operator-feedback-and-notes/`

- Problem statement: the user needs a lightweight memory system for what felt useful or noisy without editing configs ad hoc.
- In scope: save, dismiss, thumbs up/down, `good setup bad timing`, `useful context only`, notes.
- Out of scope: automatic retraining or live config mutation.
- Validation shape: UI actions, persistence checks, feedback-to-review display.

### `plans/replay-and-analytics-ui/`

- Problem statement: the user needs one place to inspect why an alert fired and whether it actually worked.
- In scope: alert stream detail, replay window, outcome summary, simulation summary, regime slices, best/avoid condition summaries.
- Out of scope: realtime alert logic itself.
- Validation shape: UI smoke tests and contract-driven rendering.

### `plans/baseline-comparison-and-tuning/`

- Problem statement: alert improvements need disciplined comparison against naive controls and safe rollback rules.
- In scope: baseline definitions, config versioning, replay comparison workflow, graduation rules, rollback rules, review reports.
- Out of scope: black-box AI optimization in the live path.
- Validation shape: baseline comparison reports and config-version reproducibility checks.
