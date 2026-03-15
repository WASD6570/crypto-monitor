# Binance Streaming Market State Runtime Integration

## Epic Summary

Integrate the already-landed Binance Spot websocket and depth runtime into `cmd/market-state-api` so current-state queries are fed by a sustained process-owned read model rather than the current bounded command-local snapshot polling seam.

## In Scope

- replace the command-local Spot snapshot reader in `cmd/market-state-api` with a long-lived runtime owner or read-model seam that continuously maintains accepted Spot state for `BTC-USD` and `ETH-USD`
- drive the existing `services/venue-binance` Spot supervisor, depth bootstrap, and depth recovery surfaces through a process-owned runtime loop instead of test-only orchestration
- keep `services/market-state-api` on the existing `/api/market-state/global` and `/api/market-state/:symbol` contract while changing only the runtime source underneath it
- preserve warm-up honesty, partial availability, and machine-readable degradation when the runtime is connecting, bootstrapping depth, resyncing, or stale
- leave enough explicit runtime state and validation seams for the later operator-observability and long-run hardening epics

## Out Of Scope

- USD-M-driven current-state or regime changes; those belong to `binance-usdm-market-state-influence`
- broad operator-facing status surfaces, logs, or runbook redesign beyond the minimum integration hooks needed to keep runtime state explicit
- `dev` and `prod` rollout defaults, compose policy changes, or broader environment hardening
- authenticated Binance endpoints, private account data, or user-data streams
- frontend-owned market-state logic or browser-side Binance integration

## Target Repo Areas

- `cmd/market-state-api`
- `services/venue-binance`
- `services/market-state-api`
- `tests/integration`
- `tests/replay`

## Validation Shape

- targeted Go tests for the sustained runtime owner, read-model updates, warm-up posture, reconnect handling, and depth bootstrap/resync integration
- integration checks that scripted Spot runtime inputs produce stable symbol and global current-state responses through the live API path
- deterministic repeated-input checks so the same accepted Spot inputs still yield stable read-model and current-state outputs
- same-origin API smoke proving the command now serves live-backed state without the command-local snapshot polling seam

## Current Repository State

- `cmd/market-state-api/live_provider.go` still owns a command-local polling reader that fetches `/api/v3/depth` on demand and only reuses `SpotDepthRecoveryOwner` as a bounded per-request state helper
- `services/market-state-api` already exposes `NewLiveSpotProvider(...)` and the stable current-state contract, so this epic should replace the reader beneath that provider rather than reopen handler semantics
- completed Binance Spot runtime work already exists for websocket supervision, depth bootstrap buffering, and depth resync/snapshot health; this epic should compose those surfaces into one sustained process-owned runtime loop instead of reimplementing them
- the archived provider cutover remains the immediate predecessor and should be treated as the temporary local-first seam this epic removes

## Major Constraints

- preserve the existing current-state consumer contract, including `world`, `usa`, bucket sections, slow-context shape, and existing route semantics
- treat this epic as the missing integration layer between the archived provider cutover in `plans/completed/binance-live-market-state-api-provider-cutover/` and later observability work; do not reopen the earlier cutover scope
- reuse the completed Spot supervisor, trade/top-of-book, depth bootstrap, and depth recovery behavior already settled under `plans/completed/`
- keep replay-sensitive runtime assumptions explicit so later refinement can preserve deterministic proof for accepted inputs
- keep Go as the live runtime path; Python remains offline-only
