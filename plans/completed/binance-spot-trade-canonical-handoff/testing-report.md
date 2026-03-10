# Testing Report: binance-spot-trade-canonical-handoff

## Environment
- Target: local Go test environment
- Date/time: 2026-03-09
- Commit/branch: working tree on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Happy trade handoff | accepted Binance Spot `trade` frame parses through the completed supervisor seam and normalizes to canonical output | one canonical `market-trade` event with Spot provenance | Unit and integration tests passed | PASS |
| Trade-time fallback | native Binance trade omits trade time but keeps event time | adapter falls back to event time deterministically before normalization | Unit tests passed | PASS |
| Timestamp degradation | native Binance trade carries an implausible exchange time versus `recvTs` | canonical trade remains emitted with degraded timestamp status | Service and integration tests passed | PASS |
| Duplicate-sensitive identity | repeated accepted Spot trade is normalized with raw writing enabled | stable `sourceRecordId` and duplicate audit facts | Service and integration tests passed | PASS |
| Supervisor seam proof | completed Spot supervisor feeds an accepted trade frame into the parser path | no second lifecycle owner is introduced | Unit and integration tests passed | PASS |

## Execution Evidence
### venue-parser-and-supervisor
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParseTradeEvent|TestParseTradeFrame|TestSpotWebsocketSupervisor' -v`
- Expected: native Spot trade parsing and the supervisor-fed frame seam both pass without widening the runtime scope
- Actual: all targeted tests passed
- Verdict: PASS

### normalizer-handoff
- Command/Request: `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(Trade|BinanceTrade)|TestNormalizerRawWriteBoundary(PreservesBinanceTradeDuplicateIdentity)?' -v`
- Expected: Binance Spot trade handoff preserves Spot provenance, timestamp behavior, and duplicate-sensitive raw compatibility
- Actual: all targeted tests passed
- Verdict: PASS

### ingestion-boundary
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestNormalizeTradeMessage|TestRawWriteBoundaryRecordsDuplicateAuditFacts' -v`
- Expected: shared trade normalization and duplicate audit boundary remain compatible with the Binance Spot handoff
- Actual: all targeted tests passed
- Verdict: PASS

### integration-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(BinanceSpotSupervisor|IngestionBinanceSpotTrade)' -v`
- Expected: completed supervisor, native trade parser, and canonical normalization work together through fixture-backed integration
- Actual: all targeted tests passed
- Verdict: PASS

### fixture-manifest
- Command/Request: `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`
- Expected: new native Binance Spot trade fixtures load and validate through the shared fixture manifest
- Actual: targeted parity validation passed
- Verdict: PASS

## Side-Effect Verification
### native-trade-fixtures
- Evidence: `tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json`, `tests/fixtures/events/binance/ETH-USD/edge-native-timestamp-degraded-trade-usdt.fixture.v1.json`, `tests/fixtures/manifest.v1.json`
- Expected state: fixture corpus now covers native Binance Spot trade happy-path and degraded timestamp parser inputs
- Actual state: new native fixtures and manifest entries were added and exercised by integration tests
- Verdict: PASS

### supervisor-parser-seam
- Evidence: `services/venue-binance/trades.go`, `services/venue-binance/trades_test.go`, `tests/integration/binance_spot_trade_test.go`
- Expected state: one explicit parser entrypoint consumes `SpotRawFrame` values emitted by the completed supervisor
- Actual state: `ParseTradeFrame` and supervisor-backed tests now prove the seam directly
- Verdict: PASS

### duplicate-sensitive-identity
- Evidence: `TestNormalizerRawWriteBoundaryPreservesBinanceTradeDuplicateIdentity`, `TestIngestionBinanceSpotTradeDuplicateIdentityStable`
- Expected state: repeated Binance Spot trades keep stable `trade:<id>` source identity and duplicate audit occurrence `2`
- Actual state: tests passed with stable IDs and duplicate audit behavior scoped to the `trades` stream family
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current smoke matrix.
- `bookTicker` canonical handoff remained separate follow-on work at the time of this report.

## Next Actions
1. Plan and implement `binance-spot-top-of-book-canonical-handoff`.
2. After both Spot stream handoff slices are done, run direct live Binance validation instead of creating a smoke-only follow-on plan.
