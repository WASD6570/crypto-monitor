# Testing: Binance USD-M Open Interest Rest Polling

## Smoke Matrix

| Case | Flow | Expected |
|---|---|---|
| Happy poll | parse one Binance REST open-interest payload and normalize it | one canonical `open-interest-snapshot` event with perpetual provenance |
| Missing exchange time | normalize a payload without trustworthy `time` | canonical event remains emitted with degraded timestamp status and `recvTs` fallback |
| Poll stale path | advance synthetic time past the configured poll freshness ceiling | feed-health becomes `STALE` with `message-stale` |
| Poll rate-limit path | exceed the configured per-minute REST poll budget | feed-health becomes `DEGRADED` with explicit rate-limit reason |

## Commands

- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalize(OpenInterest|Funding|MarkIndex|Liquidation)' -v`
- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestUSDMOpenInterest|TestUSDMOpenInterestPoller' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OpenInterest|Funding|MarkIndex|Liquidation)' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM(OpenInterest|OpenInterestTimestampDegraded|OpenInterestPollHealth)' -v`

## Verification Checklist

- canonical output keeps `BTC-USD`/`ETH-USD`, `sourceSymbol`, `quoteCurrency`, `venue=BINANCE`, and `marketType=perpetual`
- open-interest payload value is preserved in the canonical event and shared schema expectations
- missing exchange time degrades deterministically without dropping the sample
- feed-health exposes stale polling and rate-limit pressure without relying on logs
- output evidence is recorded in `plans/binance-usdm-open-interest-rest-polling/testing-report.md` while active, then archived with the completed plan directory
