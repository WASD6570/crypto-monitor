# Binance Integration Completion Feature Map

## 1. `binance-streaming-market-state-runtime-cutover`

- Goal: replace the bounded REST snapshot reader in `cmd/market-state-api` with a long-lived Spot runtime/read-model that continuously feeds current-state assembly.
- Primary repo areas: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `tests/integration`
- Why it stands alone: it is the main gap between the current live cutover and a finished Binance runtime.

## 2. `binance-runtime-health-and-operator-observability`

- Goal: expose warm-up, reconnect, snapshot/recovery, stale, and rate-limit state clearly enough for operators and local debugging.
- Primary repo areas: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `docs/runbooks`
- Why it stands alone: runtime visibility is cross-cutting and should not be hidden inside market-state semantics work.

## 3. `binance-usdm-market-state-influence`

- Goal: settle whether USD-M context remains auxiliary or actively influences current-state/regime outputs, then implement that choice cleanly.
- Primary repo areas: `services/feature-engine`, `services/regime-engine`, `services/market-state-api`, `services/venue-binance`, `tests/integration`, `tests/replay`
- Why it stands alone: it changes product semantics, not just runtime plumbing.

## 4. `binance-environment-config-and-rollout-hardening`

- Goal: define stable `local`, `dev`, and `prod` config defaults, startup expectations, and rollout-safe operator behavior for the finished runtime.
- Primary repo areas: `configs/*`, `cmd/market-state-api`, `docs/runbooks`, `docker-compose.yml`
- Why it stands alone: environment defaults and rollout posture should stay separate from core runtime implementation.

## 5. `binance-long-run-runtime-hardening`

- Goal: prove the final runtime survives reconnects, rate limits, stale periods, and repeated replay/current-state checks without drifting semantically.
- Primary repo areas: `tests/integration`, `tests/replay`, `services/venue-binance`, `services/market-state-api`
- Why it stands alone: hardening and failure-proof validation should attach to the finished runtime shape rather than be spread across earlier slices.

## Cross-Cutting Tracks

- `stable-current-state-contract`: keep the dashboard-facing JSON shape stable while runtime sourcing evolves
- `machine-readable-runtime-status`: make runtime/debug state visible in structured outputs, not log-only
- `replay-and-determinism-guardrails`: preserve deterministic audit behavior as the runtime becomes long-lived and stateful
- `rate-limit-and-reconnect-safety`: keep Binance public endpoint usage bounded and explicit across environments
