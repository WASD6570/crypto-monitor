# Testing Report: binance-spot-runtime-read-model-owner

## Environment
- Target: containerized Go test runs via `golang:1.26-alpine` because local `go` is unavailable in this workspace
- Date/time: 2026-03-14T23:48:08Z
- Commit/branch: `6d9d699` on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Startup and warm-up | `go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner|TestNewProviderWithOptions'` | Runtime owner starts cleanly, publishes no fabricated observations, keeps deterministic symbol ordering | Passed in Docker; command package tests completed successfully | PASS |
| Runtime progression and degradation | `go test ./cmd/market-state-api ./tests/integration -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState'` | Accepted inputs publish BTC/ETH observations and preserve degraded state across reconnect/resync | Passed in Docker; command and integration tests completed successfully | PASS |
| Planned determinism selector | `go test ./tests/replay -run 'TestBinanceCurrentStateDeterminism|TestBinanceSpotRuntimeDeterminism'` | Deterministic repeated-input proof executes | Package returned `ok` with `[no tests to run]`; selector no longer matches current replay test names | FAIL |
| Corrected determinism proof | `go test ./tests/replay -run 'TestReplayBinanceMarketStateDeterminism'` | Identical accepted-input sequences yield identical read-model output | Passed in Docker; current replay determinism test executed successfully | PASS |
| Full targeted regression | `go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay` | Targeted regression stays green across provider, integration, and replay surfaces | Passed after aligning replay integration fixtures with shared-partition trade manifest behavior | PASS |

## Execution Evidence
### startup-and-warm-up
- Command/Request: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner|TestNewProviderWithOptions'`
- Expected: runtime-owner startup tests and provider startup seam tests pass
- Actual: `ok   github.com/crypto-market-copilot/alerts/cmd/market-state-api  2.115s`
- Verdict: PASS

### progression-and-degradation
- Command/Request: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./cmd/market-state-api ./tests/integration -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState'`
- Expected: runtime progression, degradation, and integration snapshot tests pass
- Actual: `ok   github.com/crypto-market-copilot/alerts/cmd/market-state-api  2.097s`; `ok   github.com/crypto-market-copilot/alerts/tests/integration  0.008s`
- Verdict: PASS

### determinism-selector-from-plan
- Command/Request: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./tests/replay -run 'TestBinanceCurrentStateDeterminism|TestBinanceSpotRuntimeDeterminism'`
- Expected: targeted replay determinism test executes and passes
- Actual: `ok   github.com/crypto-market-copilot/alerts/tests/replay  0.006s [no tests to run]`
- Verdict: FAIL

### determinism-corrected-current-test-name
- Command/Request: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./tests/replay -run 'TestReplayBinanceMarketStateDeterminism'`
- Expected: current Binance market-state replay determinism test executes and passes
- Actual: `ok   github.com/crypto-market-copilot/alerts/tests/replay  0.006s`
- Verdict: PASS

### full-targeted-regression
- Command/Request: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay`
- Expected: targeted regression remains green across runtime owner, provider, integration, and replay surfaces
- Actual: `ok   github.com/crypto-market-copilot/alerts/cmd/market-state-api  12.113s`; `ok   github.com/crypto-market-copilot/alerts/services/market-state-api  0.025s`; `ok   github.com/crypto-market-copilot/alerts/tests/integration  0.031s`; `ok   github.com/crypto-market-copilot/alerts/tests/replay  0.015s`
- Verdict: PASS

## Side-Effect Verification
### startup-honesty-and-no-rest-fallback
- Evidence: `cmd/market-state-api/main_test.go:23` asserts zero observations before publishable data and zero REST depth requests during startup
- Expected state: no fabricated startup snapshot and no per-request polling fallback
- Actual state: startup command suite passed, so the zero-observation and zero-request assertions held
- Verdict: PASS

### degraded-snapshot-carry-forward
- Evidence: `cmd/market-state-api/main_test.go:66` asserts `[BTC-USD ETH-USD]` ordering, preserved last observation after disconnect, degraded `FeedHealth`, and `DepthStatus` reset to idle until resync
- Expected state: last accepted observations stay readable while machine-readable degradation is exposed
- Actual state: progression/degradation command suite passed, so degraded carry-forward assertions held
- Verdict: PASS

### provider-facing-degradation
- Evidence: `tests/integration/binance_current_state_test.go:37` asserts degraded ETH input produces degraded composite and non-available 30s bucket without schema drift
- Expected state: provider response keeps degradation visible to consumers
- Actual state: targeted integration suite passed, so provider-facing degradation assertions held
- Verdict: PASS

### repeated-input-determinism
- Evidence: `tests/replay/binance_current_state_determinism_test.go:14` deep-compares two replayed Binance market-state responses built from the same accepted input sequence
- Expected state: identical accepted inputs yield identical read-model output
- Actual state: corrected replay determinism command passed
- Verdict: PASS

### shutdown-bounds
- Evidence: passing `TestBinanceSpotRuntimeOwner*` suite includes repeated `owner.Stop(...)` assertions across startup, reconnect, snapshot refresh, and rollover scenarios in `cmd/market-state-api/main_test.go`
- Expected state: runtime shutdown completes without surfaced goroutine or lifecycle errors
- Actual state: no shutdown failures surfaced in the targeted command-package suite
- Verdict: PASS

## Blockers / Risks
- No active blockers remain from this testing pass; the replay selector was updated to `TestReplayBinanceMarketStateDeterminism` and the targeted regression baseline is green.

## Next Actions
1. Keep this report as prerequisite evidence for downstream Binance cutover work.
2. Archive the runtime-owner plan when its remaining reporting/handoff artifacts are ready.
