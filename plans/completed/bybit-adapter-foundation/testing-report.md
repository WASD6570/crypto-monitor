# Bybit Adapter Foundation Testing Report

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-bybit/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./services/venue-bybit -v`

## Results

Both commands passed on 2026-03-07 in the local workspace.

## Smoke Matrix Evidence

| Case | Evidence |
|---|---|
| Trade parse | `TestParseTradeEventFeedsCanonicalNormalization` |
| Book parse | `TestParseOrderBookEventFeedsCanonicalNormalizationHappyPath`, `TestParseOrderBookEventFeedsDeterministicGapDegradation` |
| Runtime health | `TestRuntimeEvaluateLoopStateReturnsHealthyDecision`, `TestRuntimeEvaluateLoopStateReturnsDegradedDecision`, `TestRuntimeEvaluateLoopStateReturnsStaleDecision` |

## Verification Notes

- Bybit parser boundaries stay venue-local in `services/venue-bybit`.
- Trade and order-book messages reuse shared normalization in `libs/go/ingestion`.
- Runtime decisions use the shared `ingestion.FeedHealthStatus` vocabulary for healthy, degraded, and stale states.
