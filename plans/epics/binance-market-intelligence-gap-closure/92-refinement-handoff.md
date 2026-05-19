# Refinement Handoff

## Next Recommended Action

- Run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/`.
- Why next: the active plan now defines the internal feature model, runtime read-model wiring, fixture/replay proof, and testing matrix needed to replace the fixed Binance liquidity score with observed depth-liquidity quality.

## Active Child Plan

- `plans/binance-spot-depth-liquidity-indicators/`
- Status: `ready_to_implement`
- Last archived child: `plans/completed/binance-spot-trade-flow-feature-inputs/`

## Remaining Child Queue

- `binance-usdm-derivatives-indicator-enrichment`
- `binance-market-indicator-api-and-dashboard-readiness`

## Safe Parallel Planning And Execution

- `binance-spot-depth-liquidity-indicators` is ready to implement.
- `binance-usdm-derivatives-indicator-enrichment` can be planned in parallel only if explicitly prioritized and as long as it preserves the settled USD-M cap posture.
- `binance-market-indicator-api-and-dashboard-readiness` should wait for the service-owned indicator boundaries to settle.

## Archived Child Evidence To Carry Forward

- `plans/completed/binance-spot-trade-canonical-handoff/`
- `plans/completed/binance-spot-ws-runtime-supervisor/`
- `plans/completed/binance-spot-depth-bootstrap-and-buffering/`
- `plans/completed/binance-spot-depth-resync-and-snapshot-health/`
- `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/`
- `plans/completed/binance-live-raw-append-and-feed-health-provenance/`
- `plans/completed/binance-live-replay-binance-family-determinism/`
- `plans/completed/binance-spot-runtime-read-model-owner/`
- `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/`
- `plans/completed/binance-usdm-influence-policy-and-signal/`
- `plans/completed/binance-usdm-output-application-and-replay-proof/`
- `plans/completed/binance-runtime-config-profile-parity/`
- `plans/completed/binance-rollout-compose-and-ops-handoff/`
- `plans/completed/binance-validation-baseline-reconciliation/`
- `plans/completed/binance-live-runtime-soak-and-failure-hardening/`
- `plans/completed/binance-spot-trade-flow-feature-inputs/`

## Prerequisites To Carry Forward

- keep `BTC-USD` and `ETH-USD` as the tracked symbols
- keep Go as the live runtime path and Python offline-only
- keep `/healthz` process-only and `/api/runtime-status` as the operator runtime-health surface
- keep `apps/web` from computing venue-specific market state or talking to Binance directly
- preserve replay determinism for accepted Binance inputs and config-pinned feature outputs
- keep degraded feeds, sequence gaps, rate limits, stale data, reconnect loops, and warm-up states visible in machine-readable output
- keep contract changes additive unless a child feature plan explicitly handles compatibility and consumer validation

## Assumptions And Blockers

- assumption: the current Binance public Spot and USD-M surfaces are enough for the next indicator layer; no private endpoints are needed
- assumption: richer indicators should feed later alerting, but alert generation remains out of scope for this epic
- runtime confidence gate before enriched indicator implementation: archived at `plans/completed/binance-live-runtime-soak-and-failure-hardening/`
- blocker for Compose-based proof in this environment: Docker is not currently available in WSL, so Docker checks may need a different host or later operator run
