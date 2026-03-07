# Implementation Module 1: Raw Event Persistence

## Scope

Define where canonical raw events are persisted, what metadata must travel with them, how append-only behavior is preserved, and how hot/cold retention stays replay-safe.

## Target Repo Areas

- `services/normalizer`
- `services/*` raw persistence boundary added during implementation if needed
- `libs/go`
- `configs/*`
- `schemas/json/events`
- `schemas/json/replay`
- `tests/fixtures`
- `tests/integration`

## Requirements

- Persist canonical events immediately after normalization.
- Treat persisted raw events as append-only records.
- Preserve enough metadata to reproduce later ordering, timestamp decisions, and deduplication behavior.
- Keep the storage boundary neutral to the final storage engine; do not hard-code a database product in the plan.
- Support the operating-default retention model:
  - 30 days hot raw storage
  - 365 days compressed cold raw storage
- Keep replay-critical metadata available across both hot and cold tiers.
- Keep live writes in Go and do not require Python tooling.

## Data Boundary To Lock

The raw persistence layer should store canonical facts and audit metadata, not downstream interpretation. Plan for these record groups:

- canonical event payload or canonical payload reference
- source provenance:
  - venue
  - market type
  - stream family
  - source symbol / instrument ID
  - source message ID or sequence fields when available
- timestamp provenance:
  - `exchangeTs`
  - `recvTs`
  - selected bucket timestamp source
  - timestamp degradation reason when fallback occurs
- ingest provenance:
  - ingest service identity
  - ingest attempt or connection session reference
  - normalization version or build reference
- audit flags:
  - duplicate or replayed-source marker
  - late-event marker
  - degraded-feed marker reference

Do not plan derived feature values, tradeability state, or alert outputs into the raw record.

## Partitioning And Access Strategy

- Use UTC day as the first partitioning recommendation because replay, retention, and one-day debug workflows are day-scoped.
- Partition by canonical `symbol` and `venue` within the day.
- Add stream family partitioning only where it materially improves replay cost or retention operations.
- Keep a manifest or index that maps a replay request to the exact hot or cold partitions required.
- Avoid deep speculative sharding schemes that complicate local or dev replay.

## Deduplication And Identity Recommendations

- Deduplicate at read/replay and downstream materialization boundaries using stable source identity, not destructive deletion.
- Preserve the first-seen ingest record and record later duplicates as audit facts or counters.
- Prefer this identity precedence when available:
  1. venue source message ID
  2. venue sequence/update ID + stream key
  3. deterministic canonical event ID derived from stable event fields
- The plan should require explicit handling for venues that do not provide globally unique IDs.

## Retention And Tiering Rules

- Hot-to-cold movement must preserve byte-for-byte replay inputs or a canonical equivalent with identical ordering semantics.
- Compression is allowed only if replay can stream data without changing record boundaries or event identity.
- Retention jobs must emit manifests or logs that prove which partitions were moved, retained, or expired.
- Expiry must happen only after the configured retention window and must not delete audit records required for operator investigation inside that policy window.

## Operational Recommendations

- Raw writes should fail loudly on contract/version mismatch rather than accept ambiguous payloads.
- Write acknowledgement should happen only after persistence guarantees meet the selected durability level.
- Backpressure behavior should degrade predictably and visibly; do not silently drop canonical events.
- Persist health-state references close enough to the event stream that later replay can reconstruct degraded windows.

## Unit Test Expectations

- Append-only write tests confirm later writes do not mutate existing raw records.
- Partition routing tests confirm the same event always lands in the same scoped partition for the same timestamp decision.
- Retention-tier tests confirm hot-to-cold transitions preserve replay manifests and event identity.
- Deduplication tests confirm duplicate source events are audit-visible and do not cause ambiguous downstream replay.
- Timestamp fallback tests confirm degraded timestamp choice is persisted, not inferred later.

## Contract / Fixture / Replay Impacts

- Fixtures need raw storage examples that include duplicate, late, and timestamp-degraded cases.
- Replay contracts must reference persisted raw partitions or manifests without depending on live adapter state.
- Feed-health records need stable linkage so degraded windows can be replayed alongside market events.

## Summary

This module gives the system an immutable canonical event history. The key outcome is not a specific database choice; it is a storage boundary that preserves raw facts, provenance, and retention behavior without making replay guess how the past looked.
