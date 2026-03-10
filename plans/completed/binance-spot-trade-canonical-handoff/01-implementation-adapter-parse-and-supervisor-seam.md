# Implementation: Adapter Parse And Supervisor Seam

## Module Requirements

- Confirm the Spot supervisor raw-frame seam can hand accepted `trade` frames to one bounded Binance trade parser path.
- Tighten native trade parsing only where needed to match settled Binance semantics for trade time, event time, aggressor side, and source symbol handling.
- Keep the feature scoped to `trade` only; `bookTicker` and depth work remain out of scope.

## Target Repo Areas

- `services/venue-binance`
- `plans/completed/binance-spot-ws-runtime-supervisor/`

## Unit Test Expectations

- native Binance trade payloads parse into the shared ingestion trade model
- fallback from trade time to event time is explicit and deterministic
- supervisor-fed accepted data frames can be consumed by the trade parser without creating a second lifecycle owner

## Summary

This module locks the adapter-side seam so later validation can prove Spot trade handoff uses the completed supervisor boundary instead of bypassing it.
