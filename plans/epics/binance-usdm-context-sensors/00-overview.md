# Binance USD-M Context Sensors

## Epic Summary

Refine the Binance USD-M derivatives context slice into bounded child features that add live `markPrice@1s`, funding/index extraction, `forceOrder`, and REST-polled `openInterest` under the existing Binance live initiative.

This epic inherits the Wave 1 contract seam decisions for canonical symbols, timestamp handling, source-record identity, and fixture vocabulary. It does not reopen shared schema-family or symbol-policy debates.

## In Scope

- Binance USD-M websocket handling for `markPrice@1s` across BTC and ETH
- extraction of canonical `funding-rate` and `mark-index` events from the mark price payload surface
- Binance USD-M websocket handling for `forceOrder` liquidation visibility
- REST-polled `openInterest` collection with explicit freshness and degradation semantics
- mixed WS plus REST feed-health behavior for the USD-M context sensor surface
- canonical handoff that keeps symbols asset-centric while preserving `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- fixture, validation, and planning boundaries needed so later runtime work can proceed without reopening Wave 1 decisions

## Out Of Scope

- private Binance futures endpoints, order entry, or account/user-data streams
- Spot runtime work, Spot depth bootstrap, or multi-venue derivatives rollout
- new shared schema-family design unless a concrete gap is proven during later feature planning
- raw storage, replay implementation, or market-state API cutover
- active implementation-ready feature plans under `plans/`

## Target Repo Areas

- `services/venue-binance`
- `services/normalizer`
- `configs/local/ingestion.v1.json`
- `tests/fixtures/events/binance`
- `tests/integration`
- `schemas/json/events` only if later planning proves a concrete gap rather than adapter misuse

## Validation Shape

- parser and normalization fixtures for `markPrice@1s`, funding/index extraction, `forceOrder`, and `openInterest`
- targeted integration checks for mixed WS plus REST freshness, degradation, and provenance preservation
- focused validation that canonical outputs keep `sourceSymbol`, `quoteCurrency`, and `marketType` explicit while symbols remain `BTC-USD` and `ETH-USD`
- polling and staleness checks that prove REST-derived open interest does not silently masquerade as websocket-timed data

## Major Constraints

- Keep canonical symbols asset-centric and provenance explicit through `sourceSymbol`, `quoteCurrency`, and `marketType`.
- Treat the surface as intentionally mixed: websocket is authoritative for mark/index, funding, and liquidation; REST is authoritative for open interest until a better venue-backed source is chosen.
- Preserve `exchangeTs` and `recvTs` on every accepted sensor sample, with degraded reasons when exchange time is missing, implausible, or REST-only.
- Inherit Wave 1 source-record ID and timestamp decisions instead of redefining them inside this epic.
- Respect existing health thresholds and reconnect defaults from `configs/local/ingestion.v1.json` unless later feature planning proves a bounded need to tune them.
- Keep all live runtime work in Go-owned services; Python remains offline-only.
