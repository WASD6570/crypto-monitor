# Implementation: REST Parser And Poll State

## Module Requirements

- Add a Binance REST parser for USD-M `openInterest` payloads.
- Add explicit request construction for tracked BTC/ETH perpetual symbols.
- Add a poll-state helper that tracks due polls, successful samples, and rate-limit posture per symbol.
- Keep the module scoped to `services/venue-binance` and config plumbing needed to support explicit cadence.

## Target Repo Areas

- `services/venue-binance`
- `libs/go/ingestion`
- `configs/local/ingestion.v1.json`

## Unit Test Expectations

- parser accepts happy payloads and preserves source symbol and timestamps
- parser keeps missing exchange time degradable instead of dropping the sample
- poll state marks symbols due on schedule and degrades on stale or rate-limited polling

## Summary

This module creates the REST-side runtime seam so later normalization can consume explicit open-interest samples and feed-health inputs without hiding cadence or rate-limit behavior.
