# Binance Integration Completion Overview

## Objective

Finish the Binance integration so the local and later non-local stack runs on a sustained Binance-backed runtime behind Go-owned boundaries instead of the current bounded snapshot-polling cutover.

## User Outcome

The user should be able to:

- run the stack and see BTC and ETH current state fed by a long-lived Binance runtime rather than on-demand REST snapshots
- trust that warm-up, reconnect, stale, rate-limit, and degraded-runtime states are visible to operators and machine-readable to downstream consumers
- understand whether and how Binance USD-M context changes current-state and regime decisions instead of remaining an implementation-side ambiguity
- move from local-only success to repeatable `local`, `dev`, and later `prod` defaults without redoing core Binance wiring

## In Scope

- replace the command-local snapshot reader in `cmd/market-state-api` with a long-lived Spot runtime/read-model seam for current-state queries
- make Binance runtime health, warm-up, and degradation visible enough for operators to debug failures quickly
- decide and implement the bounded role of USD-M context in current-state and regime outputs
- harden environment config, startup defaults, and rollout notes for `local`, `dev`, and `prod`
- add long-run and failure-path validation for the final Binance runtime posture

## Out Of Scope

- authenticated/private Binance endpoints
- order submission, portfolio/account state, or user-data streams
- non-Binance venue expansion
- broad frontend redesign beyond what is needed to preserve the existing Go API boundary
- Python runtime dependencies in the live path

## Current Starting Point

- The earlier `crypto-market-copilot-binance-live-market-data` initiative landed the core Spot/USD-M adapters, raw/replay support, and the archived first live `market-state-api` cutover in `plans/completed/binance-live-market-state-api-provider-cutover/`.
- The current dashboard path already consumes real Binance Spot data through Go.
- The remaining integration gap is that `cmd/market-state-api` still uses a bounded command-local Spot depth snapshot reader instead of a sustained process-owned Spot runtime/read model.
- That missing runtime integration should be refined before the later observability, USD-M semantics, rollout, and long-run hardening epics.

## High-Level System Map

- `services/venue-binance` owns native Spot and USD-M runtime behavior, reconnects, recovery, and venue-specific status
- `services/normalizer` owns canonical event output
- raw/replay remain the audit and determinism boundary
- `services/feature-engine` and `services/regime-engine` own market-state semantics
- `services/market-state-api` remains the stable consumer-facing read boundary
- `apps/web` remains a consumer of same-origin `/api/...` responses only

## Key Constraints And Assumptions

- Go remains the live runtime path; Python stays offline-only
- existing `/api/market-state/global` and `/api/market-state/:symbol` contracts should remain stable unless refinement proves a bounded contract addition is unavoidable
- `BTC-USD` and `ETH-USD` remain the tracked symbols unless a later initiative expands coverage
- replay and current-state outputs must stay deterministic for pinned inputs even after the runtime source becomes streaming
- rate limits, reconnect behavior, and degradation reasons must remain explicit rather than hidden in logs only
