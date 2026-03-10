# Binance Spot WS Runtime Supervisor

## Ordered Implementation Plan

1. Build a Spot-only websocket supervisor state surface that derives the BTC/ETH `trade` plus `bookTicker` subscription set without inheriting depth bootstrap responsibilities.
2. Implement the connection lifecycle driver for connect, Binance ping/pong, subscribe, proactive pre-24h rollover, bounded reconnect, and deterministic resubscribe.
3. Publish adapter-scoped feed-health and runtime observability from supervisor state transitions without coupling to trade or top-of-book parsing.
4. Add deterministic lifecycle tests and scripted fixture hooks that prove healthy startup, stale-message degradation, reconnect-loop degradation, and rollover recovery.
5. Validate the slice with focused Go tests and integration smoke, then hand the active plan to `feature-implementing`.

## Requirements

- Scope is only the bounded Spot websocket supervisor for BTC/ETH `trade` and `bookTicker`.
- The supervisor owns connect, subscribe, ping/pong, proactive 24h rollover, bounded reconnect, resubscribe, and adapter-scoped feed health.
- The supervisor must not absorb Spot trade parsing, Spot top-of-book parsing, or Spot depth snapshot/bootstrap logic.
- Wave 1 contract seam decisions for canonical symbols, timestamps, provenance fields, and source-record IDs are inherited, not reopened.
- Go remains the live runtime path under `services/venue-binance`; Python stays out of runtime behavior.
- Feed-health remains a first-class machine-readable output beside later canonical event handoff.
- Validation must stay targeted and high-signal, with deterministic time and scripted websocket behavior.

## Design Notes

### Bounded runtime owner

- Implement one Spot supervisor for the tracked BTC/ETH `trade` and `bookTicker` stream set so later trade and top-of-book child features share one lifecycle owner.
- Treat the supervisor as transport and state orchestration only: it emits raw frame envelopes and feed-health decisions, while later child features own payload parsing and canonical handoff.
- Filter the Binance runtime config down to Spot `trades` and `top-of-book` definitions before evaluating health so `order-book` snapshot requirements do not leak into this slice.

### Connection lifecycle

- Use one combined Binance Spot websocket session for the four desired subscriptions: `btcusdt@trade`, `ethusdt@trade`, `btcusdt@bookTicker`, and `ethusdt@bookTicker`.
- Maintain explicit desired, pending, and active subscription state so reconnect and rollover can resubscribe deterministically without parsing market payloads.
- Honor Binance-native ping/pong behavior at the supervisor boundary and treat accepted control traffic plus accepted data frames as liveness inputs.
- Force a clean rollover before Binance's 24h disconnect window using a fixed adapter deadline with headroom, then reconnect through the same bounded backoff path.
- Bound reconnect pressure with existing config defaults for backoff and connect-rate limits, and surface reconnect-loop degradation when thresholds are crossed.

### Feed health and observability

- Publish adapter-scoped feed-health from connection state, message freshness, reconnect counts, and local clock offset only; sequence-gap and snapshot semantics stay untouched in this slice.
- Keep feed-health source-record identity on the Wave 1 pattern rather than inventing supervisor-local formats.
- Expose supervisor counters and timestamps in testable state so later ops work can inspect connection start, last frame, last pong, last subscribe ack, and rollover cause.

### Live vs research boundary

- All runtime state, websocket IO, and feed-health publication stay in Go under `services/venue-binance`.
- Any future offline fixture generation may use research tooling, but no Python code belongs in the live supervisor path.

## ASCII Flow

```text
configs/local/ingestion.v1.json
          |
          v
spot supervisor config filter
  - symbols: BTC-USD, ETH-USD
  - streams: trades, top-of-book
          |
          v
spot websocket supervisor
  - connect
  - subscribe
  - ping/pong
  - rollover before 24h
  - bounded reconnect/resubscribe
          |
          +----> raw frame envelopes for later trade/bookTicker handlers
          |
          v
adapter-scoped feed-health output
  - connection state
  - message freshness
  - reconnect-loop status
  - clock degradation
```
