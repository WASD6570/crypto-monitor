# Binance Live Runtime Soak And Failure Hardening

## Ordered Implementation Plan

1. Add additive USD-M runtime visibility to `GET /api/runtime-status` so the existing operator route covers both the Spot runtime and the USD-M websocket/open-interest surfaces that already influence current state.
2. Harden the command-owned runtime through deterministic failure-sequence coverage for warm-up, reconnect, stale, depth resync, rate-limit, repeated reads, and shutdown paths.
3. Update the runtime-soak and live-boundary handoff docs so operators know the required local checks, the optional public Binance live check, and the Compose proof when Docker is available.
4. Run the validation matrix, record evidence in `plans/binance-live-runtime-soak-and-failure-hardening/testing-report.md`, then leave implementation paused for feature-testing and archive handoff.

## Outcome

Prove the sustained Binance Spot plus USD-M runtime can survive repeated failure paths without hiding degraded state, dropping deterministic current-state semantics, or depending on browser-side Binance logic, Python, mocks, or fixture-backed live runtime behavior.

## Requirements

- Keep tracked symbols fixed to `BTC-USD` and `ETH-USD`.
- Keep Go as the live runtime owner; Python remains offline-only and must not become a runtime dependency.
- Keep `/healthz` process-health only.
- Keep `GET /api/runtime-status` as the bounded operator runtime-health route.
- Keep `GET /api/market-state/global` and `GET /api/market-state/:symbol` backward-compatible consumer read routes.
- Make runtime-status changes additive only; do not remove or rename existing fields.
- Preserve existing current-state semantics, including last accepted Spot observation behavior during reconnect and conservative USD-M cap behavior.
- Keep degraded feeds, sequence gaps, reconnect loops, stale messages, depth recovery, open-interest rate limits, warm-up, and shutdown behavior machine-visible or directly validated.
- Do not add Spot trade-flow, Spot liquidity, USD-M indicator enrichment, alerting, or dashboard behavior in this feature.

## Target Repo Areas

- `services/venue-binance/runtime.go`
- `services/venue-binance/spot_ws_supervisor.go`
- `services/venue-binance/spot_depth_recovery.go`
- `services/venue-binance/usdm_runtime.go`
- `services/venue-binance/usdm_open_interest.go`
- `cmd/market-state-api/runtime_health.go`
- `cmd/market-state-api/live_provider.go`
- `cmd/market-state-api/spot_runtime_owner.go`
- `cmd/market-state-api/usdm_influence_owner.go`
- `services/market-state-api/api.go`
- `services/market-state-api/README.md`
- `docs/runbooks/binance-compose-rollout.md`
- `docs/runbooks/ingestion-feed-health-ops.md`
- `docs/runbooks/degraded-feed-investigation.md`
- `docs/runbooks/binance-runtime-soak-and-failure-check.md`
- `tests/integration`
- `tests/replay`

## ASCII Flow

```text
Binance public Spot WS + REST depth
  -> services/venue-binance Spot supervisor + depth recovery
  -> cmd/market-state-api Spot runtime owner
  -> current-state Spot observations
  -> /api/runtime-status Spot status

Binance public USD-M WS + REST open interest
  -> services/venue-binance USDM runtime + open-interest poller
  -> cmd/market-state-api USD-M influence owner
  -> conservative current-state cap inputs
  -> /api/runtime-status additive usdmStatus

Operator and consumer boundaries
  -> /healthz: process only
  -> /api/runtime-status: runtime health for BTC-USD and ETH-USD
  -> /api/market-state/*: consumer current-state reads
```

## Design Notes

### Runtime-status posture

- Preserve every existing runtime-status field and add a small `usdmStatus` object per symbol.
- Use existing USD-M runtime and open-interest health calculations rather than inventing a second policy in the HTTP handler.
- Keep per-symbol ordering deterministic: `BTC-USD`, then `ETH-USD`.
- Keep readiness driven by Spot publishability because Spot remains the current-state price source; USD-M status is auxiliary but machine-visible.

### Failure-sequence posture

- Prefer deterministic local websocket and REST test servers over real-time sleeps or public Binance calls.
- Drive repeated accepted-input sequences through the command provider and API route, not just isolated helper functions.
- Treat readiness, feed health, depth status, USD-M websocket status, open-interest poll status, and current-state output as the core invariants.
- If tests expose a shutdown, locking, or status-propagation bug, fix the smallest owning path instead of refactoring the runtime.

### Live-boundary posture

- Required validation stays local and deterministic.
- Optional live validation may use public Binance endpoints only and must be explicitly gated by `BINANCE_LIVE_VALIDATION=1`.
- Compose proof remains optional in this WSL environment because Docker is unavailable here; the runbook must still document the exact command for hosts with Docker.

## Acceptance

- `GET /api/runtime-status` continues to return the existing Spot fields and additionally reports USD-M websocket plus open-interest health for `BTC-USD` and `ETH-USD`.
- Warm-up remains `readiness=NOT_READY`; readable but degraded runtime remains `readiness=READY` with explicit feed-health reasons.
- Reconnect, stale, sequence-gap, rate-limit, depth-resync, and USD-M degraded states remain machine-visible through runtime status or canonical current-state inputs.
- Repeated deterministic failure sequences produce stable runtime-status ordering and current-state semantics.
- Provider shutdown closes Spot and USD-M owners without hanging and without leaving `/healthz` semantics overloaded.
- Required validation commands in `04-testing.md` pass or any unavailable optional live/Compose checks are recorded explicitly.

## Archive Intent

- Keep this feature active under `plans/binance-live-runtime-soak-and-failure-hardening/` while implementation and validation are in progress.
- After a passing `feature-testing` run, move the full directory and `testing-report.md` to `plans/completed/binance-live-runtime-soak-and-failure-hardening/`.
