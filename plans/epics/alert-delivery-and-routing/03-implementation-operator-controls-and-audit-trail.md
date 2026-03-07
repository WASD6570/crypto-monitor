# Operator Controls And Audit Trail

## Requirements And Scope

- Add server-authorized operator controls for delivery configuration, webhook destination management, and delivery-status review.
- Preserve immutable audit history for routing decisions, transport attempts, config changes, and manual actions.
- Expose review surfaces in `apps/web` without moving trust or authorization to the client.
- Keep the first version minimal: visibility and safe control, not workflow automation.

## Target Repo Areas

- `services/*` delivery control surface chosen during implementation
- `apps/web`
- `libs/go`
- `schemas/json/alerts`
- `tests/integration`

## Control Surface Design

### Operator Actions In Scope

- enable or disable Telegram delivery
- register, update, or disable webhook destinations
- rotate or replace webhook secret-backed configuration
- inspect delivery records and failure reasons
- optionally trigger a bounded manual re-delivery action for eligible failed attempts only if the server verifies authorization and idempotency

### Explicitly Out Of Scope

- editing canonical alert severity or permission ceilings from the client
- client-side secret storage or unsigned webhook configuration
- bulk resend campaigns, recipient groups, or escalation chains
- any action that rewrites immutable delivery history

## Audit Trail Requirements

- Store append-only records for routing decisions, transport attempts, config mutations, and manual retry requests.
- Capture actor identity, action type, before/after config fingerprint, event time, processing time, and server decision result.
- Link operator actions to affected delivery records by stable ids.
- Ensure manual retry records make it clear whether the action created a new attempt, was deduped, or was denied.

## Authorization Model

- All control actions must be authenticated and role-checked server-side.
- `apps/web` should call explicit control APIs and render denied actions as denied results, not hide server decisions behind client-only gates.
- Separate read access from mutation access where practical so review-only users can inspect delivery history without changing configuration.
- Webhook secrets or tokens must never be returned in cleartext after creation.

## Read Model Expectations

- Per-alert delivery timeline showing in-app, Telegram, and webhook statuses.
- Clear display of effective ceiling, stale suppression, and failure/suppression reason codes.
- Config summary showing which destinations are active without exposing secrets.
- Audit filters for channel, status, symbol, severity, and recent operator actions.

## Ordered Implementation Plan

1. Add append-only audit record model and read APIs.
2. Add server-authorized destination and routing-config mutation APIs.
3. Add `apps/web` review views for per-alert delivery timeline and config status.
4. Add bounded manual retry control only after the immutable audit model is in place.

## Test Expectations

- Authz tests proving unauthorized users cannot mutate routing config or retry sends.
- Audit tests proving every config change and manual retry request creates immutable records.
- UI integration tests proving delivery status reflects the server read model, including skipped and failed channels.
- Negative tests for secret re-read attempts, retry of permanently failed/blocked alerts, and client attempts to override severity.

## Summary

This module keeps delivery operations inspectable and controllable without weakening trust boundaries. The important continuation detail is that operator controls must layer on top of immutable routing and attempt records, never replace them with mutable dashboard state.
