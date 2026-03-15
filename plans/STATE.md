# Planning State

`plans/STATE.md` is the authoritative durable source of truth for repo planning and execution state.

- Use this file first when you need a quick answer for what is active, what is next, what is blocked, and what can run in parallel.
- Use initiative docs, epic docs, feature plans, and testing reports for rationale and detailed execution context.
- Use the OpenCode native task tracker only for in-session execution tasks; do not store session todo state in repo files.

## Status Vocabulary

- `seeded`: known work exists, but it has not been refined or prioritized enough for direct planning.
- `ready_to_refine`: initiative-level seed context is stable enough for `program-refining` to materialize refined epic context under `plans/epics/`.
- `ready_to_plan`: refined epic context already exists under `plans/epics/` and is ready for `feature-planning`.
- `ready_to_implement`: an active feature plan exists and can move into `feature-implementing`.
- `in_progress`: implementation or planning/testing work is actively underway.
- `ready_for_testing`: implementation is complete enough for `feature-testing`.
- `tested`: the planned validation matrix passed, but archive or follow-up state sync is still pending.
- `blocked`: the work cannot safely continue until a dependency or decision changes.
- `archived`: the feature or initiative slice is complete and lives under `plans/completed/` or equivalent historical context.

## Update Rules

- Read this file before non-micro planning, refinement, implementation, or feature-testing work.
- Update this file in the same pass whenever initiative, epic, feature, blocker, archive, next-step, or parallelization state changes.
- Keep the smallest relevant parent planning doc coherent with the status change, usually the relevant initiative `03-handoff.md` and, when needed, the epic handoff/refinement docs.
- Folder location supports state, but this file is the quick-look source of truth.

## Current Snapshot

- Last updated: 2026-03-15
- Active initiatives:
  - `initiatives/crypto-market-copilot-binance-integration-completion/`
- Active feature plans in `plans/`: none
- Ready to implement: none
- Next recommended: run `program-refining` for the `binance-runtime-health-and-operator-observability` initiative seed
- Ready in parallel after that starts: `binance-usdm-market-state-influence`
- Recently archived feature plans:
  - `plans/completed/binance-spot-runtime-read-model-owner/`
  - `plans/completed/binance-market-state-live-reader-cutover/`

## Next Recommended

1. Run `program-refining` for the `binance-runtime-health-and-operator-observability` initiative seed and materialize refined epic context.
2. Run `program-refining` for the `binance-usdm-market-state-influence` initiative seed in parallel if planning capacity allows.
3. Do not start `plans/epics/binance-environment-config-and-rollout-hardening/` until the Wave 2 epics settle the runtime-health and USD-M semantics.

## Ready In Parallel

| Item | Type | Status | Depends On | Next Action | Notes |
|---|---|---|---|---|---|
| `binance-runtime-health-and-operator-observability` | initiative seed | `ready_to_refine` | `plans/completed/binance-market-state-live-reader-cutover/` | Run `program-refining` and create `plans/epics/binance-runtime-health-and-operator-observability/` | Default next planning target |
| `binance-usdm-market-state-influence` | initiative seed | `ready_to_refine` | `plans/completed/binance-market-state-live-reader-cutover/` | Run `program-refining` in parallel when capacity allows | Parallel Wave 2 seed |

## Blocked

| Item | Type | Status | Depends On | Blocker | Next Action |
|---|---|---|---|---|---|
| `binance-environment-config-and-rollout-hardening` | initiative seed | `blocked` | Wave 2 seeds | Runtime-health and USD-M decisions are not settled yet | Revisit after Wave 2 planning and implementation |
| `binance-long-run-runtime-hardening` | initiative seed | `blocked` | Wave 2 and Wave 3 seeds | Final runtime shape and rollout defaults are not settled yet | Revisit after environment hardening is planned |

## Initiative State

### `initiatives/crypto-market-copilot-binance-integration-completion/`

- Initiative status: `in_progress`
- Parent quick-look doc: `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md`

| Item | Type | Status | Location | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|---|---|
| `binance-runtime-health-and-operator-observability` | initiative seed | `ready_to_refine` | `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` | Wave 1 complete | `binance-usdm-market-state-influence` | Run `program-refining` and create `plans/epics/binance-runtime-health-and-operator-observability/` | Default next seed |
| `binance-usdm-market-state-influence` | initiative seed | `ready_to_refine` | `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` | Wave 1 complete | `binance-runtime-health-and-operator-observability` | Refine in parallel when capacity allows | Wave 2 parallel seed |
| `binance-environment-config-and-rollout-hardening` | initiative seed | `blocked` | `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` | Wave 2 outcomes | - | Wait | Wave 3 seed |
| `binance-long-run-runtime-hardening` | initiative seed | `blocked` | `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` | Wave 2 and Wave 3 outcomes | - | Wait | Final hardening wave |

- Historical epic context retained for reference only: `plans/epics/binance-streaming-market-state-runtime-integration/`

### Other Initiative Snapshot

| Initiative | Status | Next Action | Notes |
|---|---|---|---|
| `initiatives/crypto-market-copilot-visibility-foundation/` | `archived` | Use archived plans as history only | Core visibility foundation slices are represented in `plans/completed/` |
| `initiatives/crypto-market-copilot-binance-live-market-data/` | `archived` | Use archived plans as history only | Superseded by the integration-completion follow-on initiative |
| `initiatives/crypto-market-copilot-alerting-and-evaluation/` | `ready_to_plan` | Prioritize explicitly, then run `feature-planning` from `plans/epics/alert-generation-and-hygiene/` | Refined alerting epics already exist under `plans/epics/` |

## Active Feature Plans

- None at the moment.

## Recently Archived

| Item | Archived State | Evidence | Notes |
|---|---|---|---|
| `plans/completed/binance-spot-runtime-read-model-owner/` | `archived` | `plans/completed/binance-spot-runtime-read-model-owner/testing-report.md` | Sustained Spot runtime owner and read-model seam completed |
| `plans/completed/binance-market-state-live-reader-cutover/` | `archived` | `plans/completed/binance-market-state-live-reader-cutover/testing-report.md` | Same-origin API/browser cutover and local smoke/docs completed |
