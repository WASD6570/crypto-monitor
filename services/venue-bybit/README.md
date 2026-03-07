# Venue Bybit

- Owns public Bybit spot market ingestion for BTC/ETH trades, top-of-book, and order-book recovery.
- Owns Bybit perp sensor ingestion for BTC/ETH funding, open interest, mark/index, and liquidation prints.
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
