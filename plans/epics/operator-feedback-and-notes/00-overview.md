# Operator Feedback And Notes Overview

## Ordered Implementation Plan

1. Add append-only feedback action and note-event storage for alert review records.
2. Build review workflow rules for save, dismiss, rating labels, note revisions, and conflict-safe annotation reads.
3. Expose service-owned consumer interfaces for the web review surfaces and analytics exports without mutating live alert logic.
4. Validate authz, append-only history, replay/version pinning, negative cases, and review-loop usability before handoff to implementation.

## Problem Statement

Initiative 2 needs a lightweight operator memory system so the first user can mark which alerts felt useful, noisy, mistimed, or context-only without editing configs in the moment. The feature must preserve immutable audit history, keep feedback separate from live alert decision logic, and make later review and tuning workflows smarter without turning subjective reactions into automatic threshold changes.

## Role In Initiative 2

- This is slice 6 of `crypto-market-copilot-alerting-and-evaluation`.
- It starts the subjective half of the first-user review loop after `alert-generation-and-hygiene`, `alert-delivery-and-routing`, and `outcome-evaluation` have created an alert record, delivery trail, and objective result.
- It provides durable human context that later slices consume:
  - `replay-and-analytics-ui` needs current review status plus full annotation history.
  - `baseline-comparison-and-tuning` may use feedback as one review dimension, never as automatic live logic truth.
- It must not mutate thresholds, ranking, routing, or market-state logic directly.

## Role In The First-User Review Loop

1. Alert fires and delivery/outcome records are stored with pinned versions.
2. Operator opens the alert detail in the web UI and applies one or more feedback actions.
3. The system appends immutable feedback and note events linked to the alert and actor.
4. Review surfaces render a current derived review snapshot plus the full append-only history.
5. Later analytics and tuning workflows read the stored review context alongside objective outcomes, simulations, and config versions.

## First-User Outcome

The user should be able to review a recent alert and quickly:

- save it for later follow-up
- dismiss it without deleting history
- mark thumbs up or thumbs down
- label `good setup bad timing` or `useful context only`
- add or revise notes safely without losing prior text
- trust that none of those actions silently changed live behavior

## In Scope

- operator actions: `save`, `dismiss`, `thumbs_up`, `thumbs_down`, `good_setup_bad_timing`, `useful_context_only`
- free-form operator notes with append-only revision history
- immutable storage, version pinning, provenance, and auditability requirements
- service-side derived current review snapshot built from append-only events
- consumer contracts for web review surfaces and analytics exports
- privacy, authorization, and actor attribution rules for annotations
- conflict handling and safe defaults for concurrent feedback or note edits
- validation expectations for persistence, authz, negative cases, and non-mutation of live logic

## Out Of Scope

- automatic threshold mutation, config mutation, or retraining from feedback alone
- alert generation, delivery transport, outcome math, or simulation behavior changes
- collaborative comments beyond operator notes and immutable feedback labels
- file attachments, screenshots, rich text, or external document storage
- full case management workflows, assignment queues, or moderation tooling
- concrete schema and migration implementation beyond the planning boundary

## Safe Defaults

- Default feedback model: every user action appends a new event; no event is edited or deleted in place.
- Default current-state model: consumers read a derived latest snapshot plus the full history so UI stays simple while audit history remains complete.
- Default multiple-feedback posture: contradictory events are allowed over time because review opinion can change; the latest event of each action family becomes current state.
- Default note-edit posture: note editing creates a new note revision event that references the prior note chain; previous note bodies remain readable to authorized operators.
- Default conflict posture: the service rejects stale derived-snapshot writes when the client includes an outdated review version and asks the client to refetch before appending again.
- Default privacy posture: notes and feedback are operator-visible review data, not public alert payload fields; analytics exports should exclude raw note text unless the caller is explicitly authorized.
- Default retention posture: follow initiative defaults for alerts/outcomes/review data and preserve append-only history in hot and cold storage.

## Requirements

- Keep write authority in service-owned endpoints; `apps/web` renders and submits review intent but never authors canonical review state client-side.
- Store every feedback and note action with `alertId`, `actorId`, event timestamp, event source, and pinned `configVersion`/`algorithmVersion` references from the alert being reviewed.
- Preserve append-only history and auditability for both label changes and note revisions.
- Make the derived review snapshot deterministic from ordered events and explicit conflict rules.
- Keep feedback informative for later review and tuning while explicitly preventing automatic threshold or live-policy mutation.
- Enforce server-side authorization for reading and writing operator review data.
- Keep raw note text out of unaudited analytics paths by default.
- Support replay and later correction flows by appending new review events rather than mutating old records.

## Target Repo Areas

- `apps/web`
- a review-owned service boundary under `services/`
- `services/outcome-engine` for consumer seams only
- `services/alert-engine` for consumer seams only
- `libs/go`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `schemas/json/features`
- `tests/integration`
- `tests/fixtures`
- `tests/replay`

## Design Overview

### Why Feedback Must Be Append-Only

- A first-user review loop is only trustworthy if later agents can see what the operator believed at the time, not just the latest rewritten conclusion.
- Notes often evolve after outcome and simulation data arrive; versioned revisions preserve that sequence.
- Tuning and analytics need the difference between immediate reaction and after-the-fact review, which disappears if records are overwritten.

### Feedback Families

- Lifecycle actions: `save`, `dismiss`.
- Directional sentiment: `thumbs_up`, `thumbs_down`.
- Review qualifiers: `good_setup_bad_timing`, `useful_context_only`.
- Notes: free-form text revisions attached to the alert review record.
- Families are intentionally lightweight so the user can review quickly without turning every alert into manual labeling work.

### Derived Current Snapshot

- The system should maintain a read model per alert that answers common UI questions quickly: saved state, dismissed state, latest sentiment, active qualifiers, latest note revision, last reviewed at, and last reviewed by.
- The snapshot is derived from the ordered event log and can be rebuilt deterministically.
- If event ordering ties occur, use persisted event sequence or storage-assigned monotonic review version as the tiebreaker; never depend on client clock alone.

### Note Revision Policy

- The first note creates a note thread identifier scoped to the alert.
- Editing a note appends a new revision event with the full new body and a pointer to the prior revision.
- Safe default: store whole-note revisions instead of diffs so reads and audits stay simple.
- Empty-note submissions should be rejected; clearing a note should be represented by a new revision body that is intentionally blank only if the product later chooses to support that behavior. MVP default is reject blank edits.

### Multiple Feedback Events And Conflict Handling

- The operator may save then dismiss later, or thumbs down immediately and later mark `good setup bad timing` after outcome review. Both should remain in history.
- Derived-state conflicts are resolved by family-specific latest-event wins rules:
  - lifecycle family: latest of `save` or `dismiss`
  - sentiment family: latest of `thumbs_up` or `thumbs_down`
  - qualifier family: latest presence event per qualifier, with explicit future clear events only if needed later
- For MVP planning, qualifiers are append-only positive markers with no removal UI; later removal should also be append-only through explicit clear events.
- Concurrent writes should use optimistic concurrency against the current review snapshot version. On mismatch, append is blocked and the client refetches.

### Privacy And Authorization Boundary

- Operator review data is sensitive because free-form notes may contain workflow context, personal heuristics, or incident details.
- Only authorized operator or admin roles should write feedback and read raw note bodies.
- Lower-trust analytics consumers should read redacted aggregates and structured labels, not free-form note text, unless explicitly granted review-data access.
- Delivery surfaces like Telegram or generic webhooks should not expose raw notes by default.

### Relationship To Live Logic And Tuning

- Feedback enriches review; it does not change live thresholds, severity rules, or market-state gates automatically.
- Tuning consumers may join review labels with outcomes and simulations to identify candidate hypotheses for a future config version.
- Any threshold or rule change must still become a separately reviewed config version and pass replay plus baseline validation.

## ASCII Flow

```text
emitted alert + outcome record + delivery record
                    |
                    v
             operator review UI
      - save / dismiss
      - thumbs up / down
      - good setup bad timing
      - useful context only
      - add or revise note
                    |
                    v
          service-owned review write API
      - authn/authz
      - alert/version lookup
      - optimistic concurrency check
                    |
                    v
          append-only review event log
      - feedback events
      - note revision events
      - actor + timestamps + versions
                    |
          +---------+----------+
          |                    |
          v                    v
 derived current snapshot   audit/history queries
 - latest lifecycle         - full event chain
 - latest sentiment         - note revisions
 - active qualifiers        - provenance
 - latest note
          |
          v
 review UI + analytics joins
 - operator-visible detail
 - redacted aggregate exports
 - tuning context only
 - no automatic live mutation
```
