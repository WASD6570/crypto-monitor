# Implementation: Fixture Corpus And Runbooks

## Module Requirements

- Expand Binance USD-M fixture coverage where mixed-surface validation still has a gap.
- Update fixture manifest entries if new Binance USD-M scenarios are added.
- Update runbooks so the shared health vocabulary includes the settled Binance USD-M REST reason set and mixed-surface verification notes.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`
- `docs/runbooks`

## Unit Test Expectations

- fixture manifest remains consistent with any new Binance USD-M fixtures
- runbook alignment tests include the mixed-surface vocabulary and any new degradation reasons

## Summary

This module makes the proof surface legible to later agents and operators before the integration tests consume it.
