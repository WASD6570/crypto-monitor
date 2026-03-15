# Testing Report: binance-market-state-live-reader-cutover

## Environment
- Target: local Go toolchain for Go validation, local `docker compose` stack for API/browser smoke, and Playwright container `mcr.microsoft.com/playwright:v1.58.2-noble` for browser execution
- Date/time: 2026-03-15T11:57:04Z
- Commit/branch: `6d9d699` on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Provider cutover and lifecycle | `PATH="/usr/local/go/bin:$PATH" go test ./cmd/market-state-api -run 'TestNewProviderWithOptions|TestBinanceSpotRuntimeOwner'` | Command-owned provider starts from the sustained runtime owner, preserves startup honesty, and shuts down cleanly | Passed locally; command package lifecycle and runtime-owner tests completed successfully | PASS |
| Current-state API contract proof | `PATH="/usr/local/go/bin:$PATH" go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState|TestReplayBinanceMarketStateDeterminism'` | Stable symbol/global routes remain readable, degraded behavior stays machine-readable, and replay determinism executes | Passed locally across command, integration, and replay packages | PASS |
| Same-origin browser smoke | `docker run --rm --network host -v "/home/wasd/code/crypto-monitor:/work" -w /work/apps/web mcr.microsoft.com/playwright:v1.58.2-noble bash -lc 'PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 ./node_modules/.bin/playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium --project=mobile-chrome'` | Compose-backed dashboard loads through same-origin `/api/market-state/*`, accepts warm-up fallback, and reaches the readable shell | Passed in Chromium and mobile Chrome; symbol switching and section navigation succeeded without value-pinned assertions | PASS |
| Direct API smoke | `docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >/dev/null && curl -sf http://127.0.0.1:4173/api/market-state/global && curl -sf http://127.0.0.1:4173/api/market-state/BTC-USD` | `/healthz` stays process-health only and current-state routes return JSON through the live path | Passed against the local compose stack; `/healthz` responded and both routes returned JSON | PASS |

## Execution Evidence
### provider-cutover-and-lifecycle
- Command/Request: `PATH="/usr/local/go/bin:$PATH" go test ./cmd/market-state-api -run 'TestNewProviderWithOptions|TestBinanceSpotRuntimeOwner'`
- Expected: provider startup, startup honesty, degradation carry-forward, reconnect handling, and shutdown tests pass through the runtime-owner seam
- Actual: `ok   github.com/crypto-market-copilot/alerts/cmd/market-state-api  2.379s`
- Verdict: PASS

### current-state-api-contract-proof
- Command/Request: `PATH="/usr/local/go/bin:$PATH" go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState|TestReplayBinanceMarketStateDeterminism'`
- Expected: stable provider/API path remains readable for `BTC-USD` and `ETH-USD`, degraded behavior stays visible, and replay determinism executes with the current selector
- Actual: `ok   github.com/crypto-market-copilot/alerts/cmd/market-state-api  2.115s`; `ok   github.com/crypto-market-copilot/alerts/services/market-state-api  0.013s [no tests to run]`; `ok   github.com/crypto-market-copilot/alerts/tests/integration  0.015s`; `ok   github.com/crypto-market-copilot/alerts/tests/replay  0.016s`
- Verdict: PASS

### same-origin-browser-smoke
- Command/Request: `docker run --rm --network host -v "/home/wasd/code/crypto-monitor:/work" -w /work/apps/web mcr.microsoft.com/playwright:v1.58.2-noble bash -lc 'PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 ./node_modules/.bin/playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium --project=mobile-chrome'`
- Expected: `/dashboard` loads through the compose stack, same-origin `/api/market-state/*` requests occur, the shell becomes readable, and symbol/section navigation works on desktop and mobile
- Actual: `2 passed (4.8s)` with both `chromium` and `mobile-chrome`
- Verdict: PASS

### direct-api-smoke
- Command/Request: `docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >/dev/null && curl -sf http://127.0.0.1:4173/api/market-state/global && curl -sf http://127.0.0.1:4173/api/market-state/BTC-USD`
- Expected: process health responds independently of market freshness and the same-origin current-state routes return valid JSON
- Actual: `/healthz` succeeded; `/api/market-state/global` returned `schemaVersion=market_state_current_global_v1`, `global.state=NO-OPERATE`, `symbols=2`; `/api/market-state/BTC-USD` returned `schemaVersion=market_state_current_response_v1`, `symbol=BTC-USD`, `composite.availability=partial`
- Verdict: PASS

## Side-Effect Verification
### prerequisite-runtime-owner-baseline
- Evidence: `plans/binance-spot-runtime-read-model-owner/testing-report.md` still mentions a stale selector, but the current targeted replay and regression commands passed in this run using `TestReplayBinanceMarketStateDeterminism`
- Expected state: prerequisite blocker is resolved or explicitly acknowledged before archiving the cutover
- Actual state: prerequisite regression baseline is green; the stale note remains only as historical reporting debt in the predecessor report
- Verdict: PASS

### same-origin-dashboard-path
- Evidence: `apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts` now waits for same-origin `/api/market-state/*` responses, retries through the unavailable fallback, and verifies both symbol switching and section navigation in the compose-backed UI
- Expected state: dashboard remains a same-origin consumer and reaches the readable shell without frontend mocks
- Actual state: desktop and mobile browser smoke both passed against the compose stack
- Verdict: PASS

### process-health-separation
- Evidence: compose smoke hit `wget -qO- http://127.0.0.1:8080/healthz` inside `market-state-api` and separate `curl` requests to the browser-facing current-state routes
- Expected state: `/healthz` stays process health only while current-state payloads carry availability/degradation semantics
- Actual state: `/healthz` responded successfully and the current-state routes returned valid JSON independently of partial market-state availability
- Verdict: PASS

### browser-warm-up-posture
- Evidence: warm-up-aware Playwright helper in `apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts:36`
- Expected state: smoke accepts an initial `Current State Unavailable` fallback but proves recovery into the shell once state is readable
- Actual state: the smoke path reached the readable shell during this run; an initial unavailable fallback was allowed by the test but not explicitly observed in the recorded output
- Verdict: PASS

## Blockers / Risks
- No blocking failures in the cutover testing matrix.
- Local Playwright execution still depends on containerized browser tooling in this environment because the host is missing required Chromium shared libraries.
- The predecessor report under `plans/binance-spot-runtime-read-model-owner/testing-report.md` still contains a stale selector note even though the prerequisite regression baseline is now green.

## Next Actions
1. If you want the feature fully closed, archive `plans/binance-market-state-live-reader-cutover/` to `plans/completed/binance-market-state-live-reader-cutover/` now that implementation and validation are green.
2. Optionally refresh the historical notes in `plans/binance-spot-runtime-read-model-owner/testing-report.md` so its blocker section matches the now-green prerequisite baseline.
