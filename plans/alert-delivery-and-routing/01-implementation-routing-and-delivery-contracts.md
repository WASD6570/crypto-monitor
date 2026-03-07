# Routing And Delivery Contracts

## Requirements And Scope

- Define the canonical routing decision and delivery-attempt model for `services/*`, `libs/go`, and `schemas/json/alerts`.
- Preserve upstream alert meaning while adding delivery-specific fields for channel targeting, dedupe, and audit.
- Cover in-app, Telegram, and webhook only; email and Slack stay explicitly deferred.
- Keep routing deterministic for replay by separating pure routing decisions from side-effecting transport execution.
- Ensure every contract records the effective permission ceiling and server-owned authorization context.

## Target Repo Areas

- `services/alert-engine`
- `services/*` delivery service boundary chosen during implementation
- `libs/go`
- `schemas/json/alerts`
- `tests/fixtures`
- `tests/replay`

## Module Design

### Canonical Objects

- `AlertRecord`: existing or upstream canonical alert with setup, severity, market/risk state decisions, timestamps, config versions, and reason codes.
- `DeliveryRoutingDecision`: one record per alert describing which channels were eligible, suppressed, or skipped and why.
- `DeliveryAttemptRecord`: one record per channel attempt with destination fingerprint, dedupe key, transport status, retry metadata, and transport timestamps.
- `DeliveryStatusView`: read-optimized aggregate for UI review, derived from append-only routing and attempt records.

### Contract Fields To Preserve

- `alertId`, `symbol`, `setupId`, `severity`, `effectiveSeverity`
- `marketStateDecision`, `riskStateDecision`, and final `effectivePermissionCeiling`
- `eventTs`, `routingDecisionTs`, `transportAttemptTs`, and `transportAckTs` where applicable
- `configVersion`, `algorithmVersion`, and `routingVersion`
- stable `reasonCodes` for suppression, downgrade, stale handling, config gaps, and authz denial

### Routing Decision Rules

- Build channel intents from a pure function that accepts canonical alert plus routing config snapshot.
- Produce the same result for the same inputs in replay or dry-run mode.
- Mark channels explicitly as `eligible`, `suppressed`, `not_configured`, `stale`, or `blocked_by_ceiling` before side effects run.
- Persist skipped channels too; absence of a send is still a decision that needs review.

### Transport-Level Dedupe

- Generate a stable dedupe key from canonical alert identity plus channel and destination fingerprint.
- Use that key for idempotent reprocessing, retry resume, and external idempotency headers where supported.
- Keep dedupe scoped to transport delivery, not alert generation; a new alert may still route if upstream emitted a new canonical alert.
- Prevent duplicate webhook sends caused by worker restarts or retry races.

### Authorization Boundary

- Delivery execution uses server-held credentials and secrets only.
- Routing config changes, destination registration, and webhook secret rotation require authenticated server-side operator actions.
- The delivery contracts should store actor metadata for config changes without letting clients self-assert routing permissions.

## Implementation Order

1. Define routing and attempt contract shapes in Go types and JSON schema planning notes.
2. Add pure routing decision builder with deterministic ceiling and config evaluation.
3. Add append-only persistence semantics for routing decisions and delivery attempts.
4. Expose read models needed by `apps/web` without giving the client mutation authority over delivery semantics.

## Unit Test Expectations

- Pure routing decision tests for in-app-only, Telegram-enabled, and webhook-enabled configurations.
- Determinism tests proving identical inputs produce identical routing decisions and dedupe keys.
- Ceiling tests proving `NO-OPERATE` and `STOP` keep only informational presentation and non-urgent routing.
- Contract tests proving skipped/suppressed channels still emit deterministic audit records.
- Negative tests for missing destination config, stale alert age, and duplicate processing of the same transport key.

## Summary

This module defines the canonical delivery decision layer that later channel workers and UI review tools depend on. The critical handoff is that delivery stays a service-owned, deterministic extension of the canonical alert record rather than a set of ad hoc channel-specific behaviors.
