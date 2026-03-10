# Refinement Map: Binance Spot Trades And Top Of Book Runtime

## Current Status

This Wave 2 epic is now materialized from the initiative handoff and still needs decomposition before `feature-planning`. It inherits the Wave 1 contract seam and should not reopen identity, timestamp, or source-record debates.

## What Is Already Covered

- `plans/epics/binance-live-contract-seams-and-fixtures/` defines the contract intent this runtime slice must inherit: asset-centric canonical symbols, explicit provenance fields, stream-specific time rules, and stable source identities
- `plans/completed/market-ingestion-and-feed-health/` already defines feed-health as a first-class machine-readable output, plus the rule that stale inputs and reconnect/resync failures degrade explicitly
- `services/venue-binance/README.md` already reserves the MVP Spot surface as trades, top-of-book, and later order-book recovery
- `configs/local/ingestion.v1.json` already contains the Binance reconnect, heartbeat, stale-message, and resubscribe defaults this epic should treat as operating defaults rather than test-only values
- existing Binance fixtures already cover Spot trade happy-path and timestamp degradation, but they do not yet show top-of-book runtime coverage

## What Remains

- define the bounded runtime slice for the shared Spot websocket supervisor that owns heartbeat, reconnect, rollover, and feed-health behavior for `trade` plus `bookTicker`
- define the trade-specific child feature that parses Spot trade messages and hands them to canonical normalization under the locked Wave 1 seam
- define the top-of-book-specific child feature that parses Spot `bookTicker` messages and hands them to canonical normalization under the same seam
- reserve combined Spot runtime validation as direct post-implementation testing against the real Binance API, not as a separate planned child feature

## What This Epic Must Not Absorb

- depth snapshot/bootstrap sequencing, buffered deltas, or gap recovery work
- shared schema redesign or new canonical symbol rules
- raw append/replay storage changes
- market-state API cutover or frontend behavior changes

## Refinement Waves

### Wave 2A

- `binance-spot-ws-runtime-supervisor`
- Why first: both stream-specific child features need one agreed runtime owner for heartbeat handling, reconnects, resubscribe behavior, and feed-health emission.

### Wave 2B

- `binance-spot-trade-canonical-handoff`
- `binance-spot-top-of-book-canonical-handoff`
- Why next: once the runtime boundary is clear, trade and top-of-book parsing can be planned as separate bounded handoff slices that consume the same supervisor rules.

## Notes For Future Planning

- keep feed-health output equal in importance to canonical trade and order-book-top output; this epic is not complete if only payload parsing is covered
- keep canonical symbols asset-centric as `BTC-USD` and `ETH-USD`, with provenance explicit through `sourceSymbol`, `quoteCurrency`, and `marketType`
- keep the Spot runtime limited to `trade` and `bookTicker`; if planning begins to require snapshots, buffering, or sequence repair, move that work into the later depth epic
- do not create smoke-only child features; combined runtime validation should run directly after the owning implementation slices land
- assume local config defaults are the baseline unless a later feature plan proves an environment-specific override is required
