# Alert Generation And Hygiene

## Ordered Implementation Plan

1. Define setup registry, time-horizon responsibilities, and market-state gating in Go service-owned alert generation logic.
2. Add deterministic deduplication, cooldown, clustering, and fragmentation handling so emitted alerts stay bounded.
3. Define versioned alert payload expectations and consumer contracts for delivery and outcome services.
4. Validate replay determinism, degraded-feed behavior, and negative-path suppression before delivery work begins.

## Problem Statement

Initiative 2 needs a first alerting slice that turns trusted composite and market-state outputs into high-signal notifications without recreating the noisy experience the product is supposed to replace. The system must emit only bounded, explainable alerts for setups A, B, and C, keep all live logic deterministic in Go, and preserve enough provenance for later delivery, outcome, simulation, and review features.

## Role In Initiative 2

- This is slice 1 of `crypto-market-copilot-alerting-and-evaluation`.
- It is the first feature that converts Initiative 1 market understanding into a direct first-user interruption.
- It establishes the alert-generation contract that later slices depend on: tactical risk gating, delivery routing, outcomes, simulation, and analytics.
- If this slice is noisy or opaque, every later Initiative 2 feature inherits bad inputs.

## First-User Workflow

1. The user trusts the current BTC or ETH 5m market state from Initiative 1.
2. A 30s setup candidate appears only when a deterministic setup rule sees a possible opportunity.
3. A 2m validator either confirms or rejects that candidate so short-lived noise does not become an alert.
4. The 5m symbol state and global ceiling decide whether the event is actionable or informational.
5. The user receives one bounded alert payload that says what setup fired, why it was allowed, what degraded confidence, and which horizon should matter next.

## In Scope

- setup families `A`, `B`, and `C` at planning level only
- 30s trigger, 2m validator, and 5m market-state hierarchy
- severity policy tied to setup confidence and state ceilings
- deterministic market-state gating now, with explicit hooks for later risk-state gating
- deduplication, cooldowns, clustering, fragmentation handling, and degraded-feed suppression
- config and versioning expectations for alert generation decisions
- alert payload expectations for later delivery and outcome consumers
- replay, fixture, and negative-case validation requirements

## Out Of Scope

- tactical risk-state business rules beyond defining integration points for the next slice
- delivery transport behavior beyond payload needs for Telegram, webhook, and UI consumers
- outcome scoring, simulation logic, or operator feedback workflows
- AI-generated live ranking, scoring, or explanation
- concrete schema files, migrations, or production threshold values that require implementation evidence

## Safe Defaults For Vague Areas

- Default horizon policy: 30s may nominate, 2m may validate, 5m may permit or cap; no alert becomes actionable from a single 30s observation alone.
- Default severity posture: downgrade on uncertainty; never escalate because data is missing, fragmented, or stale.
- Default setup posture: `A` and `B` can become actionable when fully validated and state-allowed; `C` starts more conservative and should require stronger confirmation before matching higher severities.
- Default fragmentation posture: elevated WORLD vs USA disagreement is suppression or downgrade input, not extra urgency.
- Default degraded-feed posture: emit informational context only when explanation value remains high; suppress duplicate action-oriented alerts during critical degradation.
- Default config posture: setup thresholds, dedupe windows, cooldown windows, clustering policy, and state/severity mappings live in versioned config snapshots.

## Requirements

- Keep all alert nomination, validation, gating, dedupe, and severity logic in Go services or shared Go helpers.
- Use Initiative 1 service outputs as source of truth; `apps/web` and delivery surfaces render service-owned alert decisions and reasons without client recomputation.
- Preserve deterministic replay for the same canonical inputs, feature outputs, config version, and code version.
- Keep event time and processing time distinct in alert records so later outcome timing stays auditable.
- Make every emitted or suppressed decision explainable through stable reason codes and contributing state references.
- Bound alert volume by design so the first user can trust interruptions instead of watching a stream of micro-events.
- Keep Python optional and offline-only for later analysis; it must not be required for live alert generation.

## Target Repo Areas

- `services/alert-engine`
- `services/feature-engine`
- `services/regime-engine`
- `services/risk-engine` for later integration only
- `libs/go`
- `configs/*`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`
- `tests/parity` only for optional offline parity fixtures

## Design Overview

### Why Alerts Must Stay Bounded And Explainable

- The product promise is reduced screen-watching, not faster noise delivery.
- Fragmented crypto markets produce repeated micro-signals from the same move; without hygiene, the user gets spammed by correlated duplicates.
- Later evaluation depends on stable causality. If multiple alerts describe the same underlying event differently, outcome attribution becomes misleading.
- Explainability is required for time-to-trust: the user should understand why the alert fired, why it was allowed, and why it was not blocked.

### Setup Intent At Planning Level

- `A`: momentum or breakout-style alert family that benefits most from 30s nomination plus 2m confirmation and 5m regime permissioning.
- `B`: mean-reversion or reclaim-style alert family that should explicitly resist fakeouts and fragmented liquidity before escalation.
- `C`: context-heavy discretionary support alert family that may start as more informational until enough evidence exists to justify stronger severities.
- Exact signal math belongs to implementation and replay evidence, not this plan; the requirement is deterministic, versioned, and explainable setup logic.

### Time-Horizon Hierarchy

- `30s` is the earliest signal horizon and should answer: is something interesting starting now?
- `2m` is the validator horizon and should answer: did the move persist enough to deserve interruption?
- `5m` is the permissioning horizon and should answer: is this market state currently fit for action, watch-only, or no-operate?
- The hierarchy must be asymmetric: a higher horizon can block a lower horizon, but a lower horizon cannot override a higher-horizon degradation.

### Severity Policy

- Base severities should be limited to `INFO`, `WATCH`, and `ACTIONABLE` during this slice, with exact delivery mappings deferred.
- `INFO` is for context worth logging or surfacing in UI even when state ceilings prevent action.
- `WATCH` is for validated but still constrained conditions where the user may want awareness without urgency.
- `ACTIONABLE` is reserved for validated setup events that pass 5m market-state gating and are not later capped by tactical risk state.
- Global `NO-OPERATE` and later risk `STOP` states must cap all alerts to `INFO`.
- Global or symbol `WATCH` states must prevent escalation above `WATCH`.

### Relationship To Market State And Later Risk State

- This slice consumes symbol and global 5m market state from `world-usa-composites-and-market-state` as a hard external input.
- Market state answers whether the market is structurally fit to act in.
- Later tactical risk state answers whether the system should permit even a structurally valid setup given recent damage or drawdown.
- Plan the alert decision object so both gates are recorded separately: `marketStateDecision` now, `riskStateDecision` later.

### Fragmentation And Degraded Feeds

- Elevated WORLD vs USA divergence, unstable venue leadership, or missing key venues should reduce trust before alert emission.
- Fragmentation should influence both validator decisions and hygiene decisions; repeated trigger flips in fragmented conditions should cluster or suppress rather than spam.
- Critical feed degradation should prefer suppression plus machine-readable reasons over speculative action alerts.
- If a degraded condition still deserves visibility, emit one bounded informational event rather than many repeated setup alerts.

### Config And Versioning

- Version config for setup enablement, threshold bundles, severity mappings, dedupe keys, cooldown durations, cluster windows, and degrade/suppress rules.
- Attach `configVersion`, `algorithmVersion`, and contract version references to every emitted alert and every persisted suppression decision that needs replay audit.
- Replays must pin the exact config snapshot used at nomination, validation, and emission time.

## ASCII Flow

```text
canonical events -> feature-engine buckets -> regime-engine 5m state
        |                    |                        |
        |                    |                        v
        |                    +---- 30s setup candidates ----+
        |                                                 |
        |                                      2m validator checks
        |                                      - persistence
        |                                      - fragmentation
        |                                      - degraded feed rules
        |                                                 |
        |                                                 v
        +--------------------------------------> alert-engine decision
                                                          |
                                              market-state gate now
                                              risk-state gate later
                                                          |
                                                          v
                                         dedupe + cooldown + clustering
                                                          |
                                  +-----------------------+------------------+
                                  |                                          |
                                  v                                          v
                        emitted bounded alert                    suppressed or downgraded event
                        - setup A/B/C                            - reason codes
                        - 30s/2m/5m evidence                     - replay/audit fields
                        - severity + versions
```
