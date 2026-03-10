# Implementation: Canonical Top Of Book Handoff

## Module Requirements

- Wire or tighten the Binance Spot top-of-book handoff into `services/normalizer` using explicit Spot metadata.
- Preserve canonical `order-book-top` output with `symbol`, `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType=spot` intact.
- Keep raw append behavior aligned with the `top-of-book` stream family, not the depth/order-book stream family.
- Tighten shared top-of-book source identity only if needed so repeated accepted Binance `bookTicker` updates do not drift solely because `recvTs` changed.

## Target Repo Areas

- `services/normalizer`
- `libs/go/ingestion`
- `services/venue-binance`

## Key Decisions

- Reuse `NormalizeOrderBookMessage(...)` and `BuildRawAppendEntryFromOrderBook(...)` rather than adding a Binance-only normalizer surface.
- Keep top-of-book normalization independent from `OrderBookSequencer` snapshot/delta semantics even if the existing shared API still accepts a sequencer parameter.
- If canonical `sourceRecordId` must prefer a stable Binance update ID over timestamps for `top-of-book`, update the shared helper in the narrowest possible way and refresh any affected non-Binance tests/fixtures in the same feature.
- Do not emit feed-health events from top-of-book normalization for timestamp-degraded but otherwise accepted messages; runtime staleness remains supervisor-owned.

## Unit Test Expectations

- canonical Binance top-of-book output keeps `BTC-USD`/`ETH-USD`, `BTCUSDT`/`ETHUSDT`, `quoteCurrency=USDT`, `venue=BINANCE`, and `marketType=spot`
- accepted top-of-book messages produce `bookAction=top-of-book` and stay distinct from snapshot/delta depth actions
- raw append entries for accepted Binance top-of-book messages use `streamFamily=top-of-book`
- repeated accepted messages preserve stable source identity under the chosen top-of-book `sourceRecordId` rule
- missing or invalid exchange time degrades to `recvTs` through the shared timestamp policy without dropping the event

## Summary

This module converts accepted Binance Spot `bookTicker` updates into canonical top-of-book outputs without reopening runtime ownership or leaking depth-specific sequencing into the first live best-bid/best-ask path.
