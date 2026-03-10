# Implementation: Integration And Health Proof

## Module Requirements

- Add targeted integration coverage for Binance USD-M open-interest happy-path normalization.
- Prove degraded timestamp fallback when REST exchange time is missing.
- Prove per-symbol poll freshness and rate-limit degradation remain machine-visible through canonical feed-health output.

## Target Repo Areas

- `tests/integration`
- `services/venue-binance`
- `services/normalizer`

## Unit Test Expectations

- integration fixture normalization matches expected canonical open-interest output
- feed-health goes `STALE` after missed polls and `DEGRADED` with explicit rate-limit reason
- no live network access is required for the test matrix

## Summary

This module proves the REST poller, normalizer handoff, and feed-health outputs work together as one bounded feature without live Binance access.
