# Refinement Handoff

## Next Recommended Child Feature

- Start `feature-planning` for `binance-spot-runtime-read-model-owner`
- Why next: `cmd/market-state-api/live_provider.go` still shows the temporary polling seam, and every later integration or observability slice depends on replacing it with one explicit sustained runtime owner first

## Safe Parallel Planning

- none yet; `binance-market-state-live-reader-cutover` depends on the runtime owner boundary and its warm-up/degradation semantics being made explicit first

## Prerequisites To Carry Forward

- preserve the archived provider-cutover contract and treat `services/market-state-api` as the stable browser-facing boundary
- reuse completed Spot runtime owners instead of creating a second websocket or depth-policy implementation
- keep `BTC-USD` and `ETH-USD` as the only supported symbols in this epic
- preserve machine-readable warm-up, degraded, partial, and unavailable behavior during startup and recovery
- keep replay-sensitive assumptions explicit enough that repeated accepted-input tests can still prove deterministic current-state outputs

## Assumptions And Blockers

- assumption: this epic remains Spot-only for current-state influence; USD-M semantics stay deferred to `binance-usdm-market-state-influence`
- assumption: `/healthz` remains process health only unless the later observability epic chooses an additive status surface
- no blocking product ambiguity remains for refinement; the current repo state supports splitting runtime ownership from provider cutover cleanly
