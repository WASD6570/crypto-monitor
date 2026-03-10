# Refinement Handoff: Binance Spot Depth Bootstrap And Recovery

## Next Recommended Child Feature

- `binance-spot-depth-bootstrap-and-buffering`

## Why This Is Next

- the epic cannot reason about resync loops safely until the startup boundary between REST snapshot state and websocket `depth@100ms` deltas is explicit
- the repo already has parser and runtime primitives, so the highest-value missing work is the bounded bootstrap owner that turns them into one trustworthy depth startup path
- settling buffered startup alignment first prevents the later recovery slice from baking in the wrong acceptance window or resync trigger assumptions

## Safe Parallel Planning

- none yet
- `binance-spot-depth-resync-and-snapshot-health` should wait until the bootstrap child feature settles accepted startup and progression semantics

## Blocked Until

- `binance-spot-depth-resync-and-snapshot-health` is blocked on the bootstrap child feature naming how buffered deltas are retained, when the first accepted delta may apply, and what state is considered synchronized

## What Already Exists

- `plans/completed/binance-spot-ws-runtime-supervisor/` already defines the Spot runtime owner for websocket lifecycle and non-depth feed-health posture
- `plans/completed/binance-spot-trade-canonical-handoff/` and `plans/completed/binance-spot-top-of-book-canonical-handoff/` already bound the non-depth Spot parser seams this epic must preserve
- `services/venue-binance/orderbook.go` already parses Binance REST snapshot and `depthUpdate` payloads into shared order-book messages
- `services/venue-binance/runtime.go` already exposes snapshot cooldown, rate-limit, reconnect-loop, and resync-loop evaluation primitives
- `tests/fixtures/events/binance/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json` already encodes the expected feed-health outcome for a sequence-gap path

## Assumptions To Preserve

- this epic covers Spot order-book bootstrap and recovery only; it must not reopen completed trade/top-of-book work
- canonical symbols remain asset-centric as `BTC-USD` and `ETH-USD`, with explicit `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- sequence gaps, stale snapshots, resync loops, and related recovery states remain machine-visible, not log-only
- Go remains the live runtime path; no Python runtime dependency is introduced
- direct live Binance validation happens after implementation and is not its own child plan

## Recommended Follow-On After This Child Feature

1. `binance-spot-depth-resync-and-snapshot-health`
2. run direct live Binance depth validation for the completed Spot depth path
3. continue to `plans/epics/binance-live-raw-storage-and-replay/` once Spot depth semantics are stable enough for replay-safe identity work
