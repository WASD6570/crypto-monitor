# Implementation: Mixed-Surface Integration

## Module Requirements

- Add targeted integration tests that exercise websocket and REST Binance USD-M sensors together.
- Prove that websocket feed-health and REST feed-health can report different states for the same symbol without ambiguity.
- Reuse settled parser and normalization behavior; this module must validate, not redesign.

## Target Repo Areas

- `tests/integration`
- `services/venue-binance`
- `services/normalizer`

## Unit Test Expectations

- mixed-surface happy path stays fixture-backed
- websocket health can remain healthy while REST polling becomes stale for the same symbol
- distinct degradation reasons remain machine-visible after canonical feed-health normalization

## Summary

This module proves the completed websocket and REST slices behave as one coherent context surface without flattening their semantics.
