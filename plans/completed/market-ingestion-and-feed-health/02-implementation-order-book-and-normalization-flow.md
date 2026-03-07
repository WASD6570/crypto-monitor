# Implementation Module 2: Order Book Integrity And Normalization Flow

## Scope

Plan how order book state is bootstrapped and maintained per venue and how raw venue payloads become canonical events without losing provenance.

## Target Repo Areas

- `services/venue-*`
- `services/normalizer`
- `libs/go`
- `tests/fixtures`
- `tests/integration`

## Requirements

- Define per-venue order book bootstrap rules:
  - snapshot acquisition
  - delta buffering if required
  - sequence alignment
  - snapshot freshness expectations
- Define explicit gap handling and resync rules.
- Define when an order book is considered unusable or degraded.
- Define the handoff from venue-specific payloads to canonical event types from `schemas/json/events`.
- Preserve:
  - source venue identity
  - market type
  - quote currency
  - exchange sequence or equivalent source ordering metadata where available
  - `exchangeTs` and `recvTs`
  - degraded reason flags when applicable

## Key Decisions To Lock

- Never guess through sequence gaps; resync instead.
- Keep normalization deterministic and side-effect free at the payload-conversion layer.
- Distinguish raw venue ordering metadata from canonical event identity so replay and debugging stay possible.
- Emit canonical events only after adapter-specific integrity checks pass or after degradation is explicitly marked.

## Deliverables

- Per-venue order book handling notes
- Canonical normalization handoff design
- Gap and resync state machine description
- Scenario map for normal, degraded, and stale-event flows

## Unit Test Expectations

- Snapshot + delta reconstruction should be testable from fixtures.
- Gap detection should trigger a deterministic degraded/resync path.
- Timestamp fallback behavior should be test-covered for invalid or implausible `exchangeTs` cases.
- Canonical normalization should produce the same output for the same fixture inputs.

## Contract / Fixture / Replay Impacts

- Event contracts and fixture corpus from the first feature must be reused rather than reinvented.
- Replay correctness depends on preserving source ordering metadata and degraded markers.
- Later market-state logic depends on knowing whether data is healthy, stale, or resynced.

## Summary

This module ensures the market stream is reconstructable and canonicalized without losing the details needed for replay, audit, and degradation-aware downstream logic.
