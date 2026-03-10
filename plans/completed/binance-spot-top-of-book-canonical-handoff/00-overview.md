# Binance Spot Top Of Book Canonical Handoff

## Ordered Implementation Plan

1. Add a bounded Binance Spot `bookTicker` parser seam that turns supervisor-accepted frames into the shared top-of-book message shape without pulling in depth bootstrap or recovery logic.
2. Wire the parsed top-of-book message through `services/normalizer` and `libs/go/ingestion` so canonical `order-book-top` output preserves Binance Spot provenance, raw-append compatibility, and explicit timestamp fallback behavior.
3. Expand Binance fixture and integration coverage so supervisor-fed `bookTicker` frames prove canonical output, stable source identity, and clear separation from depth semantics and supervisor-owned feed health.
4. Record validation evidence in `plans/binance-spot-top-of-book-canonical-handoff/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to the Spot `bookTicker` adapter-to-normalizer handoff for Binance BTC/ETH under the already-completed Spot websocket supervisor.
- Inherit the completed supervisor as the only lifecycle owner for connect, ping/pong, reconnect, rollover, resubscribe, stale-message detection, and adapter-scoped feed health.
- Preserve initiative and Wave 1 contract rules: canonical symbols stay `BTC-USD` and `ETH-USD`, while `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs` remain explicit.
- Prefer the stream's exchange timestamp when the native payload provides one; if Binance `bookTicker` does not provide a trustworthy exchange time, preserve `recvTs` and let canonical timestamp policy degrade explicitly instead of inventing a synthetic exchange clock.
- Keep top-of-book identity stable enough for raw append and replay-sensitive acceptance. If the current shared top-of-book `sourceRecordId` rule is too timestamp-dependent for Binance `bookTicker`, tighten it in the shared path and update affected fixtures/tests in the same feature.
- Keep the feature distinct from depth snapshot/bootstrap, delta sequencing, or market-state API cutover work.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Current repository state to preserve

- `plans/completed/binance-spot-ws-runtime-supervisor/` already fixes the Spot runtime owner and feed-health behavior for `trade` plus `bookTicker`.
- `services/venue-binance/spot_ws_supervisor.go` already accepts `bookTicker` frames on the shared Spot websocket session, but no Binance-specific `bookTicker` parser seam or fixture-backed integration proof exists yet.
- `libs/go/ingestion/book_normalization.go` already supports `top-of-book` normalization and raw append routing, so this feature should reuse that path rather than inventing a Binance-only canonical event type.
- `services/venue-binance/orderbook.go` and depth tests already cover snapshot/delta order-book behavior; this feature must not blur top-of-book handling with later depth sequencing or resync logic.

### Boundaries

- Keep venue-native `bookTicker` parsing in `services/venue-binance`.
- Keep canonical `order-book-top` output and raw append behavior in `services/normalizer` plus `libs/go/ingestion`.
- Keep stale-message and reconnect degradation in the completed supervisor path; normalization here should emit canonical top-of-book output, not invent a second runtime-health policy.
- If shared top-of-book identity or raw append rules need tightening for Binance update IDs, make the smallest shared-library change that preserves existing non-Binance behavior.

### Timestamp and identity posture

- Parse native best bid and best ask prices directly from Binance `bookTicker` payloads and keep the native source symbol unchanged until normalizer metadata canonicalizes the symbol.
- Prefer native exchange time when present; otherwise leave `ExchangeTs` empty and rely on strict timestamp policy to mark the canonical event degraded against `recvTs`.
- Preserve any usable Binance top-of-book update identifier in the shared message so canonical and raw-write identity can stay deterministic even when the stream lacks an exchange timestamp.

### Live vs research boundary

- All websocket frame handling, top-of-book parsing, normalization, and raw append behavior stay in Go under `services/venue-binance`, `services/normalizer`, and `libs/go/ingestion`.
- Offline fixture generation may use research tooling later, but no Python runtime dependency belongs in this live top-of-book path.

## ASCII Flow

```text
spot websocket supervisor
  - connect / subscribe / ping-pong / reconnect / rollover
          |
          v
accepted raw bookTicker frame
  - btcusdt@bookTicker / ethusdt@bookTicker
          |
          v
services/venue-binance
  - native payload parse
  - best bid / ask extraction
  - optional exchangeTs selection
  - update-id preservation
          |
          v
services/normalizer + libs/go/ingestion
  - canonical symbol + provenance
  - strict timestamp policy
  - raw append / source identity
          |
          v
canonical order-book-top event
```
