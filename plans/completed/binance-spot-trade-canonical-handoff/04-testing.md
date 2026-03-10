# Testing: Binance Spot Trade Canonical Handoff

## Smoke Matrix

| Case | Flow | Expected |
|---|---|---|
| Happy trade handoff | accepted Binance Spot `trade` payload parses and normalizes through the shared Spot trade seam | one canonical `market-trade` event with Spot provenance |
| Trade-time fallback | Binance payload omits or zeroes trade time but keeps event time | adapter chooses event time before normalization without losing provenance |
| Timestamp degradation | selected exchange time is implausible versus `recvTs` | canonical trade remains emitted with degraded timestamp status |
| Duplicate-sensitive identity | same Spot trade is accepted more than once | stable `sourceRecordId` and expected raw duplicate facts |
| Supervisor seam proof | completed Spot supervisor hands an accepted `trade` frame into the parser/normalizer path | no second lifecycle owner is introduced |

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParseTradeEvent|TestSpotWebsocketSupervisor' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalizeTrade' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalizeTrade|TestRawWriteBoundary' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(BinanceSpotSupervisor|IngestionAdapterHappyPath|IngestionBinanceSpotTrade)' -v`

## Verification Checklist

- accepted Spot trades preserve `BTC-USD`/`ETH-USD`, `BTCUSDT`/`ETHUSDT`, `quoteCurrency=USDT`, `venue=BINANCE`, and `marketType=spot`
- native trade payloads use Binance trade time first and event time second before canonical timestamp policy is applied
- degraded timestamp cases stay machine-visible and do not drop the trade
- repeated Spot trade inputs preserve stable source identity and raw duplicate expectations
- output evidence is recorded in `plans/binance-spot-trade-canonical-handoff/testing-report.md` while active, then archived with the completed plan directory
