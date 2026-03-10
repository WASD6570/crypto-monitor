# Binance Live Market Data Overview

## Objective

Replace the deterministic Binance placeholder path with real public Binance Spot and USD-M market data so the local and future live stack can produce trusted current-state outputs behind Go-owned service boundaries.

## User Outcome

The user should be able to:

- run the local stack and see BTC and ETH current state sourced from live Binance market data
- trust that spot and derivatives context arrive with explicit feed-health, freshness, and degradation signals
- inspect current-state outputs without the frontend owning venue-specific runtime logic
- replay accepted Binance inputs deterministically enough to audit state changes and operator-visible degradation

## In Scope

- Binance Spot public WS ingestion for `trade`, `bookTicker`, and `depth@100ms`
- Binance Spot REST bootstrap for `/api/v3/depth` when order-book sequencing requires a snapshot
- Binance USD-M public WS ingestion for `markPrice@1s` and `forceOrder`
- Binance USD-M REST polling for `/fapi/v1/openInterest`
- venue-native reconnect, ping/pong, 24h rollover, bounded backoff, and resubscribe behavior for Spot and USD-M
- canonical symbol normalization to repo contracts such as `BTC-USD` and `ETH-USD` while preserving `sourceSymbol`, `quoteCurrency`, and `marketType`
- stream-specific exchange-time handling with strict `recvTs` fallback and explicit degraded timestamp reasons
- raw append, replay, and feed-health compatibility for accepted Binance live inputs
- cutover of `services/market-state-api` from deterministic provider behavior to live Binance-backed current-state sourcing without frontend-owned API mocks
- config, fixtures, and runbook updates needed to operate the slice locally and in later non-local environments

## Out Of Scope

- authenticated/private Binance endpoints
- order submission, portfolio state, or account/user-data streams
- COIN-M futures, options, or non-Binance venue rollout changes
- frontend-owned venue logic or market-state recomputation in `apps/web`
- broad historical backfill or migration planning beyond the bounded live bootstrap needed for this initiative

## Recommended Defaults For This Initiative

- Streams first: Spot `trade`, `bookTicker`, and `depth@100ms`; USD-M `markPrice@1s`, `forceOrder`, and REST-polled `openInterest`
- Reconnect policy: separate Spot and USD-M supervisors with exact Binance ping/pong handling, proactive reconnect before the 24h disconnect, and bounded backoff from current config defaults
- Symbol normalization: keep canonical symbols asset-centric as `BTC-USD` and `ETH-USD`, while preserving native `sourceSymbol` and `marketType`
- Time policy: use the most semantically correct exchange timestamp per stream, then fall back to `recvTs` only when missing, invalid, or beyond the strict skew threshold already defined in `libs/go/ingestion/timestamp.go`
- Replay posture: persist accepted raw inputs and feed-health outputs with stable source IDs so replay reconstructs normalized outputs deterministically
- REST pairing: use REST only where the venue contract requires bootstrap or polling, not as a duplicate source for streams that are already authoritative enough over WS

## High-Level System Map

- `services/venue-binance` owns native Spot and USD-M connection management, parsing, feed health, and venue-specific recovery behavior
- `services/normalizer` owns canonical event output while preserving `exchangeTs`, `recvTs`, `sourceSymbol`, and degraded reasons
- raw and replay services remain the audit boundary for accepted live inputs
- feature and regime services remain the source of trusted market-state semantics
- `services/market-state-api` stays the same consumer-facing boundary while its backing data becomes live instead of deterministic
- `apps/web` remains a consumer of stable `/api/...` responses rather than a venue runtime

## Key Constraints And Assumptions

- Go remains the live runtime path; Python stays offline-only
- Binance Spot and USD-M heartbeat behavior differ and should not be flattened into one venue-agnostic connection loop by default
- order-book sequencing must preserve explicit gap and resync behavior instead of silently repairing state
- feed degradation must stay visible as canonical feed-health output, not only logs
- shared event schemas for `market-trade`, `order-book-top`, `feed-health`, `funding-rate`, `mark-index`, `open-interest-snapshot`, and `liquidation-print` already exist and should be reused unless refinement proves a contract gap
- the current web/API contract should remain stable wherever possible so live cutover does not require another frontend architecture change

## Ordered Slice Queue

1. `binance-live-contract-seams-and-fixtures`
2. `binance-spot-trades-and-top-of-book-runtime`
3. `binance-spot-depth-bootstrap-and-recovery`
4. `binance-usdm-context-sensors`
5. `binance-live-raw-storage-and-replay`
6. `binance-live-market-state-api-cutover`
