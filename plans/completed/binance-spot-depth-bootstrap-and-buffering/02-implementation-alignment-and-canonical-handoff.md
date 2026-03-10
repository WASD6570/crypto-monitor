# Implementation: Alignment And Canonical Handoff

## Module Requirements

- Implement the explicit Binance startup alignment rule between snapshot `lastUpdateId` and buffered `depthUpdate` windows.
- Emit synchronized snapshot/delta messages through the existing shared order-book sequencer and canonical normalization path.
- Preserve explicit provenance and timestamp policy on accepted depth outputs without inventing new canonical event types.
- Surface bootstrap failure as explicit state or error that the later recovery feature can consume without redefining startup semantics.

## Target Repo Areas

- `services/venue-binance`
- `libs/go/ingestion`
- `services/normalizer`

## Key Decisions

- Use Binance-native `U`/`u` semantics to locate the first buffered delta eligible after the snapshot boundary.
- Ignore buffered deltas whose final update ID is at or below the snapshot sequence, accept the first delta whose window bridges the snapshot boundary, and pass subsequent accepted deltas in deterministic order.
- Reuse `ingestion.OrderBookSequencer` and `NormalizeOrderBookMessage(...)` instead of introducing a Binance-only depth acceptance pipeline.
- Keep REST snapshot `exchangeTs` empty and let strict timestamp policy degrade explicitly against `recvTs`; websocket deltas keep native exchange time when present.

## Unit Test Expectations

- a happy bootstrap path accepts snapshot plus the correct first bridging delta window
- stale buffered deltas before the bridging window are ignored deterministically
- if no buffered delta bridges the snapshot boundary, bootstrap fails explicitly and does not emit synchronized depth output
- accepted snapshot and delta messages normalize into canonical order-book outputs with `BTC-USD`/`ETH-USD`, native `sourceSymbol`, `quoteCurrency=USDT`, `venue=BINANCE`, and `marketType=spot`

## Summary

This module turns buffered native Binance depth inputs into one explicit synchronized startup rule that downstream recovery work can inherit instead of rediscover.
