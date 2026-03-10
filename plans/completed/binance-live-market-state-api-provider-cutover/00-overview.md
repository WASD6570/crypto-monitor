# Binance Live Market State API Provider Cutover

## Ordered Implementation Plan

1. Wire `cmd/market-state-api` to the new Spot live provider by adding one command-owned reader/bootstrap path that satisfies `marketstateapi.SpotCurrentStateReader` and makes live-backed responses the default runtime behavior.
2. Update local Compose and browser validation so `/api/market-state/*` keeps the same-origin Go boundary while assertions stop depending on deterministic BTC/ETH state labels.
3. Refresh operator-facing docs to describe the live Spot-backed path, warm-up/unavailable behavior, explicit `usa` limits, and machine-readable degradation expectations.
4. Run the attached validation matrix, write `plans/binance-live-market-state-api-provider-cutover/testing-report.md`, then move the full directory to `plans/completed/binance-live-market-state-api-provider-cutover/` after implementation and validation finish.

## Requirements

- Scope is limited to entrypoint cutover, local Compose/browser validation, and operator docs for the already-implemented live query assembly seam.
- Preserve the existing HTTP contract and routes from `services/market-state-api`: `GET /healthz`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Keep the first runtime cutover Spot-driven for current-state and regime inputs; `usa` stays explicitly unavailable or partial rather than being repopulated with deterministic placeholders.
- Keep `BTC-USD` and `ETH-USD` as the only supported symbols for this feature.
- Keep `apps/web` consumer-only; do not reintroduce frontend-owned route mocks, venue logic, or client-side market-state derivation.
- Keep slow context optional and non-blocking during live cutover. Browser validation may assert the panel/fallback path, but must not require deterministic seeded values.
- Keep Go as the live runtime path; Python remains offline-only.
- Keep this slice local-first. Do not expand scope into broader environment rollout or unrelated Binance config repair outside what the local cutover needs.

## Design Notes

### Repository state to preserve

- `cmd/market-state-api/main.go` still instantiates `marketstateapi.NewDeterministicProvider()`, so the command entrypoint is the remaining cutover seam.
- `services/market-state-api/api.go` and `services/market-state-api/live_spot_provider.go` already provide the stable handler boundary plus `NewLiveSpotProvider(...)`.
- `plans/completed/binance-live-current-state-query-assembly/` already settled the Spot-first live response posture, unsupported-symbol behavior, and degradation semantics this feature must preserve.
- `docker-compose.yml` and `apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts` still reflect deterministic assumptions from the earlier Go API integration slice.
- `services/market-state-api/README.md` and `README.md` still describe deterministic local state and need to be updated once the command default changes.

### Planned runtime posture

- Keep provider ownership in `cmd/market-state-api`; do not invent a second API boundary or move runtime decisions into `apps/web`.
- Add one concrete command-owned live reader/bootstrap seam that can feed `marketstateapi.NewLiveSpotProvider(...)` from accepted Binance Spot observations without changing the `marketstateapi.Provider` contract.
- Treat the deterministic provider as a package/test utility after cutover, not the default command/runtime path.
- Allow symbol/global responses to remain honestly unavailable or partial during startup warm-up until the live reader has enough Spot observations.
- Keep `/healthz` a simple process health endpoint; market-data freshness and degradation remain visible in the JSON payloads rather than by failing healthz on first startup.

### Local-first constraint

- This feature only needs to cut over the local stack and the default command path used by Compose.
- Do not widen the slice into dev/prod Binance config rollout or immediate USD-M weighting changes.
- If later environments need non-local config loading, that should be a bounded follow-on after the local live path is proven.

### Browser validation posture

- Existing Playwright smoke currently hard-codes deterministic regime text (`BTC-USD is TRADEABLE`, `ETH-USD is WATCH`). Those assertions must become contract-oriented rather than value-pinned.
- Browser checks should prove the dashboard loads through same-origin `/api`, renders both tracked symbols, supports symbol switching, and surfaces live-backed current-state sections even when state availability is partial or degraded.

### Live vs research boundary

- All runtime provider wiring, live reader bootstrap, and compose verification stay in Go plus the existing SPA consumer.
- No Python process, notebook, or offline artifact becomes a runtime dependency for `market-state-api`.

## ASCII Flow

```text
browser
  |
  v
web origin (/dashboard, /api/*)
  |
  +--> static SPA
  |
  +--> reverse proxy /api/market-state/*
           |
           v
   cmd/market-state-api
     - load local runtime inputs
     - build live Spot reader
     - NewLiveSpotProvider(...)
           |
           v
   services/market-state-api
     - stable handlers
     - live Spot query assembly
           |
           v
stable current-state JSON
  - same routes
  - startup may be partial/unavailable
  - degradation stays machine-readable
```

## Archive Intent

- Keep this feature active under `plans/binance-live-market-state-api-provider-cutover/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/binance-live-market-state-api-provider-cutover/`.
