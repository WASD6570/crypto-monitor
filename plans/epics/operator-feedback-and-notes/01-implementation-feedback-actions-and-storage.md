# Feedback Actions And Storage

## Module Scope

Define the append-only persistence model and service-owned write/read responsibilities for operator feedback actions and note revisions linked to alerts.

## Target Repo Areas

- a review-owned Go service boundary under `services/`
- `libs/go`
- `schemas/json/alerts`
- `tests/integration`
- `tests/fixtures`
- `tests/replay`

## Module Requirements

- Support write intents for `save`, `dismiss`, `thumbs_up`, `thumbs_down`, `good_setup_bad_timing`, `useful_context_only`, and `note_revision`.
- Persist immutable events with stable identifiers and ordering metadata.
- Copy or reference the reviewed alert's pinned `configVersion`, `algorithmVersion`, and contract version at write time for audit joins.
- Preserve `exchangeTs`/`recvTs` discipline where relevant to upstream alert records; review events themselves use service receive and persistence timestamps and must not overwrite alert event time.
- Expose deterministic read APIs for full history and current snapshot.
- Keep storage and contracts simple enough for replay/audit reconstruction.

## Storage Design

### Event Log

Plan for a single append-only review event stream keyed by `alertId` with one row or document per event. Each event should include at minimum:

- `reviewEventId`
- `alertId`
- `eventType`
- `eventFamily`
- `actorId`
- `actorRole`
- `sourceSurface` such as `web`
- `persistedAt`
- `clientObservedAt` when supplied, stored only as metadata
- `reviewVersion` monotonic per alert
- `alertConfigVersion`
- `alertAlgorithmVersion`
- `alertSchemaVersion`
- `payload`

### Event Payload Expectations

- Lifecycle actions carry no extra payload beyond type and provenance.
- Sentiment and qualifier actions carry structured enums only; avoid free-form labels.
- Note revision payload carries the full note body plus optional metadata such as note title later if introduced. MVP default is body-only.
- Keep note payload text in the protected review store, not in general alert payload tables or delivery records.

### Derived Snapshot

Plan a materialized read model or service-level projection keyed by `alertId` that stores:

- `currentLifecycleState`
- `currentSentimentState`
- `activeQualifiers`
- `latestNoteRevisionId`
- `latestNotePreview`
- `lastReviewedAt`
- `lastReviewedBy`
- `currentReviewVersion`

This read model is disposable and rebuildable from the append-only event log.

## Ordering And Versioning Rules

- Use persistence order or service-generated monotonic `reviewVersion` as canonical ordering for derived-state rebuilds.
- Persist the reviewed alert's version references on every event so history remains interpretable after config changes.
- If alert replay or correction produces a superseding alert/outcome record later, feedback remains attached to the original alert identifier unless a future review-link migration plan explicitly rethreads it.

## Safe Defaults

- Default write idempotency: accept a client request id or equivalent write token so network retries do not duplicate the same intended event.
- Default duplicate-click posture: identical requests with the same idempotency token return the original success result; identical requests without a token create distinct events and remain auditable.
- Default note length posture: cap note size conservatively and reject oversize bodies rather than truncating silently.
- Default blank-note posture: reject blank or whitespace-only revisions.

## Negative Cases To Cover In Implementation

- write against unknown `alertId`
- write against alert records the actor is not authorized to review
- stale `currentReviewVersion` supplied by client
- unsupported action enum or malformed note payload
- note body too large or blank
- repeated transport retry without idempotency token handling

## Unit Test Expectations

- append-only inserts preserve prior events unchanged
- snapshot rebuild from event log matches incremental projection result
- latest-event-wins rules apply correctly by family
- note revision chain preserves prior bodies and ordering
- idempotent retry returns stable response without duplicate event creation
- stale version write is rejected predictably

## Summary

This module defines the durable truth for operator review: an append-only event log plus a rebuildable current snapshot. It keeps history immutable, version-aware, and safe for later audit, replay, and analytics joins.
