# Testing Report

## Summary

- Status: feature-testing passed; archive is complete.
- Scope: Binance validation baseline reconciliation only; no enriched indicator, alerting, dashboard, live runtime, or schema-shape migration behavior was added.
- Residual state: archived under `plans/completed/binance-validation-baseline-reconciliation/`; next Binance child is `binance-live-runtime-soak-and-failure-hardening`.

## Changes Validated

- Reconciled Binance and venue runtime tests with checked-in prod-like config values instead of stale hard-coded reconnect and cooldown expectations.
- Reconciled replay run-result contract-family validation with the existing `runId`, `status`, `inputCounters`, `startedAt`, `finishedAt`, and `manifestDigest` schema/Go producer shape.
- Reconciled replay seed deterministic ordering with exact fixture canonical `sourceRecordId` values.

## Feature-Testing Evidence

The required smoke matrix was rerun on 2026-05-11.

| Command | Status | Evidence |
|---|---|---|
| `go test ./services/venue-binance -run 'TestRuntimeReconnectDelayUsesBinanceConfigBounds|TestRuntimeReconnectDelayClampsAtConfiguredMaximum|TestRuntimeSnapshotRecoveryStatusReportsRemainingCooldown|TestRuntimeSnapshotRecoveryStatusAllowsRetryAfterCooldown|TestRuntimeEvaluateReconnectLoopReturnsNormalBelowThreshold|TestRuntimeEvaluateResyncLoopReturnsNormalBelowThreshold|TestRuntimeAdapterHealthSnapshotComposesDegradedStatuses|TestSpotDepthRecoveryOwnerMarksSequenceGapAndBlocksOnCooldown|TestSpotDepthRecoveryOwnerRecoversWithReplacementSnapshot'` | Passed | `ok github.com/crypto-market-copilot/alerts/services/venue-binance (cached)` |
| `go test ./...` | Passed | All Go packages passed from cache, including `services/venue-binance`, venue peers, `tests/integration`, `tests/parity`, and `tests/replay`. |
| `make contracts-validate` | Passed | `Contract family manifests and docs validate successfully.` |
| `pnpm test contracts` | Passed | `TypeScript runtime contract tests passed.` |
| `make fixtures-validate` | Passed | `Fixture corpus and replay seeds validate successfully.` |
| `CONTRACT_FIXTURES=1 make replay-smoke` | Passed | `Replay smoke checks passed.` |

## Tested Flows And Inputs

- Runtime drift: config-derived reconnect backoff, loop thresholds, snapshot cooldown, and depth recovery cooldown behavior.
- Contract baseline: replay run-result required-field policy aligned with the existing schema and Go producer shape.
- Fixture replay ordering: seed `expectedDeterminism.orderedSourceRecordIds` matched fixture canonical `sourceRecordId` values.
- No `test-helpers/` assets were present under the active plan directory; the plan files and checked-in configs, schemas, fixtures, and scripts supplied all inputs.

## Determinism And Idempotency

- Runtime and integration tests now derive expected backoff, threshold, and cooldown values from deterministic checked-in configs or explicit test-local overrides.
- Replay smoke checks compute the same materialized ordering checksum twice for each seed in one run.
- No validation depends on live Binance responses, wall-clock time, Docker, browser execution, or Python in the live runtime path.

## Side Effects And State

- No persisted database records, exchange state, browser state, or live runtime side effects were created.
- `plans/STATE.md`, `plans/epics/binance-market-intelligence-gap-closure/92-refinement-handoff.md`, and `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` were updated to mark this feature archived and point to `binance-live-runtime-soak-and-failure-hardening` as the next child.

## Notes

- The first full `go test ./...` rerun exposed additional stale config-driven expectations outside the initial focused test list; those were reconciled in the same baseline-validation slice and then `go test ./...` passed.
- Compose validation was not part of this feature's required matrix and was not rerun here; prior environment context indicated Docker is unavailable in this WSL environment.
- No blockers remain for this baseline feature.
