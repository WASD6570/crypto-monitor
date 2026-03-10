# Refinement Handoff: Binance Spot Trades And Top Of Book Runtime

## Next Recommended Child Feature

- `binance-spot-ws-runtime-supervisor`

## Why This Is Next

- both stream-specific handoff slices need one explicit owner for Binance Spot websocket lifecycle behavior before they can be planned safely
- feed-health is a first-class output for this epic, and the supervisor slice is where heartbeat, reconnect-loop, and stale-runtime degradation semantics belong
- defining the supervisor first keeps later child plans focused on stream parsing and canonical handoff instead of reopening runtime recovery behavior in two places

## Safe Parallel Planning

- `binance-spot-trade-canonical-handoff` and `binance-spot-top-of-book-canonical-handoff` are safe to plan in parallel after the supervisor child feature is refined

## Blocked Until

- `binance-spot-trade-canonical-handoff` is blocked on the supervisor child feature naming the shared connection, heartbeat, reconnect, and feed-health lifecycle it consumes
- `binance-spot-top-of-book-canonical-handoff` is blocked on the same supervisor decisions so it does not invent a second runtime path or drift into depth semantics

## What Already Exists

- `initiatives/crypto-market-copilot-binance-live-market-data/03-handoff.md` already frames this epic as the first Spot runtime slice for `trade` and `bookTicker`
- `plans/epics/binance-live-contract-seams-and-fixtures/` already carries the symbol, timestamp, provenance, and source-ID decisions this epic must inherit
- `plans/completed/market-ingestion-and-feed-health/` already supplies feed-health vocabulary and normalization expectations for degraded runtime behavior
- `services/venue-binance/README.md` and `configs/local/ingestion.v1.json` already reserve the Spot stream inventory and runtime default thresholds
- `tests/fixtures/events/binance/BTC-USD/happy-trade-usdt.fixture.v1.json` and `tests/fixtures/events/binance/ETH-USD/edge-timestamp-degraded-usdt.fixture.v1.json` already provide starter trade coverage, while top-of-book coverage is still absent

## Assumptions To Preserve

- this epic covers Spot websocket runtime only for `trade` and `bookTicker`; depth bootstrap and recovery remain in `plans/epics/binance-spot-depth-bootstrap-and-recovery/`
- Wave 1 contract seam decisions are inherited, not reopened
- canonical symbols stay asset-centric as `BTC-USD` and `ETH-USD`, with explicit `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- feed-health remains machine-readable and first-class beside canonical event handoff
- Go remains the live runtime path; no Python runtime dependency is introduced

## Recommended Follow-On After This Child Feature

1. `binance-spot-trade-canonical-handoff`
2. `binance-spot-top-of-book-canonical-handoff`
3. run direct live Binance validation for the completed Spot runtime slices; do not create a smoke-only child plan
