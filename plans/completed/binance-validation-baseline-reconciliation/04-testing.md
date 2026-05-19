# Testing Plan

## Required Validation Matrix

| Area | Command | Expected Result |
|---|---|---|
| Focused Binance runtime drift | `go test ./services/venue-binance -run 'TestRuntimeReconnectDelayUsesBinanceConfigBounds|TestRuntimeReconnectDelayClampsAtConfiguredMaximum|TestRuntimeSnapshotRecoveryStatusReportsRemainingCooldown|TestRuntimeSnapshotRecoveryStatusAllowsRetryAfterCooldown|TestRuntimeEvaluateReconnectLoopReturnsNormalBelowThreshold|TestRuntimeEvaluateResyncLoopReturnsNormalBelowThreshold|TestRuntimeAdapterHealthSnapshotComposesDegradedStatuses|TestSpotDepthRecoveryOwnerMarksSequenceGapAndBlocksOnCooldown|TestSpotDepthRecoveryOwnerRecoversWithReplacementSnapshot'` | Failing runtime and depth recovery tests pass with deterministic timestamps |
| Full Go baseline | `go test ./...` | All Go packages pass |
| Contract families | `make contracts-validate` | Replay schema and family validation pass without unrelated validator weakening |
| Contract consumers | `pnpm test contracts` | TypeScript contract checks remain green if available in the repo scripts |
| Fixture structure | `make fixtures-validate` | Fixture and replay seed JSON remains valid |
| Replay determinism | `CONTRACT_FIXTURES=1 make replay-smoke` | All replay seeds match materialized fixture ordering and stable checksum checks pass |

## Not Required For This Feature

- `pnpm --dir apps/web test` and `pnpm --dir apps/web build` are not required unless implementation unexpectedly touches `apps/web`.
- Compose validation is not a required pass gate in the current WSL environment because Docker is unavailable; record it as unavailable if rerun.
- No live Binance credentials or private endpoint checks are required.

## Critical Negative Cases

- Runtime tests must still reject invalid reconnect attempts, negative reconnect/resync counts, future snapshot attempt timestamps, and missing current times.
- Contract validation must still fail for schemas missing required top-level schema keys, required field declarations, or property definitions.
- Replay smoke must still fail when seed event counts or ordered source-record IDs do not match referenced fixture canonical output.

## Idempotency And Determinism Checks

- Rerunning `go test ./services/venue-binance -run ...` should produce the same result because all timestamps are deterministic.
- Rerunning `CONTRACT_FIXTURES=1 make replay-smoke` should produce the same ordered IDs and checksum outcomes for unchanged fixture inputs.
- Reconciliation must not depend on wall-clock time, live Binance responses, Docker availability, or browser execution.

## Expected Side Effects

- Test expectations, validator policy, schema files, replay seed expectations, and planning docs may change.
- Live runtime behavior should not change unless implementation proves existing runtime code is inconsistent with config semantics.
- No persisted database state, external API state, or live exchange state is mutated.

## Testing Report

Implementation and feature-testing evidence is archived at:

```text
plans/completed/binance-validation-baseline-reconciliation/testing-report.md
```

The report includes commands run, pass/fail status, key output snippets, unavailable checks, and residual risk.

## Archive Intent

Implementation passed the validation matrix and the follow-up `feature-testing` run is complete. The full active plan directory and `testing-report.md` were moved to:

```text
plans/completed/binance-validation-baseline-reconciliation/
```

`plans/STATE.md`, `plans/epics/binance-market-intelligence-gap-closure/92-refinement-handoff.md`, and `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` now point to `binance-live-runtime-soak-and-failure-hardening` as the next child feature.
