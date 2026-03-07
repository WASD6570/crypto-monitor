# Refinement Handoff: World USA Composites And Market State

## Recommended Next Child Feature

- none; child queue complete

## What Just Finished

- `world-usa-composite-snapshots` is implemented, validated, and archived under `plans/completed/world-usa-composite-snapshots/`.
- `market-quality-and-divergence-buckets` is implemented, validated, and archived under `plans/completed/market-quality-and-divergence-buckets/`.
- `symbol-and-global-regime-state` is implemented, validated, and archived under `plans/completed/symbol-and-global-regime-state/`.
- `market-state-current-query-contracts` is implemented, validated, and archived under `plans/completed/market-state-current-query-contracts/`.
- `market-state-history-and-audit-reads` is implemented, validated, and archived under `plans/completed/market-state-history-and-audit-reads/`.

## Why This Is Next

- Composite assembly, bucket outputs, regime outputs, current-state contracts, and replay-aware historical audit reads now exist for this epic.
- The next repo-level planning focus can move to `plans/epics/visibility-dashboard-core/92-refinement-handoff.md`.
- History and audit reads extended the completed current-state contract family without inventing a parallel read model.

## Recommended `feature-planning` Order

1. No further child planning remains in this epic.

## Parallelism Guidance

- No further child planning remains in this epic.
- Repo-level follow-up should continue in `plans/epics/visibility-dashboard-core/92-refinement-handoff.md`.

## Completed Prerequisites To Preserve

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`
- `plans/completed/raw-event-log-boundary/`
- `plans/completed/dashboard-shell-and-summary-strip/`
- `plans/completed/replay-run-manifests-and-ordering/`
- `plans/completed/world-usa-composite-snapshots/`
- `plans/completed/market-quality-and-divergence-buckets/`
- `plans/completed/symbol-and-global-regime-state/`
- `plans/completed/market-state-current-query-contracts/`

## Key Dependency Notes

- `plans/completed/raw-storage-and-replay-foundation/` now satisfies replay/storage stabilization for downstream history and audit consumers.
- `plans/epics/visibility-dashboard-core/92-refinement-handoff.md` is waiting on this epic's service-owned current-state query contracts; current-state consumer surfaces are therefore plan-ready once regime outputs are bounded.
- Replay and storage stabilization no longer block this epic; remaining sequencing is driven by composite -> buckets -> regime -> current-state contracts -> history reads.

## Assumptions To Preserve

- Go remains the live source of truth for composite construction, bucket math, and regime classification.
- `apps/web` stays read-only and presentational: it may format, sort, or label service output, but it must not recompute weights, fragmentation, market quality, or regime state.
- All thresholds, venue membership, stablecoin proxy allowlists, and downgrade logic remain config-versioned and replay-pinned.
- Conservative trust handling remains mandatory: degraded or fragmented conditions must lower trust quickly and never default toward `TRADEABLE`.

## Blocker Statement

- No blocker remains inside this epic; all child features are implemented.
- No replay/storage blocker remains.
- Next planning work is outside this epic.

## Suggested Validation Shape For The Next Planner

- Use this epic as completed prerequisite context for downstream dashboard or alerting consumers.
- Do not reopen the child queue unless a new bounded market-state feature is explicitly added.
