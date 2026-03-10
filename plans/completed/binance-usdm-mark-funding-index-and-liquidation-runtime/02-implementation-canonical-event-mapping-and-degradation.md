# Implementation Module: Canonical Event Mapping And Degradation

## Scope

- Parse Binance USD-M `markPrice@1s` payloads into canonical `funding-rate` and `mark-index` candidates.
- Parse Binance USD-M `forceOrder` payloads into canonical `liquidation-print` candidates.
- Normalize the resulting messages through shared Go ingestion helpers and `services/normalizer` while preserving inherited symbol, provenance, source-record ID, and timestamp rules.

## Target Repo Areas

- `services/venue-binance/usdm_mark_price.go`
- `services/venue-binance/usdm_mark_price_test.go`
- `services/venue-binance/usdm_force_order.go`
- `services/venue-binance/usdm_force_order_test.go`
- `libs/go/ingestion/derivatives_normalization.go`
- `libs/go/ingestion/derivatives_normalization_test.go`
- `services/normalizer/service.go`
- `services/normalizer/service_test.go`

## Requirements

- Map Binance USD-M source symbols such as `BTCUSDT` and `ETHUSDT` to canonical `BTC-USD` and `ETH-USD` while preserving `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType=perpetual`.
- Preserve `exchangeTs` and `recvTs` on every accepted message and let strict timestamp policy mark degraded fallbacks when exchange timestamps are missing, invalid, or implausibly skewed.
- Reuse the existing `funding-rate`, `mark-index`, and `liquidation-print` schemas rather than redefining canonical shapes.
- Inherit the Wave 1 source-record ID patterns for mark/index, funding, and liquidation outputs instead of inventing new prefixes or entropy.
- Reject unsupported symbols or malformed payloads explicitly in adapter tests.

## Key Decisions

- `markPrice@1s` stays coupled: one accepted native payload yields two canonical candidates, `funding-rate` and `mark-index`, because the venue provides those values on the same timestamp and provenance surface.
- `markPrice@1s` exchange time should come from the venue event time field selected by the inherited Binance time policy; implementation must not substitute `recvTs` unless strict timestamp resolution degrades it.
- `forceOrder` should use the liquidation event/order timestamp selected by the inherited Binance time policy, not socket receive time, for event identity and canonical event time when valid.
- Canonical downstream symbols remain asset-centric; source-specific futures identity stays visible through `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`.
- If the inherited Wave 1 identity helper for USD-M source-record IDs is not present when implementation begins, stop and align with that dependency before coding this module rather than encoding a competing ID rule here.

## Unit Test Expectations

- Happy-path `markPrice@1s` parsing emits one funding candidate and one mark-index candidate with the same provenance and exchange timestamp.
- Happy-path `forceOrder` parsing emits one liquidation candidate with the expected symbol mapping and canonical metadata.
- Missing or invalid exchange timestamps degrade timestamp status deterministically without losing `recvTs`.
- Unsupported source symbols are rejected instead of silently normalized.
- Normalizer tests prove canonical events preserve `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` for all three event families.
- Duplicate raw payloads produce stable source-record IDs.

## Contract / Fixture / Replay Impacts

- Expected schema impact is none; reuse existing event schemas unless implementation proves a concrete missing required field.
- Add Binance fixture pairs for:
  - happy `markPrice@1s` -> `funding-rate`
  - happy `markPrice@1s` -> `mark-index`
  - degraded-timestamp `markPrice@1s`
  - happy `forceOrder` -> `liquidation-print`
  - duplicate/re-delivery `forceOrder` or `markPrice@1s` proving stable source-record IDs
- Replay-sensitive acceptance should stay deterministic because canonical IDs, timestamps, and provenance are fixed by inherited seam rules.

## Summary

This module converts the two USD-M websocket payload families into canonical derivatives context events while preserving the repo's existing symbol, provenance, and degraded-time rules and explicitly blocking any attempt to smuggle in REST open-interest behavior.
