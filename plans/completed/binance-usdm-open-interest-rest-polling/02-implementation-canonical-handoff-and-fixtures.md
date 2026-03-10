# Implementation: Canonical Handoff And Fixtures

## Module Requirements

- Add shared Go normalization for `open-interest-snapshot`.
- Extend the normalizer raw-append boundary for canonical open-interest events.
- Add or update Binance fixtures so happy and degraded timestamp cases are covered explicitly.
- Update shared schema and fixture expectations only as required by the concrete Binance payload semantics.

## Target Repo Areas

- `libs/go/ingestion`
- `services/normalizer`
- `schemas/json/events`
- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`

## Unit Test Expectations

- canonical normalization preserves `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- missing or skewed exchange time degrades deterministically to `recvTs`
- raw append entries partition open-interest separately and keep stable source-record identity

## Summary

This module turns parsed REST samples into canonical events and deterministic fixtures so later replay and mixed-surface integration work inherit one stable open-interest contract.
