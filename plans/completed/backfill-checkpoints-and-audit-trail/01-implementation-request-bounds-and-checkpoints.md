# Implementation Request Bounds And Checkpoints

## Module Requirements And Scope

- Target repo areas: `services/replay-engine`, `services/*` replay-control boundary if added, `libs/go`, `schemas/json/replay`, `configs/*`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Implement one bounded request/control module for replay and backfill operations in Go.
- Keep the module focused on scope validation, request identity, checkpoint shape, resume eligibility, and deterministic overlap handling.

## In Scope

- request payload and CLI/service fields for symbol, venue, stream family, UTC start/end, mode, initiator, and reason
- linkage between replay request identity and run manifest identity from completed prerequisites
- checkpoint record fields for partition position, last materialized event identity, retry lineage, and pinned snapshot references
- resume acceptance/rejection rules when scope, config, contract, or build references drift
- deterministic overlap detection for inspect, rebuild, compare, and apply-capable requests
- fixture-backed tests for bounded requests, resume idempotency, and conflict handling

## Out Of Scope

- final audit record promotion policy details beyond the checkpoint facts this module emits
- external publish execution or downstream sink integration
- storage-engine tables, locking primitives, or migration design
- retention continuity tests beyond ensuring checkpoint records carry logical partition references

## Recommended Repo Breakdown

- `services/replay-engine`: request parsing, scope resolution against replay manifests, conflict checks, and checkpoint-aware orchestration.
- `services/*` replay-control boundary if introduced: CLI/job/API entrypoint that normalizes operator or service requests into one internal command shape.
- `libs/go`: pure helpers for request validation, scope normalization, conflict-key derivation, checkpoint comparison, and resume guards.
- `schemas/json/replay`: versioned request/checkpoint contracts or reserved contract families for internal/external audit-safe serialization.
- `configs/*`: mode allowlists, max window defaults, stream-family allowlists, and apply eligibility flags when runtime policy needs config backing.
- `tests/fixtures`, `tests/integration`, `tests/replay`: pinned request windows, failed-run fixtures, and deterministic resume/conflict scenarios.

## Request Boundary Rules

- Require explicit values for:
  - `symbol`
  - `venueScope`
  - `streamFamilyScope`
  - `startTsUtc`
  - `endTsUtc`
  - `mode`
  - `initiator`
  - `reasonCode`
- Reject wildcard or implicit "repair everything" requests in this slice.
- Normalize equivalent scopes into one canonical request key so conflict checks and audit lookups do not drift on formatting differences.
- Enforce bounded window limits from config or policy, with smaller defaults for apply-capable requests than for inspect-only requests.

## Request And Run Identity

- Preserve the stable replay run ID from `plans/completed/replay-run-manifests-and-ordering/`.
- Add one request identity layer above the run manifest so retries and resumed execution can be linked without inventing a new audit story on every restart.
- Recommended identity fields:
  - canonical scope key
  - requested mode
  - config snapshot digest
  - contract version set digest
  - build provenance digest
  - initiator identity
- A retried request with identical pinned inputs should link to the existing request identity and append retry lineage, not create a logically distinct recovery request.

## Checkpoint Shape

- Every durable checkpoint should record:
  - `requestId`
  - `runId`
  - logical partition or manifest reference
  - last fully materialized event ID
  - optional last scanned event ID when scan and materialize positions differ
  - config snapshot reference
  - contract version reference
  - build provenance reference
  - output mode
  - checkpoint sequence number or monotonic step index
  - retry count
  - last error class
  - recorded-at timestamp
- Keep the checkpoint storage-neutral: it must describe logical progress, not vendor-specific offsets.
- Prefer checkpoint lineage that another agent can inspect without reading process logs.

## Resume Rules

1. Load the prior request identity and most recent durable checkpoint.
2. Re-resolve the scoped manifest set and pinned snapshot references.
3. Reject resume if any of these differ from the original request:
   - canonical scope key
   - config snapshot digest
   - contract version set
   - build provenance
   - output mode
4. Restart from the last fully materialized event boundary, not from an ambiguous partially processed offset.
5. Require downstream materialization paths to use replay-aware idempotency keys so reprocessing from the checkpoint boundary does not duplicate derived artifacts.

## Overlap And Conflict Handling

- Derive one deterministic conflict key from canonical scope plus time window plus correction surface.
- `inspect` and `compare` requests may run concurrently with the same scope when they do not materialize outputs.
- `rebuild` requests for the same correction surface should be serialized or rejected if they would write to the same isolated artifact target.
- apply-capable requests for overlapping windows on the same symbol/venue/stream family should be rejected or queued deterministically; do not rely on best-effort operator coordination.
- If a narrower request overlaps a wider active request, policy should prefer one explicit result:
  - reject the newer request with the active conflict key, or
  - queue it behind the active request using the same normalized scope model
- The plan should not assume concrete locking technology; it should define the logical behavior implementation must preserve.

## Determinism Notes

- Normalize all scopes before deriving IDs, checkpoints, or conflict keys.
- Avoid wall-clock dependent request IDs when logical identity is what matters.
- Keep retry lineage append-only and audit-visible.
- Do not infer resume state from logs, in-memory caches, or mutable current config.

## Unit And Integration Test Expectations

- `go test ./libs/go/... -run 'Test(ReplayRequestValidation|ReplayCheckpointResumeGuard|ReplayConflictKeyNormalization)'`
- `go test ./services/replay-engine/... -run 'TestReplayRequestRejectsUnboundedScope|TestReplayResumeUsesLastMaterializedCheckpoint|TestReplayResumeRejectsConfigSnapshotDrift|TestReplayRejectsOverlappingApplyRequests'`
- `go test ./tests/integration/... -run 'TestBackfillResumeAfterFailure|TestBackfillOverlapConflictHandling'`

## Summary

This module gives implementation one bounded operational control surface: every replay/backfill request is explicit, every resume decision is pinned to immutable inputs, and every conflicting correction request is handled deterministically before later audit and apply work begins.
