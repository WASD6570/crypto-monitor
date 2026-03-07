# Implementation Current State Contracts And Schema Seams

## Module Requirements And Scope

- Target repo areas: `schemas/json/features`, `libs/go`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Define the versioned read-model schemas and envelope seams for current market state.
- Package current composite, divergence, quality, and regime outputs so consumers read one authoritative service contract instead of stitching internal service artifacts together.

## In Scope

- schema families for current symbol state, current global state, and bounded recent-context sections
- machine-readable degraded reason, provenance, and availability fields for composite, bucket, and regime subsections
- version metadata fields including schema family version, `configVersion`, `algorithmVersion`, and source bucket timestamps
- fixture seams and schema validation hooks that let integration and replay tests assert contract shape deterministically
- explicit placeholder seams that later history/audit reads can extend without changing current-state semantics

## Out Of Scope

- arbitrary time-range history schemas or pagination contracts
- replay lookup request shapes keyed by bucket/version context beyond naming reserved metadata fields
- endpoint routing, auth, caching, or transport mechanics
- UI labels, formatting strings, or dashboard layout concerns

## Target Contract Family

- Keep the family under `schemas/json/features` and version it explicitly.
- Recommended schema set:
  - `market_state_current_symbol_v1`
  - `market_state_current_global_v1`
  - `market_state_recent_context_v1`
  - `market_state_current_response_v1` as the top-level symbol read model
- Keep schemas additive and conservative; if semantics change materially, add `v2` instead of overloading fields.

## Required Symbol Response Sections

- identity: `symbol`, `asOf`, `schemaVersion`, `configVersion`, `algorithmVersion`
- composite section: current WORLD and USA composite snapshots, contributor coverage, exclusion reasons, timestamp trust, and availability markers
- bucket summary section: current 30s, 2m, and 5m divergence/fragmentation/coverage/market-quality summaries already computed by services
- regime section: current symbol regime, current global ceiling snapshot applied to the symbol, transition reason codes, and trust posture metadata
- recent context section: a fixed-size sequence of the most recent closed windows needed for sparklines and transition context, not an open-ended history API
- provenance section: source bucket identifiers or timestamps, replay/config pinning metadata, and machine-readable degraded reason families

## Required Global Response Sections

- global state identity and `asOf`
- current global ceiling classification with reason codes and transition metadata
- per-symbol summary references for BTC and ETH showing the capped current state at read time
- shared trust or degradation causes that explain why the ceiling is restrictive
- version metadata matching the symbol response family for cross-consumer consistency

## Recent-Context Boundary

- Keep the recent-context window fixed and small so it remains a current-state convenience, not a hidden history API.
- Safe planning default: expose only the latest closed 30s, 2m, and 5m summaries needed for operator context and dashboard sparklines.
- Include ordering and completeness markers so consumers know whether a bucket is missing, unavailable, or degraded.
- Reserve downstream extension points for history/audit work by carrying bucket keys and version metadata now, without adding query parameters or retention expectations here.

## Schema And Consumer Decisions

- Prefer machine-readable enums/codes for regime states, degradation reasons, availability states, and timestamp trust rather than UI prose.
- Keep summary values derived upstream; this module packages them but does not introduce new scoring or recomputation rules.
- Use one top-level symbol response contract that nests the current-state sections so dashboard adapters and service consumers read a single payload family.
- Ensure symbol and global contracts can be fixture-rendered and replay-validated without live network dependencies.

## Test Expectations

- schema validation tests for every new current-state schema
- fixture-backed tests asserting degraded reasons, version metadata, and unavailable-state markers are present when expected
- integration tests proving a consumer can render current state from one response contract without additional joins
- replay checks proving repeated runs emit the same contract payload for the same pinned fixtures and versions

## Summary

This module defines the schema family for current market state, not the transport or storage. The key seam is a small, versioned, service-owned response that packages already-computed composite, bucket, and regime outputs while keeping future history/audit retrieval clearly separate.
