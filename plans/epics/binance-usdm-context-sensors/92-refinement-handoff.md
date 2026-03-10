# Refinement Handoff: Binance USD-M Context Sensors

## Next Recommended Child Feature

- `binance-usdm-mark-funding-index-and-liquidation-runtime`

## Why This Is Next

- it captures the highest-value live derivatives context first through the websocket surface already reserved by the initiative
- it depends only on inherited Wave 1 contract seams, not on new schema design or unresolved REST cadence choices
- it gives later planning a concrete websocket baseline before mixed-surface integration and REST freshness proof are added

## Safe Parallel Planning

- `binance-usdm-open-interest-rest-polling` can be planned in parallel with the websocket child feature because both inherit the same Wave 1 contract rules but touch different acquisition modes
- `binance-usdm-context-sensor-fixtures-and-integration` should wait until the first two child features settle their concrete semantics

## Blocked Until

- `binance-usdm-context-sensor-fixtures-and-integration` is blocked on the websocket child feature defining accepted `markPrice@1s` and `forceOrder` fixture cases
- `binance-usdm-context-sensor-fixtures-and-integration` is blocked on the REST child feature defining polling cadence, freshness ceilings, and degraded timestamp expectations for `openInterest`

## What Already Exists

- `plans/completed/canonical-contracts-and-fixtures/00-overview.md` already defines the shared event family and fixture foundation
- `plans/completed/market-ingestion-and-feed-health/00-overview.md` already defines the health vocabulary and explicit degradation posture
- `plans/epics/binance-live-contract-seams-and-fixtures/00-overview.md` already defines the Binance-specific contract seam expectations this epic must inherit
- `services/venue-binance/README.md` and `configs/local/ingestion.v1.json` already reserve the USD-M sensor inventory and operational defaults
- existing schemas and starter fixtures already cover part of the mark/index path, so later plans should extend them instead of treating the surface as greenfield

## Assumptions To Preserve

- canonical downstream symbols stay `BTC-USD` and `ETH-USD`
- provenance stays explicit through `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- `recvTs` stays mandatory for every accepted websocket event and REST poll sample
- websocket-derived and REST-derived freshness semantics remain distinct and explicit
- this epic inherits Wave 1 contract seam decisions rather than reopening them

## Recommended Follow-On After This Child Feature

1. `binance-usdm-open-interest-rest-polling`
2. `binance-usdm-context-sensor-fixtures-and-integration`
