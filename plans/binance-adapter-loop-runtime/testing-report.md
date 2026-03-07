# Binance Adapter Loop Runtime Testing Report

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestAdapterLoopState|TestRuntimeEvaluateLoopState|TestAdapterLoop' -v`

## Results

Both commands passed on 2026-03-07 in the local workspace.

## Smoke Matrix Evidence

| Case | Evidence |
|---|---|
| Healthy loop | `TestRuntimeEvaluateLoopStateReturnsHealthyDecision`, `TestAdapterLoopDecisionTracksHealthyDegradedStaleAndRecoveryTransitions` |
| Gap degradation | `TestRuntimeEvaluateLoopStateTracksDegradedTransition`, `TestAdapterLoopDecisionTracksHealthyDegradedStaleAndRecoveryTransitions` |
| Stale transition | `TestRuntimeEvaluateLoopStateTracksStaleTransition`, `TestAdapterLoopDecisionTracksHealthyDegradedStaleAndRecoveryTransitions` |
| Recovery reset | `TestRuntimeEvaluateLoopStateRecoversAfterClearingGapAndResettingCounters`, `TestAdapterLoopDecisionTracksHealthyDegradedStaleAndRecoveryTransitions` |
| Snapshot recovery window | `TestAdapterLoopStatePruneSnapshotRecoveryHistoryRemovesAttemptsOutsideWindow`, `TestAdapterLoopDecisionMatchesRuntimeHelpers` |

## Verification Notes

- Loop-state helpers now cover direct mutation for connection state, message/snapshot timestamps, sequence gaps, reconnect/resync counters, and snapshot recovery history.
- Decision outputs remain `ingestion.FeedHealthStatus` values built through the shared runtime helpers.
- No live websocket, REST, goroutine, or channel dependency was introduced in this slice.
