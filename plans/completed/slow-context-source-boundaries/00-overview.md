# Slow Context Source Boundaries

## Ordered Implementation Plan

1. Add one Go-owned slow-context ingestion boundary for CME volume, CME open interest, and ETF daily flow without tying the live path to a single vendor.
2. Add scheduled polling, publication-state classification, idempotent re-poll handling, and source-health outputs that stay separate from realtime venue feed health.
3. Add deterministic CME and ETF fixtures plus targeted Go validation for published, repeated, delayed, corrected, and failed source reads.

## Problem Statement

`plans/epics/slow-context-panel/` is intentionally late and advisory, but the epic still needs one implementation-ready entry point for slower USA context. Right now there is no bounded live-path seam for CME volume/open interest or ETF daily flow, so later query and dashboard work would have to guess where scheduled fetches live, how source publication states are classified, and how slow-source failures stay isolated from the realtime market-state path.

This feature defines that first seam only: how slower institutional context enters the system in Go, how repeated polling stays idempotent, and how source health remains visible without becoming a hidden dependency for `TRADEABLE`, `WATCH`, or `NO-OPERATE`.

## Bounded Scope

- one dedicated slow-context ingestion boundary in Go for slower scheduled source families
- source-family adapters for CME volume, CME open interest, and ETF daily flow
- scheduled polling defaults and publish-window classification for `published_new_value`, `published_same_value_or_same_asof`, and `not_yet_published`
- source timestamp, publication timestamp, ingest timestamp, and stable dedupe-key capture
- operator-visible slow-source health outputs such as `healthy`, `delayed_publication`, `source_unavailable`, and `parse_failed`
- deterministic fixtures and targeted tests for repeated polling, correction handling, and delayed publication behavior

## Out Of Scope

- normalized storage/query response design beyond the adapter handoff needed for the next child feature
- dashboard rendering, UI copy, or client state changes
- vendor-specific auth commitments, billing decisions, or production credentials
- using slow context as a hard realtime gate
- Python in the live runtime path
- concrete shared schema rollout unless a later feature proves one is required

## Requirements

- Keep the live path in Go and preserve the monorepo boundary between `services/*` and offline Python work.
- Prefer one explicit slow-context service boundary over scattering polling logic across venue adapters or dashboard consumers.
- Keep source handling source-family based and provider-agnostic; later implementation can plug in one vendor without baking that choice into the core abstractions.
- Preserve idempotency for repeated polling of unchanged daily/session data.
- Capture both source timing and local ingest timing so later freshness classification does not guess.
- Keep slow-source health separate from existing realtime feed-health semantics and payloads.
- Preserve replay/audit friendliness by making correction handling explicit rather than silently overwriting prior accepted points.
- Keep configuration shallow and boring: publish-window defaults, polling cadence, and rate/backoff knobs only where needed for the live Go path.

## Target Repo Areas

- `services/slow-context` (new dedicated service boundary for scheduled slow-source acquisition)
- optional shared helpers under `libs/go` only if adapter or polling primitives are reused cleanly
- optional minimal config entries under `configs/*` if scheduling/backoff defaults need environment-specific values
- `tests/fixtures/slow-context`

## Module Breakdown

### 1. Slow-Context Service Boundary And Source Interfaces

- Create one shallow `services/slow-context` home for scheduled source acquisition and adapter registration.
- Define source-family adapter interfaces for CME session/daily context and ETF daily-flow context.
- Keep provider parsing behind the adapter boundary so downstream code sees only classified candidate records plus timing metadata.

### 2. Polling, Publish-State, And Source-Health Handling

- Add the scheduler/poller seam for expected publish windows, coarse off-hours backoff, and repeated-poll idempotency.
- Emit source-family health/state outputs that remain advisory and isolated from realtime venue-health streams.
- Make correction handling explicit for same-as-of republishes.

### 3. Deterministic Fixtures And Validation

- Add representative CME and ETF fixtures for new publication, repeated polling, delayed publication, corrected publication, and source failure.
- Add targeted Go tests for parsing, idempotency, delayed-publication classification, and auditable correction handling.

## Acceptance Criteria

- Another agent can implement the live-path slow-context acquisition seam without reopening the parent epic.
- The plan names the exact repo areas for the new service boundary, fixtures, and optional config.
- The plan keeps storage/query semantics for the next child feature instead of bundling them into source-boundary work.
- Validation commands use explicit Go binaries and prove idempotency plus source-health isolation.

## ASCII Flow

```text
scheduled poller in Go
        |
        v
services/slow-context
  - source-family adapter registry
  - publish-window scheduler
  - retry/backoff rules
  - source-health outputs
        |
        +--> CME volume / OI adapter
        |      - published_new_value
        |      - published_same_value_or_same_asof
        |      - not_yet_published
        |
        +--> ETF daily flow adapter
               - published_new_value
               - published_same_value_or_same_asof
               - not_yet_published
        |
        v
classified slow-context candidates
  - sourceKey
  - metric family
  - asOf / published / ingest timestamps
  - dedupe identity
  - source health state
        |
        v
next child feature:
slow-context-query-surface-and-freshness
```

## Live-Path Boundary

- All scheduled fetches, publication classification, and source-health outputs stay in Go under `services/slow-context`.
- Python may assist with offline fixture inspection later, but it is not part of the live acquisition path.
- The dashboard and query layers consume later service-owned outputs only; this feature does not add UI behavior.
