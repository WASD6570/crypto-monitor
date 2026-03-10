# Testing: Binance Spot Depth Bootstrap And Buffering

## Validation Matrix

| Case | Flow | Expected |
|---|---|---|
| Happy bootstrap alignment | buffered `depthUpdate` frames + one `/api/v3/depth` snapshot produce a bridging first delta | synchronized snapshot and accepted deltas normalize into canonical order-book output |
| Stale buffered deltas | buffered deltas finish at or before snapshot `lastUpdateId` | stale deltas are ignored until a bridging delta is found |
| Missing bridging delta | snapshot arrives but no buffered delta window bridges the snapshot boundary | bootstrap fails explicitly and does not emit synchronized depth output |
| Canonical handoff | synchronized snapshot/delta outputs reach the shared sequencer and normalizer | asset-centric symbol, explicit provenance, and strict timestamp fallback remain intact |
| Runtime boundary preservation | completed Spot supervisor remains the generic websocket lifecycle owner | no second connection manager or non-depth lifecycle duplication is introduced |

Expected output artifact: `plans/binance-spot-depth-bootstrap-and-buffering/testing-report.md`

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParseOrderBook|TestSpotDepthBootstrap' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestOrderBookSequencer|TestNormalizeOrderBookMessage' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalizeOrderBook' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepthBootstrap' -v`
- `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`

## Verification Checklist

- startup alignment uses explicit Binance snapshot and delta sequence semantics rather than ad hoc local ordering
- synchronized bootstrap output preserves `BTC-USD`/`ETH-USD`, `BTCUSDT`/`ETHUSDT`, `quoteCurrency=USDT`, `venue=BINANCE`, and `marketType=spot`
- REST snapshots keep explicit degraded timestamp fallback through `recvTs` instead of synthetic exchange time
- failure to find a bridging delta does not emit accepted depth output or silently clear recovery needs
- validation evidence is recorded in `plans/binance-spot-depth-bootstrap-and-buffering/testing-report.md` while active, then archived with the completed plan directory
