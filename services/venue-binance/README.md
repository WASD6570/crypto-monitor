# Venue Binance

- Owns public Binance spot market ingestion for BTC/ETH trades, top-of-book, and order-book bootstrap plus delta recovery.
- Owns Binance USD-M perp sensor ingestion for BTC/ETH funding, open interest, mark/index, and liquidation prints.
- Emits adapter-scoped feed-health updates for reconnect loops, stale messages, sequence gaps, snapshot drift, and local clock degradation.
- Hands normalized payload candidates to `services/normalizer` only after preserving `exchangeTs`, `recvTs`, source symbol, and degraded reasons.

## MVP Streams

- `trades` (`spot`)
- `top-of-book` (`spot`)
- `order-book` (`spot`, snapshot required)
- `funding-rate` (`perpetual`)
- `open-interest` (`perpetual`)
- `mark-index` (`perpetual`)
- `liquidation` (`perpetual`)

## Runtime Notes

- REST recovery is limited to order-book bootstrap and bounded resync.
- Reconnects must use bounded backoff and resubscribe the active stream set.
- Health output is a first-class stream alongside canonical payload handoff.
- `ParseTradeEvent` turns websocket trade payloads into the shared ingestion trade model before canonical normalization.
- `ParseOrderBookSnapshot` and `ParseOrderBookDelta` turn native Binance REST snapshot and websocket `depthUpdate` payloads into the shared ingestion order-book model before sequencing and canonical normalization.
