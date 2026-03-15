# Testing

## Validation Matrix

### 1. Provider cutover and lifecycle

- Goal: prove the command builds the live provider from the sustained runtime owner and shuts it down cleanly.
- Command: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./cmd/market-state-api -run 'TestNewProviderWithOptions|TestBinanceSpotRuntimeOwner'`
- Verify:
  - provider startup succeeds with the runtime-owner path
  - startup does not fabricate symbol observations before publishable data exists
  - shutdown remains bounded and no per-request polling fallback is exercised

### 2. Current-state API contract proof

- Goal: prove symbol and global current-state responses still assemble through the stable API/provider boundary after the cutover.
- Command: `docker run --rm -v "/home/wasd/code/crypto-monitor:/src" -w /src -e GOCACHE=/tmp/gocache golang:1.26-alpine go test ./cmd/market-state-api ./services/market-state-api ./tests/integration ./tests/replay -run 'TestBinanceSpotRuntime|TestIngestionBinanceCurrentState|TestReplayBinanceMarketStateDeterminism'`
- Verify:
  - `BTC-USD` and `ETH-USD` remain readable through the live provider path
  - startup/unavailable and degraded states stay machine-readable rather than silently masked
  - the determinism selector in the active plan matches the actual replay proof that executes in the repo

### 3. Same-origin browser smoke

- Goal: prove the compose-backed dashboard can load the shell through `/api/market-state/*` without frontend mocks or value-pinned live assertions.
- Command: `docker compose up --build -d && PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 pnpm --dir apps/web exec playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium --project=mobile-chrome`
- Verify:
  - the dashboard loads from the compose stack at `/dashboard`
  - both tracked symbol cards render and symbol switching updates the focused panel
  - the test accepts honest warm-up/fallback behavior but proves the shell once readable state is available

### 4. Direct API smoke

- Goal: prove the local compose stack serves the stable routes and keeps `/healthz` separate from market-data freshness.
- Command: `docker compose up --build -d && docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >/dev/null && curl -sf http://127.0.0.1:4173/api/market-state/global && curl -sf http://127.0.0.1:4173/api/market-state/BTC-USD`
- Verify:
  - `/healthz` responds successfully once the process is up
  - current-state routes respond with JSON through the live path
  - payloads can remain partial or unavailable during warm-up without failing the process health endpoint

## Notes For Testing Agent

- Read `plans/binance-spot-runtime-read-model-owner/testing-report.md` first and confirm its prerequisite regression blockers are resolved or explicitly acknowledged before treating this cutover as ready to archive.
- Prefer containerized Go commands in this workspace because local `go` may be unavailable.
- If the compose-backed browser smoke needs a short readiness wait, keep it bounded and document the exact condition used.
- Record whether the dashboard started in the unavailable state before transitioning to the full shell.
- Write results to `plans/binance-market-state-live-reader-cutover/testing-report.md` while the feature is active, then move the full directory to `plans/completed/binance-market-state-live-reader-cutover/` after implementation and validation finish.
