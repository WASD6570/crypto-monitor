# Coinbase Adapter Foundation Testing Report

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-coinbase/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./services/venue-coinbase -v`

## Results

Both commands passed on 2026-03-07 in the local workspace.

## Smoke Matrix Evidence

| Case | Evidence |
|---|---|
| Trade parse | `TestParseTradeEventProducesSharedTradeMessage`, `TestParseTradeEventFeedsCanonicalNormalization` |
| Book parse | `TestParseTopOfBookEventFeedsCanonicalNormalizationHappyPath`, `TestParseTopOfBookEventPreservesQuoteVariantMetadata` |
| Runtime health | `TestRuntimeEvaluateLoopStateReturnsHealthyDecision`, `TestRuntimeEvaluateLoopStateReturnsDegradedDecision`, `TestRuntimeEvaluateLoopStateReturnsStaleDecision` |

## Verification Notes

- Coinbase parser boundaries remain venue-local in `services/venue-coinbase`.
- Trade and top-of-book outputs feed the shared ingestion normalization helpers.
- Runtime decisions use the shared `ingestion.FeedHealthStatus` vocabulary for healthy, degraded, and stale states.
