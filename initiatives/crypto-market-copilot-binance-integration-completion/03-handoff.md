# Binance Integration Completion Handoff

## Refined Epic Queue

- `plans/epics/binance-market-intelligence-gap-closure/` has active child plan `plans/binance-spot-depth-liquidity-indicators/` ready to implement and supersedes the narrower `binance-long-run-runtime-hardening` seed.
- `plans/epics/binance-environment-config-and-rollout-hardening/` still has the dev-only live-reload workflow implementation evidence, but its active plan directory needs reconciliation before archive testing can be driven from `plans/binance-live-reload-dev-workflow/`.

## Execution State

- Initiative status: `in_progress`
- Completed prerequisite epic context: `plans/epics/binance-streaming-market-state-runtime-integration/` (historical reference only)
- Next recommended execution step: run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/`
- Parallel-safe steps: plan `binance-usdm-derivatives-indicator-enrichment` only if explicitly prioritized; restore or recreate the dev-only live-reload testing matrix if archive testing remains desired

| Item | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/binance-runtime-health-and-operator-observability/` | `archived` | `plans/completed/binance-runtime-health-snapshot-owner/` | - | Use archived child evidence as the settled operator runtime-health surface | The runtime-status endpoint and ops-handoff child is complete and archived |
| `plans/epics/binance-usdm-market-state-influence/` | `archived` | Wave 1 complete and `plans/epics/binance-usdm-context-sensors/` | - | Use archived USD-M child evidence as the settled market-state semantics reference | Both child plans are complete and archived |
| `plans/epics/binance-environment-config-and-rollout-hardening/` | `blocked` | Wave 2 decisions are now settled | `plans/epics/binance-market-intelligence-gap-closure/` | Restore/recreate `plans/binance-live-reload-dev-workflow/` before archive testing | Checked-in configs stay prod-like everywhere and the rollout/startup child is archived; dev-only live-reload implementation evidence exists, but the active plan directory is absent |
| `plans/epics/binance-market-intelligence-gap-closure/` | `in_progress` | Wave 2 and Wave 3 outcomes archived | `plans/epics/binance-environment-config-and-rollout-hardening/` | Run `feature-implementing` for `plans/binance-spot-depth-liquidity-indicators/` | Depth-liquidity plan is ready to implement while alerting remains out of scope |

## Planning Waves

### Wave 1

- `binance-streaming-market-state-runtime-integration`
- Why now: the archived provider cutover still leaves `cmd/market-state-api` on a bounded snapshot seam; every later slice should consume the finished sustained runtime source rather than plan against a temporary local seam.

### Wave 2

- `binance-runtime-health-and-operator-observability`
- `binance-usdm-market-state-influence`
- Why parallel: both consume the sustained runtime integration but solve different problems; one is operator visibility, the other is market-state semantics.

### Wave 3

- `binance-environment-config-and-rollout-hardening`
- Why later: rollout defaults should follow the settled runtime and observability posture rather than invent their own behavior.

### Wave 4

- `binance-market-intelligence-gap-closure`
- Why later: full Binance market-intelligence closure should validate the final runtime shape after the streaming cutover, observability surface, USD-M semantics, and environment defaults are all clear, then add the missing Spot liquidity, trade-flow, and derivatives indicator layers before alerting work starts.

## Refined Epics

### `plans/epics/binance-environment-config-and-rollout-hardening/`

- Problem statement: the Binance runtime now has a settled live shape, but the repo still needs its checked-in configs and rollout docs to preserve one prod-like startup posture everywhere instead of drifting into environment-specific behavior
- In scope: checked-in `local`/`dev`/`prod` config parity with prod-like behavior everywhere, existing override guardrails in `cmd/market-state-api`, and compose plus runbook rollout handoff
- Out of scope: runtime-health surface redesign, USD-M semantics redesign, infrastructure-specific deployment tooling, or long-run soak validation
- Child queue:
  - blocked reconciliation: `plans/binance-live-reload-dev-workflow/`
- Archived child evidence:
  - `plans/completed/binance-runtime-config-profile-parity/`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/`
- Active child plan: none in the current worktree; `plans/binance-live-reload-dev-workflow/` is referenced but absent
- Next recommended action: restore/recreate the dev-only live-reload testing matrix only if archive testing remains desired; otherwise continue with `plans/epics/binance-market-intelligence-gap-closure/`
- Parallel work inside the epic: none; the only remaining work inside this epic is reconciling the missing dev-workflow plan/testing artifact
- Next child to plan: none inside this epic
- Artifact pointers:
  - `plans/epics/binance-environment-config-and-rollout-hardening/00-overview.md`
  - `plans/epics/binance-environment-config-and-rollout-hardening/90-refinement-map.md`
  - `plans/epics/binance-environment-config-and-rollout-hardening/91-child-plan-seeds.md`
  - `plans/epics/binance-environment-config-and-rollout-hardening/92-refinement-handoff.md`
  - `plans/binance-live-reload-dev-workflow/00-overview.md`
  - `plans/binance-live-reload-dev-workflow/04-testing.md`
  - `plans/completed/binance-runtime-config-profile-parity/00-overview.md`
  - `plans/completed/binance-runtime-config-profile-parity/04-testing.md`
  - `plans/completed/binance-runtime-config-profile-parity/testing-report.md`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/00-overview.md`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/04-testing.md`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/testing-report.md`

### `plans/epics/binance-runtime-health-and-operator-observability/`

- Problem statement: current runtime success and failure states are still indirect because the sustained Spot runtime exposes internal health inputs, but operators do not yet have one explicit bounded status surface for warm-up, reconnect, stale, recovery, and rate-limit posture.
- In scope: settle the additive runtime-health boundary, keep `/healthz` process-only, preserve current-state contract stability, and update operator-facing docs plus validation around the chosen status surface.
- Out of scope: changing current-state or regime semantics, environment rollout defaults, long-run soak validation, or dashboard redesign.
- Archived child evidence:
  - `plans/completed/binance-runtime-health-snapshot-owner/`
  - `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/`
- Remaining child queue: none
- Active child plan: none
- Next recommended action: use the archived runtime-status feature as the settled operator runtime-health reference while Wave 3 rollout planning proceeds.
- Parallel implementation inside the initiative: none; this epic is complete.
- Artifact pointers:
  - `plans/epics/binance-runtime-health-and-operator-observability/00-overview.md`
  - `plans/epics/binance-runtime-health-and-operator-observability/90-refinement-map.md`
  - `plans/epics/binance-runtime-health-and-operator-observability/91-child-plan-seeds.md`
  - `plans/epics/binance-runtime-health-and-operator-observability/92-refinement-handoff.md`
  - `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/00-overview.md`
  - `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/04-testing.md`
  - `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/testing-report.md`
  - `plans/completed/binance-runtime-health-snapshot-owner/00-overview.md`
  - `plans/completed/binance-runtime-health-snapshot-owner/testing-report.md`

### `plans/epics/binance-usdm-market-state-influence/`

- Problem statement: USD-M sensors already exist, but the live stack still has no settled, deterministic rule for whether derivatives context should stay auxiliary or influence current-state and regime semantics.
- In scope: define the bounded influence policy, apply it in Go-owned market-state and regime logic, preserve `/api/market-state/*` by default, and expand deterministic replay plus focused API proof.
- Out of scope: new USD-M acquisition work, runtime-health surface design, environment defaults, or long-run soak validation.
- Child queue:
  - none
- Archived child evidence:
  - `plans/completed/binance-usdm-influence-policy-and-signal/`
  - `plans/completed/binance-usdm-output-application-and-replay-proof/`
- Active child plan: none
- Next recommended action: use the archived USD-M output-application feature as the settled semantics reference while Wave 3 planning begins
- Parallel work inside the epic: none; this epic is complete
- Safe parallel implementation outside the epic: none; the runtime-health Wave 2 epic is archived
- Next child to plan: none; the epic is complete
- Artifact pointers:
  - `plans/epics/binance-usdm-market-state-influence/00-overview.md`
  - `plans/epics/binance-usdm-market-state-influence/90-refinement-map.md`
  - `plans/epics/binance-usdm-market-state-influence/91-child-plan-seeds.md`
  - `plans/epics/binance-usdm-market-state-influence/92-refinement-handoff.md`
  - `plans/completed/binance-usdm-influence-policy-and-signal/00-overview.md`
  - `plans/completed/binance-usdm-influence-policy-and-signal/04-testing.md`
  - `plans/completed/binance-usdm-influence-policy-and-signal/testing-report.md`
  - `plans/completed/binance-usdm-output-application-and-replay-proof/00-overview.md`
  - `plans/completed/binance-usdm-output-application-and-replay-proof/04-testing.md`
  - `plans/completed/binance-usdm-output-application-and-replay-proof/testing-report.md`

## Initiative Seeds

### `binance-long-run-runtime-hardening`

- Superseded by `plans/epics/binance-market-intelligence-gap-closure/`.
- The original long-run/failure-path confidence gate is retained as the child seed `binance-live-runtime-soak-and-failure-hardening` inside the broader epic.

### `plans/epics/binance-market-intelligence-gap-closure/`

- Problem statement: finishing Binance integration now requires both runtime confidence and richer market-intelligence indicators; the current stack has live Spot best bid/ask and conservative USD-M caps, but not yet trade-flow, real liquidity scoring, derivatives indicator enrichment, or a green full validation baseline.
- In scope: validation baseline reconciliation, long-run/failure-path hardening, Spot trade-flow feature inputs, Spot depth liquidity indicators, USD-M derivatives indicator enrichment, and additive service-owned indicator readiness surfaces.
- Out of scope: private Binance endpoints, order submission, account state, non-Binance expansion, browser-side venue logic, or Python live-runtime dependencies.
- Target repo areas: `services/venue-binance`, `cmd/market-state-api`, `services/market-state-api`, `services/feature-engine`, `libs/go/features`, `libs/go/ingestion`, `schemas/json/features`, `tests/integration`, `tests/replay`, `tests/fixtures`, `apps/web` only after service-owned contracts settle.
- Contract/fixture/parity/replay implications: enriched indicators must stay additive, deterministic, replayable, and backed by fixture/integration evidence before alerting consumes them.
- Last archived child plan: `plans/completed/binance-spot-trade-flow-feature-inputs/`.
- Active child plan: `plans/binance-spot-depth-liquidity-indicators/` (`ready_to_implement`).
- Next child to implement: `plans/binance-spot-depth-liquidity-indicators/`.
- Next child to plan, only if explicitly prioritized for parallel planning: `binance-usdm-derivatives-indicator-enrichment`.

## Open Questions That Still Matter

- Which optional live or Compose runtime checks should be rerun on a Docker-capable host now that the required deterministic local runtime-hardening matrix is archived?
- Which enriched Binance indicators should be first-class current-state fields versus internal alert-readiness inputs?
