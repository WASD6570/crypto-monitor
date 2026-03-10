# Testing Report: Binance Live Market State API Provider Cutover

## Result

- Status: passed
- Scope: command/runtime cutover, compose proxy path, browser smoke, and operator doc alignment

## Commands

| Command | Purpose | Result |
|---|---|---|
| `"/usr/local/go/bin/go" test ./services/market-state-api ./cmd/market-state-api -v` | validate stable handlers plus live-backed command wiring | Passed |
| `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceCurrentState.*' -v` | validate Spot-first current-state behavior remains stable | Passed |
| `pnpm --dir apps/web test -- --run dashboardDecoders` | validate web decoder contract against Go payload shape | Passed |
| `docker compose config` | validate compose renders with the updated `market-state-api` image/runtime path | Passed |
| `docker compose up --build -d` | start local stack with live-backed `market-state-api` command | Passed |
| `curl http://127.0.0.1:4173/api/market-state/global` | verify same-origin global API path returns Go-owned JSON | Passed |
| `curl http://127.0.0.1:4173/api/market-state/BTC-USD` | verify same-origin symbol API path returns schema-stable symbol JSON | Passed |
| `LD_LIBRARY_PATH="/home/personal/.local/lib/playwright-deps/usr/lib/x86_64-linux-gnu:/home/personal/.local/lib/playwright-deps/usr/lib/x86_64-linux-gnu/nss${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}" PLAYWRIGHT_EXTERNAL_SERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 pnpm --dir apps/web exec playwright test tests/e2e/dashboard-compose-api-smoke.spec.ts --project=chromium` | validate dashboard loads from the real Go API without deterministic-state assertions | Passed |
| `docker compose down` | clean up local validation stack | Passed |

## Notes

- `cmd/market-state-api` now defaults to `marketstateapi.NewLiveSpotProvider(...)` through a command-owned Binance Spot snapshot reader.
- The live reader performs on-demand Binance Spot depth snapshot fetches for `BTC-USD` and `ETH-USD`, caches observations briefly, and returns honest unavailable responses when live observations are not yet available.
- The first cutover remains Spot-driven: `usa` stays explicit and unavailable, and slow context remains optional/non-blocking.
- The sampled live API response returned real Binance-derived world prices while still showing `timestamp-fallback` and incomplete-window degradation, which is expected for this bounded first cutover.
- Playwright required the existing local runtime-library workaround via `LD_LIBRARY_PATH`; no repo files were added for those system libraries.
