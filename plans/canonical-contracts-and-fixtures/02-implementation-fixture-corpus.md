# Implementation Module 2: Fixture Corpus And Replay Seeds

## Scope

Design the deterministic fixture corpus used for schema validation, normalization checks, replay smoke tests, and later parity work.

## Target Repo Areas

- `tests/fixtures`
- `tests/replay`
- `docs/specs`

## Requirements

- Define fixture naming and directory layout by family, venue, symbol, and scenario.
- Include happy-path fixtures for:
  - trades
  - top of book
  - order book snapshot + delta sequences
  - funding
  - open interest
  - mark/index
  - liquidations where supported
- Include edge-case fixtures for:
  - sequence gaps
  - forced resync
  - stale or missing feed messages
  - timestamp degradation
  - late or out-of-order events
  - quote-currency variants across `USD`, `USDT`, and `USDC`
- Define a small replay seed set that later agents can use to prove determinism:
  - one normal microstructure window
  - one fragmented USA vs WORLD window
  - one degraded-feed window

## Key Decisions To Lock

- Fixtures must be deterministic and small enough for fast local validation.
- Raw source payload examples and canonical expected outputs should be paired where that reduces ambiguity.
- Replay seeds should favor high-signal windows over long bulk dumps.
- Fixture metadata should record source venue, scenario purpose, and intended consumer checks.

## Deliverables

- Fixture directory specification
- Scenario catalog for normal and degraded cases
- Replay seed catalog with deterministic expectations
- Fixture manifest format or equivalent indexing approach

## Unit Test Expectations

- Fixture manifests must detect missing expected outputs.
- Replay seeds must preserve event ordering and timestamp provenance.
- Edge-case fixtures must exercise degraded paths intentionally rather than incidentally.

## Contract / Fixture / Replay Impacts

- Later normalization and replay implementations rely on these fixtures to prove correctness.
- Visibility and alerting tests should be able to reuse the same replay seeds without redefining scenario meaning.
- Optional Python parity work should consume the same fixture corpus, not a separate copy.

## Summary

This module turns abstract schemas into concrete, replayable examples. The fixture corpus is the shared truth set that later implementations use to prove they interpret contracts the same way.
