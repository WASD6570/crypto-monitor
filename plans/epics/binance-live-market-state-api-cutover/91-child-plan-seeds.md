# Child Plan Seeds: Binance Live Market State API Cutover

## `binance-live-current-state-query-assembly`

- Outcome: one bounded slice replaces deterministic bundle building with a service-owned live query assembler that turns accepted Binance Spot state into stable `SymbolCurrentStateQuery` and `GlobalCurrentStateQuery` inputs for `BTC-USD` and `ETH-USD` without changing the consumer contract.
- Primary repo areas: `services/market-state-api`, `services/feature-engine`, `services/regime-engine`, `tests/integration`, `tests/replay`
- Dependencies: completed Binance Spot runtime, depth recovery, raw append, and replay archives; current-state builders in `libs/go/features/market_state_current.go`; the stable API boundary from `plans/completed/market-state-api-compose-integration/`.
- Validation shape: targeted Go tests for live query assembly, unsupported symbol handling, explicit unavailable or degraded section behavior, and deterministic repeated-run checks using pinned Binance accepted inputs for symbol and global current-state responses.
- Why it stands alone: provider cutover, compose checks, and browser verification should not proceed until live current-state semantics are settled behind the existing API seam.

## `binance-live-market-state-api-provider-cutover`

- Outcome: one bounded slice wires `cmd/market-state-api`, compose, and browser verification to the live-backed provider while preserving same-origin `/api` behavior, the existing response routes, and operator-visible degradation messaging.
- Primary repo areas: `services/market-state-api`, `cmd/market-state-api`, `docker-compose.yml`, `apps/web/tests/e2e`, `services/market-state-api/README.md`, `README.md`
- Dependencies: `binance-live-current-state-query-assembly` and the existing compose/browser seams from `plans/completed/market-state-api-compose-integration/`.
- Validation shape: targeted `market-state-api` and command tests, `docker compose config`, compose startup against the live-backed API, direct `curl` checks for `/api/market-state/global` and `/api/market-state/BTC-USD`, and focused Playwright coverage for the dashboard path that already reads same-origin `/api`.
- Why it stands alone: consumer cutover and operator validation are distinct from query-assembly semantics, but they are only safe after the live provider posture is fixed.

## Validation Note

- Do not create a separate smoke-only or integration-only child feature for this epic.
- Attach compose, API, and browser verification directly to `binance-live-market-state-api-provider-cutover`.
- If product later requires USD-M context to alter current-state or regime decisions, open a new bounded slice after the Spot-driven cutover rather than expanding the first child plan implicitly.
