# Refinement Map: Binance USD-M Context Sensors

## Current Status

This Wave 2 epic is newly materialized from the Binance live initiative and is still too broad for direct `feature-planning` as one slice.

It should be decomposed into child features that separate websocket sensor runtime work, REST polling semantics, and the mixed-surface validation needed to prove they behave as one coherent context feed.

## What Is Already Covered

- `plans/completed/canonical-contracts-and-fixtures/00-overview.md` already fixes the shared contract families, canonical symbol vocabulary, and timestamp/degraded-state expectations.
- `plans/completed/market-ingestion-and-feed-health/00-overview.md` already fixes the feed-health vocabulary and the rule that stale or degraded inputs must stay machine-visible.
- `plans/epics/binance-live-contract-seams-and-fixtures/00-overview.md` already establishes that Binance live work must preserve `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs`.
- `services/venue-binance/README.md` and `configs/local/ingestion.v1.json` already reserve the MVP USD-M inventory: funding-rate, open-interest, mark-index, and liquidation.
- Existing schemas already exist for `funding-rate`, `mark-index`, `open-interest-snapshot`, and `liquidation-print`.
- Existing Binance fixtures already seed mark/index and degraded timestamp coverage, even though the USD-M surface is not yet complete.

## What Remains

- split the websocket-derived USD-M work into a bounded child feature that covers `markPrice@1s`, funding extraction, index/mark extraction, and `forceOrder`
- split REST-polled `openInterest` into its own bounded child feature so cadence, freshness, and provenance can be planned without entangling websocket runtime work
- define the validation and fixture slice that proves WS plus REST sensors coexist without timestamp ambiguity or hidden degradation
- preserve Wave 1 contract seam decisions rather than reopening source-record ID, canonical symbol, or timestamp policy in later child plans

## What Should Not Be Reopened

- asset-centric canonical symbols remain `BTC-USD` and `ETH-USD`
- provenance remains explicit via `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- shared event family ownership remains under `schemas/json/events`
- generic feed-health vocabulary and timestamp fallback rules remain inherited from completed and Wave 1 work

## Refinement Waves

### Wave 2A

- `binance-usdm-mark-funding-index-and-liquidation-runtime`
- Why first: most derivatives context value comes from the websocket path, and it establishes the primary USD-M runtime handoff before REST polling is layered in.

### Wave 2B

- `binance-usdm-open-interest-rest-polling`
- Why next: `openInterest` is operationally distinct because it is REST-polled, freshness-sensitive, and rate-limit-sensitive even though it shares the same canonical symbol and provenance policy.

### Wave 2C

- `binance-usdm-context-sensor-fixtures-and-integration`
- Why last: mixed-surface fixture and integration planning should validate the settled websocket and REST semantics instead of inventing them early.

## Notes For Future Planning

- Keep the websocket and REST acquisition modes explicit in every child plan; do not flatten them into one generic "sensor" loop.
- Child plans may reuse existing event schemas, but they should treat schema edits as exceptional and justified by a concrete Binance gap.
- `openInterest` planning must call out polling cadence, freshness ceiling, and degraded behavior separately from websocket staleness.
- Funding and mark/index extraction should stay coupled because Binance exposes them together through `markPrice@1s` and they share timestamp/provenance handling.
- `forceOrder` planning should stay with the websocket runtime slice because liquidation visibility depends on the same USD-M connection and health behavior.
