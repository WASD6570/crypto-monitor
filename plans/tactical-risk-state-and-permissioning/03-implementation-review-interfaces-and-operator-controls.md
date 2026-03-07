# Implementation: Review Interfaces And Operator Controls

## Module Requirements And Scope

- Expose operator-facing review surfaces for tactical state history, current ceilings, and review-required items.
- Define authorized control hooks for acknowledgement, note capture, and review clearance.
- Keep the UI read-heavy and audit-focused; do not design live execution or trading workflows.

## Target Repo Areas

- `apps/web`
- `services/*` operator-control or API boundary
- `schemas/json/alerts`
- `tests/integration`

## Review Surface Goals

- Show current global and symbol tactical state next to the market-state ceiling that produced the effective restriction.
- Show why the current ceiling exists using stable reason codes, thresholds, timestamps, and config version.
- Let an authorized operator inspect pending review-required transitions without editing raw config.
- Make prior manual actions and notes visible so later review does not rely on memory.

## Recommended Operator Surfaces

### Tactical State Panel

- current tactical state for global and symbol scopes
- effective permission ceiling and whether it comes from market state, tactical state, or both
- active review-required indicator
- last transition time, reason codes, and config version

### Transition History View

- ordered transition list with filters for symbol, state, reason code, and review-required status
- expandable record details showing threshold values, market-state snapshot references, actor, and linked notes
- replay tag or correction indicator when the entry comes from replay validation rather than live operation

### Review Queue Surface

- pending items for `STOP` entries, manual overrides, and reset-clearance requests
- explicit operator actions: acknowledge, add note, approve clearance, reject clearance
- visible authorization errors and immutable audit trail for every action

## Operator Control Hooks

- `GET` current tactical permission status by scope
- `GET` transition history and decision-log summaries
- `GET` review-required items
- `POST` acknowledge review item with note
- `POST` approve or reject clearance for review-required transitions

These hooks should be server-authorized control-plane actions only. They should not imply trading, order placement, or direct market mutation.

## Human-Review Boundaries

- Entering `DE-RISK` from a soft limit may be auto-applied without approval, but should still be visible and annotatable.
- Entering `STOP` requires operator visibility immediately and any later clearance must be manually approved.
- Manual attempts to restore `NORMAL` during an active weekly stop should be rejected server-side.
- If market state remains `NO-OPERATE`, review clearance may update tactical state history but cannot elevate the effective alert ceiling above informational.

## Security And Authz Expectations

- Server-side role checks gate all acknowledgement and clearance actions.
- UI must not assume permission based on hidden buttons alone; backend rejection paths must be first-class and tested.
- Every operator action logs actor id, action type, timestamp, note, and reviewed transition references.

## UX Constraints

- Keep the web UI mobile-readable and dense enough for market operations review.
- Prefer status-first summaries with drill-down details over complex workflow builders.
- Use existing product visual language; this feature adds control surfaces, not a new standalone admin product.

## Integration Test Expectations

- authorized operator can view and act on review-required items
- unauthorized actor receives clear rejection and no state mutation
- approving clearance updates review status but still respects active `NO-OPERATE` ceilings
- repeated acknowledgement requests are idempotent or visibly deduplicated
- transition history renders deterministic reason-code and threshold data from persisted records

## Summary

This module adds the minimum operator review layer needed to make tactical restrictions trustworthy. The UI and APIs should explain current ceilings, expose pending review decisions, and enforce human approval boundaries without drifting into execution tooling.
