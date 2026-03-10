# Testing: Binance USD-M Context Sensor Fixtures And Integration

## Smoke Matrix

| Case | Flow | Expected |
|---|---|---|
| Mixed-surface happy path | normalize websocket-derived funding/mark/liquidation plus REST open-interest fixtures together | canonical events preserve provenance and fixture expectations across both acquisition modes |
| Distinct health visibility | websocket stays fresh while REST polling goes stale for the same symbol | canonical feed-health outputs remain separate and machine-readable |
| Mixed degradation visibility | websocket reconnect degradation and REST rate-limit degradation are both representable | degradation reasons do not collapse into one generic sensor state |
| Replay-sensitive identity | repeat websocket and REST normalization with raw writing enabled | stable `sourceRecordId` values and duplicate facts per stream family |

## Commands

- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestion(BinanceUSDM|RunbookAlignmentUsesSharedHealthVocabulary)' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OpenInterest|Funding|MarkIndex|Liquidation)' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestRawWriteBoundaryPersistsOpenInterestStreamFamily' -v`

## Verification Checklist

- fixture corpus covers the mixed-surface Binance USD-M paths referenced by the integration smoke
- runbooks mention the shared health state names and the settled `rate-limit` reason
- websocket and REST feed-health outputs stay distinct through canonical normalization
- repeated inputs preserve stable source identities and duplicate audit behavior
- output evidence is recorded in `plans/binance-usdm-context-sensor-fixtures-and-integration/testing-report.md` while active, then archived with the completed plan directory
