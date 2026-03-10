# Testing Report: binance-spot-depth-bootstrap-and-buffering

## Environment
- Target: local Go test environment plus public Binance REST probe
- Date/time: 2026-03-09
- Commit/branch: working tree on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Happy bootstrap alignment | buffered Binance `depthUpdate` frames plus one `/api/v3/depth` snapshot | stale pre-bootstrap deltas are ignored, the first bridging window is accepted, and canonical depth outputs emit in order | Unit, service, and integration tests passed | PASS |
| Missing bridging delta | snapshot arrives but buffered deltas never bridge `lastUpdateId + 1` | bootstrap stays explicit and unsynchronized without emitting accepted depth output | Unit and integration tests passed | PASS |
| Shared sequencer handoff | first accepted Binance delta spans a `U`/`u` window over the snapshot boundary | shared order-book sequencing accepts the bridged delta without a Binance-only canonical path | Ingestion and normalizer tests passed | PASS |
| REST snapshot timestamp posture | REST snapshot omits `exchangeTs` and real Binance depth responses omit `symbol` | parser uses the bootstrap owner source symbol, preserves empty `exchangeTs`, and degrades against `recvTs` | Unit tests plus live REST probe confirmed the shape | PASS |
| Fixture/parity alignment | happy bootstrap fixture joins the shared manifest | Go fixture consumer still validates the Binance fixture corpus | Parity test passed | PASS |

## Execution Evidence
### bootstrap-owner
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestParseOrderBook|TestSpotDepthBootstrap' -v`
- Expected: snapshot parsing, source-symbol fallback, buffered startup state, bridge alignment, and explicit failure posture all pass in the venue-owned module
- Actual: all targeted tests passed
- Verdict: PASS

### shared-sequencer-and-normalizer
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestOrderBookSequencer|TestNormalizeOrderBookMessage' -v && "/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalizeOrderBook' -v`
- Expected: bridging Binance delta windows are accepted through shared sequencing and canonical order-book normalization without new event types
- Actual: all targeted tests passed
- Verdict: PASS

### integration-proof
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepthBootstrap' -v`
- Expected: completed Spot supervisor frames feed the bootstrap owner, synchronized startup emits canonical order-book outputs, and non-bridging startup remains unsynchronized
- Actual: happy and failure-adjacent integration tests passed
- Verdict: PASS

### fixture-parity
- Command/Request: `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`
- Expected: the new happy Binance depth bootstrap fixture loads through the shared manifest and validates like the rest of the corpus
- Actual: targeted parity validation passed
- Verdict: PASS

### direct-live-rest-probe
- Command/Request: `curl -fsS "https://api.binance.com/api/v3/depth?symbol=BTCUSDT&limit=5"`
- Expected: confirm the live snapshot response shape used by bootstrap work and catch any mismatch before archiving the feature
- Actual: live Binance returned `lastUpdateId`, `bids`, and `asks` without a `symbol` field; the parser/owner path was updated to supply the known source symbol explicitly
- Verdict: PASS

## Side-Effect Verification
### live-shape-compatible-bootstrap
- Evidence: `services/venue-binance/orderbook.go`, `services/venue-binance/spot_depth_bootstrap.go`, `services/venue-binance/orderbook_test.go`
- Expected state: the bootstrap owner can consume real Binance REST snapshot payloads while preserving the known `BTCUSDT` or `ETHUSDT` source symbol and empty snapshot `exchangeTs`
- Actual state: `ParseOrderBookSnapshotWithSourceSymbol(...)` now accepts live REST payload shape and venue tests cover the fallback path
- Verdict: PASS

### fixture-backed-startup-proof
- Evidence: `tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/edge-depth-bootstrap-missing-bridge-usdt.fixture.v1.json`, `tests/integration/binance_spot_depth_bootstrap_test.go`
- Expected state: the repo proves both happy startup alignment and explicit non-bridging failure before the later recovery slice
- Actual state: fixture-backed integration now covers both paths with supervisor-fed depth frames
- Verdict: PASS

### shared-sequencer-reuse
- Evidence: `libs/go/ingestion/orderbook.go`, `libs/go/ingestion/book_normalization_test.go`, `services/normalizer/service_test.go`
- Expected state: Binance bootstrap alignment reuses the shared sequencer and normalizer instead of introducing a venue-only acceptance path
- Actual state: shared sequencing now accepts bridging delta windows through `FirstSequence` while preserving prior snapshot, stale, and gap behavior
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current validation matrix.
- Recurring resync loops, snapshot refresh cadence, and snapshot-stale degradation remain intentionally out of scope for the next child feature `binance-spot-depth-resync-and-snapshot-health`.

## Next Actions
1. Implement `binance-spot-depth-resync-and-snapshot-health` on top of the settled startup bootstrap owner and bridge semantics.
2. Reuse the live REST shape check when wiring the real snapshot requester so the runtime path keeps matching `/api/v3/depth` responses.
