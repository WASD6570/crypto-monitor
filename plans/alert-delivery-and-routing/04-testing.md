# Alert Delivery And Routing Testing

## Goal

Validate that Initiative 2 delivery routing reliably creates the first-user interruption loop without breaking server-side authorization, deterministic delivery records, or bounded transport behavior.

## Output Artifact

- Write execution results to `plans/alert-delivery-and-routing/testing-report.md` during implementation/testing work.

## Required Environments And Inputs

- local or CI environment with alert-generation fixtures and tactical-risk fixtures available
- test Telegram destination or stub transport
- test webhook endpoint or local mock receiver
- seeded operator auth identities for read-only and mutation-capable roles
- config snapshots covering in-app-only, Telegram-enabled, webhook-enabled, and capped `NO-OPERATE` / `STOP` cases

## Validation Commands

- `go test ./services/... -run TestDeliveryRoutingDecision`
- `go test ./services/... -run TestDeliveryAttemptPersistence`
- `go test ./services/... -run TestTelegramDelivery`
- `go test ./services/... -run TestWebhookDelivery`
- `go test ./services/... -run TestDeliveryControlAuthz`
- `go test ./tests/replay/... -run TestAlertDeliveryReplayDeterminism`
- `pnpm --dir apps/web test -- --run delivery`

## Smoke Matrix

### 1. In-App Source Of Truth

- Input: actionable alert, Telegram configured, webhook disabled.
- Verify: in-app record exists, routing decision shows Telegram eligible, delivery timeline links all records by `alertId`.
- Verify: canonical severity and effective ceiling match upstream alert decision.

### 2. Telegram First Push

- Input: actionable alert, Telegram configured and healthy.
- Verify: one send attempt succeeds, dedupe key is stable, no duplicate send on worker retry.
- Verify: in-app still records the same alert and attempt status.

### 3. Webhook Secondary Delivery

- Input: watch-level alert, Telegram configured, webhook configured.
- Verify: Telegram and webhook both receive channel-appropriate payloads derived from the same canonical alert.
- Verify: webhook failure does not block Telegram or in-app.

### 4. Effective Ceiling Preservation

- Input: canonical actionable alert with global `NO-OPERATE` or tactical `STOP` ceiling.
- Verify: all channels present informational-only routing semantics, no urgent push behavior, and records show the cap reason.

### 5. Stale Push Suppression

- Input: alert older than the configured freshness window before transport execution.
- Verify: in-app persists, Telegram/webhook record `stale`, and no external send occurs.

### 6. Degraded Channel Fallback

- Input: healthy in-app path, Telegram transport timeout, webhook disabled.
- Verify: bounded retries occur, final status is transient failure or deferred according to policy, and in-app remains available as fallback.

## Negative Cases

- Missing Telegram destination config returns `not_configured` without retry.
- Invalid webhook secret or malformed endpoint fails permanently and is audited.
- Duplicate worker processing of the same delivery intent produces one external send and deterministic duplicate suppression records.
- Unauthorized operator cannot enable channels, rotate webhook config, or trigger manual retry.
- Client attempts to submit a higher severity or bypass a permission ceiling are rejected server-side.
- Quiet-hours config absent means no quiet-hours suppression is applied.
- Replayed alerts regenerate identical routing decisions and dedupe keys without performing real external sends.
- Permanently failed delivery attempts are not retried indefinitely.

## Replay And Determinism Checks

- Run the replay suite with pinned alert and routing fixtures.
- Verify identical routing decisions, dedupe keys, suppression reasons, and attempt-record shapes for the same alert/config inputs.
- Verify replay mode disables live transport side effects while still producing comparable audit artifacts.

## Manual Review Checklist

- Confirm per-alert delivery timeline is understandable in `apps/web`.
- Confirm operator audit log links config changes and manual retries to affected alerts.
- Confirm no channel can present a more urgent meaning than the canonical alert plus effective ceiling.
- Confirm secret material is never exposed in UI or logs.

## Exit Criteria

- In-app, Telegram, and optional webhook routing behave according to config and ceiling rules.
- Transport-level dedupe and retry behavior are bounded and observable.
- Authorization checks protect all delivery control actions.
- Delivery records remain deterministic, append-only, and suitable for later outcome/feedback joins.
