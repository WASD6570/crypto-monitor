# Implementation History Contracts And Lookup Keys

## Module Requirements And Scope

- Target repo areas: `schemas/json/features`, `schemas/json/replay`, `libs/go`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Extend the completed current-state contract family with bounded history and audit schemas.
- Define exact lookup keys for historical retrieval without adding a parallel read model or open-ended query language.

## In Scope

- versioned schemas for historical symbol state, historical global state, and audit provenance responses
- shared lookup metadata covering `symbol`, scope, bucket interval, closed-window key or `asOf`, `configVersion`, `algorithmVersion`, and replay identity
- response sections that reuse current-state composite, bucket, regime, recent-context, and provenance shapes where they still apply
- explicit unavailable and superseded markers when a requested historical artifact is missing or has been corrected by replay
- fixture seams so integration and replay tests can assert contract compatibility deterministically

## Out Of Scope

- transport routing, auth, caching, or storage indexing mechanics
- adding new business fields that are not already part of the current-state family or replay provenance
- arbitrary pagination, cursoring, trend analytics, or cross-symbol batch reports

## Contract Family Decisions

- Keep the new family under `schemas/json/features` and name it as an extension of the current read seam, for example:
  - `market_state_history_symbol_v1`
  - `market_state_history_global_v1`
  - `market_state_audit_provenance_v1`
- Reuse the completed current-state nested sections wherever semantics match; only add history-specific wrappers for lookup and lineage.
- Keep history responses additive and versioned. If a later feature materially changes semantics, add a new schema version instead of mutating the v1 contract.

## Required Lookup Metadata

- request identity: `symbol` for symbol scope or explicit `global` scope
- closed-window selector: bucket family (`30s`, `2m`, `5m`) plus bucket key or exact closed `asOf`
- version pin: `configVersion`, `algorithmVersion`, and replay run identifier or manifest reference
- resolution status: `exact`, `unavailable`, or `superseded`
- provenance anchors: source composite snapshot ids, bucket artifact ids, regime artifact ids, and replay correction ids when present

## Response Structure Expectations

- `state` section reuses the current-state contract sections instead of restating market logic in a second schema family
- `lookup` section records the requested bucket/context and the resolved authoritative artifact ids
- `audit` section records correction status, superseded lineage, and reason codes explaining why a replay answer differs from a previously observed live answer
- `availability` section uses machine-readable codes for missing artifacts, retention gaps, mismatched version pins, or incomplete replay context

## Test Expectations

- schema validation for each new history and audit schema
- fixture checks that history responses embed the same current-state subsections and version fields as the completed current-state family
- replay checks that unavailable and superseded markers are stable for the same pinned fixtures and replay manifests

## Summary

This module extends the existing current-state contract family with closed-window lookup metadata and audit lineage. The key rule is reuse: historical and audit reads must package the same authoritative state sections, not introduce a competing market-state schema.
