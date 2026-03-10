# Binance Spot Trade Canonical Handoff

## Ordered Implementation Plan

1. Tighten the Binance Spot trade adapter seam so native websocket `trade` payloads map cleanly into the shared ingestion trade model behind the completed Spot supervisor.
2. Wire the trade handoff through `services/normalizer` with explicit Spot metadata, stable source identity, raw append compatibility, and strict timestamp fallback behavior.
3. Expand Binance Spot fixture and integration coverage so accepted supervisor-fed trade frames prove happy-path, timestamp degradation, and duplicate-sensitive identity stability.
4. Record validation evidence in `plans/binance-spot-trade-canonical-handoff/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to the Spot `trade` adapter-to-normalizer handoff for Binance BTC/ETH under the already-completed Spot websocket supervisor.
- Inherit the completed supervisor as the only lifecycle owner for connect, ping/pong, reconnect, rollover, resubscribe, and adapter-scoped feed-health; do not reintroduce a second Spot runtime path here.
- Preserve Wave 1 and initiative contract rules: canonical symbols stay `BTC-USD` and `ETH-USD`, while `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs` remain explicit.
- Prefer Binance trade time as `exchangeTs`, fall back to the event time when needed, and rely on strict canonical timestamp policy to degrade to `recvTs` only when the selected exchange time is missing, invalid, or implausible.
- Preserve stable duplicate-sensitive source identity for repeated Spot trades and keep raw append/replay compatibility in scope for accepted trade events.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Current repository state to preserve

- `services/venue-binance/trades.go` and `services/venue-binance/trades_test.go` already parse native Binance Spot trade payloads into the shared ingestion trade model.
- `services/normalizer/service.go` already exposes the shared Spot trade normalization seam.
- `plans/completed/binance-spot-ws-runtime-supervisor/` already fixes the Spot runtime owner and feed-health behavior for `trade` plus `bookTicker`.
- Existing Binance Spot fixtures include a happy trade canonical case and a degraded timestamp case, but the feature still needs one active implementation plan that makes the supervisor-to-parser-to-normalizer boundary explicit and fully testable.

### Boundaries

- Keep all venue-native Spot trade parsing in `services/venue-binance`.
- Keep canonical trade output and raw append behavior in `services/normalizer` and `libs/go/ingestion`.
- Keep lifecycle validation out of this feature except where it proves the supervisor raw-frame seam can feed trade parsing without widening scope.
- Do not absorb `bookTicker` parsing or depth bootstrap/recovery work.

## ASCII Flow

```text
spot websocket supervisor
  - connect / subscribe / ping-pong / reconnect / rollover
          |
          v
accepted raw trade frame
  - btcusdt@trade / ethusdt@trade
          |
          v
services/venue-binance/trades.go
  - native payload parse
  - trade time -> exchangeTs selection
  - sourceSymbol preservation
          |
          v
services/normalizer
  - canonical symbol + provenance
  - timestamp policy
  - raw append compatibility
          |
          v
canonical market-trade event
```
