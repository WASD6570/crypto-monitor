# Polling, Publication State, And Source Health

## Module Requirements And Scope

Target repo areas:

- `services/slow-context`
- optional config under `configs/local`, `configs/dev`, and `configs/prod`

This module defines how slow-context adapters are polled, how publication windows are interpreted, and how slow-source health stays visible without contaminating realtime feed-health semantics.

In scope:

- scheduled polling cadence defaults for expected publish windows and off-hours backoff
- source-family health categories and publication-delay detection
- idempotent handling of repeated polling when no new point is published
- explicit correction handling for same-as-of republishes
- operator-visible metrics or status outputs for last successful poll and last new publication

Out of scope:

- persistence schema or query response packaging
- dashboard copy and panel rendering
- hidden reuse of realtime venue-health categories for slow-context state

## Planning Guidance

### Polling Rules

- During expected publish windows, poll no faster than every 15 minutes by default.
- Outside expected publish windows, back off to a coarse cadence such as hourly.
- Keep cadence configuration simple and environment-owned; do not overfit to one provider's exact publishing schedule.

### Publication-State Handling

- `published_new_value` accepts a new candidate and emits a new-publication event or internal handoff.
- `published_same_value_or_same_asof` must update operational freshness only if policy explicitly allows it; it must not create duplicate accepted records.
- `not_yet_published` must be observable and may transition the source health to `delayed_publication` once the expected window passes.

### Health Boundary

- Keep slow-source health categories separate from realtime venue health:
  - `healthy`
  - `delayed_publication`
  - `source_unavailable`
  - `parse_failed`
- Record at least:
  - last successful poll
  - last successfully parsed publication
  - last new publication
  - consecutive delayed polls
  - consecutive failure count
- None of these states may mark the core market-state path unhealthy by themselves.

### Idempotency And Corrections

- Dedupe on source family, metric family, asset/instrument identity, and publication/as-of identity.
- Same-as-of corrected values should emit an explicit correction marker or revision signal for the next child feature rather than silently overwriting the previous accepted point.
- Repeated unchanged polls should be safe to retry and replay-auditable.

## Unit Test Expectations

- repeated polling of unchanged publication stays idempotent
- delayed publication classification starts only after the expected window is crossed
- source fetch failures increment slow-source failure state without affecting realtime feed-health outputs
- corrected same-as-of publication produces explicit correction metadata
- poll scheduling honors publish-window cadence versus off-hours backoff under a pinned clock

## Summary

This module locks the operational behavior of slow-source polling: when to poll, how to classify delayed publication, and how to expose source health honestly. The next module can then add deterministic fixtures and targeted validation against these exact behaviors.
