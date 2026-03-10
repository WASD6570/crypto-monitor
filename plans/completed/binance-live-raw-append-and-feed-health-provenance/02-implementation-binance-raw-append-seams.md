# Implementation: Binance Raw Append Seams

## Module Requirements

- Add the minimum Binance-owned seam helpers required to turn completed Spot and USD-M canonical outputs into shared raw append entries with the correct connection/session/degraded-feed context.
- Reuse already-settled runtime and supervisor ownership rather than introducing a second provenance source.
- Keep Spot websocket, Spot depth recovery, USD-M websocket, and USD-M REST provenance explicit and distinct where they already differ in live operation.

## Target Repo Areas

- `services/venue-binance`
- `services/venue-binance/spot_ws_supervisor.go`
- `services/venue-binance/spot_depth_recovery.go`
- `services/venue-binance/usdm_runtime.go`
- `services/venue-binance/usdm_open_interest.go`

## Key Decisions

- Build venue-owned helper functions that prepare `ingestion.RawWriteContext` and call the shared raw append builders instead of embedding raw entry logic directly into integration tests.
- Keep Spot depth health separate from generic Spot websocket health where source record identity and degradation origin already differ.
- Preserve USD-M mixed-surface separation so websocket and REST raw append entries remain distinguishable to later replay and operator audit tools.
- Prefer helper coverage per family or per surface over one oversized catch-all Binance raw append function.

## Data And Provenance Notes

- Spot websocket families should inherit connection/session references from the completed Spot supervisor boundary.
- Spot depth recovery families should carry the dedicated depth health source record and any degraded-feed linkage emitted during recovery.
- USD-M websocket-derived families should keep websocket session provenance; REST-polled open-interest should carry a stable polling/session provenance without pretending to be websocket-originated.
- Degraded feed-health raw entries should preserve the source record ID that downstream audit or replay checks will use to tie degraded conditions back to the originating live surface.

## Unit / Integration Test Expectations

- Binance seam helpers produce raw entries for at least one family on each live surface:
  - Spot trade or top-of-book
  - Spot depth or Spot depth feed-health
  - USD-M funding or mark-index
  - USD-M open-interest
- connection/session references are present and stable for repeated identical inputs
- degraded feed-health retains explicit `degradedFeedRef` when applicable
- no helper reopens canonical payload content already settled in completed feature archives

## Contract / Fixture / Replay Notes

- Reuse completed Binance fixtures where practical; add only the minimum raw-append-specific fixtures or expectations needed to prove provenance.
- Keep helper outputs replay-ready but do not implement replay-engine logic here.

## Summary

This module adds venue-owned raw append seams so completed Binance live families can reach the shared raw writer with the correct transport and degradation provenance intact.
