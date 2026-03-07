# Testing Plan

## Goal

Prove that operator feedback and notes are append-only, authorized, version-aware, and useful in review surfaces without mutating live alert logic.

## Validation Commands

- `go test ./services/... -run Review`
- `go test ./tests/integration -run Review`
- `go test ./tests/replay -run Review`
- `pnpm --dir apps/web test -- --run review`
- `pnpm --dir apps/web test -- --run alert-detail`

## Smoke Matrix

### 1. Quick Feedback Persistence

- Create or load fixture alert with pinned outcome and delivery records.
- Submit `save`, `thumbs_up`, `good_setup_bad_timing`, and `dismiss` as separate append requests.
- Verify full history contains four immutable events in deterministic order.
- Verify current snapshot reflects latest lifecycle state, latest sentiment state, and active qualifier.
- Verify alert payload and outcome payload remain unchanged.

### 2. Note Revision History

- Submit first note revision, then a second revised note body.
- Verify both note revisions remain readable in chronological history for an authorized user.
- Verify snapshot points to the latest note revision and preview only.
- Verify prior note text is not overwritten.

### 3. Conflict Handling

- Read snapshot version `N` from two simulated clients.
- Client A appends a note or quick action and advances to `N+1`.
- Client B submits with stale `N`.
- Verify stale write is rejected with conflict metadata and no silent overwrite.
- Verify client retry with refreshed version succeeds as a new append event.

### 4. Idempotent Retry

- Submit a feedback write with a client idempotency token.
- Replay the same request after simulated network timeout.
- Verify only one review event exists and the response is stable.
- Submit the same semantic action without the token.
- Verify a second distinct event is preserved when policy allows duplicate manual actions.

### 5. Authorization And Privacy

- Authorized operator can read/write review labels and raw notes.
- Non-operator or lower-privilege user cannot write feedback.
- Lower-privilege analytics caller receives structured labels only, with raw notes redacted or absent.
- Delivery/webhook payload tests confirm notes are excluded by default.

### 6. Replay And Audit Integrity

- Rebuild the current snapshot from the review event log and compare it to the stored projection.
- Run replay-oriented validation to ensure ordered events derive the same snapshot deterministically.
- Verify appended review events preserve reviewed alert version references even after unrelated config changes.

## Negative Cases

- unknown `alertId`
- malformed action enum
- blank note body
- oversize note body
- stale `currentReviewVersion`
- unauthorized raw note read
- unauthorized feedback write
- duplicate transport retry without idempotency token expectations
- attempt to derive live threshold mutation from feedback event processing

## Assertions

- append-only history is preserved for every feedback family and note revision
- current snapshot is deterministic from ordered events
- live alert logic, severity, and thresholds do not change as a side effect of feedback writes
- note text stays behind operator-authorized interfaces
- audit/history views expose actor, timestamp, and reviewed version provenance

## Expected Artifacts

- `plans/operator-feedback-and-notes/testing-report.md`

## Handoff Notes For Implementation

- Prefer fixture-driven tests that reuse alert and outcome records from prior Initiative 2 slices.
- Include at least one scenario where operator sentiment changes after objective outcome review so history vs snapshot behavior is explicit.
- Include one dashboard-query performance smoke check to confirm review snapshot reads do not require loading full note histories.
