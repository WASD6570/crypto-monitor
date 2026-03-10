# Binance Live Market Data Feature Map

## 1. `binance-live-contract-seams-and-fixtures`

- Goal: lock the canonical symbol, timestamp, source-ID, and fixture conventions for Binance Spot and USD-M before runtime code spreads those assumptions.
- Primary repo areas: `schemas/json/events`, `libs/go/contracts`, `tests/fixtures`, `tests/integration`, `docs/runbooks`
- Why it stands alone: later ingestion and replay slices should consume one explicit semantics package instead of redefining live Binance behavior independently.

## 2. `binance-spot-trades-and-top-of-book-runtime`

- Goal: establish real Spot WS connection management plus normalized `trade` and `bookTicker` handoff for the tracked symbols.
- Primary repo areas: `services/venue-binance`, `services/normalizer`, `configs/*`, `tests/integration`
- Why it stands alone: it delivers the first live current-state inputs without the extra sequencing complexity of depth recovery.

## 3. `binance-spot-depth-bootstrap-and-recovery`

- Goal: add `depth@100ms` sequencing, REST snapshot bootstrap, bounded resync, and snapshot refresh behavior for Spot order-book integrity.
- Primary repo areas: `services/venue-binance`, `libs/go/ingestion`, `configs/*`, `tests/integration`, `tests/fixtures/events/binance`
- Why it stands alone: snapshot/resync logic is riskier than simple stream intake and needs isolated validation.

## 4. `binance-usdm-context-sensors`

- Goal: ingest USD-M `markPrice@1s`, funding/index data, liquidation snapshots, and polled open interest as canonical derivatives context.
- Primary repo areas: `services/venue-binance`, `services/normalizer`, `schemas/json/events`, `configs/*`, `tests/integration`
- Why it stands alone: USD-M context mixes WS and REST sensor behavior and should not be hidden inside the Spot path.

## 5. `binance-live-raw-storage-and-replay`

- Goal: ensure accepted Binance Spot and USD-M events, plus degraded feed-health outputs, flow into raw storage and replay with stable identities.
- Primary repo areas: `libs/go/ingestion`, `services/replay-engine`, `tests/replay`, `tests/integration`
- Why it stands alone: live state is not trustworthy if the accepted inputs and degradation signals cannot be reproduced.

## 6. `binance-live-market-state-api-cutover`

- Goal: replace deterministic provider behavior with live Binance-backed current-state sourcing while keeping the `services/market-state-api` boundary stable for the web app.
- Primary repo areas: `services/market-state-api`, `services/feature-engine`, `services/regime-engine`, `cmd/market-state-api`, `docker-compose.yml`, `apps/web/tests/e2e`
- Why it stands alone: consumer cutover and operator validation should happen after live ingestion semantics are already stable.

## Cross-Cutting Tracks

- `symbol-and-market-type-policy`: keep canonical symbol names stable while preserving venue-native identifiers and `spot` vs `perpetual` separation.
- `time-and-freshness-policy`: keep exchange-time selection, skew handling, and freshness thresholds explicit and shared.
- `feed-health-observability`: preserve reconnect, stale, gap, resync, and clock degradation as first-class output.
- `local-to-live-boundary`: keep `services/market-state-api` and same-origin `/api` stable while the backing data source changes.
