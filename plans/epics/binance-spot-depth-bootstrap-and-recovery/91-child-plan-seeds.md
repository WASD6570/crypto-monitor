# Child Plan Seeds: Binance Spot Depth Bootstrap And Recovery

## `binance-spot-depth-bootstrap-and-buffering`

- Outcome: one bounded Spot depth slice bootstraps `/api/v3/depth`, buffers `depth@100ms` deltas during startup, aligns the first accepted delta window to the snapshot boundary, and emits sequencer-ready snapshot/delta messages for BTC/ETH without reopening trade or top-of-book logic.
- Primary repo areas: `services/venue-binance`, `tests/fixtures/events/binance`, `tests/integration`
- Dependencies: `plans/completed/binance-spot-ws-runtime-supervisor/`, existing Binance depth parsers in `services/venue-binance/orderbook.go`, inherited Wave 1 identity/time policy, and Binance snapshot defaults in `configs/*/ingestion.v1.json`
- Validation shape: targeted Go tests for snapshot bootstrap, buffered delta retention, first accepted delta alignment, and canonical normalization of accepted snapshot/delta outputs; direct Binance depth API validation after implementation
- Why it stands alone: startup alignment is the first trustworthy depth boundary and should settle before gap recovery and refresh policy add more moving parts

## `binance-spot-depth-resync-and-snapshot-health`

- Outcome: one bounded recovery slice handles sequence-gap triggered resync, snapshot cooldown/rate-limit enforcement, snapshot refresh cadence, snapshot staleness or drift degradation, and machine-readable feed-health output for the live Spot depth path.
- Primary repo areas: `services/venue-binance`, `libs/go/ingestion`, `configs/*/ingestion.v1.json`, `tests/integration`
- Dependencies: `binance-spot-depth-bootstrap-and-buffering`, existing runtime health primitives in `services/venue-binance/runtime.go`, inherited feed-health vocabulary from `plans/completed/market-ingestion-and-feed-health/`, and Binance depth config defaults
- Validation shape: deterministic gap/resync tests, snapshot cooldown/rate-limit tests, snapshot refresh and staleness checks, and direct live validation that degraded recovery states remain visible without silent repair
- Why it stands alone: recovery policy is rollout-sensitive and couples runtime state, config thresholds, and feed-health semantics more tightly than the initial bootstrap slice

## Validation Note

- Do not create a separate smoke-only or integration-only child feature for this epic.
- After the bootstrap and recovery slices are implemented, run direct live Binance depth validation and record the result in the current handoff or implementation report.
