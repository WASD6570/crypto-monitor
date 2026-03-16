# Child Plan Seeds

## `binance-usdm-influence-policy-and-signal`

- Outcome: add a deterministic Go-owned evaluator that turns the existing USD-M context for `BTC-USD` and `ETH-USD` into one bounded per-symbol influence signal with explicit no-context and degraded behavior
- Primary repo area: `services/feature-engine`, `services/venue-binance`, `tests/replay`
- Dependencies: `plans/epics/binance-usdm-context-sensors/`, `plans/completed/binance-spot-runtime-read-model-owner/`, `plans/completed/binance-market-state-live-reader-cutover/`
- Validation shape: targeted evaluator tests plus replay fixtures showing identical pinned Spot plus USD-M inputs always yield the same influence signal, and absent USD-M context preserves current Spot-only behavior
- Why it stands alone: this settles the policy seam first so later current-state and regime work does not guess at semantics or API consequences

## `binance-usdm-output-application-and-replay-proof`

- Outcome: apply the settled USD-M influence signal to current-state and regime assembly, keep `/api/market-state/*` stable by default, and add only the smallest additive provenance metadata if the new semantics would otherwise be opaque
- Primary repo area: `services/regime-engine`, `services/market-state-api`, `services/feature-engine`, `tests/integration`, `tests/replay`
- Dependencies: `binance-usdm-influence-policy-and-signal`
- Validation shape: focused API and integration checks for `BTC-USD` and `ETH-USD`, plus deterministic replay proof for the final consumer-facing behavior
- Why it stands alone: consumer-facing application, compatibility proof, and any additive API seam should follow the settled influence signal instead of being bundled into the lower-level evaluator slice
