# Binance Spot Depth Bootstrap And Buffering

## Ordered Implementation Plan

1. Add a bounded Spot depth bootstrap owner in `services/venue-binance` that requests `/api/v3/depth`, retains buffered `depth@100ms` deltas during startup, and exposes synchronized startup state without re-owning generic websocket lifecycle behavior.
2. Wire accepted bootstrap snapshot plus buffered delta progression through the shared ingestion order-book sequencer and canonical normalization path so accepted depth outputs preserve Binance Spot provenance and explicit timestamp handling.
3. Expand deterministic Binance depth fixtures and integration coverage so bootstrap alignment, stale-pre-bootstrap deltas, and accepted-first-window behavior are proven before the later resync/snapshot-health slice.
4. Record validation evidence in `plans/binance-spot-depth-bootstrap-and-buffering/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to the initial trustworthy Spot depth startup path for Binance BTC/ETH: REST snapshot fetch, websocket delta buffering, first accepted delta-window alignment, and synchronized handoff into existing sequencing/normalization code.
- Inherit the completed Spot websocket supervisor as the only generic Spot lifecycle owner for connect, subscribe, ping/pong, reconnect, rollover, and non-depth feed health.
- Reuse existing Binance depth parsers in `services/venue-binance/orderbook.go` and existing shared sequencing semantics in `libs/go/ingestion/orderbook.go`; this feature should orchestrate them rather than redesign them.
- Preserve Wave 1 and completed Spot contract rules: canonical symbols stay `BTC-USD` and `ETH-USD`, while `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs` remain explicit.
- Keep REST snapshot exchange time empty and let strict timestamp policy degrade explicitly against `recvTs` instead of inventing a synthetic snapshot exchange timestamp.
- Do not absorb resync-loop handling, snapshot refresh cadence, snapshot-stale degradation, or cooldown/rate-limit recovery policy beyond the minimum bootstrap attempt needed to enter a synchronized depth state; those belong to `binance-spot-depth-resync-and-snapshot-health`.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Current repository state to preserve

- `plans/completed/binance-spot-ws-runtime-supervisor/` already fixes the Spot lifecycle owner for websocket transport and feed-health posture outside depth-specific sequencing.
- `services/venue-binance/orderbook.go` already parses Binance REST snapshots and `depthUpdate` payloads into `ingestion.OrderBookMessage` with snapshot `Sequence=lastUpdateId` and delta `Sequence=u`.
- `libs/go/ingestion/orderbook.go` already supports snapshot acceptance, sequential delta acceptance, stale delta ignore, and resync-required decisions for gaps.
- `services/venue-binance/runtime.go` and config defaults already expose snapshot cooldown and rate-limit primitives that later recovery work will use; this feature should only touch them if startup bootstrap needs the smallest explicit owner boundary.

### Bounded startup owner

- Introduce a Spot depth bootstrap owner that lives beside the completed supervisor rather than inside `services/normalizer` or shared ingestion code.
- Treat the owner as startup orchestration only: collect buffered deltas, fetch and parse one snapshot, decide which buffered deltas are eligible after the snapshot boundary, and emit synchronized snapshot/delta messages for the existing sequencer.
- Keep the websocket data source unchanged: accepted `depthUpdate` frames still originate from the Spot supervisor/runtime path.

### Alignment posture

- Buffer native `depthUpdate` deltas until a REST snapshot is available.
- Make the startup acceptance rule explicit in code and tests using Binance's `U`/`u` window semantics relative to snapshot `lastUpdateId`.
- Drop or ignore buffered deltas that are definitively stale relative to the snapshot, accept the first window that bridges the snapshot boundary, then pass accepted updates into the shared sequencer in deterministic order.
- If no buffered delta can bridge the snapshot boundary, fail startup explicitly into the later recovery path instead of silently declaring synchronization.

### Live vs research boundary

- All snapshot bootstrap, delta buffering, and synchronized handoff state stay in Go under `services/venue-binance`, with canonical output still produced by `services/normalizer` plus `libs/go/ingestion`.
- Offline fixture generation may use research tooling later, but no Python runtime dependency belongs in this live depth path.

## ASCII Flow

```text
spot websocket supervisor
  - owns depth stream transport
          |
          v
accepted depthUpdate frames
  - buffered before sync
          |
          +-------------------+
          |                   |
          v                   |
services/venue-binance        |
depth bootstrap owner         |
  - request /api/v3/depth     |
  - parse snapshot            |
  - retain buffered deltas    |
  - find first bridging delta |
          |                   |
          +---------> synchronized snapshot + accepted deltas
                                   |
                                   v
services/normalizer + libs/go/ingestion
  - shared sequencer
  - canonical order-book-top output
  - explicit degraded timestamp fallback
```
