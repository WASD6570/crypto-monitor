# Testing Report: Binance Spot Top Of Book Canonical Handoff

## Result

- Status: passed
- Date: 2026-03-09

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParse(TopOfBookEvent|TopOfBookFrame)|TestSpotWebsocketSupervisor' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(BinanceTopOfBook|OrderBook)' -v`
- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalizeOrderBookMessage|TestRawWriteBoundary|TestBuildRawAppendEntryFromTopOfBook' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(BinanceSpotSupervisor|IngestionBinanceSpotTopOfBook)' -v`
- `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`

## Evidence

- Binance Spot `bookTicker` frames now parse through `ParseTopOfBookEvent(...)` and `ParseTopOfBookFrame(...)` under the completed Spot supervisor seam.
- Shared top-of-book normalization now uses Binance update-sequence-backed identity when exchange time is absent, preserving stable duplicate handling for native Spot top-of-book inputs.
- Raw append coverage confirms accepted Binance Spot top-of-book events keep `streamFamily=top-of-book` and duplicate audit behavior.
- Native Binance top-of-book happy and degraded fixtures validate against the shared contract after widening the `order-book-top` schema for `bookAction=top-of-book` and empty `exchangeTs` fallback.
- Supervisor-backed integration proves happy-path, timestamp-degraded, and duplicate-identity flows without introducing depth bootstrap logic.
