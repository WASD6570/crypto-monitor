# Implementation: Fixtures And Integration Proof

## Module Requirements

- Expand or tighten Binance Spot trade fixtures so at minimum happy-path and degraded timestamp behavior are exercised from the trade handoff perspective.
- Add targeted integration coverage that proves the completed Spot supervisor can feed accepted trade frames into canonical normalization.
- Add duplicate-sensitive proof for repeated Binance Spot trade inputs where the feature materially touches raw append or replay-sensitive identity behavior.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/integration`
- `tests/fixtures/manifest.v1.json`

## Unit Test Expectations

- fixture-backed integration proves accepted Spot trade frames normalize into canonical `market-trade` events
- degraded timestamp coverage remains explicit and machine-visible
- repeated Spot trade inputs preserve stable source identity and any expected raw duplicate behavior

## Summary

This module closes the proof gap between parser unit tests and supervisor-backed integration so the trade slice is ready for later direct live Spot runtime validation.
