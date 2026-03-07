# Crypto Market Copilot Program Overview

## First User

The first user is you: a single operator who wants the system to watch BTC and ETH markets continuously while you focus on other work.

The product must help you answer, quickly and reliably:

- what is happening right now across WORLD and USA liquidity
- whether the market is tradeable or should be ignored
- whether a new alert deserves attention right now
- whether past alerts actually worked after costs and market conditions

## Core Product Loop

1. Trust the screen: open the product and understand current market state in under a minute.
2. Trust the gate: know when the system says `TRADEABLE`, `WATCH`, or `NO-OPERATE`, and why.
3. Trust the alert: when an alert arrives, see the reason, levels, and context immediately.
4. Trust the review: later confirm whether the alert worked, failed, or was too costly after realistic assumptions.
5. Tune deliberately: compare against baselines and adjust thresholds through replay, not intuition alone.

## Program Structure

This MVP is intentionally split into two initiatives:

1. `crypto-market-copilot-visibility-foundation`
   - build the trusted market-state and visualization substrate first
2. `crypto-market-copilot-alerting-and-evaluation`
   - layer alerts, outcomes, simulation, and operator feedback on top of that substrate

This split is deliberate. If the screen, replay, regime state, and feed integrity are not trustworthy first, alerting will only compound confusion.

## In Scope

- BTC and ETH only
- WORLD spot and perp sensors from Binance and Bybit
- USA spot feeds from Coinbase and Kraken
- slow USA institutional context from CME volume/OI and ETF daily flows
- canonical event normalization, replay, dashboards, alerts, outcomes, simulation, and review surfaces
- alert-only MVP with simulated execution only

## Out Of Scope

- live trading in the MVP
- private or inside data
- black-box AI in the live decision path
- single-venue alert logic as the product default
- HFT or sub-millisecond assumptions

## Constraints

- `services/*` own live and realtime logic in Go
- `apps/web` owns operator UX in React + Vite
- Python remains offline-only for research, parity checks, and future analysis
- shared contracts live under `schemas/json/...`
- replay and evaluation must be deterministic for the same data and config
- feed degradation must visibly reduce trust and tradeability, not silently continue

## High-Level System Map

```text
venue ws/rest feeds
        |
        v
services/venue-* -> services/normalizer -> canonical events
        |                                  |
        |                                  v
        +--------------------------> raw append-only store
                                           |
                                           v
                                 services/replay-engine
                                           |
                                           v
                         services/feature-engine + services/regime-engine
                                           |
                            initiative 1 ends with trusted state + dashboards
                                           |
                                           v
                           services/alert-engine + services/risk-engine
                                           |
                                           v
                      services/outcome-engine + services/simulation-api
                                           |
                                           v
                                  apps/web review surfaces
```

## Planning Assumptions

- Bybit remains the second offshore venue for the initial WORLD view.
- USA institutional context is valuable but should not delay the first trusted market-state release.
- AI-assisted explanation, summarization, and tuning suggestions are phase-two enhancements, not MVP critical-path logic.
