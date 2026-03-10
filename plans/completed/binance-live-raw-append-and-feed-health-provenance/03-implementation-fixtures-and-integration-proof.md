# Implementation: Fixtures And Integration Proof

## Module Requirements

- Add deterministic proof that completed Binance live outputs append into stable raw entries before the later replay feature consumes them.
- Keep validation attached to the raw append implementation slice rather than spinning off a validation-only child feature.
- Cover both happy accepted-input paths and degradation-sensitive paths that materially change raw provenance.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/integration`
- `libs/go/ingestion`
- `services/venue-binance`

## Key Decisions

- Prefer reusing the completed Binance fixture corpus as the seed for raw append assertions instead of creating a second parallel event vocabulary.
- Add integration assertions for provenance-sensitive paths that the completed feature tests do not already cover:
  - Spot depth recovery health with explicit degraded-feed linkage
  - USD-M mixed-surface feed-health identity separation
  - duplicate-input stability for at least one Binance family with venue message ID and one family with sequence identity
- Keep replay-sensitive checks in this slice focused on raw append output facts; full replay determinism remains the next child feature.

## Integration Expectations

- one integration test proves Spot websocket-originated families append with stable websocket context
- one integration test proves Spot depth health appends with explicit degraded-feed provenance when depth leaves synchronized state
- one integration test proves USD-M websocket and REST surfaces append with distinct source identity and partition behavior
- one duplicate-input check proves raw duplicate facts are deterministic for repeated Binance inputs

## Unit Test Expectations

- fixture-backed raw append assertions cover at least one canonical payload from each completed Binance surface
- appended raw entries preserve stream-family routing, bucket timestamp source, and duplicate identity precedence
- deterministic repeated appends do not drift in partition key or source identity

## Summary

This module closes the audit gap for raw append by proving that completed Binance live outputs can be stored with stable provenance and replay-ready identity before the replay epic continues.
