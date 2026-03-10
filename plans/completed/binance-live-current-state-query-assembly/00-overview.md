# Binance Live Current State Query Assembly

## Ordered Implementation Plan

1. Introduce a service-owned live query-source seam inside `services/market-state-api` so current-state assembly is separated from the existing HTTP `Provider` and no longer depends on deterministic bundle builders.
2. Build a Spot-driven Binance current-state assembler that turns accepted top-of-book, depth recovery, and feed-health state for `BTC-USD` and `ETH-USD` into stable `SymbolCurrentStateQuery` and `GlobalCurrentStateQuery` inputs.
3. Add deterministic fixture-backed proof in `services/market-state-api`, `tests/integration`, and `tests/replay` so repeated runs over the same accepted Binance inputs produce stable symbol and global current-state responses.
4. Run the attached validation matrix, write `plans/binance-live-current-state-query-assembly/testing-report.md`, then move the full directory to `plans/completed/binance-live-current-state-query-assembly/` after implementation and validation finish.

## Requirements

- Scope is limited to the first live query-assembly seam behind `services/market-state-api`; do not wire `cmd/market-state-api` or compose over to the new provider in this feature.
- Preserve the existing `Provider` and HTTP route contracts so the follow-on cutover feature can swap providers without changing `/api/market-state/global`, `/api/market-state/:symbol`, or `/healthz`.
- Keep the first live cutover Spot-driven for composite and regime inputs. USD-M remains out of scope for this feature except as future context.
- Keep canonical symbols fixed as `BTC-USD` and `ETH-USD`.
- Preserve the existing current-state consumer contract, including `world`, `usa`, 30s/2m/5m bucket sections, recent context, slow-context shape, and reserved history/audit seam.
- Keep absent `usa` contributors explicit through normal unavailable or partial semantics rather than introducing new deterministic placeholders.
- Preserve machine-readable degradation from Binance Spot feed-health and depth recovery inputs.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Repository state to preserve

- `services/market-state-api/api.go` already defines the stable `Provider` interface and current HTTP boundary, but its package-local `DeterministicProvider` still owns bundle construction.
- `services/feature-engine/service.go` already exposes `BuildCompositeSnapshot`, `ObserveWorldUSABucket`, `AdvanceWorldUSABuckets`, and `QueryCurrentStateWithSlowContext`.
- `services/regime-engine/service.go` already exposes `Observe` and `QueryCurrentGlobalState`.
- `libs/go/features/market_state_current.go` already fixes the symbol/global response contract and derives availability from unavailable or degraded composite and bucket inputs.
- `plans/completed/binance-spot-top-of-book-canonical-handoff/` and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` already settle the accepted Binance Spot price, timestamp, and depth-health inputs this feature should consume.

### First live assembly posture

- Use accepted Binance Spot top-of-book as the price input for the `world` composite group.
- Treat `usa` as genuinely unavailable in the first live cutover because no live USA contributor exists in this initiative slice.
- Use the existing bucket and regime builders rather than inventing a market-state-specific classifier.
- Drive degradation from accepted feed-health and depth recovery posture so unsynchronized depth, stale messages, or timestamp fallback remain visible in the assembled query results.
- Keep slow context optional and non-blocking through the existing `feature-engine` seam.

### Boundaries

- Keep query assembly ownership in `services/market-state-api`; this feature creates the live read-model seam there rather than pushing venue state logic into `apps/web` or hiding it inside generic libs.
- Reuse `services/feature-engine` and `services/regime-engine` as shared builders, extracting only tiny helper seams there if current package boundaries make assembly impossible.
- Do not reopen Binance parsing, raw append identity, or replay manifest behavior already settled upstream.
- Do not absorb command wiring, compose updates, or browser checks; those belong to `binance-live-market-state-api-provider-cutover`.

### Live vs research boundary

- All live query assembly, bucketing, regime evaluation, and deterministic proof stay in Go under `services/market-state-api`, `services/feature-engine`, `services/regime-engine`, and `tests/`.
- Offline analysis may inspect replay artifacts later, but no Python runtime dependency belongs in this live current-state path.

## ASCII Flow

```text
accepted Binance Spot state
  - top-of-book price
  - supervisor feed health
  - depth recovery status
          |
          v
services/market-state-api live query source
  - latest symbol inputs
  - world/usa snapshot assembly
  - bucket advancement
  - regime evaluation
          |
          +--> services/feature-engine.QueryCurrentStateWithSlowContext
          |
          +--> services/regime-engine.QueryCurrentGlobalState
          |
          v
stable SymbolCurrentStateQuery / GlobalCurrentStateQuery outputs
  - existing current-state contract preserved
  - unavailable usa stays explicit
  - degradation remains machine-readable
```

## Archive Intent

- Keep this feature active under `plans/binance-live-current-state-query-assembly/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/binance-live-current-state-query-assembly/`.
