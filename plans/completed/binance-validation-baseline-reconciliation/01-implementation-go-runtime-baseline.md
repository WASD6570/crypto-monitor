# Implementation: Go Runtime Baseline

## Requirements And Scope

- Reconcile failing Binance runtime and depth recovery tests with current config-driven behavior.
- Prefer updating stale tests over changing runtime code when the runtime already follows the checked-in config.
- Do not change checked-in `configs/*/ingestion.v1.json` unless the implementation proves those config values are the source of the defect.
- Keep runtime-health states machine-readable and preserve the existing `/api/runtime-status` boundary.

## Target Files

- `services/venue-binance/runtime_test.go`
- `services/venue-binance/spot_depth_recovery_test.go`
- `services/venue-binance/runtime.go` only if implementation proves runtime logic is wrong
- `services/venue-binance/spot_depth_recovery.go` only if implementation proves depth recovery logic is wrong
- `configs/local/ingestion.v1.json`, `configs/dev/ingestion.v1.json`, and `configs/prod/ingestion.v1.json` as read-only source-of-truth checks unless a deliberate config correction is required

## Current Failing Areas

- `TestRuntimeReconnectDelayUsesBinanceConfigBounds`
- `TestRuntimeReconnectDelayClampsAtConfiguredMaximum`
- `TestRuntimeSnapshotRecoveryStatusReportsRemainingCooldown`
- `TestRuntimeSnapshotRecoveryStatusAllowsRetryAfterCooldown`
- `TestRuntimeEvaluateReconnectLoopReturnsNormalBelowThreshold`
- `TestRuntimeEvaluateResyncLoopReturnsNormalBelowThreshold`
- `TestRuntimeAdapterHealthSnapshotComposesDegradedStatuses`
- `TestSpotDepthRecoveryOwnerMarksSequenceGapAndBlocksOnCooldown`
- `TestSpotDepthRecoveryOwnerRecoversWithReplacementSnapshot`

## Implementation Steps

1. Confirm the intended Binance config values across `configs/local`, `configs/dev`, and `configs/prod` before editing tests.
2. Update reconnect-delay tests to assert the current `ReconnectBackoffMin` and `ReconnectBackoffMax` behavior or derive expected values from the loaded config.
3. Update reconnect-loop and resync-loop tests to assert the current `ReconnectLoopThreshold` and `ResyncLoopThreshold` behavior or derive expected values from the loaded config.
4. Update snapshot cooldown tests to drive timestamps from `runtime.config.SnapshotCooldown` instead of stale hard-coded cooldown assumptions.
5. For depth recovery tests whose purpose is recovery mechanics, either advance test timestamps past the configured cooldown or use an explicit test-local cooldown override.
6. Run the focused runtime/depth test command, then run `go test ./...` after the contract and replay slices are complete.

## Unit Test Expectations

- Reconnect backoff tests cover non-zero attempts, exponential growth, and clamping at the configured maximum.
- Cooldown tests cover first attempt ready, in-cooldown blocked, and after-cooldown ready states using deterministic timestamps.
- Reconnect and resync loop tests cover below-threshold, at-threshold, above-threshold, and invalid negative counts.
- Depth recovery tests keep sequence gaps visible while blocked and clear them only after successful replacement snapshot recovery.

## Next-Agent Summary

The likely minimal patch is in tests: stale expectations should follow the current prod-like config. Touch runtime code only if focused tests expose behavior that contradicts `ingestion.VenueRuntimeConfig` semantics.
