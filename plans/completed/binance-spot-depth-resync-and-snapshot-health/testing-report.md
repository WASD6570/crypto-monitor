# Testing Report: binance-spot-depth-resync-and-snapshot-health

## Environment
- Target: local Go test environment plus public Binance REST/live integration probe
- Date/time: 2026-03-09
- Commit/branch: working tree on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Gap-triggered resync | synchronized depth detects a sequence gap and buffers post-gap deltas | depth stops claiming synchronized state, then re-enters synchronized state only after an eligible replacement snapshot aligns | Venue and integration tests passed | PASS |
| Cooldown-blocked retry | gap occurs too soon after the last accepted snapshot | replacement snapshot request stays blocked and feed health remains explicitly degraded | Venue and integration tests passed | PASS |
| Rate-limit-blocked retry | per-minute snapshot recovery allowance is exhausted | retry stays blocked with visible `rate-limit` degradation and no silent repair | Venue, normalizer, and integration tests passed | PASS |
| Snapshot refresh and stale posture | replacement snapshot becomes due, then later stale while depth messages still arrive | refresh due is tracked, stale precedence remains `STALE`, and snapshot freshness stays machine-visible | Venue and integration tests passed | PASS |
| Canonical feed-health handoff | recovery owner emits Binance Spot depth health through the normalizer | canonical feed-health output preserves stable source identity plus recovery reasons | Normalizer tests passed | PASS |
| Live Binance shape check | direct public `/api/v3/depth` probe and env-gated Go validation | current live snapshot payload still parses with source-symbol fallback and positive sequence | Curl probe and live integration test passed | PASS |

## Execution Evidence
### venue-runtime-and-recovery
- Command/Request: `"/usr/local/go/bin/go" test ./services/venue-binance -run 'Test(Runtime|SpotDepth)' -v`
- Expected: recovery owner state transitions, refresh cadence, snapshot staleness, and runtime health composition all pass in the venue-owned module
- Actual: all targeted tests passed
- Verdict: PASS

### shared-handoff
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'Test(OrderBook|FeedHealth)' -v && "/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OrderBook|FeedHealth)' -v`
- Expected: replacement snapshot handoff and recovery-originated feed-health outputs remain compatible with shared sequencing and canonical normalization
- Actual: all targeted tests passed
- Verdict: PASS

### fixture-backed-integration
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepth(Recovery|Bootstrap)' -v`
- Expected: deterministic Binance depth bootstrap and recovery flows prove successful resync, blocked retry states, and snapshot-stale visibility through the supervisor-to-normalizer boundary
- Actual: targeted integration tests passed
- Verdict: PASS

### fixture-parity
- Command/Request: `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`
- Expected: newly added Binance depth recovery fixtures load through the shared manifest and validate against canonical event requirements
- Actual: targeted parity validation passed
- Verdict: PASS

### direct-live-validation
- Command/Request: `curl -fsS "https://api.binance.com/api/v3/depth?symbol=BTCUSDT&limit=5" && BINANCE_LIVE_VALIDATION=1 "/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepthLive' -v`
- Expected: live Binance depth snapshot shape still matches parser expectations and the env-gated Go probe succeeds against the real public API boundary
- Actual: live Binance returned `lastUpdateId`, `bids`, and `asks`; the env-gated integration test passed with `ParseOrderBookSnapshotWithSourceSymbol(...)`
- Verdict: PASS

## Side-Effect Verification
### recovery-owner-boundary
- Evidence: `services/venue-binance/spot_depth_recovery.go`, `services/venue-binance/spot_depth_recovery_test.go`
- Expected state: Binance Spot depth now has a venue-owned recovery state machine for gap handling, blocked retries, refresh cadence, and feed-health emission without reopening transport ownership
- Actual state: recovery owner landed with synchronized, resyncing, cooldown-blocked, rate-limit-blocked, and bootstrap-failed posture plus deterministic tests
- Verdict: PASS

### runtime-refresh-visibility
- Evidence: `services/venue-binance/runtime.go`, `services/venue-binance/runtime_test.go`
- Expected state: runtime exposes refresh-due status alongside existing cooldown, rate-limit, and stale-snapshot primitives
- Actual state: `SnapshotRefreshStatus(...)` now exists and is covered by targeted runtime tests
- Verdict: PASS

### fixture-backed-recovery-proof
- Evidence: `tests/fixtures/events/binance/BTC-USD/happy-native-depth-resync-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/edge-depth-recovery-cooldown-blocked-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/edge-depth-recovery-rate-limit-usdt.fixture.v1.json`, `tests/fixtures/events/binance/BTC-USD/edge-depth-snapshot-stale-usdt.fixture.v1.json`, `tests/integration/binance_spot_depth_recovery_test.go`
- Expected state: the repo proves successful resync, cooldown/rate-limit blocked recovery, and snapshot-stale degradation through deterministic integration fixtures
- Actual state: recovery fixtures and integration coverage were added and passed
- Verdict: PASS

### canonical-depth-health-identity
- Evidence: `services/normalizer/service_test.go`
- Expected state: recovery-originated feed-health outputs preserve a stable source identity and explicit degradation reasons through canonical normalization
- Actual state: normalizer test passed with stable `runtime:binance-spot-depth:BTCUSDT` identity and preserved sequence-gap / connection-not-ready reasons
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current validation matrix.
- This slice still stops at recovery semantics; raw append/replay rollout and market-state API cutover remain owned by later Binance epics.

## Next Actions
1. Move to `plans/epics/binance-live-raw-storage-and-replay/` now that Spot depth bootstrap and recovery semantics are settled.
2. Reuse the env-gated live depth validation path whenever the real snapshot requester or later replay-sensitive Binance work changes parser assumptions.
