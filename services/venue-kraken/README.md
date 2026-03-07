# Venue Kraken

- Owns public Kraken spot ingestion for BTC/ETH trades, top-of-book, and L2 order-book reconstruction.
- Exists as an explicit service boundary because Kraken integrity handling should not be hidden inside another venue runtime.
- Emits adapter-scoped feed-health updates for reconnect loops, stale messages, sequence gaps, snapshot drift, resync loops, and local clock degradation.
- Hands normalized payload candidates to `services/normalizer` only after preserving `exchangeTs`, `recvTs`, source symbol, and degraded reasons.

## MVP Streams

- `trades` (`spot`)
- `top-of-book` (`spot`)
- `order-book` (`spot`, snapshot required)

## Runtime Notes

- Kraken order-book recovery is bounded, observable, and must hard-fail into degraded state on sequence uncertainty.
- Reconnects must use bounded backoff and resubscribe the active stream set.
- Health output is a first-class stream alongside canonical payload handoff.
