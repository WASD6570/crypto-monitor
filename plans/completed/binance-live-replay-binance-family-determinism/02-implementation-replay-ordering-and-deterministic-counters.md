# Implementation: Replay Ordering And Deterministic Counters

## Module Requirements

- Prove that replay execution remains deterministic for the completed Binance family set across inspect, rebuild, and compare modes.
- Preserve duplicate-input visibility, degraded timestamp evidence, late-event counters, and distinct feed-health identities without adding a Binance-only replay mode.
- Keep ordering behavior driven by persisted raw-entry facts rather than recomputing venue semantics during replay.

## Target Repo Areas

- `services/replay-engine`
- `services/replay-engine/runtime.go`
- `services/replay-engine/runtime_test.go`
- `tests/replay`

## Key Decisions

- Preserve the current shared ordering rule: `bucketTimestamp -> sequence when present -> canonicalEventID` unless Binance raw evidence proves one narrow shared gap.
- Treat duplicate Binance raw entries as replay-visible inputs that increment counters and affect compare/rebuild digests deterministically.
- Preserve `BucketTimestampSource`, `TimestampDegradationReason`, `Late`, and `DegradedFeedRef` as replay evidence rather than replay-time heuristics.

## Data And Algorithm Notes

- Add deterministic coverage for representative Binance cases:
  - Spot top-of-book vs Spot trade with mixed sequence availability at the same bucket timestamp
  - Spot depth feed-health entries that carry degraded source identity and degraded-feed linkage
  - USD-M websocket and REST-originated families with distinct stream keys and duplicate identities
  - repeated duplicate inputs for one message-ID-backed family and one sequence-backed family
- Validate that compare mode classifies identical repeated runs as `match` and concrete duplicate/degraded changes as `drift` in a stable way.
- If helper refactors are needed in replay tests, keep them minimal and directly tied to Binance family coverage.

## Unit Test Expectations

- repeated replay execution over the same Binance raw entries yields identical ordered IDs, output digest, and input counters
- mixed Binance entries with and without sequences sort in a stable, explainable order
- replay counters continue to reflect duplicates, late events, and degraded timestamp events for Binance inputs
- compare mode surfaces deterministic drift when one retained Binance raw entry is added, removed, or reordered by changed persisted evidence

## Contract / Fixture / Replay Notes

- Do not reopen the raw duplicate identity precedence settled by the prior child feature; replay should consume it unchanged.
- Keep replay artifacts and audit records generic so downstream audit consumers do not need Binance-specific branches.

## Summary

This module locks replay execution onto persisted Binance raw-entry facts so repeated runs and audit counters stay deterministic across the completed live family set.
