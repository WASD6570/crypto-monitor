# Testing Report: binance-usdm-open-interest-rest-polling

## Environment
- Target: local Go test environment
- Date/time: 2026-03-09
- Commit/branch: working tree on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Happy poll | Parse one Binance `/fapi/v1/openInterest` payload and normalize it | one canonical `open-interest-snapshot` with perpetual provenance | Unit and integration tests passed | PASS |
| Missing exchange time | Normalize an open-interest payload without Binance `time` | canonical event stays emitted with degraded timestamp status and `recvTs` fallback | Unit and integration tests passed | PASS |
| Poll stale path | Advance synthetic time past the freshness ceiling after last successful sample | `STALE` feed-health with `message-stale` | Unit and integration tests passed | PASS |
| Poll rate-limit path | Exceed the configured local open-interest poll budget | `DEGRADED` feed-health with explicit `rate-limit` reason | Venue tests passed | PASS |

## Execution Evidence
### ingestion-normalization
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'Test(LoadEnvironmentConfigParsesLocalRuntimeConfig|Normalize(OpenInterest|Funding|MarkIndex|Liquidation)|RawWriteBoundaryPersistsOpenInterestStreamFamily)' -v`
- Expected: config, open-interest normalization, and raw append partitioning pass
- Actual: all targeted tests passed
- Verdict: PASS

### venue-poller
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestUSDM(OpenInterest|MarkPrice|ForceOrder|SubscriptionShape|WebsocketRuntime)' -v`
- Expected: open-interest parsing and polling health pass without regressing existing USD-M runtime behavior
- Actual: all targeted tests passed
- Verdict: PASS

### normalizer-service
- Command/Request: `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OpenInterest|Funding|MarkIndex|Liquidation)' -v`
- Expected: service-owned normalization preserves perpetual provenance for open-interest and existing USD-M event families
- Actual: all targeted tests passed
- Verdict: PASS

### integration-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM(HappyPath|TimestampDegraded|Reconnect|NoLiquidationStale|OpenInterestTimestampDegraded|OpenInterestPollHealth)' -v`
- Expected: fixture-backed integration proves parser, normalizer, and poll-health behavior together
- Actual: all targeted tests passed
- Verdict: PASS

## Side-Effect Verification
### canonical-open-interest
- Evidence: `TestNormalizeOpenInterest`, `TestServiceNormalizeOpenInterest`, `TestIngestionBinanceUSDMHappyPath`
- Expected state: canonical event keeps `symbol`, `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, and the open-interest value
- Actual state: tests passed with preserved provenance and value
- Verdict: PASS

### degraded-missing-time
- Evidence: `TestNormalizeOpenInterestFallsBackWhenExchangeTimestampMissing`, `TestIngestionBinanceUSDMOpenInterestTimestampDegraded`
- Expected state: missing Binance `time` degrades deterministically while preserving `recvTs`
- Actual state: tests passed with degraded timestamp behavior and stable `oi:` source-record IDs
- Verdict: PASS

### rest-poll-health
- Evidence: `TestUSDMOpenInterestPollerMarksStaleAndRateLimit`, `TestIngestionBinanceUSDMOpenInterestPollHealth`
- Expected state: per-symbol polling emits machine-visible stale and rate-limit degradation instead of log-only signals
- Actual state: tests passed with explicit `message-stale` and `rate-limit` reasons
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current smoke matrix.
- Mixed-surface USD-M validation can now build on explicit websocket plus REST semantics without reopening timestamp or provenance policy.

## Next Actions
1. Plan and implement `binance-usdm-context-sensor-fixtures-and-integration` to prove websocket plus REST sensors together at the fixture/runbook layer.
2. Wire the new open-interest poller into whichever live Binance adapter entrypoint consumes the bounded USD-M runtime slices next.
