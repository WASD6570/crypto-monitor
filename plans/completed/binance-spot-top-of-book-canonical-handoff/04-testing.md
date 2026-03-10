# Testing: Binance Spot Top Of Book Canonical Handoff

## Smoke Matrix

| Case | Flow | Expected |
|---|---|---|
| Happy top-of-book handoff | accepted Binance Spot `bookTicker` payload parses and normalizes through the shared Spot top-of-book seam | one canonical `order-book-top` event with Spot provenance and `bookAction=top-of-book` |
| Missing exchange time fallback | native `bookTicker` payload lacks a trustworthy exchange timestamp | canonical top-of-book event still emits with degraded timestamp status and visible fallback reason |
| Stable source identity | same accepted top-of-book update is processed more than once | chosen `sourceRecordId` rule stays stable and raw duplicate facts remain predictable |
| Stream-family routing | accepted Binance top-of-book output is written through raw append | raw append entry uses `streamFamily=top-of-book`, not depth/order-book |
| Supervisor seam proof | completed Spot supervisor hands an accepted `bookTicker` frame into the parser/normalizer path | no second lifecycle owner or depth bootstrap logic is introduced |

Expected output artifact: `plans/binance-spot-top-of-book-canonical-handoff/testing-report.md`

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParse(TopOfBookEvent|TopOfBookFrame)|TestSpotWebsocketSupervisor' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(BinanceTopOfBook|OrderBook)' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalizeOrderBookMessage|TestRawWriteBoundary' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(BinanceSpotSupervisor|IngestionBinanceSpotTopOfBook)' -v`

## Verification Checklist

- accepted Spot top-of-book messages preserve `BTC-USD`/`ETH-USD`, `BTCUSDT`/`ETHUSDT`, `quoteCurrency=USDT`, `venue=BINANCE`, and `marketType=spot`
- canonical output remains `order-book-top` with `bookAction=top-of-book`, distinct from later depth snapshot/delta work
- timestamp-degraded cases remain machine-visible when `bookTicker` lacks or fails exchange-time validation
- repeated accepted top-of-book inputs preserve the chosen stable source identity and expected raw duplicate behavior
- supervisor-driven stale/reconnect degradation remains covered by supervisor/runtime tests instead of being duplicated inside the top-of-book parser feature
- output evidence is recorded in `plans/binance-spot-top-of-book-canonical-handoff/testing-report.md` while active, then archived with the completed plan directory
