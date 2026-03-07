# Implementation Module: Coinbase Parser And Runtime Baseline

## Requirements And Scope

- Add one explicit Coinbase trade parser boundary and one explicit book/top-of-book boundary.
- Feed both through shared ingestion normalization helpers.
- Add the small Coinbase runtime health surface needed for reconnects, staleness, and degraded-state decisions.
- Add deterministic tests using Coinbase fixtures.

## Target Repo Areas

- `services/venue-coinbase`
- `libs/go/ingestion`
- `tests/fixtures/events/coinbase`

## Key Decisions

- Keep Coinbase message handling explicit and venue-local.
- Reuse shared runtime health helpers and thresholds rather than inventing Coinbase-only health vocabulary.

## Unit Test Expectations

- Trade parsing and normalization are deterministic.
- Book/top-of-book parsing and normalization are deterministic.
- Runtime state degrades predictably under stale and reconnect scenarios.

## Contract / Fixture / Replay Impacts

- Fixtures should preserve Coinbase source metadata needed for replay.
- No new canonical contract family is introduced.

## Summary

This module establishes Coinbase as a real ingestion participant with the same parser/runtime baseline already prototyped for Binance.
