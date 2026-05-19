# Implementation: Runtime Soak And Failure Paths

## Requirements And Scope

- Add high-signal deterministic coverage for the final live runtime posture.
- Exercise the command provider and API route where practical so the test proves actual runtime ownership boundaries.
- Keep tests local and deterministic by using `httptest` REST servers, local websocket servers, injected clocks, and test-local runtime config overrides.
- Do not rely on public Binance, Docker, wall-clock long soaks, Python, or browser execution for the required pass gate.

## Target Files

- `cmd/market-state-api/main_test.go`
- `cmd/market-state-api/runtime_health_test.go`
- `services/venue-binance/runtime_test.go`
- `services/venue-binance/spot_ws_supervisor_test.go`
- `services/venue-binance/spot_depth_recovery_test.go`
- `services/venue-binance/usdm_runtime_test.go`
- `services/venue-binance/usdm_open_interest_test.go`
- `tests/integration/binance_spot_depth_recovery_test.go`
- `tests/integration/binance_usdm_runtime_test.go`
- `tests/integration/binance_current_state_test.go`
- `tests/replay/binance_current_state_determinism_test.go`

## Failure Paths To Cover

| Path | Expected Behavior |
|---|---|
| Warm-up | `/api/runtime-status` returns both symbols as `NOT_READY`; `/healthz` stays process-only; current-state may be unavailable or partial without causing extra REST bursts. |
| Ready steady state | Spot observations publish in `BTC-USD`, `ETH-USD` order; runtime-status is `READY`; repeated reads are deterministic for the same accepted inputs and injected time. |
| Spot reconnect | Last accepted Spot observation stays stable until depth re-sync; runtime-status shows non-connected connection state and canonical degradation reasons. |
| Spot stale | Readiness remains `READY` after prior publish, but feed health becomes `STALE` with `message-stale` or `snapshot-stale` reasons. |
| Depth sequence gap and resync | Sequence gaps move depth status into a visible recovery state; successful replacement snapshot returns to synchronized. |
| Depth recovery rate limit | `depthStatus.state=rate-limit-blocked` and `feedHealth.reasons` includes `rate-limit`; no silent healthy status is emitted. |
| USD-M websocket reconnect or stale | Additive `usdmStatus.websocket` reports degraded or stale state without changing Spot readiness. |
| USD-M open-interest rate limit or stale | Additive `usdmStatus.openInterest` reports `rate-limit` or stale state and current-state cap behavior remains bounded. |
| Shutdown | `providerWithRuntime.Close(ctx)` stops USD-M and Spot owners without hanging; repeated close should be safe or explicitly tested to return a stable error if implementation keeps one-shot semantics. |

## Algorithm And Test Notes

- Prefer table-driven scenario scripts for command-level tests when the same local REST and websocket fixtures can drive warm-up, ready, degradation, and recovery.
- Use existing helpers such as `newSpotWebsocketTestServer`, `depthSnapshotPayload`, `waitFor`, and test-local runtime config overrides.
- Keep sleeps short and tied to injected config values; never encode production cooldowns as literal test waits.
- When testing repeated runs, execute the same scripted inputs twice and compare the runtime-status payload plus current-state output after normalizing generated times that are intentionally different.
- If failures expose current data races, add focused locking or snapshot-copy fixes in the owner that owns the mutable state.

## Unit And Integration Test Expectations

- `services/venue-binance` focused tests cover reconnect backoff, rate limits, stale evaluation, depth recovery, USD-M websocket health, and open-interest health.
- `cmd/market-state-api` tests cover the composed live provider, runtime-status route, current-state preservation, and provider close behavior.
- `tests/integration` coverage confirms current-state and canonical feed-health evidence still align with the runbooks.
- `tests/replay` coverage confirms deterministic current-state and Binance replay families remain stable for pinned fixture inputs.

## Replay And Determinism Impact

- This feature should not change replay schema shape or accepted fixture source identities.
- Replay validation remains required because current-state semantics and USD-M cap inputs are replay-sensitive.
- If implementation touches canonical events, fixtures, replay seeds, or schemas, update the affected fixture, schema, producer, and consumer validation in the same implementation slice.

## Summary For Next Agent

After `usdmStatus` exists, add deterministic command-level tests that script warm-up, ready, degraded, recovered, and shutdown states through the real provider and route boundary. Fix only bugs exposed by those tests.
