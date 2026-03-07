# Implementation Module 1: Raw Write Boundary

## Scope

Implement the append-only write boundary that persists canonical events immediately after `services/normalizer` and preserves the exact provenance later replay depends on.

## Target Repo Areas

- `services/normalizer`
- `libs/go/ingestion`
- `libs/go/contracts`
- `configs/*`
- `tests/fixtures`
- `tests/integration`

## Module Requirements

- Add one explicit raw writer interface owned by Go code and called from the `services/normalizer` success path.
- Persist one immutable raw append entry per normalized event.
- Persist enough fields to make later replay ordering and timestamp choice auditable without consulting live adapter state.
- Treat duplicate arrivals as audit-visible facts instead of destructive overwrites or silent drops.
- Keep storage-engine details behind the writer boundary; no table design or migration planning belongs in this feature.
- Fail writes on contract/version mismatch rather than accepting ambiguous raw records.

## Implementation Decisions To Lock

- The write boundary starts after normalization, not inside venue adapters and not inside downstream feature engines.
- The writer contract should accept canonical events plus ingest metadata, not venue-specific raw payload structs.
- The persisted entry is append-only even when the source event is duplicate, late, or degraded.
- Duplicate suppression for downstream rebuilds belongs later; this module only preserves the immutable write history and duplicate audit metadata.

## Required Persisted Fields

- canonical identity:
  - canonical event ID
  - venue message ID when present
  - venue sequence or update ID when present
  - stream key used for identity fallback
- market provenance:
  - symbol
  - venue
  - market type
  - stream family
  - source instrument identifier when different from canonical symbol
- time provenance:
  - `exchangeTs`
  - `recvTs`
  - selected bucket timestamp source
  - timestamp degradation reason
  - late-event marker
- ingest provenance:
  - normalizer service identity
  - connection/session reference
  - normalization or build version reference
- audit metadata:
  - duplicate marker or duplicate counter linkage
  - degraded-feed linkage or reference

## Recommended Package Shape

- Keep the normalizer-facing writer interface small and boring.
- Put shared append-entry types and routing helpers in `libs/go/ingestion` unless they are purely service-local.
- Keep canonical contract validation or schema-version guards aligned with `libs/go/contracts` and `schemas/json/replay` references when the write boundary stores replay-facing metadata.

## Delivery Checklist

- normalizer-to-writer handoff with explicit success and failure behavior
- append-entry type definition and validation rules
- duplicate identity precedence implementation notes
- config surface for durability level, duplicate audit behavior, and late-event tagging thresholds if not already inherited

## Unit Test Expectations

- `TestRawWriteBoundaryAppendsImmutableEntries` verifies later writes never mutate previously written entries.
- `TestRawWriteBoundaryPersistsTimestampProvenance` verifies `exchangeTs`, `recvTs`, selected bucket source, and degraded reason are stored explicitly.
- `TestRawWriteBoundaryRecordsDuplicateAuditFacts` verifies duplicate inputs stay audit-visible.
- `TestRawWriteBoundaryRejectsContractMismatch` verifies unknown or mismatched contract versions fail fast.

## Acceptance Criteria

- Another agent can implement the write boundary without making storage-engine choices.
- The normalizer handoff is explicit and singular.
- Persisted raw entries retain the timestamp and identity facts needed for later deterministic replay.
- Duplicate and late-event handling preserve auditability without introducing live-side mutation.

## Summary

This module fixes the immutable write-side boundary: canonical events leave `services/normalizer` once, enter the raw log as append-only facts, and carry enough provenance that replay never has to reconstruct how the event was interpreted during live ingest.
