# Tactical Risk State And Permissioning Overview

## Ordered Implementation Plan

1. Define the service-owned tactical risk state machine, threshold inputs, and deterministic transition rules on top of market-state outputs.
2. Add transition persistence and structured decision logging so every state change and permission ceiling is replayable and auditable.
3. Expose read-only review interfaces and operator controls for acknowledging, annotating, and approving review-required transitions without introducing live execution flows.
4. Validate deterministic replay, negative-path threshold handling, authorization boundaries, and `NO-OPERATE` ceiling behavior.

## Role In Initiative 2

- This is slice 2 of `crypto-market-copilot-alerting-and-evaluation` and the control layer between market-state outputs and downstream alert delivery.
- It turns trusted market-state outputs into tactical alert-permission ceilings so bad trading conditions reduce noise instead of increasing it.
- It does not create live trading behavior; it only governs alert posture, operator review requirements, and auditability.

## Relation To Market-State Outputs

- Inputs come from service-owned market-state outputs produced by `world-usa-composites-and-market-state`, especially symbol and global `TRADEABLE` / `WATCH` / `NO-OPERATE` classifications plus degradation reasons.
- Tactical risk state never replaces market state. It applies an additional ceiling to alert eligibility and severity.
- If market state is stricter than tactical risk state, market state wins. If tactical risk state is stricter, tactical risk state wins.

## Problem Statement

Alerting needs a deterministic tactical overlay that reduces or halts attention when losses, degraded conditions, or review-required situations indicate that the operator should slow down. Without this layer, a difficult session can produce more alerts exactly when the operator needs fewer.

## In Scope

- service-owned tactical states `NORMAL`, `DE-RISK`, and `STOP`
- soft daily limits, hard daily limits, and weekly stop rules
- interaction between tactical state and symbol/global `NO-OPERATE`
- alert permission ceilings and review-required posture
- transition persistence and structured decision logging
- operator review surfaces and hooks for state inspection and acknowledgements
- deterministic replay and test planning for state transitions and authz boundaries

## Out Of Scope

- live order placement, flattening, or broker/exchange actions
- client-computed risk decisions or client-only authorization
- automatic threshold mutation from operator feedback
- speculative PnL source design beyond consuming an existing trusted realized-PnL or outcome input
- workflow automation that bypasses human review for protected transitions

## Safe Defaults For Vague Areas

- Default threshold posture: missing, stale, or invalid threshold config fails closed to the stricter effective state and emits a structured configuration error.
- Default daily handling: soft daily breach moves to `DE-RISK`; hard daily breach moves to `STOP` and remains until an explicit reset window and review condition are satisfied.
- Default weekly handling: weekly stop forces `STOP` until the next configured weekly boundary and operator review clears it.
- Default persistence: append immutable transition records with event-time, processing-time, config version, market-state snapshot, trigger reason, and actor metadata; derive current state from the latest valid record plus replay inputs.
- Default review boundary: any transition into or out of `STOP`, any manual override, and any reset after hard or weekly breach is review-required.
- Default market-state interaction: `NO-OPERATE` does not mutate tactical state by itself, but it forces an alert ceiling no weaker than informational delivery.

## Requirements

- Keep tactical risk evaluation in Go service-owned logic; `apps/web` renders and submits authorized review actions only.
- Consume market-state outputs as inputs rather than recomputing tradeability or degradation logic in the risk layer.
- Preserve deterministic transitions for the same ordered inputs, config version, and replay window.
- Keep transition history and decision logs replayable, queryable, and human-auditable.
- Treat authorization as server-side policy; review actions must be authenticated and role-checked before they mutate state.
- Keep Python optional and offline-only for analysis or parity fixtures, never for live tactical-state evaluation.

## Target Repo Areas

- `services/*` risk or alert-control service boundary to be chosen during implementation
- `libs/go`
- `schemas/json/alerts`
- `schemas/json/outcomes` if shared references are needed for realized-loss inputs
- `apps/web`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Design Overview

### Core Model

- Market state answers whether the market should be trusted.
- Tactical risk state answers how aggressive alerting may be given recent losses and review posture.
- The effective permission ceiling is the strictest result across global market state, symbol market state, tactical risk state, and review-required flags.

### State Definitions

- `NORMAL`: alerts may use the full market-state ceiling.
- `DE-RISK`: alerts remain allowed only within reduced severity, frequency, or setup ceilings defined by config.
- `STOP`: alerts become informational only, with review-required controls exposed to operators.

### Threshold Families

- Soft daily limits reduce permission from `NORMAL` to `DE-RISK`.
- Hard daily limits escalate to `STOP`.
- Weekly stop enforces a longer-lived `STOP` regardless of daily recovery attempts.
- Threshold inputs should be explicit and versioned, using trusted realized outcomes or other server-owned loss accounting, not client-entered values.

### Effective Ceiling Rules

- Global or symbol `NO-OPERATE` always forces informational-only alert handling.
- `STOP` also forces informational-only handling, even if market state is `TRADEABLE`.
- `DE-RISK` may permit only the most conservative setups or lower severities, but exact ceilings stay config-owned and replay-pinned.
- Review-required flags may prevent re-escalation even after the numeric threshold condition clears.

## ASCII Flow

```text
market-state outputs          trusted realized loss / outcome inputs
(global + symbol ceilings)              + config snapshot
            |                                   |
            +-------------------+---------------+
                                v
                 tactical risk state evaluator
                 - validate config and timestamps
                 - apply soft daily / hard daily rules
                 - apply weekly stop rules
                 - mark review-required conditions
                                |
                                v
              append-only transition + decision log records
                                |
                                v
     effective permission ceiling = strictest of:
     market state, tactical state, review-required gate
                                |
                 +--------------+--------------+
                 |                             |
                 v                             v
      alert-generation consumers      operator review surfaces
      read-only ceiling input         inspect / acknowledge / approve
```

## Ordered Delivery Notes

- Implement the state machine first so later logging and UI work depend on stable semantics.
- Add persistence and decision logs before operator controls so manual actions never create opaque state changes.
- Add review surfaces last, and keep them read-only plus authorized control hooks rather than execution workflows.
