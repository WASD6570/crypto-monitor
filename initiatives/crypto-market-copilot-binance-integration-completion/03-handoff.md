# Binance Integration Completion Handoff

## Refined Epic Queue

No actionable refined epic queue exists yet for Wave 2. The next items are initiative seeds that still need `program-refining` before they can enter `plans/epics/`.

## Execution State

- Initiative status: `in_progress`
- Completed prerequisite epic context: `plans/epics/binance-streaming-market-state-runtime-integration/` (historical reference only)
- Next recommended seed: `binance-runtime-health-and-operator-observability`
- Parallel-safe seed after that starts: `binance-usdm-market-state-influence`

| Item | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `binance-runtime-health-and-operator-observability` | `ready_to_refine` | Wave 1 complete | `binance-usdm-market-state-influence` | Run `program-refining` and materialize `plans/epics/binance-runtime-health-and-operator-observability/` | First Wave 2 seed and default next planning target |
| `binance-usdm-market-state-influence` | `ready_to_refine` | Wave 1 complete | `binance-runtime-health-and-operator-observability` | Refine in parallel when capacity allows and materialize `plans/epics/binance-usdm-market-state-influence/` | Second Wave 2 seed |
| `binance-environment-config-and-rollout-hardening` | `blocked` | Wave 2 decisions | - | Wait for Wave 2 planning and implementation outcomes before refining | Wave 3 seed |
| `binance-long-run-runtime-hardening` | `blocked` | Wave 2 and Wave 3 outcomes | - | Wait for runtime shape and rollout defaults to settle before refining | Wave 4 seed |

## Planning Waves

### Wave 1

- `binance-streaming-market-state-runtime-integration`
- Why now: the archived provider cutover still leaves `cmd/market-state-api` on a bounded snapshot seam; every later slice should consume the finished sustained runtime source rather than plan against a temporary local seam.

### Wave 2

- `binance-runtime-health-and-operator-observability`
- `binance-usdm-market-state-influence`
- Why parallel: both consume the sustained runtime integration but solve different problems; one is operator visibility, the other is market-state semantics.

### Wave 3

- `binance-environment-config-and-rollout-hardening`
- Why later: rollout defaults should follow the settled runtime and observability posture rather than invent their own behavior.

### Wave 4

- `binance-long-run-runtime-hardening`
- Why later: long-run/failure hardening should validate the final runtime shape after the streaming cutover, observability surface, USD-M semantics, and environment defaults are all clear.

## Initiative Seeds

### `binance-runtime-health-and-operator-observability`

- Problem statement: current success/failure states are difficult to interpret because runtime health is mostly implicit and logs are sparse on the happy path.
- In scope: expose runtime warm-up, reconnect, stale, recovery, and rate-limit state through bounded status surfaces, docs, and operator-facing guidance.
- Out of scope: changing the main browser contract for market-state payloads unless refinement proves a tiny additive status seam is needed.
- Target repo areas: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `docs/runbooks`
- Contract/fixture/parity/replay implications: status outputs must remain machine-readable and should not break existing consumers.
- Likely validation shape: targeted runtime tests, status endpoint or payload checks, and operator-runbook verification.

### `binance-usdm-market-state-influence`

- Problem statement: USD-M sensors exist, but it is still unresolved whether they should influence current-state and regime semantics or remain auxiliary context.
- In scope: settle the product/algorithm role of USD-M inputs, implement the bounded behavior, and preserve deterministic validation.
- Out of scope: adding new derivative venues, private futures endpoints, or unrelated dashboard redesign.
- Target repo areas: `services/feature-engine`, `services/regime-engine`, `services/market-state-api`, `services/venue-binance`, `tests/integration`, `tests/replay`
- Contract/fixture/parity/replay implications: highly replay-sensitive; any semantic changes need fixture and deterministic proof updates.
- Likely validation shape: targeted current-state/regime tests, replay determinism checks, and focused API verification.

### `binance-environment-config-and-rollout-hardening`

- Problem statement: the current cutover is local-first, but the final Binance integration needs explicit environment defaults and rollout-safe startup behavior.
- In scope: define `local`, `dev`, and `prod` runtime defaults, config loading expectations, rollout notes, and startup/health behavior.
- Out of scope: unrelated platform deployment automation or non-Binance rollout work.
- Target repo areas: `configs/*`, `cmd/market-state-api`, `docker-compose.yml`, `docs/runbooks`
- Contract/fixture/parity/replay implications: config defaults are rollout-sensitive and should not silently alter replay or current-state semantics.
- Likely validation shape: config parsing tests, local compose proof, and runbook-driven startup verification.

### `binance-long-run-runtime-hardening`

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
