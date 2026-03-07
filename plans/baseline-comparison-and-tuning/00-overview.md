# Baseline Comparison And Tuning Overview

## Ordered Implementation Plan

1. Define deterministic baseline families, comparability keys, and evaluation joins shared across production and control alerts.
2. Add config-version snapshots, candidate manifests, and a lightweight graduation workflow that always uses replay-backed evidence.
3. Add reporting outputs, approval rules, and rollback rules so config tuning stays auditable and reversible.
4. Validate determinism, join correctness, negative cases, and candidate promotion or rollback behavior with replay-first tests.

## Problem Statement

Initiative 2 needs a disciplined way to prove that the alert stack is better than simple controls before thresholds change in production. The system should compare production alerts against explicit naive baselines, pin every comparison to a config snapshot, and let operators graduate or roll back config versions without drifting into live AI optimization or ad hoc threshold edits.

## Role In Initiative 2

- This is slice 8 of `crypto-market-copilot-alerting-and-evaluation`.
- It closes the initiative trust loop after alert generation, delivery, outcomes, simulation, feedback, and replay review exist.
- It turns review data into a conservative tuning workflow instead of letting operators tweak configs from instinct.
- It is the guardrail that keeps Initiative 2 improvements measurable against the required baselines in `docs/specs/crypto-market-copilot-program/02-product-success.md`.

## Role In The Disciplined Tuning Loop

1. Production and baseline alerts emit comparable records.
2. Outcome evaluation and later simulation attach the same vocabulary to both.
3. Replay runs compare the active config and a candidate config over pinned windows plus a recent rolling window.
4. Reports show whether the candidate improves or preserves precision and net viability without worsening fragmented-market false positives.
5. A lightweight human approval promotes the candidate to active or rejects it; later rolling degradation triggers rollback to the last known-good snapshot.

## In Scope

- deterministic control definitions for `naive-breakout`, `naive-vwap-reversion`, and `single-venue-trigger`
- comparable joins across alert, baseline, outcome, and simulation records
- config-version snapshots and immutable candidate manifests
- replay-backed promotion, graduation, and rollback workflow for alert configs
- lightweight approval rules for config changes with clear defaults
- reporting outputs for replay, rolling-window, baseline-delta, and rollback decisions
- explicit boundaries between live deterministic tuning inputs and optional offline analysis

## Out Of Scope

- black-box AI optimization, auto-mutating thresholds, or model-selected live configs
- changing baseline logic from user feedback alone
- heavyweight governance, approval committees, or ticket workflows
- implementing new alert families beyond the existing production setup families and required baselines
- replacing outcome evaluation or simulated execution with a tuning-specific scoring system
- requiring Python or notebooks in the live path

## Safe Defaults

- Candidate promotion default: `draft -> candidate -> active`, with `draft` for local iteration, `candidate` for replay-complete snapshots, and `active` for the single live config.
- Replay windows default: at least 14 recent days for rolling evaluation plus 3 pinned replay fixtures that include one clean trend day, one fragmented day, and one degraded-feed day.
- Promotion threshold default: candidate must improve or preserve precision and positive net viability, and must not worsen fragmented-market false positives.
- Failure threshold default: reject or roll back if candidate loses more than 3 absolute precision points, loses more than 5 relative percent of positive net viability, or raises fragmented-market false positives by more than 2 absolute points versus the active config on the same comparable window.
- Approval default: automated replay and report checks must pass, then one human approver from the repo maintainer or designated operator role may activate the config; no second committee layer is required unless production policy later adds one.
- Research boundary default: Python may produce offline analysis notes or candidate suggestions, but only versioned config snapshots and Go-owned replay evidence may affect live graduation.

## Requirements

- Keep live comparison, config selection, and rollback decisions deterministic and replay-backed.
- Ensure baseline alerts, production alerts, outcomes, and simulation results are directly comparable by symbol, setup family, horizon, regime, and config version.
- Preserve immutable config snapshots so later replays use the exact policy state that produced a report or promotion.
- Keep one active config per environment and one explicit rollback target to avoid ambiguous live state.
- Separate human review assistance from decision authority; no AI or notebook output may directly change live configs.
- Keep Python optional and offline-only.

## Target Repo Areas

- `services/alert-engine`
- `services/outcome-engine`
- `services/replay-engine`
- `services/simulation-api`
- `libs/go`
- `configs/*`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `schemas/json/replay`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`
- `tests/parity` for optional offline parity helpers only

## Design Overview

### Why Baselines Must Be Explicit

- Initiative 2 promises regime-aware alert quality, not just more complex alert logic.
- The required baselines are the honesty check for whether validation, multi-venue confirmation, and market-state gating actually help.
- Baselines should stay simple, deterministic, and intentionally weaker than the production stack without becoming toy logic that cannot be compared fairly.

### Baseline Families

- `naive-breakout`: level break trigger without 2m validator and without 5m market-state gating; used mainly against production setup `A`.
- `naive-vwap-reversion`: fixed-distance-from-intraday-VWAP trigger without deeper absorption, fakeout, or regime context; used mainly against production setup `B`.
- `single-venue-trigger`: one-venue trigger without WORLD vs USA confirmation or fragmentation awareness; used to prove multi-venue confirmation adds value across setups, especially fragmented conditions.
- All baselines must emit the same minimum alert fields needed by outcome evaluation and reporting, even if their logic is intentionally simpler.

### Comparable Evaluation Joins

- Comparison should join on stable dimensions, not on fuzzy narrative similarity.
- Minimum join keys should include `symbol`, `direction`, `setupFamily` or mapped baseline family, `horizon`, `openBucket`, `configVersion`, `evaluationSource`, and regime slice fields.
- A production alert and a baseline alert do not need the same `alertId`; they need the same comparable evaluation window and outcome vocabulary.
- Join outputs should support both pairwise comparisons and aggregate summaries by regime, symbol, and timeframe.

### Config Snapshots And Candidate Manifests

- Every tuning run should materialize an immutable snapshot that includes alert thresholds, gating policy, dedupe policy, baseline parameters, outcome settings referenced by the run, and simulation cost inputs if used.
- Candidate manifests should record parent active version, replay windows used, report artifact paths, and whether the candidate is `draft`, `candidate`, `active`, `rolled-back`, or `rejected`.
- Promotion should activate the whole snapshot, not cherry-pick individual threshold fields after the report is generated.

### Graduation And Rollback

- Graduation should be evidence-based and lightweight: pass deterministic replay checks, pass rolling-window comparison, attach report artifacts, and obtain one human approval.
- Rollback should be mechanical: point the environment back to the last known-good config snapshot and preserve the failed candidate's report for later diagnosis.
- Rolling degradation should never be fixed by silent in-place edits; it should always create a new candidate or revert to a prior active snapshot.

### Live Vs Research Boundary

- Optional offline research may suggest candidate thresholds or explain report patterns.
- Live graduation authority comes only from deterministic replay outputs, versioned config snapshots, and human approval.
- No notebook, Python script, or AI agent may directly mutate active production config.

## ASCII Flow

```text
active config ----+----------------------------------------------+
                  |                                              |
candidate config -+--> replay-engine runs pinned + rolling windows|
                                                                 |
production alerts --------------------+                          |
baseline alerts ----------------------+--> comparable joins -----+--> tuning report
outcome records ----------------------+                          |    - precision deltas
simulation records (optional input) --+                          |    - fragmentation false positives
regime slices ------------------------+                          |    - net viability deltas
                                                                 |
                                                                 v
                                                     approval + thresholds
                                                     - pass -> promote snapshot active
                                                     - fail -> reject or roll back
                                                                 |
                                                                 v
                                                    immutable config history
                                                    - active
                                                    - prior known-good
                                                    - rejected / rolled-back
```
