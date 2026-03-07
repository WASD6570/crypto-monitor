# Implementation Module: Bybit Parser And Runtime Baseline

## Requirements And Scope

- Add explicit native payload structs for one MVP trade path and one MVP book path.
- Map them into shared ingestion types already used by Binance tests.
- Add a small runtime helper surface for Bybit that reuses shared backoff/health/config logic.
- Add deterministic tests for happy-path normalization, gap degradation where applicable, and runtime health evaluation.

## Target Repo Areas

- `services/venue-bybit`
- `libs/go/ingestion`
- `tests/fixtures/events/bybit`

## Key Decisions

- Follow the Binance service shape where practical, but keep Bybit-specific naming explicit.
- Prefer one parser boundary per stream family instead of one huge decoder.
- Only introduce shared helpers in `libs/go` when at least Binance and Bybit both need them.

## Unit Test Expectations

- Deterministic trade normalization.
- Deterministic book normalization or degraded gap path.
- Runtime helper decisions for healthy/degraded/stale states.

## Contract / Fixture / Replay Impacts

- Reuse canonical event and feed-health vocabulary.
- Add only the minimum fixtures needed for Bybit MVP coverage.

## Summary

This module brings Bybit up to the same baseline shape Binance already has: native parsers, shared normalization handoff, and a small runtime health surface.
