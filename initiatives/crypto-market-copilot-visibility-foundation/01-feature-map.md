# Visibility Foundation Feature Map

## 1. `canonical-contracts-and-fixtures`

- Goal: define stable payload families and deterministic fixtures for events, features, replay, and market-state outputs.
- Primary repo areas: `schemas/json`, `tests/fixtures`, `tests/parity`, `docs/specs`
- Why it stands alone: all visibility surfaces depend on stable contracts and replayable examples.

## 2. `market-ingestion-and-feed-health`

- Goal: ingest Binance, Bybit, Coinbase, and Kraken with resilient WS/REST behavior, order-book integrity checks, and feed health signals.
- Primary repo areas: `services/venue-*`, `services/normalizer`, `libs/go`, `configs/*`
- Why it stands alone: trusted state starts with trusted ingress and visible degradation.

## 3. `raw-storage-and-replay-foundation`

- Goal: persist canonical events append-only and replay them deterministically for audits and state reconstruction.
- Primary repo areas: `services/replay-engine`, `services/backfill-engine`, `libs/go`, `tests/replay`
- Why it stands alone: no visibility layer is trustworthy if it cannot be reproduced.

## 4. `world-usa-composites-and-market-state`

- Goal: compute per-venue and composite features plus the 5m market state and fragmentation metrics.
- Primary repo areas: `services/feature-engine`, `services/regime-engine`, `schemas/json/features`, `tests/integration`
- Why it stands alone: this turns raw market data into a product-level worldview.

## 5. `visibility-dashboard-core`

- Goal: present symbol overview, microstructure, derivatives context, feed health, and market-state views in the web app.
- Primary repo areas: `apps/web`, `libs/ts`, `tests/e2e`
- Why it stands alone: operators need a polished monitoring loop, not just backend correctness.

## 6. `slow-context-panel`

- Goal: surface CME volume/OI and ETF flow context as clearly slower, contextual signals.
- Primary repo areas: `services/feature-engine`, `apps/web`, `apps/research`
- Why it stands alone: it enriches visibility but should not block the core live state.

## Cross-Cutting Tracks

- `config-versioning`: ensure threshold and regime config is externalized and versioned.
- `time-policy-and-bucketing`: ensure visibility outputs use the same event-time policy everywhere.
- `feed-observability`: make reconnects, lag, staleness, and gap recovery visible in both logs and UI.
