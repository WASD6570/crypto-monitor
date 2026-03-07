# Alert Delivery And Routing Overview

## Ordered Implementation Plan

1. Define the service-owned delivery routing model, alert delivery contract, and persistence rules that preserve severity, permission ceilings, and deterministic audit fields.
2. Implement channel adapters for in-app delivery first, Telegram second, and generic webhook third with transport-level dedupe, retry policy, stale-alert handling, and degraded-mode behavior.
3. Add operator-facing delivery controls and audit surfaces so routing changes, failures, and manual actions stay server-authorized and reviewable.
4. Validate delivery contracts, replay-safe persistence, negative-path authz handling, channel failure recovery, and deterministic record generation before later Initiative 2 review work depends on this slice.

## Role In Initiative 2

- This is slice 3 of `crypto-market-copilot-alerting-and-evaluation`.
- It turns a valid alert into the first reliable user interruption loop: generate -> cap -> deliver -> review later.
- It depends on `alert-generation-and-hygiene` for bounded alert payloads and `tactical-risk-state-and-permissioning` for the effective permission ceiling.
- It creates the delivery evidence that later outcome, feedback, and analytics slices need.

## Problem Statement

An alert engine without reliable delivery still forces the user to watch dashboards manually. This slice must route service-owned alert decisions to the web UI, Telegram, and optional webhook without losing severity context, permission ceilings, or auditability. Delivery must stay deterministic enough for replay and review, but practical enough to retry transient failures and suppress noisy duplicate transport sends.

## First-User Interruption Loop

1. `services/alert-engine` emits a bounded alert with setup, severity, state ceilings, and reason codes.
2. Delivery routing computes the effective delivery plan from alert severity, market-state ceiling, tactical-risk ceiling, operator routing config, and channel health.
3. The alert is written to the in-app stream as the source-of-truth review surface.
4. If push delivery is allowed, Telegram receives the first push notification and webhook receives a secondary integration event.
5. Every attempted send writes a deterministic delivery record so the user can later answer what should have been sent, what was sent, and what failed.

## In Scope

- delivery planning for `apps/web` in-app stream, Telegram, and optional generic webhook
- service-owned routing decisions using alert severity plus effective permission ceiling
- channel-specific payload shaping without changing the canonical alert meaning
- transport-level dedupe and idempotent retry behavior
- delivery persistence and audit trail for attempts, outcomes, suppression, stale drops, and degraded fallbacks
- safe defaults for routing, fallback behavior, quiet-hours posture, stale alerts, and degraded-channel states
- server-side authorization for operator-controlled routing config and webhook secrets
- deterministic delivery record generation for replay and review

## Out Of Scope

- email or Slack delivery by default
- client-computed routing, permissioning, or severity changes
- outcome scoring, simulation, or operator feedback semantics beyond storing delivery references
- escalation trees, paging rotations, on-call scheduling, or multi-recipient workflow automation
- rich template management beyond stable, versioned channel formatting needs
- implementation of concrete schema files or storage migrations without supporting code work

## Safe Defaults

- In-app delivery is always on for persisted alerts unless the alert is fully suppressed before delivery routing.
- Telegram is the first push surface when a valid operator destination is configured; otherwise the system records `not_configured` and relies on in-app only.
- Webhook is opt-in and secondary; delivery failure there must not block in-app or Telegram.
- `INFO` alerts never escalate because of transport failure, missing config, or degraded state.
- `WATCH` and `ACTIONABLE` alerts preserve their service-owned severity in payloads, but channel eligibility still honors the effective permission ceiling.
- `NO-OPERATE` or tactical `STOP` cap all channel payloads to informational presentation and non-urgent routing.
- Quiet hours are off by default; if later enabled, the default policy should suppress only push delivery for `INFO` and `WATCH`, never mutate the canonical alert record.
- Alerts older than a conservative freshness window at send time should be marked stale and skipped for push transports while remaining visible in-app with the stale reason recorded.
- If channel health is degraded, the router should prefer bounded suppression or reduced fanout over repeated retries that create transport spam.

## Requirements

- Keep routing, transport eligibility, retry policy, and persistence in Go service-owned logic.
- Treat `apps/web` as the source-of-truth review surface and read model, not a routing authority.
- Preserve the canonical alert meaning across channels: same alert ID, same setup intent, same severity ceiling inputs, same reason codes.
- Separate canonical alert time, routing decision time, and per-attempt transport processing time.
- Make every delivery outcome explainable through stable status codes such as `sent`, `suppressed`, `stale`, `not_configured`, `failed_transient`, and `failed_permanent`.
- Ensure retries and duplicate webhook/Telegram submissions are idempotent through stable transport dedupe keys.
- Preserve deterministic delivery records for the same alert input, config snapshot, routing snapshot, and replay mode.
- Enforce all operator controls, destination changes, and credential-backed webhook actions with server-side authorization.
- Keep Python out of the live routing path.

## Target Repo Areas

- `services/alert-engine`
- `services/*` delivery or alert-control boundary selected during implementation
- `libs/go`
- `schemas/json/alerts`
- `apps/web`
- `configs/*`
- `tests/integration`
- `tests/replay`
- `tests/fixtures`

## Design Overview

### Delivery Model

- The alert record remains the canonical object produced upstream.
- Delivery routing derives one or more channel delivery intents from that canonical object plus the effective permission ceiling and operator config snapshot.
- Each delivery intent produces append-only delivery attempt records instead of mutating prior history.
- Channel formatting may shorten or reshape fields, but it cannot invent new severity or authorization meaning.

### Channel Priority

- In-app is mandatory and first because it anchors later review and audit.
- Telegram is the first push surface because the program defaults call for one lightweight interruption path before broader integrations.
- Webhook is secondary and generic, intended for external automation or mirrors without becoming a primary review source.

### Effective Ceiling Handling

- The delivery router consumes the already-computed alert severity and the effective ceiling inputs from market state and tactical risk state.
- The strictest ceiling wins for channel presentation and urgency.
- Delivery routing may suppress a channel because of config, stale age, or degraded transport, but it may never raise severity above the service-owned alert decision.

### Deterministic Delivery Records

- Persist one canonical routing decision record per alert plus one append-only transport attempt record per channel attempt.
- Use stable ids such as `alertId + channel + destination fingerprint + routing version` for dedupe and idempotency.
- Record both the chosen routing snapshot and the reason a channel was skipped, retried, or degraded.
- Replay mode should be able to regenerate the same routing decisions without performing real external sends.

### Failure And Degraded-State Posture

- Transient failures should retry with a bounded schedule and an idempotency key.
- Permanent failures should stop retrying quickly and preserve operator-visible reasons.
- If Telegram or webhook is unavailable, in-app still succeeds and becomes the fallback review surface.
- Repeated channel outages should not multiply delivery attempts for the same alert beyond configured caps.

## ASCII Flow

```text
alert-engine emits canonical alert
  - alertId
  - setup/severity
  - marketStateDecision
  - riskStateDecision
              |
              v
delivery router
  - load routing config snapshot
  - apply effective permission ceiling
  - check destination readiness
  - check stale window / quiet-hours / channel health
  - create per-channel intents
              |
    +---------+-------------------+
    |         |                   |
    v         v                   v
 in-app    Telegram            webhook
 always    first push          secondary opt-in
 record     send/skip           send/skip
    |         |                   |
    +---------+-------------------+
              v
append-only delivery records
  - routing decision
  - attempt outcome
  - dedupe key
  - retry count
  - failure/suppression reason
              |
              v
apps/web review + later outcome/feedback joins
```
