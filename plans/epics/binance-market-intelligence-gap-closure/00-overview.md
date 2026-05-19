# Binance Market Intelligence Gap Closure

## Epic Summary

Close the remaining Binance integration gaps so `BTC-USD` and `ETH-USD` current state is backed by a sustained, operationally proven Binance Spot plus USD-M runtime and richer market-intelligence inputs instead of only a live best-bid/best-ask price path with conservative derivatives caps.

## User Outcome

- The user can run the stack and trust that BTC and ETH Spot prices come from the live Go-owned Binance runtime.
- The user can see whether Spot depth, spread, trade flow, and USD-M context are healthy enough to trust.
- The system can turn Binance Spot and USD-M inputs into deterministic liquidity and market indicators that later alerting plans can consume.
- Replay, runtime health, and current-state outputs stay deterministic and auditable for the same inputs and config.

## In Scope

- reconcile current validation drift before adding new Binance indicator behavior
- prove the sustained Binance runtime through repeated reconnect, stale, rate-limit, depth-resync, and warm-up checks
- promote live Spot trade prints from parse/raw support into feature-engine inputs where they materially improve current-state and later alert logic
- derive real Spot liquidity indicators from top-of-book and depth data instead of the current fixed liquidity score
- enrich USD-M derivatives context beyond the current conservative cap into explicit indicator inputs such as basis, funding pressure, open-interest change, and liquidation intensity
- expose any new indicator outputs through additive service-owned contracts or current-state fields only when feature plans prove the exact boundary
- preserve Go as the live runtime owner and keep `apps/web` as a same-origin API consumer

## Out Of Scope

- authenticated Binance endpoints, account/user-data streams, order placement, portfolio state, or exchange credentials
- COIN-M futures, options, non-Binance venue expansion, or new traded symbols beyond `BTC-USD` and `ETH-USD`
- browser-side Binance access, frontend-computed market state, or fixture-backed live runtime behavior
- speculative AI scoring or ranking in the live path
- concrete schemas, migrations, or API contract changes before child feature planning identifies the minimal additive boundary

## Target Repo Areas

- `services/venue-binance`
- `cmd/market-state-api`
- `services/market-state-api`
- `services/feature-engine`
- `services/regime-engine`
- `libs/go/features`
- `libs/go/ingestion`
- `schemas/json/features`
- `schemas/json/events`
- `schemas/json/replay`
- `tests/integration`
- `tests/replay`
- `tests/fixtures`
- `apps/web` only for additive rendering once service contracts are stable
- `docs/runbooks`
- `plans/STATE.md`

## Current Repository State

- Binance Spot and USD-M low-level ingestion surfaces already exist for `BTC-USD` and `ETH-USD`.
- Spot runtime support includes websocket lifecycle, `trade`, `bookTicker`, `depth@100ms`, REST depth bootstrap, depth recovery, feed-health states, and runtime-status exposure.
- USD-M support includes `markPrice@1s`, funding, mark/index, `forceOrder` liquidations, REST-polled open interest, and conservative USD-M influence caps.
- The current live market-state path uses Spot best bid/ask as the price contributor and applies a fixed `LiquidityScore: 100` rather than real liquidity analytics.
- Spot trade parsing and raw/replay seams exist, but live trade flow is not yet a first-class current-state or alert-readiness input.
- Depth is operationally important for synchronization and health, but spread, size, imbalance, slippage proxy, and depth pressure are not yet first-class feature outputs.
- USD-M derivatives data is used mostly as conservative provenance/cap logic; richer derivatives indicators such as open-interest delta, funding pressure, basis regime, and liquidation intensity remain missing.
- Local validation currently has drift: full Go tests, contract-family validation, and fixture replay smoke need reconciliation before new Binance indicator work should be trusted.
- `plans/STATE.md` previously tracked `binance-long-run-runtime-hardening` as a narrow seed; this epic supersedes that seed by folding long-run hardening into the broader market-intelligence gap closure.

## Major Constraints

- Keep `BTC-USD` and `ETH-USD` fixed unless a later initiative explicitly expands coverage.
- Keep `/healthz` process-health only and `/api/runtime-status` as the operator runtime-health route.
- Keep `/api/market-state/global` and `/api/market-state/:symbol` service-owned; any new fields must be additive and backward-compatible unless a feature plan explicitly handles rollout.
- Do not trust client-computed market state, liquidity, alert, risk, or outcome decisions.
- Preserve event time versus processing time distinction across accepted Binance inputs, feature buckets, replay, and current-state output.
- Replay and repeated runtime checks must be deterministic for the same accepted raw inputs, config, and code version.
- Degraded feeds, stale data, sequence gaps, reconnect loops, cooldowns, and rate limits must stay machine-visible instead of being hidden in logs.

## Validation Shape

- `go test ./...` must pass before the enriched Binance feature work is considered stable.
- Contract validation must pass with the current replay schemas and family manifests.
- Fixture-backed replay smoke must pass with deterministic ordering and source-record identity.
- Targeted Binance runtime tests must cover reconnect, stale, depth recovery, rate-limit, and warm-up behavior.
- Enriched Spot and USD-M indicators must have deterministic integration and replay coverage before they feed alerting.
- Compose smoke and, where practical, a documented live-boundary check should prove the same-origin Go-owned market path without mocks.
