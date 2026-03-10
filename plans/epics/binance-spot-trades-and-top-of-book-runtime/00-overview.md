# Binance Spot Trades And Top Of Book Runtime

## Epic Summary

Materialize the first live Binance Spot runtime slice that keeps one websocket runtime alive for `trade` and `bookTicker`, emits adapter-scoped feed-health as a first-class output, and hands trade plus top-of-book candidates to canonical normalization using the Wave 1 contract seam without reopening symbol, timestamp, or source-ID policy.

This epic is the first live price-path unlock for the initiative. It should prove that BTC and ETH Spot data can enter the Go runtime with explicit provenance, degradation visibility, and bounded reconnect behavior before depth bootstrap and later API cutover work begin.

## In Scope

- Spot websocket runtime behavior for the selected `trade` and `bookTicker` streams
- Binance-specific ping/pong handling, proactive pre-24h reconnect, bounded backoff, and resubscribe behavior
- adapter-scoped feed-health outputs for connection state, stale-message detection, reconnect loops, and local clock degradation relevant to this stream set
- parsing and adapter handoff for Spot `trade` payloads using the locked Wave 1 identity, timestamp, and source-record rules
- parsing and adapter handoff for Spot `bookTicker` payloads using the same locked contract seam
- canonical symbol handling that stays asset-centric as `BTC-USD` and `ETH-USD` while preserving `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- config and integration-test updates needed to exercise this runtime slice locally

## Out Of Scope

- Spot depth snapshot bootstrap, buffered delta sequencing, gap repair, or `/api/v3/depth` recovery
- schema-family redesign or reopening the Wave 1 contract decisions
- raw append/replay rollout, USD-M sensors, or `services/market-state-api` live cutover
- frontend behavior changes or client-owned Binance runtime logic

## Target Repo Areas

- `services/venue-binance`
- `services/normalizer`
- `configs/local/ingestion.v1.json` and sibling environment configs when later planning requires parity
- `tests/integration`
- `tests/fixtures/events/binance` for stream-specific runtime fixtures as needed

## Validation Shape

- targeted runtime tests for websocket lifecycle behavior: bounded reconnect, resubscribe, heartbeat timeout handling, and proactive rollover before Binance disconnects the connection
- fixture-backed parser and normalization checks for Spot `trade` and Spot `bookTicker`
- direct live validation that shows feed-health and canonical handoff both move when the Spot runtime receives healthy and degraded inputs
- explicit checks that canonical symbols remain asset-centric and provenance fields remain visible on emitted outputs

## Major Constraints

- inherit Wave 1 contract-seam decisions; do not reopen canonical symbol, timestamp precedence, or source-record identity policy inside this epic
- keep feed-health first-class; degraded runtime states must be visible as machine-readable outputs, not only logs
- keep Spot runtime work strictly limited to `trade` and `bookTicker`; depth belongs to `plans/epics/binance-spot-depth-bootstrap-and-recovery/`
- preserve `exchangeTs`, `recvTs`, `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` through the adapter-to-normalizer handoff
- keep the live runtime in Go and within `services/venue-binance`; Python remains offline-only
- prefer minimal, bounded child features that let later planners validate lifecycle behavior separately from per-stream parsing work
