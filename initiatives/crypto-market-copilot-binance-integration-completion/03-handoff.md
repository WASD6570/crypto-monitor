# Binance Integration Completion Handoff

## Refined Epic Queue

- `plans/epics/binance-environment-config-and-rollout-hardening/` has a new bounded child ready for implementation: `plans/binance-live-reload-dev-workflow/`.

## Execution State

- Initiative status: `in_progress`
- Completed prerequisite epic context: `plans/epics/binance-streaming-market-state-runtime-integration/` (historical reference only)
- Next recommended execution step: run `feature-testing` for `plans/binance-live-reload-dev-workflow/`
- Parallel-safe step after that starts: `binance-long-run-runtime-hardening` can still move through `program-refining`

| Item | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/binance-runtime-health-and-operator-observability/` | `archived` | `plans/completed/binance-runtime-health-snapshot-owner/` | - | Use archived child evidence as the settled operator runtime-health surface | The runtime-status endpoint and ops-handoff child is complete and archived |
| `plans/epics/binance-usdm-market-state-influence/` | `archived` | Wave 1 complete and `plans/epics/binance-usdm-context-sensors/` | - | Use archived USD-M child evidence as the settled market-state semantics reference | Both child plans are complete and archived |
| `plans/epics/binance-environment-config-and-rollout-hardening/` | `in_progress` | Wave 2 decisions are now settled | `binance-long-run-runtime-hardening` | Run `feature-testing` for `plans/binance-live-reload-dev-workflow/` | Checked-in configs stay prod-like everywhere, the rollout/startup child is archived, and the bounded dev-only live-reload follow-up is now implemented and reviewed |
| `binance-long-run-runtime-hardening` | `ready_to_refine` | Wave 2 and Wave 3 outcomes archived | `plans/epics/binance-environment-config-and-rollout-hardening/` | Run `program-refining` when planning capacity exists | Wave 4 seed is still unblocked |

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

- `binance-long-run-runtime-hardening`
- Why later: long-run/failure hardening should validate the final runtime shape after the streaming cutover, observability surface, USD-M semantics, and environment defaults are all clear.

## Refined Epics

### `plans/epics/binance-environment-config-and-rollout-hardening/`

- Problem statement: the Binance runtime now has a settled live shape, but the repo still needs its checked-in configs and rollout docs to preserve one prod-like startup posture everywhere instead of drifting into environment-specific behavior
- In scope: checked-in `local`/`dev`/`prod` config parity with prod-like behavior everywhere, existing override guardrails in `cmd/market-state-api`, and compose plus runbook rollout handoff
- Out of scope: runtime-health surface redesign, USD-M semantics redesign, infrastructure-specific deployment tooling, or long-run soak validation
- Child queue:
  - active: `plans/binance-live-reload-dev-workflow/`
- Archived child evidence:
  - `plans/completed/binance-runtime-config-profile-parity/`
  - `plans/completed/binance-rollout-compose-and-ops-handoff/`
- Active child plan: `plans/binance-live-reload-dev-workflow/`
- Next recommended action: run `feature-testing` for the dev-only live-reload workflow without changing the default prod-like Compose path
- Parallel work inside the epic: none after the active child starts; the only remaining work inside this epic is the active dev-workflow child
- Next child to plan: none beyond the active child
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

- Problem statement: finishing the integration requires confidence that the final runtime survives reconnects, stale periods, and repeated validation without semantic drift.
- In scope: long-run/failure-path checks, reconnect/rate-limit/staleness validation, and final replay/current-state regression coverage for the settled runtime.
- Out of scope: standalone smoke-only planning or unrelated observability platform work.
- Target repo areas: `tests/integration`, `tests/replay`, `services/venue-binance`, `services/market-state-api`
- Contract/fixture/parity/replay implications: this is the final confidence gate for runtime determinism and operator trust.
- Likely validation shape: repeated runtime tests, replay checks, focused live-path failure simulations, and post-cutover compose verification.

## Open Questions That Still Matter

- How much of long-run hardening belongs in CI versus a documented manual/operator validation flow?
