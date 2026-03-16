# Implementation Module 1: Runtime Status HTTP Surface

## Scope

- Define the public runtime-status response contract and add the additive operator route in `services/market-state-api`.
- Cover handler success, unsupported-provider behavior, and route-level stability for `/healthz` plus `/api/market-state/*`.
- Exclude command startup/lifecycle wiring and runbook publication.

## Target Repo Areas

- `services/market-state-api`
- `services/market-state-api` tests

## Requirements

- Add `GET /api/runtime-status` without changing the existing handler behavior for `/healthz` or `/api/market-state/*`.
- Keep the public response deterministic and machine-readable for `BTC-USD` and `ETH-USD` only.
- Preserve shared feed-health states and reasons exactly as emitted by the internal runtime-health snapshot.
- Keep readiness, feed-health, connection posture, depth posture, timestamps, and reconnect counters explicit in the public projection.
- Return a clear unsupported response when the provider does not implement the runtime-status seam.

## Key Decisions

- Introduce public response types in `services/market-state-api` so the handler owns JSON shape while `cmd/market-state-api` keeps internal snapshot ownership.
- Add a narrow optional interface, separate from `Provider`, for runtime-status reads.
- Use the existing `Handler.Routes()` mux to register `GET /api/runtime-status` next to the current routes.
- Keep unsupported-provider behavior explicit and testable instead of forcing deterministic or non-Binance providers to implement runtime status.

## Unit Test Expectations

- The new route returns a stable response with both tracked symbols in deterministic order.
- Shared feed-health states and reasons survive JSON serialization unchanged.
- `/healthz` still returns only the process-health payload.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain unchanged.
- Unsupported runtime-status providers return the planned error status and body.

## Contract / Fixture / Replay Impacts

- This module adds one new public API surface but does not change the existing current-state schema family.
- No shared schema update is expected because the route is service-owned rather than a cross-language contract.
- Keep the response shape small, explicit, and easy to reuse in focused integration smoke.

## Summary

This module establishes the bounded public operator route and its JSON contract so later wiring and docs can depend on one explicit runtime-status surface without widening the current-state provider boundary.
