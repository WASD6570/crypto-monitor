# Implementation Module 2: Command Provider Wiring And Live Route Proof

## Scope

- Wire `cmd/market-state-api` live provider ownership into the new runtime-status route through the existing runtime-health snapshot seam.
- Cover command-level route proof for warm-up, degraded, stale, recovery, and rate-limit visibility.
- Exclude runbook prose changes.

## Target Repo Areas

- `cmd/market-state-api`
- `services/market-state-api`
- `cmd/market-state-api` tests
- `tests/integration` only if focused API smoke is needed beyond command-level proof

## Requirements

- Reuse the existing `providerWithRuntime` and command-owned runtime snapshot owner; do not add new Binance network paths or per-request snapshot polling.
- Keep the runtime-status route cheap and concurrency-safe by reading the settled internal snapshot seam.
- Preserve the current market-state provider behavior while adding the new route in parallel.
- Prove warm-up, reconnect, stale, recovery, and rate-limit status can be observed through the route while `/healthz` stays process-only.
- Keep lifecycle behavior explicit: startup without publishable observations should still expose deterministic not-ready symbol entries.

## Key Decisions

- Keep runtime-status reads command-local because the command already owns runtime startup, shutdown, and provider lifecycle.
- Bridge the handler to the existing `RuntimeHealthSnapshot(ctx, now)` seam instead of duplicating status assembly logic in `services/market-state-api`.
- Extend existing command route tests to cover the new endpoint alongside the current-state routes.
- Prefer command-level smoke over broader live integration unless the handler boundary needs additional proof.

## Unit Test Expectations

- A live provider-backed handler serves `GET /api/runtime-status` after startup.
- Warm-up returns deterministic `NOT_READY` entries before publishable observations arrive.
- Degraded and stale runtime transitions preserve canonical reasons through the public route.
- Runtime recovery clears no-longer-active reasons and updates time-relative fields consistently.
- Existing current-state route tests remain green with no extra Binance REST pressure.

## Contract / Fixture / Replay Impacts

- No replay contract changes are expected in this module.
- Reuse existing runtime and websocket test harnesses for deterministic status transitions.
- Add focused command fixtures only if they improve route-level proof for rate-limit or stale timing.

## Summary

This module makes the additive route real in the live command path by reusing the settled runtime-health seam and proving operators can read accurate runtime posture without changing the current market-state contract.
