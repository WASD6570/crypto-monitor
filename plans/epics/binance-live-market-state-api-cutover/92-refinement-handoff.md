# Refinement Handoff: Binance Live Market State API Cutover

## Next Recommended Child Feature

- Run `feature-planning` for `binance-live-current-state-query-assembly` first.
- Reason: there is no live current-state read model yet, and the API/provider cutover is not safe to plan in detail until live symbol/global assembly semantics are explicit.

## Parallel Planning Status

- No safe parallel child planning yet.
- `binance-live-market-state-api-provider-cutover` depends on the first child slice settling:
  - the live provider seam inside `services/market-state-api`
  - how the existing `world` and `usa` contract sections remain honest under a Spot-first live source
  - the degradation and unsupported-symbol posture that compose and browser checks must assert

## Already-Settled Inputs To Preserve

- Binance live ingestion and audit behavior from:
  - `plans/completed/binance-spot-trade-canonical-handoff/`
  - `plans/completed/binance-spot-top-of-book-canonical-handoff/`
  - `plans/completed/binance-spot-depth-bootstrap-and-buffering/`
  - `plans/completed/binance-spot-depth-resync-and-snapshot-health/`
  - `plans/completed/binance-live-raw-append-and-feed-health-provenance/`
  - `plans/completed/binance-live-replay-binance-family-determinism/`
- Current API and dashboard boundary from:
  - `plans/completed/market-state-api-compose-integration/`
  - `services/market-state-api/api.go`
  - `apps/web/src/api/dashboard/dashboardClient.ts`
  - `apps/web/src/api/dashboard/dashboardDecoders.ts`
- Current-state builders from:
  - `services/feature-engine/service.go`
  - `services/regime-engine/service.go`
  - `libs/go/features/market_state_current.go`

## Assumptions

- The first live cutover is Spot-driven for current-state and regime inputs; it should not invent new deterministic replacements for missing `usa` or USD-M contributors.
- Slow context remains non-blocking, and the reserved history/audit seam in the current-state contract remains unchanged.
- If product later requires USD-M data to change current-state or regime decisions, that should become a new bounded follow-on slice rather than implicit scope growth here.
- No MCP servers are configured in the current session, so refinement used repo-local context only.

## Blockers

- No blocking product decision is visible if the Spot-first assumption holds.
- The main implementation blocker is structural: no live current-state read model exists yet, so the first child feature must create that source seam before API cutover work begins.
- If the product requires immediate USD-M influence on current-state behavior, this epic should be re-refined because `configs/dev/ingestion.v1.json` and `configs/prod/ingestion.v1.json` do not yet carry the open-interest polling defaults used in `local`.
