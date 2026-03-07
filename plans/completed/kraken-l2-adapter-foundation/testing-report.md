# Kraken L2 Adapter Foundation Testing Report

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-kraken/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./services/venue-kraken -v`

## Results

Both commands passed on 2026-03-07 in the local workspace.

## Smoke Matrix Evidence

| Case | Evidence |
|---|---|
| Trade parse | `TestParseTradeEventFeedsCanonicalNormalization` |
| L2 happy path | `TestParseOrderBookEventFeedsCanonicalNormalizationHappyPath` |
| L2 gap path | `TestParseOrderBookEventMarksGapAsResync` |
| Runtime health | `TestRuntimeEvaluateLoopStateReturnsHealthyDecision`, `TestRuntimeEvaluateLoopStateReturnsDegradedDecision`, `TestRuntimeEvaluateLoopStateReturnsStaleDecision` |

## Verification Notes

- Kraken trade and L2 parser boundaries stay venue-local in `services/venue-kraken`.
- L2 integrity handling is explicit through `L2IntegrityState`, which hard-switches to resync on sequence uncertainty.
- Runtime decisions use shared `ingestion.FeedHealthStatus` output and surface L2 integrity loss through shared degradation reasons.
