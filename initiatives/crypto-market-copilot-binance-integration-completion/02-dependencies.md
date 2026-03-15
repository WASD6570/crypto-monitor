# Binance Integration Completion Dependencies

## Suggested Order

1. `binance-streaming-market-state-runtime-integration`
2. `binance-runtime-health-and-operator-observability`
3. `binance-usdm-market-state-influence`
4. `binance-environment-config-and-rollout-hardening`
5. `binance-long-run-runtime-hardening`

## Dependency Notes

### `binance-streaming-market-state-runtime-integration`

- Depends on: completed Binance adapter, replay, current-state assembly work, and the archived provider cutover in `plans/completed/binance-live-market-state-api-provider-cutover/`
- Unlocks: the missing sustained Spot runtime/read-model path for market-state queries
- Risk: high

### `binance-runtime-health-and-operator-observability`

- Depends on: the streaming runtime seam being explicit enough to observe
- Unlocks: debuggable startup, warm-up, stale, reconnect, and degraded-runtime behavior
- Risk: medium-high

### `binance-usdm-market-state-influence`

- Depends on: a stable streaming Spot runtime plus already-landed USD-M sensor surfaces
- Unlocks: final product semantics for whether derivatives context changes current-state/regime decisions
- Risk: high

### `binance-environment-config-and-rollout-hardening`

- Depends on: settled runtime and operator-health posture
- Unlocks: reliable `local`, `dev`, and `prod` defaults plus safe rollout notes
- Risk: medium-high

### `binance-long-run-runtime-hardening`

- Depends on: the finished runtime shape, observability posture, and any USD-M semantic changes
- Unlocks: confidence that the final Binance integration is resilient enough to operate without hidden drift
- Risk: high

## Cross-Cutting Dependency Notes

- `services/market-state-api` must remain the stable browser-facing contract while runtime sourcing changes underneath it.
- Any change that makes USD-M influence current-state or regime outputs is replay-sensitive and must preserve deterministic behavior for pinned fixtures.
- Runtime health work is rollout-sensitive because `/healthz` currently reflects process readiness only; refinement must decide what belongs in healthz versus payload/status surfaces.
- Environment hardening should consume established runtime behavior rather than invent separate semantics per environment.

## Inputs Already Available From Existing Work

- completed Spot runtime, depth recovery, USD-M sensor, raw append, replay, and current-state cutover artifacts under `plans/completed/`
- local Compose and browser validation for the current Go-owned `/api` boundary
- working public Binance Spot consumption in the local app path via the bounded snapshot reader
