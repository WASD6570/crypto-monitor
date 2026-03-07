# Review Workflow And Annotations

## Module Scope

Define the review workflow behavior in `apps/web` and the service/API contract needed to capture lightweight operator annotations without turning the review loop into manual case management.

## Target Repo Areas

- `apps/web`
- a review-owned service boundary under `services/`
- `schemas/json/alerts`
- `tests/integration`
- `tests/fixtures`

## Module Requirements

- Present review actions directly on alert detail and recent-alert review surfaces.
- Keep interaction fast enough for a first user scanning many alerts.
- Read the derived review snapshot for current state and request the full event history or note revision history only when needed.
- Guard writes with server-side authz and optimistic concurrency.
- Make audit history visible enough that the operator can tell when a label or note changed and by whom.

## Workflow Design

### Primary Review Journey

1. User opens an alert with delivery, outcome, and current review snapshot.
2. User taps one of the quick actions or opens note editing.
3. Client submits an append request with `alertId`, action payload, idempotency token, and `currentReviewVersion`.
4. Service validates actor rights and review version, appends the event, and returns the updated snapshot.
5. UI updates current state immediately from the returned snapshot and offers history drill-down.

### Quick Actions

- `save` and `dismiss` should be one-tap actions on both desktop and mobile.
- `thumbs_up`, `thumbs_down`, `good setup bad timing`, and `useful context only` should be visible but secondary to the alert facts.
- Actions should show current active state from the snapshot, not by recomputing in the client.

### Note Editing Safe Default

- Default note UX is single latest note body with explicit "save revision" semantics.
- Do not auto-save partial text to the canonical review log.
- Keep unsaved draft text local to the browser session or ephemeral client state only.
- When another review event lands before save, the note submit path should detect the stale snapshot version and prompt a refresh/merge decision.

### Conflict Handling

- If the server returns stale-version conflict, UI should preserve the local draft, refetch the latest snapshot/history, and ask the operator to resubmit against the newest state.
- Safe default for note conflict is manual review instead of automatic text merge.
- Safe default for quick-action conflict is refetch then reapply because those events are lightweight and append-only.

### Annotation History Presentation

- Show a compact current-state header first.
- Provide a chronological history list with action label, actor, timestamp, and version reference.
- Provide note revision history with prior revision timestamps and author metadata.
- Redact or hide raw note text from users without review-note permission even if they can see aggregate labels.

## Accessibility And Product Constraints

- Keep tap targets at least 44px in mobile layouts.
- Do not hide critical alert facts behind the note panel.
- Prefer dense but scannable controls because the user may review many alerts quickly.
- Avoid decorative motion; review state changes should emphasize reliability over animation.

## Negative Cases To Cover In Implementation

- unauthorized user can see alert but not review controls
- user loses write race after typing a note
- transient network retry after a successful write
- user attempts to submit empty note or unsupported label
- review history unavailable while alert detail still loads

## Unit And UI Test Expectations

- snapshot-driven button states render correctly for each feedback family
- note save sends versioned append request, not overwrite request
- stale-version response preserves draft and triggers refetch path
- authorized users can inspect note history; unauthorized users see redaction or denial
- repeated click with same idempotency token does not duplicate rendered history entries after refresh

## Summary

This module keeps the operator workflow lightweight: quick actions for immediate reaction, explicit note revisions for richer context, and conflict-safe reads/writes that preserve auditability without making the UI compute canonical review state.
