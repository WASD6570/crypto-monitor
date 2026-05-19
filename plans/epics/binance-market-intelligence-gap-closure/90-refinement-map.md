# Refinement Map

## Already Done

- `plans/completed/binance-spot-trade-canonical-handoff/` implemented Spot trade parsing and canonical handoff fixtures.
- `plans/completed/binance-spot-ws-runtime-supervisor/` implemented Spot websocket lifecycle, subscriptions, reconnect posture, and feed-health hooks.
- `plans/completed/binance-spot-depth-bootstrap-and-buffering/` and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` implemented depth bootstrap, delta alignment, resync, cooldown, rate-limit, snapshot refresh, and feed-health visibility.
- `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/` implemented USD-M mark/funding/index and liquidation runtime foundations.
- `plans/completed/binance-live-raw-append-and-feed-health-provenance/` and `plans/completed/binance-live-replay-binance-family-determinism/` implemented raw/replay seams for accepted Binance inputs.
- `plans/completed/binance-live-market-state-api-provider-cutover/` and `plans/completed/binance-spot-runtime-read-model-owner/` moved the API path to a sustained Spot-backed runtime/read model.
- `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/` settled `/api/runtime-status` as the operator runtime-health surface.
- `plans/completed/binance-usdm-influence-policy-and-signal/` and `plans/completed/binance-usdm-output-application-and-replay-proof/` settled the conservative USD-M provenance and watch-cap behavior.
- `plans/completed/binance-runtime-config-profile-parity/` and `plans/completed/binance-rollout-compose-and-ops-handoff/` settled prod-like checked-in config and default Compose startup posture.

## Gaps This Epic Closes

- validation drift currently prevents treating the full repo as stable after the latest Binance config/runtime changes
- long-run Binance runtime behavior still lacks a final failure-path and repeated-run confidence gate
- live Spot trade prints are parsed but do not yet materially influence current-state or alert-readiness features
- live Spot depth is used for synchronization and health but not yet converted into market indicators such as spread, top-level size, imbalance, depth pressure, or slippage proxy
- current-state assembly uses a fixed Binance liquidity score instead of observed liquidity quality
- USD-M context has basic funding, basis, liquidation, and open-interest inputs but lacks richer derivatives indicator outputs such as open-interest deltas, liquidation intensity, and funding/basis regimes
- API/dashboard surfaces do not yet present a coherent Binance indicator summary beyond existing current-state, provenance, and runtime-status information

## Refinement Waves

### Wave 1: Stabilize The Baseline

- `binance-validation-baseline-reconciliation`
- Why first: new indicator features should not be layered on top of failing full-suite, contract, or replay smoke baselines.

### Wave 2: Prove Runtime Trust

- `binance-live-runtime-soak-and-failure-hardening`
- Why next: enriched market indicators are only useful if the underlying Spot/USD-M runtime survives reconnects, stale periods, rate limits, and repeated validation without semantic drift.

### Wave 3: Promote Spot Inputs Into Indicators

- `binance-spot-trade-flow-feature-inputs`
- `binance-spot-depth-liquidity-indicators`
- Why parallel after Wave 2: both consume settled Spot runtime inputs but derive different feature surfaces; trade-flow work should not block depth-liquidity work unless they choose a shared additive contract.

### Wave 4: Promote USD-M Inputs Into Indicators

- `binance-usdm-derivatives-indicator-enrichment`
- Why later: USD-M enrichment should build on the already settled conservative cap and should not reopen acquisition or runtime-health design.

### Wave 5: Surface Indicator Readiness

- `binance-market-indicator-api-and-dashboard-readiness`
- Why last: API and UI surfaces should follow stable service-owned indicators, not invent client-side calculations or premature contracts.

## Overlap And Non-Goals

- do not create a validation-only standalone feature if the remaining work is just running an existing smoke; validation baseline reconciliation is valid only because it resolves known contract/test/replay drift
- do not reopen Spot/USD-M acquisition scope; the child plans should use existing Binance public surfaces unless refinement proves one small missing public field is required
- do not introduce account/private data, execution, positions, or exchange credentials
- do not use Python in the live runtime path
- do not require frontend changes before service-owned indicator contracts are stable

## Direct Post-Implementation Checks

- run `go test ./...`
- run `make contracts-validate`
- run `CONTRACT_FIXTURES=1 make replay-smoke`
- run targeted Binance integration/replay tests for any touched stream family
- run `pnpm --dir apps/web test` and `pnpm --dir apps/web build` only when dashboard/API rendering surfaces change
- run `make compose-smoke` or `make compose-dev-smoke` when Docker is available and the child feature changes startup, runtime wiring, or proxy behavior
