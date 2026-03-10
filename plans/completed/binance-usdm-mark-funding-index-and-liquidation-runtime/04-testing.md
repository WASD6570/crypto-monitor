# Testing Plan: Binance USD-M Mark Funding Index And Liquidation Runtime

Expected output artifact: `plans/binance-usdm-mark-funding-index-and-liquidation-runtime/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Happy mark/index + funding | Parse one Binance `markPrice@1s` payload for BTC or ETH, normalize both derived events | one `funding-rate` and one `mark-index` canonical event with `marketType=perpetual`, preserved provenance, and normal timestamp status | package unit tests + fixture-backed integration output |
| Happy liquidation | Parse one Binance `forceOrder` payload, normalize one liquidation event | one `liquidation-print` canonical event with stable source-record ID and preserved source symbol | package unit tests + fixture-backed integration output |
| Degraded timestamp fallback | Feed `markPrice@1s` payload with missing, invalid, or skewed exchange time | canonical events are still emitted with degraded timestamp status and fallback reason preserved | unit tests + integration output |
| Websocket stale path | Advance synthetic time beyond `messageStaleAfterMs` after last `markPrice@1s` update | feed-health becomes `STALE` with `message-stale` and no contract ambiguity | runtime unit tests |
| Sparse liquidation does not stale | Run runtime without `forceOrder` traffic while `markPrice@1s` remains fresh | feed-health stays healthy or only reflects non-liquidation reasons | runtime unit tests |
| Reconnect and resubscribe degradation | Simulate disconnect/reconnect loop and resubscribe path | bounded reconnect behavior stays within config and degraded reasons remain machine-visible | runtime unit tests |
| Duplicate identity stability | Reprocess identical raw `markPrice@1s` or `forceOrder` payload | derived canonical source-record IDs are identical across repeats | parser/normalizer unit tests |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestUSDM(MarkPrice|ForceOrder|Runtime)' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalize(Funding|MarkIndex|Liquidation)|TestResolveCanonicalTimestamp' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(Funding|MarkIndex|Liquidation)' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM(HappyPath|TimestampDegraded|Reconnect|NoLiquidationStale)' -v`

## Verification Checklist

- No REST `openInterest` logic or polling tests are introduced in this feature.
- All canonical events keep `symbol`, `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` explicit.
- `markPrice@1s` drives freshness expectations; `forceOrder` sparsity does not create false stale alarms.
- Connection and reconnect degradation are visible through feed-health output, not only logs.
- Timestamp fallback stays deterministic and preserves both `exchangeTs` and `recvTs`.
- Validation evidence is written to `plans/binance-usdm-mark-funding-index-and-liquidation-runtime/testing-report.md` before archive handoff.
