# Implementation: Fixtures And Integration Proof

## Module Requirements

- Add Binance Spot top-of-book fixtures that exercise both accepted canonical output and explicit timestamp fallback behavior.
- Add targeted integration coverage that proves the completed Spot supervisor can feed accepted `bookTicker` frames into canonical normalization.
- Keep proof focused on top-of-book handoff, stable identity, and stream-family routing; depth bootstrap/recovery remains out of scope.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`
- `tests/integration`

## Key Decisions

- Add at least one BTC happy-path/native top-of-book fixture and one ETH edge fixture for missing, invalid, or skewed exchange time depending on the settled parser behavior.
- If the shared top-of-book `sourceRecordId` rule changes, update any affected existing fixtures or assertions in the same patch so canonical expectations stay explicit.
- Follow the trade feature pattern by proving supervisor `AcceptDataFrame(...) -> ParseTopOfBookFrame(...) -> NormalizeOrderBook(...)` without needing live Binance access.
- Include one duplicate-sensitive or repeated-message proof if the implementation changes shared top-of-book identity or raw append behavior.

## Unit Test Expectations

- fixture-backed integration proves accepted Spot `bookTicker` frames normalize into canonical `order-book-top` events
- timestamp-degraded coverage remains machine-visible and does not require a separate feed-health event for accepted top-of-book messages
- raw append entries stay in the `top-of-book` stream family and preserve the chosen source identity
- top-of-book proofs stay separate from depth snapshot/delta fixtures and do not assert any depth resync behavior

## Summary

This module closes the proof gap between parser unit tests and supervisor-backed integration so the Binance Spot top-of-book slice is ready for later direct live Spot runtime validation.
