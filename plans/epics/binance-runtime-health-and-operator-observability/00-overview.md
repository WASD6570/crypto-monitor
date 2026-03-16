# Binance Runtime Health And Operator Observability

## Epic Summary

Expose Binance runtime warm-up, reconnect, stale, recovery, and rate-limit posture through bounded machine-readable operator surfaces without reopening the existing market-state response contract or `/healthz` process-only semantics.

## In Scope

- aggregate sustained Spot runtime owner, supervisor, and depth recovery state into one explicit operator-oriented health snapshot for `BTC-USD` and `ETH-USD`
- expose additive runtime status surface(s) from `cmd/market-state-api` and `services/market-state-api` while keeping the existing market-state routes stable
- preserve shared feed-health vocabulary and carry operator-meaningful timestamps, counters, and reasons through the new surface
- update runbooks and focused validation so operators can distinguish warm-up, reconnect, stale, rate-limit, and recovery states quickly

## Out Of Scope

- changing current-state or regime semantics; that stays in `binance-usdm-market-state-influence`
- using `/healthz` for market-data freshness gating or breaking existing `/api/market-state/*` payloads
- environment rollout defaults, broader config policy, or long-run soak validation
- dashboard redesign or direct browser-to-Binance access

## Target Repo Areas

- `cmd/market-state-api`
- `services/venue-binance`
- `services/market-state-api`
- `docs/runbooks`
- `tests/integration`

## Validation Shape

- targeted Go tests for runtime-health snapshot assembly and state transitions across warm-up, reconnect, stale, recovery, and rate-limit paths
- handler and integration checks for the additive status surface while `/api/market-state/*` and `/healthz` stay stable
- local API and runbook proof covering process-health versus runtime-health separation and shared health vocabulary usage

## Current Repository State

- `plans/completed/binance-spot-runtime-read-model-owner/` and `plans/completed/binance-market-state-live-reader-cutover/` already landed the sustained Spot runtime seam and the live API cutover
- `services/venue-binance` already has machine-readable supervisor and depth-recovery health inputs, but they are not yet surfaced as one explicit operator-facing runtime status boundary
- `services/market-state-api` currently keeps `/healthz` process-only and relies on current-state payload degradation for user-facing honesty, which is useful but still indirect for operator debugging
- existing docs explain warm-up and degraded payload behavior, but they do not yet settle the bounded runtime-health surface this epic needs

## Major Constraints

- keep `/healthz` process health only; runtime-health visibility should land as an additive bounded surface
- preserve `BTC-USD` and `ETH-USD` as the only tracked symbols for this epic
- reuse the shared feed-health vocabulary from `docs/runbooks/ingestion-feed-health-ops.md`; logs are supporting evidence, not the primary contract
- keep current-state contracts backward-compatible; any payload metadata must be additive and not the only operator seam
- keep Go as the live runtime path; Python remains offline-only
