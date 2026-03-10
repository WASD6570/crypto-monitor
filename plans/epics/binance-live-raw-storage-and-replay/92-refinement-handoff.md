# Refinement Handoff: Binance Live Raw Storage And Replay

## Next Recommended Child Feature

- Run `feature-planning` for `binance-live-raw-append-and-feed-health-provenance` first.
- Reason: the replay slice is not safe to plan in detail until Binance raw append identity, partition routing, and degraded-feed provenance are settled.

## Parallel Planning Status

- No safe parallel child planning yet.
- `binance-live-replay-binance-family-determinism` depends on the raw append slice fixing:
  - final Binance stream-family partition posture
  - duplicate identity precedence for accepted live families
  - degraded feed-health retention format reused by replay tests

## Already-Settled Inputs To Preserve

- Spot canonical families and depth recovery semantics from:
  - `plans/completed/binance-spot-trade-canonical-handoff/`
  - `plans/completed/binance-spot-top-of-book-canonical-handoff/`
  - `plans/completed/binance-spot-depth-bootstrap-and-buffering/`
  - `plans/completed/binance-spot-depth-resync-and-snapshot-health/`
- USD-M canonical families and mixed-surface health semantics from:
  - `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/`
  - `plans/completed/binance-usdm-open-interest-rest-polling/`
  - `plans/completed/binance-usdm-context-sensor-fixtures-and-integration/`
- Shared raw/replay primitives from:
  - `libs/go/ingestion/raw_event_log.go`
  - `services/replay-engine/runtime.go`
  - `tests/replay/replay_retention_safety_test.go`

## Assumptions

- The existing shared raw append contract is sufficient for Binance live families unless the first child feature finds one concrete provenance gap.
- Replay determinism should reuse the existing ordering policy and audit trail path instead of introducing a Binance-only replay mode.
- Degraded feed-health and timestamp fallback facts are part of the replay audit boundary, not incidental metadata.

## Blockers

- No blocking product decision is visible from current initiative context.
- The main dependency blocker is implementation sequencing: replay planning should wait for the raw append child feature to settle concrete Binance partition and identity behavior.
