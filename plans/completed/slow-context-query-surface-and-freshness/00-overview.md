# Slow Context Query Surface And Freshness

## Ordered Implementation Plan

1. Add one service-owned slow-context record and store/query seam that preserves as-of, published, ingest, revision, cadence, and last-known-value history without reopening the source-boundary slice.
2. Add deterministic freshness classification and explicit unavailable handling for CME and ETF slow context under pinned clocks.
3. Add non-blocking current-state integration so slow-context lookup failure degrades only the slow-context block, not the core market-state response.
4. Add deterministic fixtures, targeted Go tests, and focused integration coverage for freshness thresholds, explicit unavailable payloads, and current-state isolation.

## Problem Statement

`plans/completed/slow-context-source-boundaries/` established how slower institutional context enters the system, but there is still no service-owned read model for later consumers. Without this slice, later dashboard work would have to guess how to interpret revisions, which timestamps matter for freshness, when a record becomes `delayed` or `stale`, and how slow-context failures should appear alongside existing current-state responses.

This feature defines that missing seam only: one normalized slow-context record/query model, one deterministic freshness policy, and one non-blocking integration path that keeps slow context advisory while the realtime market-state path stays authoritative.

## Bounded Scope

- one Go-owned slow-context record model for CME volume, CME open interest, and ETF daily flow
- append-safe latest-value and correction-aware query/store semantics needed for current reads
- deterministic `fresh`, `delayed`, `stale`, and `unavailable` classification under source-family rules
- one dedicated slow-context response block and/or current-state integration seam owned by services
- explicit unavailable/error-scoped behavior when no trusted record exists or slow-context lookup fails
- deterministic fixtures and focused Go validation for freshness thresholds, unavailable handling, and current-state isolation

## Out Of Scope

- dashboard layout, panel copy, or `apps/web` implementation
- changing slow-source acquisition semantics already completed in `plans/completed/slow-context-source-boundaries/`
- using slow context as a hard dependency for `TRADEABLE`, `WATCH`, or `NO-OPERATE`
- speculative migration/backfill orchestration beyond the bounded latest/history semantics needed for this query slice
- Python in the live runtime path

## Requirements

- Reuse `plans/completed/slow-context-source-boundaries/` as the authoritative source-family, publication-state, correction, and failure-isolation boundary.
- Keep the live path in Go under `services/*` and shared Go helpers only; Python remains offline-only.
- Preserve distinct `asOfTs`, `publishedTs`, and `ingestTs` fields so freshness and operator copy never collapse naturally slow data into operational lag.
- Compute freshness in services using source-family defaults, not in the UI:
  - CME volume/open interest: `fresh` through the next expected publish window, `delayed` after that window, `stale` after 36 hours
  - ETF daily flow: `fresh` through the next expected daily publish window, `delayed` after that window, `stale` after 48 hours
- Reserve `unavailable` for no accepted record yet or lookup failure that prevents a trustworthy last-known value.
- Keep slow-context response behavior conservative and explicit: no silent omission, no optimistic fallback, no effect on core current-state success when slow-context lookup fails.
- Preserve append-safe correction history and revision visibility so later audit/replay work can explain superseded slow-context values.
- Add shared schema files only if this slice must expose a non-Go consumer contract now; otherwise keep the response seam Go-owned and ready for a later contract extraction.

## Target Repo Areas

- `services/slow-context`
- `services/feature-engine`
- optional shared Go types under `libs/go/features` or `libs/go/slowcontext` if a reusable read model emerges cleanly
- optional `schemas/json/features` only if a consumer-facing JSON contract is required during implementation
- `tests/fixtures/slow-context`
- `tests/integration`
- optional `tests/replay` if implementation makes the slow-context store replay-visible

## Module Breakdown

### 1. Normalized Record, Store, And Freshness Policy

- Extend the slow-context service from acquisition-only parsing into a normalized accepted-record and latest/history query seam.
- Preserve revision-aware latest-value lookup and explicit source-family freshness thresholds.
- Keep storage expectations append-safe and audit-friendly without choosing a heavyweight migration or warehouse design.

### 2. Query Surface And Current-State Isolation

- Add one service-owned slow-context response block with value, cadence, freshness, age, revision, and message key/text fields.
- Integrate that block into the current-state path only through an isolated optional seam so slow-context lookup failure cannot fail BTC/ETH market-state delivery.
- Reuse the existing current-state availability and provenance conventions where helpful, but keep slow-context semantics advisory and clearly separate.

### 3. Fixtures, Integration, And Determinism

- Add pinned-clock fixtures and focused integration tests for freshness boundaries, unavailable handling, and non-blocking current-state fallback.
- Add replay-focused checks only if the implementation introduces replay-visible persistence earlier than planned.

## Acceptance Criteria

- Another agent can implement the service-owned slow-context query slice without reopening the parent epic or reworking the completed source-boundary feature.
- The plan names exact repo areas for the slow-context service, feature-engine integration seam, fixtures, and optional shared contracts.
- The plan keeps UI work deferred to `slow-context-dashboard-panel` and keeps source-acquisition work in the completed archive.
- Validation commands are concrete, Go-first, and cover freshness classification, explicit unavailable payloads, and current-state isolation.

## ASCII Flow

```text
completed source-boundary slice
services/slow-context adapters + tracker
        |
        v
accepted slow-context records
  - contextFamily
  - sourceKey
  - asset
  - asOf / published / ingest timestamps
  - revision
  - value + unit
        |
        v
slow-context query assembler in Go
  - freshness classification
  - latest trusted value
  - unavailable fallback
  - cadence + age metadata
  - operator-safe message key/text
        |
        +--> dedicated slow-context response seam
        |
        +--> optional feature-engine current-state inclusion
                 - slow-context block succeeds or degrades alone
                 - core market-state response still succeeds
        |
        v
next child feature:
slow-context-dashboard-panel
```

## Live-Path Boundary

- All normalization, freshness classification, and query assembly stay in Go under `services/slow-context` and `services/feature-engine`.
- Python may support offline fixture inspection later, but it is not part of the live query/runtime path.
- `apps/web` remains a later read-only consumer and does not derive freshness or gating semantics locally.
