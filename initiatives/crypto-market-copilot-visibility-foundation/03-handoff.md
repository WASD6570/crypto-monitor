# Visibility Foundation Handoff

## Refined Epic Queue

No active queue remains. This initiative is archived.

## Execution State

- Initiative status: `archived`
- Historical reference only: `plans/epics/world-usa-composites-and-market-state/`, `plans/epics/visibility-dashboard-core/`, `plans/epics/slow-context-panel/`
- Completed feature plans live under `plans/completed/`
- Next recommended epic: none
- Parallel-safe now: none

| Epic | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/world-usa-composites-and-market-state/` | `archived` | Completed | - | Use archived feature evidence only | Includes `plans/completed/market-state-history-and-audit-reads/` |
| `plans/epics/visibility-dashboard-core/` | `archived` | Completed | - | Use archived feature evidence only | Includes `plans/completed/dashboard-fixture-smoke-matrix/` |
| `plans/epics/slow-context-panel/` | `archived` | Completed | - | Use archived feature evidence only | Includes `plans/completed/slow-context-dashboard-panel/` |

## Planning Waves

### Wave 1

- `canonical-contracts-and-fixtures` (completed)
- Why now: every later slice depends on shared event, feature, replay, and market-state vocabulary.

### Wave 2

- `market-ingestion-and-feed-health` (completed)
- Why now: trusted live inputs and degradation signals must exist before replay, state, and dashboard consumers can be implemented safely.

### Wave 3

- `raw-storage-and-replay-foundation` (completed)
- `visibility-dashboard-core`
- Why parallel: replay depended on ingestion semantics and contracts; dashboard IA/query planning could proceed once the market-state contract boundary was understood, without redefining ingestion or replay semantics.

### Wave 4

- `world-usa-composites-and-market-state`
- Why later: this slice depends on contracts and ingestion semantics being stable; replay-backed history is now unblocked, but current-state consumer sequencing still depends on regime outputs.

### Wave 5

- `slow-context-panel`
- Why later: this is intentionally non-blocking and should not shape the core realtime state path.

## Refined Epics

No active refined epics remain for this archived initiative. Use `## Execution State` for historical references and `plans/completed/` for implementation history.
