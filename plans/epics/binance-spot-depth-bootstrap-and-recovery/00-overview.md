# Binance Spot Depth Bootstrap And Recovery

## Epic Summary

This epic covers the remaining Spot order-book integrity work after the completed trade, top-of-book, and websocket-supervisor slices. It must add `/api/v3/depth` bootstrap, `depth@100ms` delta intake, buffered startup alignment, bounded resync behavior, snapshot refresh handling, and explicit feed-health degradation without reopening already-set Spot runtime or contract decisions.

## In Scope

- Binance Spot `/api/v3/depth` snapshot bootstrap for tracked BTC/ETH symbols
- Binance Spot `depth@100ms` websocket delta intake and parsing
- buffered startup alignment between REST snapshot state and websocket deltas
- accepted sequence rules for initial depth handoff and later delta acceptance
- bounded resync behavior after sequence gaps, stale snapshots, or drift
- snapshot refresh cadence, cooldown, and rate-limit handling using existing config defaults
- machine-readable feed-health output for sequence-gap, resync-loop, snapshot-stale, and related degraded states

## Out Of Scope

- Spot trade or `bookTicker` handoff work already completed in archived child features
- replay/raw-storage rollout work owned by `plans/epics/binance-live-raw-storage-and-replay/`
- market-state API or dashboard cutover work owned by `plans/epics/binance-live-market-state-api-cutover/`
- schema redesign beyond narrow changes proven necessary by concrete Binance depth behavior
- private/authenticated Binance endpoints or multi-venue order-book redesign

## Target Repo Areas

- `services/venue-binance`
- `libs/go/ingestion`
- `configs/*/ingestion.v1.json`
- `tests/fixtures/events/binance`
- `tests/integration`

## Validation Shape

- targeted Go tests for REST snapshot parsing, websocket delta parsing, and buffered startup alignment
- deterministic gap/resync, snapshot cooldown, rate-limit, and snapshot-staleness tests
- normalization checks proving accepted depth snapshot/delta messages preserve asset-centric symbols and explicit provenance
- direct live validation against Binance Spot `depth@100ms` and `/api/v3/depth` once the owning implementation slices land

## Major Constraints

- inherit completed Spot runtime ownership from `plans/completed/binance-spot-ws-runtime-supervisor/`; do not invent a second generic Spot connection manager if the existing supervisor can be extended safely
- inherit existing Spot contract rules: canonical symbols stay `BTC-USD` and `ETH-USD`, with explicit `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs`
- preserve explicit gap and resync visibility; order-book repair must never be silent
- keep raw/replay identity concerns scoped to accepted depth outputs, but defer raw-storage rollout mechanics to the later replay epic
- do not create smoke-only child features; live validation belongs after implementation, not as its own plan
- keep Go as the live runtime path; Python remains offline-only
