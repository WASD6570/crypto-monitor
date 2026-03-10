# Implementation: Canonical Trade Handoff

## Module Requirements

- Wire or tighten the Binance Spot trade handoff into `services/normalizer` using explicit Spot metadata.
- Preserve stable source-record identity, provenance fields, and raw append compatibility for accepted Spot trades.
- Keep timestamp degradation behavior aligned with the existing strict canonical timestamp policy rather than adding Binance-specific fallback rules beyond trade-time selection.

## Target Repo Areas

- `services/normalizer`
- `libs/go/ingestion`
- `services/venue-binance`

## Unit Test Expectations

- canonical Spot trade output keeps `symbol`, `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType=spot`
- repeated accepted Binance trades keep stable `sourceRecordId` values and duplicate-sensitive raw facts
- degraded timestamp handling remains deterministic when the selected exchange time is implausible

## Summary

This module converts accepted Binance Spot trades into canonical outputs without losing auditability or reopening shared timestamp and identity rules.
