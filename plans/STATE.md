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

- Last updated: 2026-05-19
- Active initiatives:
  - `initiatives/crypto-market-copilot-binance-integration-completion/`
- Active feature plans in `plans/`:
  - `plans/binance-spot-depth-liquidity-indicators/`
- Ready to refine: none
- Ready to plan: `binance-usdm-derivatives-indicator-enrichment` only if explicitly prioritized for parallel planning
- Ready to implement:
  - `plans/binance-spot-depth-liquidity-indicators/`
- In progress: none
- Ready for testing: none
- Next recommended: run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/`
- Ready in parallel: USD-M derivatives enrichment can be planned in parallel only if explicitly prioritized; run `feature-testing` for the dev-only live-reload workflow only if the missing active plan directory is restored or the testing matrix is recreated from implementation evidence
- Recently archived feature plans:
  - `plans/completed/binance-spot-trade-flow-feature-inputs/`
  - `plans/completed/binance-live-runtime-soak-and-failure-hardening/`
  - `plans/completed/binance-validation-baseline-reconciliation/`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/`
  - `plans/completed/binance-usdm-output-application-and-replay-proof/`
  - `plans/completed/binance-usdm-influence-policy-and-signal/`
  - `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/`
  - `plans/completed/binance-spot-runtime-read-model-owner/`
  - `plans/completed/binance-market-state-live-reader-cutover/`

## Next Recommended

1. Use `plans/completed/binance-spot-trade-flow-feature-inputs/testing-report.md` as the archived internal trade-flow input evidence for later indicator/API work.
2. Run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/`.
3. USD-M derivatives enrichment can be planned in parallel only if explicitly prioritized; otherwise run it after Spot depth-liquidity implementation/testing settles.
4. Keep alerting and dashboard readiness out of scope until service-owned enriched indicator boundaries are planned, implemented, tested, and archived.

## Ready In Parallel

| Item | Type | Status | Depends On | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/binance-market-intelligence-gap-closure/` | refined epic | `in_progress` | Wave 2 and Wave 3 outcomes archived, trade-flow archived | Run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/` | Depth-liquidity plan is ready to implement; USD-M derivatives enrichment can be planned in parallel only if explicitly prioritized |

## Blocked

| Item | Type | Status | Depends On | Blocker | Next Action |
|---|---|---|---|---|---|
| `binance-live-reload-dev-workflow` | feature-testing artifact | `blocked` | `plans/completed/binance-rollout-compose-and-ops-handoff/` | `plans/binance-live-reload-dev-workflow/` is referenced by durable state but absent from the current worktree | Restore/recreate the active plan directory or recreate the testing matrix from implementation evidence before archive testing |

## Initiative State

### `initiatives/crypto-market-copilot-binance-integration-completion/`

- Initiative status: `in_progress`
- Parent quick-look doc: `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md`

| Item | Type | Status | Location | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|---|---|
| `plans/epics/binance-runtime-health-and-operator-observability/` | refined epic | `archived` | `plans/epics/binance-runtime-health-and-operator-observability/` | `plans/completed/binance-runtime-health-snapshot-owner/` | - | Use archived child evidence as the settled operator runtime-health surface | The endpoint-and-ops-handoff child is complete and archived |
| `plans/epics/binance-usdm-market-state-influence/` | refined epic | `archived` | `plans/epics/binance-usdm-market-state-influence/` | Wave 1 complete and `plans/epics/binance-usdm-context-sensors/` | - | Use archived USD-M child evidence as the settled market-state semantics reference | Both USD-M child plans are complete and archived |
| `plans/epics/binance-environment-config-and-rollout-hardening/` | refined epic | `blocked` | `plans/epics/binance-environment-config-and-rollout-hardening/` | Wave 2 runtime-health and USD-M semantics complete | `plans/epics/binance-market-intelligence-gap-closure/` | Restore/recreate `plans/binance-live-reload-dev-workflow/` before archive testing | The prod-like rollout posture is archived, dev-only live-reload implementation evidence exists, but the referenced active feature plan directory is absent |
| `plans/epics/binance-market-intelligence-gap-closure/` | refined epic | `in_progress` | `plans/epics/binance-market-intelligence-gap-closure/` | Wave 2 and Wave 3 outcomes archived | - | Run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/` | Depth-liquidity plan is ready to implement while alerting remains out of scope |

- Historical epic context retained for reference only: `plans/epics/binance-streaming-market-state-runtime-integration/`

### Other Initiative Snapshot

| Initiative | Status | Next Action | Notes |
|---|---|---|---|
| `initiatives/crypto-market-copilot-visibility-foundation/` | `archived` | Use archived plans as history only | Core visibility foundation slices are represented in `plans/completed/` |
| `initiatives/crypto-market-copilot-binance-live-market-data/` | `archived` | Use archived plans as history only | Superseded by the integration-completion follow-on initiative |
| `initiatives/crypto-market-copilot-alerting-and-evaluation/` | `ready_to_plan` | Prioritize explicitly, then run `feature-planning` from `plans/epics/alert-generation-and-hygiene/` | Refined alerting epics already exist under `plans/epics/` |

## Active Feature Plans

| Feature Plan | Status | Depends On | Next Action | Notes |
|---|---|---|---|---|
| `plans/binance-spot-depth-liquidity-indicators/` | `ready_to_implement` | `plans/completed/binance-live-runtime-soak-and-failure-hardening/`, `plans/completed/binance-spot-depth-bootstrap-and-buffering/`, `plans/completed/binance-spot-depth-resync-and-snapshot-health/`, `plans/completed/binance-spot-runtime-read-model-owner/` | Run `feature-implementing` for this plan | Internal Go-owned Spot depth-liquidity indicators and observed current-state liquidity score; public API/dashboard indicator exposure remains deferred |

## Recently Archived

| Item | Archived State | Evidence | Notes |
|---|---|---|---|
| `plans/completed/binance-spot-trade-flow-feature-inputs/` | `archived` | `plans/completed/binance-spot-trade-flow-feature-inputs/testing-report.md` | Internal Go-owned Spot trade-flow feature inputs passed unit/service, runtime, integration, replay, full Go, contract, fixture, replay-smoke, and whitespace validation |
| `plans/completed/binance-live-runtime-soak-and-failure-hardening/` | `archived` | `plans/completed/binance-live-runtime-soak-and-failure-hardening/testing-report.md` | Additive USD-M runtime-status visibility, deterministic failure-path coverage, runtime-soak runbook handoff, and required local feature-testing matrix completed; optional public Binance and Compose checks remain environment-dependent |
| `plans/completed/binance-validation-baseline-reconciliation/` | `archived` | `plans/completed/binance-validation-baseline-reconciliation/testing-report.md` | Focused Binance runtime tests, `go test ./...`, contract-family validation, TypeScript contract tests, fixture validation, and deterministic replay smoke passed; no live runtime behavior was added |
| `plans/completed/binance-rollout-compose-and-ops-handoff/` | `archived` | `plans/completed/binance-rollout-compose-and-ops-handoff/testing-report.md` | Compose now pins one prod-like startup posture, the operator rollout runbook matches the live stack, and the repeatable smoke proof plus manual handoff checks passed |
| `plans/completed/binance-runtime-config-profile-parity/` | `archived` | `plans/completed/binance-runtime-config-profile-parity/testing-report.md` | Checked-in local/dev/prod configs now stay prod-like and identical in runtime behavior, with real-file ingestion invariants, provider config-path consumption proof, and focused Binance USD-M smoke passing |
| `plans/completed/binance-usdm-output-application-and-replay-proof/` | `archived` | `plans/completed/binance-usdm-output-application-and-replay-proof/testing-report.md` | Conservative USD-M watch-cap application, additive symbol/global provenance, live provider wiring, and deterministic replay proof completed |
| `plans/completed/binance-usdm-influence-policy-and-signal/` | `archived` | `plans/completed/binance-usdm-influence-policy-and-signal/testing-report.md` | Deterministic internal USD-M influence contract, venue input seam, bounded evaluator, and replay/current-state regression proof completed |
| `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/` | `archived` | `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/testing-report.md` | Additive operator runtime-status route, live wiring, and ops handoff completed |
| `plans/completed/binance-spot-runtime-read-model-owner/` | `archived` | `plans/completed/binance-spot-runtime-read-model-owner/testing-report.md` | Sustained Spot runtime owner and read-model seam completed |
| `plans/completed/binance-market-state-live-reader-cutover/` | `archived` | `plans/completed/binance-market-state-live-reader-cutover/testing-report.md` | Same-origin API/browser cutover and local smoke/docs completed |
