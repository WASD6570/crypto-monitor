# Visibility Foundation Overview

## Objective

Build the trusted market-visibility layer first so the user can understand BTC and ETH state across WORLD and USA liquidity without watching raw venue feeds all day.

## User Outcome

Within 60 seconds, the user should be able to answer:

- is BTC tradeable right now or should it be ignored
- is ETH tradeable right now or should it be ignored
- are USA and WORLD aligned or fragmented
- which feeds are healthy, degraded, or stale
- what changed recently enough to deserve attention

## In Scope

- shared canonical contracts for events, features, replay, and regime state
- deterministic fixtures and replay seeds
- resilient ingestion for Binance, Bybit, Coinbase, and Kraken
- raw append-only storage and deterministic replay
- per-venue and composite features for 30s, 2m, and 5m
- WORLD vs USA divergence and `TRADEABLE/WATCH/NO-OPERATE` market state
- dashboards for symbol overview, microstructure, derivatives context, and feed health
- CME and ETF context only if it does not block the core trusted-state path

## Out Of Scope

- alert setup logic
- push delivery surfaces
- alert outcome evaluation
- simulated execution
- operator feedback on alerts

## Exit Criteria

- repeated replay runs produce identical state outputs
- feed degradation is visible and traceable
- current market state is understandable in under 60 seconds
- the user can see WORLD vs USA divergence, market quality, and regime state without opening external tools

## Ordered Slice Queue

1. `canonical-contracts-and-fixtures`
2. `market-ingestion-and-feed-health`
3. `raw-storage-and-replay-foundation`
4. `world-usa-composites-and-market-state`
5. `visibility-dashboard-core`
6. `slow-context-panel`

## Constraints

- favor stability and explainability over extra sources or extra panels
- do not rederive market logic in the UI
- do not make Python part of the live path
- no alert logic should be hidden inside dashboard-only transforms
