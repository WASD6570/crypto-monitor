# Testing Report: binance-usdm-mark-funding-index-and-liquidation-runtime

## Environment
- Target: local Go test environment
- Date/time: 2026-03-09
- Commit/branch: `23eb71a2eda558b37353a0dea66c21d0a882ad4d` on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Happy mark/index + funding | Parse one Binance `markPrice@1s` payload and normalize derived funding + mark-index events | one canonical `funding-rate` and one canonical `mark-index` event with perpetual provenance | Unit and integration tests passed | PASS |
| Happy liquidation | Parse one Binance `forceOrder` payload and normalize one liquidation event | one canonical `liquidation-print` event with stable source-record identity | Unit and integration tests passed | PASS |
| Degraded timestamp fallback | Normalize a skewed `markPrice@1s` payload | canonical event remains emitted with degraded timestamp status and fallback reason | Unit and integration tests passed | PASS |
| Websocket stale path | Advance synthetic time beyond `messageStaleAfterMs` after last `markPrice@1s` frame | feed-health becomes `STALE` with `message-stale` | Runtime tests passed | PASS |
| Sparse liquidation does not stale | Keep `forceOrder` sparse while `markPrice@1s` stays fresh | feed-health remains healthy | Runtime and integration tests passed | PASS |
| Reconnect and resubscribe degradation | Simulate repeated disconnect/reconnect attempts | bounded reconnect behavior remains visible through degraded feed-health | Runtime and integration tests passed | PASS |
| Duplicate identity stability | Reprocess the same raw derivatives payload | stable canonical source-record IDs across repeats | Normalization tests passed | PASS |

## Execution Evidence
### venue-runtime
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestUSDM(MarkPrice|ForceOrder|Runtime)' -v`
- Expected: USD-M parser and runtime tests pass for mark price, liquidation, staleness, reconnect, and subscription shape
- Actual: all targeted tests passed
- Verdict: PASS

### ingestion-normalization
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalize(Funding|MarkIndex|Liquidation)|TestResolveCanonicalTimestamp' -v`
- Expected: shared derivatives normalization and timestamp fallback remain deterministic
- Actual: all targeted tests passed
- Verdict: PASS

### normalizer-service
- Command/Request: `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(Funding|MarkIndex|Liquidation)' -v`
- Expected: service-owned normalization preserves perpetual provenance for funding, mark/index, and liquidation events
- Actual: all targeted tests passed
- Verdict: PASS

### integration-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM(HappyPath|TimestampDegraded|Reconnect|NoLiquidationStale)' -v`
- Expected: fixture-backed integration proves parser + normalizer + runtime health behavior together
- Actual: all targeted tests passed
- Verdict: PASS

## Side-Effect Verification
### canonical-provenance
- Evidence: `TestIngestionBinanceUSDMHappyPath`, `TestServiceNormalizeFunding`, `TestServiceNormalizeMarkIndex`, `TestServiceNormalizeLiquidation`
- Expected state: canonical events keep `symbol`, `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType=perpetual`
- Actual state: all targeted tests passed with preserved provenance
- Verdict: PASS

### degraded-timestamp-behavior
- Evidence: `TestNormalizeFundingFallsBackWhenExchangeTimestampSkews`, `TestIngestionBinanceUSDMTimestampDegraded`
- Expected state: skewed exchange time degrades deterministically without losing `recvTs`
- Actual state: all targeted tests passed with degraded timestamp status and fallback reason preserved
- Verdict: PASS

### usdm-health-boundary
- Evidence: `TestUSDMWebsocketRuntimeUsesMarkPriceForFreshnessAndNotForceOrderSparsity`, `TestUSDMWebsocketRuntimeTracksReconnectLoopAndFeedHealthInputs`, `TestIngestionBinanceUSDMNoLiquidationStale`, `TestIngestionBinanceUSDMReconnect`
- Expected state: `markPrice@1s` drives freshness; sparse liquidation traffic does not create false stale alarms; reconnect loops remain machine-visible
- Actual state: all targeted tests passed with the expected health semantics
- Verdict: PASS

### fixture-corpus
- Evidence: `tests/fixtures/events/binance/BTC-USD/happy-funding-usdt.fixture.v1.json`, `tests/fixtures/events/binance/ETH-USD/happy-mark-index-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/happy-liquidation-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/edge-timestamp-degraded-funding-usdt.fixture.v1.json`, and `tests/fixtures/manifest.v1.json`
- Expected state: Binance USD-M fixtures exist for happy funding, happy mark-index, happy liquidation, and degraded timestamp coverage
- Actual state: fixture corpus and manifest entries were added and exercised by integration tests
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current smoke matrix.
- `openInterest` REST polling remains intentionally out of scope and still needs its own feature before the full USD-M context surface is complete.

## Next Actions
1. Run `feature-testing` for this feature after any follow-up implementation adjustments.
2. Implement `binance-usdm-open-interest-rest-polling` so the remaining USD-M context sensor arrives behind a separate, explicit freshness model.
