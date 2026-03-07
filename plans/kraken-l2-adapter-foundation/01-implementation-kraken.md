# Implementation Module: Kraken Parser And L2 Integrity Baseline

## Requirements And Scope

- Add one Kraken trade parser boundary.
- Add one Kraken L2/book parser boundary with explicit sequencing or equivalent ordering checks.
- Drive shared normalization and degraded health outputs from deterministic fixtures.
- Add the Kraken runtime helper surface needed for staleness, reconnects, resyncs, and gap handling.

## Target Repo Areas

- `services/venue-kraken`
- `libs/go/ingestion`
- `tests/fixtures/events/kraken`

## Key Decisions

- Never guess through Kraken L2 gaps.
- Keep the L2 integrity handling explicit even if other venues are simpler.
- Reuse shared health state and degradation reasons.

## Unit Test Expectations

- Deterministic trade normalization.
- Deterministic L2 happy path.
- Deterministic degraded gap/resync path.
- Runtime state correctly reflects repeated resync pressure.

## Contract / Fixture / Replay Impacts

- Kraken fixtures should capture the exact ordering metadata needed for replay.
- Later replay should be able to distinguish healthy vs degraded Kraken book periods.

## Summary

This module gives Kraken a bounded plan for the harder L2 integrity work without dragging the rest of the umbrella plan into the same implementation pass.
