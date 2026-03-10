# Testing Report: binance-spot-ws-runtime-supervisor

## Environment
- Target: local Go test environment
- Date/time: 2026-03-09
- Commit/branch: `23eb71a2eda558b37353a0dea66c21d0a882ad4d` on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Healthy startup | Spot supervisor connect -> subscribe -> ack -> fresh traffic | `HEALTHY` feed-health and subscribed state | Matching unit and integration tests passed | PASS |
| Message stale path | Connected supervisor with traffic paused past stale threshold | `STALE` with `message-stale` | Matching unit and integration tests passed | PASS |
| Reconnect loop degradation | Repeated disconnects through reconnect threshold | `DEGRADED` with `connection-not-ready` and `reconnect-loop` | Matching unit and integration tests passed | PASS |
| Resubscribe after reconnect | Disconnect then reconnect and resubscribe | Desired stream set restored exactly once | Matching unit and integration tests passed | PASS |
| Proactive 24h rollover | Advance synthetic time to rollover deadline | Intentional reconnect before venue cutoff | Matching unit and integration tests passed | PASS |
| Feed-health normalization seam | Supervisor-generated health input through `services/normalizer` | Canonical feed-health preserves timestamps and provenance | Matching unit and integration tests passed | PASS |

## Execution Evidence
### unit-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestSpotWebsocketSupervisor|TestSpotSubscriptionState|TestSpotFeedHealth' -v`
- Expected: Spot subscription, lifecycle, stale, reconnect-loop, and normalization-seam tests all pass
- Actual: all targeted tests passed
- Verdict: PASS

### runtime-health-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestRuntimeEvaluateLoopState|TestRuntimeEvaluateAdapterInput' -v`
- Expected: shared runtime helpers remain compatible with supervisor-driven health evaluation
- Actual: all targeted tests passed
- Verdict: PASS

### integration-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestBinanceSpotSupervisor' -v`
- Expected: supervisor lifecycle and normalizer handoff work together without live Binance access
- Actual: integration smoke passed
- Verdict: PASS

## Side-Effect Verification
### feed-health-output-shape
- Evidence: `TestSpotFeedHealthNormalizationSeamPreservesProvenance` and `TestBinanceSpotSupervisor`
- Expected state: canonical feed-health keeps `BTC-USD`/`ETH-USD`, `BTCUSDT`/`ETHUSDT`, Spot market type, and stable source-record IDs
- Actual state: tests passed with preserved provenance and healthy/stale/degraded transitions
- Verdict: PASS

### spot-only-health-policy
- Evidence: `TestSpotFeedHealthUsesSpotOnlyRuntimeAndRecoversAfterFreshTraffic`
- Expected state: snapshot freshness stays `NOT_APPLICABLE` and no depth semantics leak into this feature
- Actual state: test passed with Spot-filtered health evaluation
- Verdict: PASS

### no-live-network-dependency
- Evidence: all smoke cases executed through synthetic time and scripted supervisor transitions in `services/venue-binance/spot_ws_supervisor_test.go` and `tests/integration/binance_spot_supervisor_test.go`
- Expected state: feature tests do not require a live Binance websocket session
- Actual state: all tests passed locally without external connectivity
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current smoke matrix.
- Remaining risk is downstream integration: later trade and `bookTicker` parser features still need to consume the supervisor raw-frame seam without widening this feature's scope.

## Next Actions
1. Run `feature-testing` after `binance-spot-trade-canonical-handoff` lands so the supervisor can be exercised with real Spot trade parsing.
2. Run `feature-testing` after `binance-spot-top-of-book-canonical-handoff` lands so the same supervisor path is verified with `bookTicker` normalization.
