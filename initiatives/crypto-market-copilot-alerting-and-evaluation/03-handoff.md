# Alerting And Evaluation Handoff

## Refined Epic Queue

These slices already exist as refined epic context under `plans/epics/` and are ready for `feature-planning` when this initiative is prioritized.

1. `plans/epics/alert-generation-and-hygiene/`
2. `plans/epics/tactical-risk-state-and-permissioning/`

## Execution State

- Initiative status: `ready_to_plan`
- Next recommended epic: `plans/epics/alert-generation-and-hygiene/`
- Parallel-safe now: `plans/epics/tactical-risk-state-and-permissioning/`

| Epic | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/alert-generation-and-hygiene/` | `ready_to_plan` | Visibility foundation complete | `plans/epics/tactical-risk-state-and-permissioning/` | Run `feature-planning` | Default Wave 1 epic |
| `plans/epics/tactical-risk-state-and-permissioning/` | `ready_to_plan` | Visibility foundation complete | `plans/epics/alert-generation-and-hygiene/` | Plan in parallel when capacity allows | Second Wave 1 epic |
| `plans/epics/alert-delivery-and-routing/` | `blocked` | Wave 1 epic outputs | - | Wait for alert payloads and permissioning posture to settle | Wave 2 epic |
| `plans/epics/outcome-evaluation/` | `blocked` | Wave 1 epic outputs | - | Wait for alert payloads to settle | Wave 2 epic |
| `plans/epics/simulated-execution/` | `blocked` | Wave 2 outputs | `plans/epics/operator-feedback-and-notes/` | Wait for alerts and outcomes to settle | Wave 3 epic |
| `plans/epics/operator-feedback-and-notes/` | `blocked` | Wave 2 outputs | `plans/epics/simulated-execution/` | Wait for alerts and outcomes to settle | Wave 3 epic |
| `plans/epics/replay-and-analytics-ui/` | `blocked` | Wave 2 and Wave 3 outputs | `plans/epics/baseline-comparison-and-tuning/` | Wait for alert, outcome, and review data to settle | Wave 4 epic |
| `plans/epics/baseline-comparison-and-tuning/` | `blocked` | Wave 2 and Wave 3 outputs | `plans/epics/replay-and-analytics-ui/` | Wait for alert, outcome, and review data to settle | Wave 4 epic |

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

## Refined Epics

### `plans/epics/alert-generation-and-hygiene/`

- Problem statement: the product needs to notify the user only when conditions deserve attention, not every time price twitches.
- In scope: setups A/B/C, 30s trigger and 2m validator wiring, 5m market-state gating, dedupe, cooldown, clustering, severity.
- Out of scope: outcome metrics, simulation, UI analytics.
- Validation shape: replay-based alert scenarios, duplicate suppression checks, setup-level fixture tests.

### `plans/epics/tactical-risk-state-and-permissioning/`

- Problem statement: alert severity must respect both market state and tactical risk state so bad days do not create more noise.
- In scope: `NORMAL/DE-RISK/STOP`, soft and hard drawdown rules, weekly stop, no-operate interaction, transition logging.
- Out of scope: real order flattening, live execution auth.
- Validation shape: deterministic state-transition tests and structured log checks.

### `plans/epics/alert-delivery-and-routing/`

- Problem statement: an alert product without reliable delivery still forces the user to stare at dashboards.
- In scope: in-app stream, Telegram, optional webhook, severity/state-preserving payloads, delivery audit trail.
- Out of scope: email and Slack by default.
- Validation shape: delivery smoke checks and payload contract verification.

### `plans/epics/outcome-evaluation/`

- Problem statement: every alert needs objective proof of what happened afterward.
- In scope: target/invalidation ordering, MAE/MFE, time-to-hit, timeout, regime-tagged outcome records, net viability fields.
- Out of scope: simulated fills and leverage assumptions beyond the outcome boundary.
- Validation shape: deterministic outcome fixtures and replay-driven scenario checks.

### `plans/epics/simulated-execution/`

- Problem statement: directional correctness is not enough if costs and latency erase the edge.
- In scope: spot long and perp long/short simulation, latency defaults, fee defaults, slippage model, low-confidence labels, saved runs.
- Out of scope: real orders and exchange credentials.
- Validation shape: saved simulation payload checks, L2-walk vs fallback cases, cost-model scenario tests.

### `plans/epics/operator-feedback-and-notes/`

- Problem statement: the user needs a lightweight memory system for what felt useful or noisy without editing configs ad hoc.
- In scope: save, dismiss, thumbs up/down, `good setup bad timing`, `useful context only`, notes.
- Out of scope: automatic retraining or live config mutation.
- Validation shape: UI actions, persistence checks, feedback-to-review display.

### `plans/epics/replay-and-analytics-ui/`

- Problem statement: the user needs one place to inspect why an alert fired and whether it actually worked.
- In scope: alert stream detail, replay window, outcome summary, simulation summary, regime slices, best/avoid condition summaries.
- Out of scope: realtime alert logic itself.
- Validation shape: UI smoke tests and contract-driven rendering.

### `plans/epics/baseline-comparison-and-tuning/`

- Problem statement: alert improvements need disciplined comparison against naive controls and safe rollback rules.
- In scope: baseline definitions, config versioning, replay comparison workflow, graduation rules, rollback rules, review reports.
- Out of scope: black-box AI optimization in the live path.
- Validation shape: baseline comparison reports and config-version reproducibility checks.
