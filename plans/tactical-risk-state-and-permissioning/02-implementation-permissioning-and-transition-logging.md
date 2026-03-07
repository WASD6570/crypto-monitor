# Implementation: Permissioning And Transition Logging

## Module Requirements And Scope

- Persist tactical risk transitions and effective permission decisions in an auditable form.
- Define structured decision-log fields for automated transitions and manual review actions.
- Ensure permission ceilings are queryable by alert-generation and review surfaces.
- Preserve idempotency and replayability for repeated or late-arriving inputs.

## Target Repo Areas

- `services/*` persistence or control-plane boundary
- `libs/go`
- `schemas/json/alerts`
- `schemas/json/outcomes` only if referenced inputs need stable linkage
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Transition Persistence Expectations

- Use append-only transition records; do not overwrite prior state history.
- Store both the derived tactical state and the effective permission ceiling seen by downstream consumers.
- Persist enough context to reconstruct why a transition occurred without re-running ad hoc business logic from logs alone.

## Safe Default Record Shape

Each transition record should carry at least:

- stable transition id
- symbol scope or global scope
- previous tactical state
- new tactical state
- effective permission ceiling before and after
- transition reason code set such as `soft_daily_breach`, `hard_daily_breach`, `weekly_stop`, `manual_review_clear`, `config_invalid`, `input_stale`
- triggering loss values and threshold values used in the decision
- market-state snapshot references and key reason codes
- `eventTime`, `processedTime`, and evaluation window ids
- `configVersion`, `algorithmVersion`, and schema version
- actor metadata for manual actions, or `system` for automatic transitions
- replay correlation id or fixture id when generated from replay

## Structured Decision Log Expectations

- Emit one structured decision log entry for every evaluation, even when no state transition occurs, so "why nothing changed" remains auditable.
- Emit a second log entry when a transition record is appended, linked by evaluation id.
- Keep reason codes enumerable and stable so UI filters and tests do not depend on free-form text.
- Free-form operator notes belong in dedicated note fields, not primary machine-readable reason codes.

## Permissioning Model

### Consumer Contract

Expose a read-only permission object that downstream alert components can consume directly, including:

- tactical state label
- effective ceiling label
- review-required boolean
- allowed setup classes or severity bands
- blocking reason codes
- current review ticket or review item reference when applicable

### Safe Defaults

- If any required state input or config is invalid, emit the strictest effective ceiling and mark the permission object degraded.
- If the same evaluated state repeats, avoid duplicate transition records but still emit an idempotent decision log entry.
- If late replay data changes historical state, record replay-specific correction artifacts rather than mutating prior live logs invisibly.

## Idempotency And Ordering

- Deduplicate transition writes by stable evaluation key plus resulting state.
- Preserve ordered evaluation by event time with explicit handling for late data.
- Never allow a duplicated replay or retry to create multiple manual-review artifacts for the same transition.

## Review-Required Behavior

- Entering `STOP` from hard daily or weekly breaches must mark the state as review-required.
- Manual overrides that raise permission above the automatic ceiling must always be review-required and separately logged.
- Clearing review-required status must reference the reviewed transition ids and the authorized actor.

## Unit Test Expectations

- append-only transition persistence with duplicate suppression
- no-transition evaluations still producing structured decision logs
- replay of identical inputs yielding identical transition ids or deterministic dedupe keys
- invalid config or stale inputs forcing degraded strict ceilings
- manual review actions creating linked audit records
- late-data correction path preserving prior live history and replay traceability

## Summary

This module makes tactical risk decisions inspectable instead of implicit. Transition records and structured decision logs should be immutable, enumerable, and idempotent so alert generation, operator review, and replay all consume the same explainable permission state.
