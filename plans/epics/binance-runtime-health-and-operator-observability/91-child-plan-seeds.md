# Child Plan Seeds

## `binance-runtime-health-snapshot-owner`

- Outcome: add one command-owned runtime-health snapshot that merges supervisor state, depth recovery posture, and read-model readiness into stable per-symbol operator status for `BTC-USD` and `ETH-USD`
- Primary repo area: `cmd/market-state-api`, `services/venue-binance`
- Dependencies: completed `binance-spot-ws-runtime-supervisor`, `binance-spot-runtime-read-model-owner`, and `binance-market-state-live-reader-cutover`
- Validation shape: targeted Go tests for warm-up, reconnect, stale, rate-limit, recovery, and deterministic status mapping
- Why it stands alone: this settles the machine-readable runtime-health contract before any endpoint or docs commit to it

## `binance-runtime-status-endpoint-and-ops-handoff`

- Outcome: expose the runtime-health snapshot through one additive operator-facing API/status surface, keep `/healthz` process-only, and update runbooks plus focused API smoke
- Primary repo area: `services/market-state-api`, `cmd/market-state-api`, `docs/runbooks`, `tests/integration`
- Dependencies: `binance-runtime-health-snapshot-owner`
- Validation shape: handler and integration tests for the status surface, direct API smoke proving `/healthz` separation, and runbook verification using shared feed-health vocabulary
- Why it stands alone: it is the consumer and operator exposure slice, so it should stay separate from lower-level status aggregation logic
