# Implementation Module 2: Command-Owned Snapshot Owner And Lifecycle

## Scope

- Plan the owner that builds and serves the runtime-health snapshot inside `cmd/market-state-api`.
- Cover startup, snapshot refresh/update triggers, concurrent snapshot reads, and shutdown behavior.
- Exclude HTTP handlers, public endpoint wiring, and runbook publication.

## Target Repo Areas

- `cmd/market-state-api`
- `services/venue-binance`
- `cmd/market-state-api` tests and any focused integration test file needed for command-level status proof

## Requirements

- Start alongside the sustained Spot runtime owner and stop cleanly with the command lifecycle.
- Read existing runtime state without adding new Binance network paths or per-request polling.
- Keep snapshot reads concurrency-safe and cheap for the later status endpoint feature.
- Preserve the current market-state provider behavior while the snapshot owner runs in parallel as an internal seam.
- Keep failure handling explicit: if part of the underlying runtime state is temporarily unavailable, surface that honestly in the snapshot rather than synthesizing healthy defaults.

## Key Decisions

- Keep the owner command-local because the command already coordinates runtime startup, shutdown, and provider wiring.
- Prefer pull-based snapshot assembly from the sustained runtime seam when it keeps state authoritative and testable; introduce cached aggregation only if required for consistency or concurrency safety.
- Reuse the completed supervisor and depth-recovery health outputs as source signals; add narrow helpers there only when command consumption would otherwise duplicate logic.
- Keep the owner separate from handler code so the later endpoint feature can consume the same snapshot without redoing lifecycle logic.

## Unit Test Expectations

- Startup with no observations still returns deterministic snapshot entries for both tracked symbols.
- Healthy progression after accepted runtime data updates readiness and timestamps without losing stable ordering.
- Reconnect-loop, stale, and rate-limit states appear in snapshot reads after source-state changes.
- Concurrent snapshot reads during updates do not race or return partially built symbol sets.
- Shutdown leaves no hanging goroutines or blocked snapshot reads.

## Contract / Fixture / Replay Impacts

- No new public contract or fixture family is required in this slice.
- Focus test coverage on deterministic status propagation and command-lifecycle behavior rather than live exchange smoke.
- If command-level fixture hooks are added, keep them narrow and aligned with existing runtime test helpers.

## Summary

This module turns the snapshot contract into a reusable command-owned capability, so the next feature can expose runtime health through an API surface without mixing handler concerns into runtime orchestration.
