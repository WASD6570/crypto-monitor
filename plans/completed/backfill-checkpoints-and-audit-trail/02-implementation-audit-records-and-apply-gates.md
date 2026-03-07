# Implementation Audit Records And Apply Gates

## Module Requirements And Scope

- Target repo areas: `services/replay-engine`, `services/*` replay-control or operator-recovery boundary if added, `libs/go`, `schemas/json/replay`, `configs/*`, `tests/integration`, `tests/replay`, `docs/runbooks`
- Implement the structured audit and promotion-control layer that sits around bounded replay/backfill execution.
- Keep the module focused on audit record families, checkpoint lineage visibility, approval/apply seams, and default side-effect safety.

## In Scope

- structured request, execution, checkpoint, outcome, and promotion audit records
- explicit approval/apply context for any run that could promote corrected artifacts
- idempotent apply-gate seams that later downstream sinks or output promotion paths must consume
- runbook-facing operational expectations for inspect, rebuild, compare, resume, reject, and promote paths
- negative-path coverage for missing approval, duplicate promotion attempts, and audit incompleteness

## Out Of Scope

- concrete notification, webhook, or alert resend implementations
- retention-phase smoke matrices for hot/cold replay continuity
- operator UI or authorization product flows beyond naming required server-side context fields
- storage-engine tables, append logs, or message-bus topics

## Recommended Repo Breakdown

- `services/replay-engine`: emit structured execution and outcome records, attach checkpoint lineage, and enforce apply-gate checks before any promotion path.
- `services/*` replay-control or operator-recovery boundary if introduced: capture initiator identity, request source, approval context, and operator notes.
- `libs/go`: shared audit structs, reason enums, promotion token helpers, and idempotency-key composition.
- `schemas/json/replay`: versioned audit and promotion-gate contract families or reserved spaces for request/execution/outcome serialization.
- `configs/*`: approval policy toggles, mode-to-gate rules, and sink-side effect deny-by-default settings.
- `docs/runbooks`: operator guidance for safe request submission, resume behavior, rejected apply attempts, and audit lookup during incident review.

## Audit Record Families

- `request audit record`
  - initiator identity
  - request source
  - reason code
  - operator note
  - normalized scope key
  - requested mode
  - apply intent flag
- `execution audit record`
  - request and run linkage
  - resolved manifest references
  - config snapshot reference
  - contract version set
  - build provenance
  - start/end timestamps
  - counters for processed, late, duplicate, degraded, skipped, and rejected events
- `checkpoint audit record`
  - checkpoint sequence
  - logical partition reference
  - last materialized event ID
  - retry lineage
  - failure class when applicable
- `outcome audit record`
  - terminal status: no-op, compare-only, isolated rebuild, rejected apply, promoted correction
  - artifact references for manifests, diffs, or isolated rebuild outputs
  - prior materialized output reference when a replacement is promoted
  - replacement output reference when promotion succeeds

## Audit Behavior Rules

- Emit structured records for rejected requests as well as successful execution; a denied apply attempt is still an auditable action.
- Keep checkpoint lineage queryable as records rather than reconstructing it from free-form log text.
- Preserve enough snapshot and build metadata to explain why two runs with the same scope could differ.
- Keep audit record vocabulary enumerable and machine-readable so later runbooks and tools can filter by status or risk class.

## Apply-Gate Seam

- `inspect`, `compare`, and isolated `rebuild` must not require approval beyond normal request authorization because they do not promote canonical corrected outputs.
- Any promotion or apply-capable path must require:
  - explicit apply intent on the request
  - server-side authorization context
  - approval or policy context reference
  - replay-aware idempotency key or promotion token
- The apply gate should sit between isolated correction artifacts and canonical downstream materialization so promotion is explicit and auditable.
- Replaying the same approved promotion request must be idempotent: a second attempt should return the existing promoted outcome or a deterministic duplicate-action rejection.

## Side-Effect Safety Rules

- No default replay/backfill mode may emit alerts, notifications, or webhooks.
- Promotion of corrected internal artifacts must remain distinct from later external side effects.
- If a later sink can trigger external behavior, its interface must consume the apply-gate token and its own idempotency key; that downstream implementation remains out of scope here.
- Audit records must state whether a run stopped at isolated outputs or crossed the promotion gate.

## Runbook Expectations

- Document the safe default request forms for inspect, compare, and isolated rebuild.
- Document what operators must provide for apply-capable requests and how rejected approvals surface in audit records.
- Document resume expectations after crash or partial failure, including when a fresh request is required instead of resume.
- Document where to look for request, checkpoint, and outcome records during incident review.

## Unit And Integration Test Expectations

- `go test ./libs/go/... -run 'Test(ReplayAuditRecordSchema|ReplayApplyGateIdempotencyKey|ReplayOutcomeRecord)'`
- `go test ./services/replay-engine/... -run 'TestReplayAuditRecordsCaptureRequestAndOutcome|TestReplayApplyModeRequiresApprovalContext|TestReplayApplyGateIsIdempotent|TestReplayDryRunNeverPublishesSideEffects'`
- `go test ./tests/integration/... -run 'TestBackfillAuditTrailCompleteness|TestBackfillApplyGateNegativePaths'`

## Summary

This module locks the recovery audit story and the no-surprises promotion seam. It ensures replay/backfill work leaves structured evidence of what was requested, what ran, what changed, and whether any correction crossed an explicit apply boundary.
