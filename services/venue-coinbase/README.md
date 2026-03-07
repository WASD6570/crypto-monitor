# Venue Coinbase

- Owns public Coinbase spot ingestion for BTC/ETH trades and best bid/ask updates.
- Treats Coinbase as a spot-only venue in the MVP; no perp sensors are in scope here.
- Emits adapter-scoped feed-health updates for reconnect loops, stale messages, and local clock degradation.
- Hands normalized payload candidates to `services/normalizer` only after preserving `exchangeTs`, `recvTs`, source symbol, and degraded reasons.

## MVP Streams

- `trades` (`spot`)
- `top-of-book` (`spot`)

## Runtime Notes

- No order-book snapshot bootstrap is required in the first slice; the adapter focuses on continuous BBO trust.
- Reconnects must use bounded backoff and resubscribe the active stream set.
- Health output is a first-class stream alongside canonical payload handoff.
