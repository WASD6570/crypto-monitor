# Consumer Interfaces And Analytics Hooks

## Module Scope

Define the read contracts and analytics seams that expose operator feedback safely to review surfaces, future tuning workflows, and reporting jobs without leaking raw note text or mutating live decision logic.

## Target Repo Areas

- a review-owned read/query boundary under `services/`
- `apps/web`
- `services/outcome-engine` for join expectations only
- `services/alert-engine` for join expectations only
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `tests/integration`
- `tests/fixtures`

## Module Requirements

- Expose a review snapshot contract for alert detail and list views.
- Expose a history contract for drill-down and audit inspection.
- Expose analytics-safe aggregates that include structured labels but exclude raw notes by default.
- Make joins to alerts, outcomes, simulations, and future baselines version-aware and append-only friendly.
- Keep feedback strictly advisory for review and analysis; no consumer may treat it as automatic threshold-change input.

## Consumer Contracts

### Review Snapshot Read

Plan a lightweight contract keyed by `alertId` with fields such as:

- lifecycle state
- sentiment state
- active qualifiers
- latest note preview or note-present boolean
- last reviewed at/by
- current review version
- counts by feedback family if helpful for history badges

### History Read

Plan a paginated chronological contract that returns:

- immutable event metadata
- actor attribution
- structured action payloads
- note revisions for authorized callers
- version references needed for audit and replay context

### Analytics Export / Query Hook

Plan structured fields suitable for dashboards and tuning analysis:

- `setupFamily`
- symbol and alert timestamps
- outcome result and regime context
- review lifecycle state
- sentiment label
- qualifier flags
- note-present boolean
- first-review latency and last-review latency
- actor cohort or role only if privacy rules allow

Default exclusion: raw note text, client metadata, and personally identifying free text stay out of generic analytics exports.

## Join Strategy

- Join review snapshot to alert detail by immutable `alertId`.
- Join objective outcome and later simulation data by alert identifier and pinned version context.
- For analytics, distinguish event-time alert fields from review-time annotation fields so delayed feedback does not masquerade as immediate signal quality.
- When multiple feedback events exist, analytics should preserve both the current snapshot and counts/timing where needed rather than flattening history into one irreversible label.

## Safe Defaults

- Default UI list posture: fetch review snapshot with alert list rows; fetch full history lazily on detail expansion.
- Default analytics posture: aggregate structured labels daily or on query-time, but keep the append-only event store as the source of truth.
- Default privacy posture: expose raw notes only through operator-authorized review APIs, never through public dashboards or webhook payloads.
- Default tuning posture: feedback can flag hypotheses such as "many thumbs down under fragmentation" but cannot change active config until a separate versioned tuning workflow approves it.

## Negative Cases To Cover In Implementation

- analytics query accidentally requests raw note bodies without permission
- alert list view over-fetches full note histories and misses query targets
- replay or backfill correction changes alert/outcome context but review joins remain ambiguous
- consumer attempts to infer current state from unordered history without review version semantics

## Unit And Integration Test Expectations

- snapshot contract stays stable when multiple historical events exist
- history read returns deterministic ordering and proper auth redaction
- analytics export omits raw note text by default
- alert-detail join shows review context without changing underlying alert/outcome payloads
- query shape supports dashboard targets from operating defaults

## Summary

This module defines how the rest of Initiative 2 consumes operator review safely: fast snapshot reads for the UI, explicit history for audits, and redacted structured exports for analytics and tuning context only.
