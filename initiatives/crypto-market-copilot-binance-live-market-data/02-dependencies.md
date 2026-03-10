# Binance Live Market Data Dependencies

## Suggested Order

1. `binance-live-contract-seams-and-fixtures`
2. `binance-spot-trades-and-top-of-book-runtime`
3. `binance-usdm-context-sensors`
4. `binance-spot-depth-bootstrap-and-recovery`
5. `binance-live-raw-storage-and-replay`
6. `binance-live-market-state-api-cutover`

## Dependency Notes

### `binance-live-contract-seams-and-fixtures`

- Depends on: current event families, replay rules, and market-state consumer expectations already in the repo
- Unlocks: every later live Binance slice
- Risk: high

### `binance-spot-trades-and-top-of-book-runtime`

- Depends on: contract, symbol, timestamp, and source-ID decisions from the first slice
- Unlocks: first live price path, feed-health verification, and later market-state cutover work
- Risk: high

### `binance-usdm-context-sensors`

- Depends on: contract and timestamp decisions from the first slice
- Unlocks: derivatives context, funding/index data, liquidation visibility, and open-interest integration
- Risk: medium-high

### `binance-spot-depth-bootstrap-and-recovery`

- Depends on: the Spot runtime foundation plus agreed snapshot/resync semantics
- Unlocks: robust order-book integrity, richer microstructure inputs, and bounded resync behavior
- Risk: high

### `binance-live-raw-storage-and-replay`

- Depends on: accepted live Spot and USD-M outputs from prior slices
- Unlocks: deterministic auditability, replay safety, and future live/backfill confidence
- Risk: high

### `binance-live-market-state-api-cutover`

- Depends on: stable live ingestion semantics, replay-safe identities, and a known way to query current live state
- Unlocks: end-to-end live dashboard behavior behind the existing API boundary
- Risk: high

## Cross-Cutting Dependency Notes

- `schemas/json/events/*` already reserve the main public-event families needed by this initiative, but refinement must confirm whether source-ID and field-level expectations are sufficient for live Binance behavior.
- `configs/*/ingestion.v1.json` already encodes bounded reconnect, snapshot, and health thresholds; refinement should treat those as defaults, not incidental test values.
- `services/market-state-api` should remain consumer-stable while the backing provider changes.
- Spot order-book bootstrap is rollout-sensitive because sequence gaps must degrade explicitly instead of being silently hidden.
- USD-M open interest is rollout-sensitive because the current official source is REST polling rather than an obvious matching market-data WS stream.

## Inputs Already Available From Existing Work

- canonical event schemas for trade, order-book top, feed health, funding, mark/index, open interest, and liquidation
- venue-runtime config and feed-health vocabulary under `libs/go/ingestion`
- deterministic current-state and dashboard API boundary via `services/market-state-api`
- compose-backed web-to-Go integration path already verified locally
