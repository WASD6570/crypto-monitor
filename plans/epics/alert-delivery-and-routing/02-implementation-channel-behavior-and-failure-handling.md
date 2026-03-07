# Channel Behavior And Failure Handling

## Requirements And Scope

- Implement channel adapters and execution rules for in-app, Telegram, and generic webhook delivery.
- Preserve severity and permission ceilings across channel formatting and urgency handling.
- Define bounded retries, stale-alert handling, degraded-channel behavior, and fallback rules.
- Keep transport execution idempotent and auditable.
- Avoid adding product complexity such as escalation chains, rich schedules, or channel-specific business logic beyond what is needed for reliable first-user interruption.

## Target Repo Areas

- `services/*` delivery worker or dispatcher boundary selected during implementation
- `libs/go`
- `apps/web`
- `configs/*`
- `tests/integration`
- `tests/fixtures`

## Channel Policies

### In-App

- Always persist and surface the alert in-app when the canonical alert survives upstream suppression.
- In-app does not depend on Telegram or webhook success.
- In-app presentation should show canonical severity, effective ceiling, and delivery status for push channels.
- In-app is the default fallback surface during any push-channel degradation.

### Telegram

- Telegram is the first push surface and should receive concise, operator-readable messages.
- Preserve canonical severity labels, but map urgency conservatively when the effective ceiling downgrades the alert.
- Use bot/API idempotency support where possible plus internal dedupe keys.
- If Telegram config is absent, invalid, or unauthorized, record `not_configured` or `failed_permanent` and do not block in-app.

### Webhook

- Webhook is opt-in and secondary.
- Send a structured JSON payload that mirrors the canonical delivery contract rather than a Telegram-shaped template.
- Require signed or secret-authenticated outbound configuration owned by the server.
- Treat repeated non-2xx responses as delivery failures without changing the canonical alert meaning.

## Failure Handling

### Retry Policy

- Retry transient network or rate-limit failures with bounded exponential backoff and a small max-attempt count.
- Do not retry permanent failures such as invalid destination config, revoked credentials, or structurally invalid payload construction.
- Each retry appends a new attempt record while preserving the same transport dedupe key.

### Stale Alerts

- Define a conservative max age for push delivery measured from canonical alert event time or routing decision time, chosen explicitly during implementation.
- If the alert is stale before a push attempt starts, skip push sends and write `stale` status.
- Keep stale alerts visible in-app so later review still reflects the canonical alert history.

### Quiet Hours

- Default posture is no quiet hours.
- If the repository later introduces quiet hours, apply them only to push channels and only as suppression metadata; do not mutate severity or the in-app record.
- Actionable alerts should not silently disappear under a future quiet-hours model without explicit operator configuration.

### Degraded States

- Channel health degradation should reduce send fanout and retry pressure before it increases operational noise.
- During sustained Telegram or webhook degradation, keep writing audit records and surface channel health in-app.
- If the delivery worker is degraded, prefer `queued` or `deferred` records over pretending a send happened.

## Ordered Implementation Plan

1. Deliver in-app persistence and read-model updates first.
2. Add Telegram adapter with deterministic formatting, dedupe, and retry bounds.
3. Add webhook adapter with signed config, idempotent sends, and bounded retry handling.
4. Add channel-health gates, stale suppression, and degraded fallback semantics.

## Unit And Integration Test Expectations

- In-app smoke tests proving every eligible alert appears even when push channels fail.
- Telegram adapter tests for formatting, dedupe, stale suppression, and transient vs permanent failure classification.
- Webhook tests for signed payload construction, retry handling on 5xx/timeouts, and no retry on invalid config.
- End-to-end delivery tests proving channel failures do not duplicate sends or mutate canonical severity.
- Negative tests for repeated worker restart, expired alert age, and push disabled by effective ceiling.

## Summary

This module turns routing decisions into real channel behavior while keeping in-app as the stable fallback surface. The main constraint for the next agent is to keep retries bounded, transport sends idempotent, and all degraded behavior visible through delivery records instead of hidden channel-specific state.
