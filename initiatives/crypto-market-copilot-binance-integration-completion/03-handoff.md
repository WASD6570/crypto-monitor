# Binance Integration Completion Handoff

## Epic Queue

1. `plans/epics/binance-streaming-market-state-runtime-cutover/`
2. `plans/epics/binance-runtime-health-and-operator-observability/`
3. `plans/epics/binance-usdm-market-state-influence/`
4. `plans/epics/binance-environment-config-and-rollout-hardening/`
5. `plans/epics/binance-long-run-runtime-hardening/`

## Planning Waves

### Wave 1

- `binance-streaming-market-state-runtime-cutover`
- Why now: the current remaining gap is the bounded snapshot reader in `cmd/market-state-api`; every later slice should consume the finished runtime source rather than plan against a temporary local seam.

### Wave 2

- `binance-runtime-health-and-operator-observability`
- `binance-usdm-market-state-influence`
- Why parallel: both consume the streaming runtime but solve different problems; one is operator visibility, the other is market-state semantics.

### Wave 3

- `binance-environment-config-and-rollout-hardening`
- Why later: rollout defaults should follow the settled runtime and observability posture rather than invent their own behavior.

### Wave 4

- `binance-long-run-runtime-hardening`
- Why later: long-run/failure hardening should validate the final runtime shape after the streaming cutover, observability surface, USD-M semantics, and environment defaults are all clear.

## Epic Seeds

### `plans/epics/binance-streaming-market-state-runtime-cutover/`

- Problem statement: the dashboard is live-backed today, but `cmd/market-state-api` still reads Binance through a bounded on-demand snapshot seam instead of a sustained streaming runtime.
- In scope: replace the snapshot reader with a process-owned Spot runtime/read model, wire it into `market-state-api`, preserve warm-up honesty, and keep current routes/contracts stable.
- Out of scope: USD-M-driven regime changes, broad environment rollout, or frontend redesign.
- Target repo areas: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `tests/integration`
- Contract/fixture/parity/replay implications: current-state payload shape stays stable; replay-sensitive runtime assumptions must stay explicit.
- Likely validation shape: targeted Go tests, integration checks against the live runtime path, and same-origin API smoke.

### `plans/epics/binance-runtime-health-and-operator-observability/`

- Problem statement: current success/failure states are difficult to interpret because runtime health is mostly implicit and logs are sparse on the happy path.
- In scope: expose runtime warm-up, reconnect, stale, recovery, and rate-limit state through bounded status surfaces, docs, and operator-facing guidance.
- Out of scope: changing the main browser contract for market-state payloads unless refinement proves a tiny additive status seam is needed.
- Target repo areas: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `docs/runbooks`
- Contract/fixture/parity/replay implications: status outputs must remain machine-readable and should not break existing consumers.
- Likely validation shape: targeted runtime tests, status endpoint or payload checks, and operator-runbook verification.

### `plans/epics/binance-usdm-market-state-influence/`

- Problem statement: USD-M sensors exist, but it is still unresolved whether they should influence current-state and regime semantics or remain auxiliary context.
- In scope: settle the product/algorithm role of USD-M inputs, implement the bounded behavior, and preserve deterministic validation.
- Out of scope: adding new derivative venues, private futures endpoints, or unrelated dashboard redesign.
- Target repo areas: `services/feature-engine`, `services/regime-engine`, `services/market-state-api`, `services/venue-binance`, `tests/integration`, `tests/replay`
- Contract/fixture/parity/replay implications: highly replay-sensitive; any semantic changes need fixture and deterministic proof updates.
- Likely validation shape: targeted current-state/regime tests, replay determinism checks, and focused API verification.

### `plans/epics/binance-environment-config-and-rollout-hardening/`

- Problem statement: the current cutover is local-first, but the final Binance integration needs explicit environment defaults and rollout-safe startup behavior.
- In scope: define `local`, `dev`, and `prod` runtime defaults, config loading expectations, rollout notes, and startup/health behavior.
- Out of scope: unrelated platform deployment automation or non-Binance rollout work.
- Target repo areas: `configs/*`, `cmd/market-state-api`, `docker-compose.yml`, `docs/runbooks`
- Contract/fixture/parity/replay implications: config defaults are rollout-sensitive and should not silently alter replay or current-state semantics.
- Likely validation shape: config parsing tests, local compose proof, and runbook-driven startup verification.

### `plans/epics/binance-long-run-runtime-hardening/`

- Problem statement: finishing the integration requires confidence that the final runtime survives reconnects, stale periods, and repeated validation without semantic drift.
- In scope: long-run/failure-path checks, reconnect/rate-limit/staleness validation, and final replay/current-state regression coverage for the settled runtime.
- Out of scope: standalone smoke-only planning or unrelated observability platform work.
- Target repo areas: `tests/integration`, `tests/replay`, `services/venue-binance`, `services/market-state-api`
- Contract/fixture/parity/replay implications: this is the final confidence gate for runtime determinism and operator trust.
- Likely validation shape: repeated runtime tests, replay checks, focused live-path failure simulations, and post-cutover compose verification.

## Open Questions That Still Matter

- Should USD-M inputs change current-state/regime outputs immediately, or only cap/degrade them until more explicit weighting rules are proven?
- Should operator visibility land as a small new status surface, additive metadata on existing payloads, richer logs, or some combination?
- What environment-specific polling/reconnect defaults are acceptable for `dev` and `prod` without creating avoidable Binance pressure?
- How much of long-run hardening belongs in CI versus a documented manual/operator validation flow?
