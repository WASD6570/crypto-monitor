# Implementation Module: Binance Runtime Loop

## Requirements And Scope

- Extend `AdapterLoopState` with the remaining explicit mutators needed by a future loop:
  - mark/clear sequence gap
  - increment/reset reconnect count
  - increment/reset resync count
  - prune snapshot recovery history to the active one-minute window when evaluating rate limits
- Add one thin harness/driver type that owns `Runtime` plus `AdapterLoopState` and exposes a small decision method.
- Keep all behavior side-effect free except for state mutation on the loop-state struct itself.

## Target Repo Areas

- `services/venue-binance/runtime.go`
- `services/venue-binance/runtime_test.go`

## Key Decisions

- Do not add goroutines, channels, or sockets in this slice.
- Keep state transitions explicit rather than hiding them in generic event reducers.
- Prefer immutable decision outputs and mutable state inputs.

## Unit Test Expectations

- Each mutator updates exactly one logical part of the state.
- Reconnect/resync reset behavior is deterministic.
- Sequence-gap recovery is visible in the emitted decision.
- Snapshot recovery history pruning is deterministic under synthetic time.
- The harness returns the same health decision as the lower-level runtime helpers.

## Contract / Fixture / Replay Impacts

- No new contracts are required.
- This slice should improve the ability to drive future replay and integration tests from a single loop-state surface.

## Summary

This module finishes the Binance-local state-machine surface so future runtime code can update state incrementally and ask for one shared health decision without rebuilding input structs by hand.
