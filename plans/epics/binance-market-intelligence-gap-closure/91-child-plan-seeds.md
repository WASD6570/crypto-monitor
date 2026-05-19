# Child Plan Seeds

## `binance-validation-baseline-reconciliation`

- Outcome: restore the current Binance validation baseline so full Go tests, contract-family validation, and deterministic replay smoke pass before new market-intelligence behavior is added
- Primary repo area: `services/venue-binance`, `tests/integration`, `tests/replay`, `schemas/json/replay`, `scripts/dev`, `plans/STATE.md`
- Dependencies: current Binance runtime/config work and the validation drift observed after the prod-like config updates
- Validation shape: `go test ./...`, `make contracts-validate`, `CONTRACT_FIXTURES=1 make replay-smoke`, plus targeted Binance runtime tests for any changed expectations
- Why it stands alone: it establishes a trustworthy baseline and resolves state/schema/replay drift without bundling new indicator semantics

## `binance-live-runtime-soak-and-failure-hardening`

- Outcome: prove the sustained Binance Spot plus USD-M runtime survives repeated reconnect, stale, rate-limit, depth-resync, warm-up, and shutdown paths without dropping deterministic current-state semantics
- Primary repo area: `cmd/market-state-api`, `services/venue-binance`, `services/market-state-api`, `tests/integration`, `tests/replay`, `docs/runbooks`
- Dependencies: `binance-validation-baseline-reconciliation`; archived runtime-status, Spot runtime owner, USD-M output-application, config parity, and rollout evidence
- Validation shape: focused runtime failure tests, deterministic replay/current-state regression checks, repeated runtime-status snapshots, and Compose or documented live-boundary smoke when Docker is available
- Why it stands alone: it is the final trust gate for the existing live runtime before richer indicators consume the same surfaces

## `binance-spot-trade-flow-feature-inputs`

- Outcome: turn accepted Binance Spot trade prints into deterministic feature inputs for trade-flow and price-action context instead of only parsing and replaying them
- Primary repo area: `services/venue-binance`, `services/feature-engine`, `libs/go/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Dependencies: `binance-live-runtime-soak-and-failure-hardening`; archived Spot trade canonical handoff and raw/replay seams
- Validation shape: fixture-backed trade-flow feature tests, integration proof that live trade observations remain ordered and deterministic, and replay proof for the same accepted raw trade inputs
- Why it stands alone: trade-flow indicators use the existing trade stream and can be developed independently from depth-derived liquidity metrics

## `binance-spot-depth-liquidity-indicators`

- Outcome: derive real Spot liquidity indicators from top-of-book and synchronized depth state, replacing the current fixed Binance liquidity contribution with observed spread, size, imbalance, depth pressure, and degradation-aware liquidity quality
- Primary repo area: `services/venue-binance`, `services/feature-engine`, `libs/go/features`, `services/market-state-api`, `tests/integration`, `tests/replay`, `schemas/json/features` if an additive contract is required
- Dependencies: `binance-live-runtime-soak-and-failure-hardening`; archived depth bootstrap/recovery and runtime-status evidence
- Validation shape: depth fixture tests, sequence-gap and resync degradation checks, deterministic liquidity-feature replay, and current-state regression proving degraded depth cannot silently improve liquidity quality
- Why it stands alone: depth-derived liquidity has different state, failure, and replay risks than trade-flow features

## `binance-usdm-derivatives-indicator-enrichment`

- Outcome: enrich USD-M context into explicit derivatives indicators such as funding pressure, basis regime, open-interest change, and liquidation intensity while preserving the conservative current-state cap posture
- Primary repo area: `services/venue-binance`, `services/feature-engine`, `libs/go/features`, `services/market-state-api`, `tests/integration`, `tests/replay`, `schemas/json/features` if an additive contract is required
- Dependencies: `binance-live-runtime-soak-and-failure-hardening`; archived USD-M influence policy and output-application proof
- Validation shape: deterministic USD-M fixture scenarios for missing, stale, degraded, and elevated derivatives context; replay proof for OI/funding/basis/liquidation indicators; current-state regression proving caps remain bounded
- Why it stands alone: USD-M enrichment uses different surfaces and freshness rules from Spot trade/depth indicators and should not reopen Spot runtime behavior

## `binance-market-indicator-api-and-dashboard-readiness`

- Outcome: expose the settled Binance Spot and USD-M indicator readiness through additive service-owned current-state/API output and dashboard rendering so later alerting plans can consume and inspect the same truth
- Primary repo area: `services/market-state-api`, `services/feature-engine`, `libs/go/features`, `apps/web`, `schemas/json/features`, `tests/integration`, `tests/replay`
- Dependencies: `binance-spot-trade-flow-feature-inputs`, `binance-spot-depth-liquidity-indicators`, and `binance-usdm-derivatives-indicator-enrichment`
- Validation shape: API contract tests, dashboard decoder/view-model tests, replay/current-state compatibility checks, `pnpm --dir apps/web test`, and `pnpm --dir apps/web build`
- Why it stands alone: the user-facing/readiness surface should only be planned once the underlying service-owned indicators are stable
