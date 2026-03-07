# Slow Context Panel

## Ordered Implementation Plan

1. Define slow-feed ingestion boundaries for CME volume/open interest and ETF daily flow context without locking the system to a single provider.
2. Define normalization, persistence, freshness, and context-only query semantics in service-owned Go paths.
3. Define non-blocking dashboard integration, timestamp/cadence labeling, stale handling, and operator messaging in `apps/web`.
4. Validate fixture-backed missing, delayed, and stale slow-context cases before implementation handoff.

## Feature Role In The Visibility Initiative

`slow-context-panel` is the sixth slice in `crypto-market-copilot-visibility-foundation`, after trusted realtime ingestion, replay, market state, and the core dashboard are already usable.

It is intentionally later and non-blocking because CME participation and ETF flow context help explain USA institutional posture, but they update far more slowly than the live market-state loop. The first operator must still be able to trust `TRADEABLE`, `WATCH`, and `NO-OPERATE` decisions when slow context is absent, late, or stale.

## Scope

- planning for slower-cadence USA context surfaces covering CME volume, CME open interest, and ETF daily net flow
- service-owned ingestion boundaries and normalization guidance in Go for live/runtime paths
- storage and query semantics for slow context snapshots or daily points
- timestamp, as-of, source-time, and cadence labeling rules
- stale, missing, delayed, and unavailable behavior
- non-blocking integration seams in the visibility dashboard for BTC and ETH operator workflows
- operator messaging that makes context advisory rather than gating in MVP
- targeted fixture and smoke validation guidance

## Out Of Scope

- changing realtime tradeability, regime, or feed-health gating to depend on slow context in MVP
- speculative provider-specific lock-in, vendor contracts, or source-specific auth designs
- Python in the live runtime path
- new alert logic, outcome logic, or simulation behavior
- deep historical analytics, backtests, or research notebooks as part of this feature plan
- concrete schema implementation beyond planning guidance

## Key Requirements

- Slow context must remain visually and semantically separate from realtime market-state surfaces.
- CME volume/open interest and ETF daily flow must show explicit as-of timestamps and expected cadence.
- The UI must state when data is fresh, stale, delayed, or unavailable.
- Missing slow context must degrade explanation quality only; it must not silently become a hard dependency for market-state gating.
- Service outputs remain the source of truth for normalized slow context and freshness classification.
- Query behavior should be conservative and cheap enough to avoid harming current-state dashboard responsiveness.
- Python may support offline enrichment or validation, but implementation must keep live runtime responsibility in Go services and the web app.

## Assumptions

- `visibility-dashboard-core` already defines the main dashboard reading order and can reserve one explicit slot for slower institutional context.
- CME context is represented as end-of-session or delayed-publish values rather than realtime tick-by-tick derivatives microstructure.
- ETF flow context is represented as daily net flow snapshots, with delayed publication relative to intraday trading.
- Initial MVP uses service-owned polling or scheduled fetches rather than a streaming model for slow sources.

## Safe Default Policies

### Stale Handling

- CME volume and open interest default to `stale` once the latest published point is older than 36 hours.
- ETF daily flow defaults to `stale` once the latest published point is older than 48 hours.
- Before crossing the stale threshold, content may be marked `delayed` when it is beyond its expected publish window but still inside the maximum tolerated age.
- When stale or missing, the system preserves the last known value only with explicit age labeling and advisory messaging.

### Query Cadence

- Service ingestion for slow context should default to scheduled polling no faster than every 15 minutes during expected publish windows.
- Outside expected publish windows, services should back off to a coarse cadence such as hourly to avoid waste and provider coupling.
- Web queries should reuse dashboard current-state requests or a dedicated lightweight slow-context endpoint, but should not poll faster than once per minute for MVP.

### Operator Messaging

- Default label: `Context only`.
- Fresh content message: `Institutional context updated on slower cadence; use it to explain, not gate, the live state.`
- Delayed content message: `Slow context publish is delayed; live market state still reflects realtime feeds.`
- Stale or missing content message: `Slow context unavailable or stale; do not treat this as a realtime market-state signal.`

## Design Overview

### Trust Boundary

- Go services own source fetch timing, normalization, freshness classification, and query outputs.
- `apps/web` renders those outputs, labels cadence, and explains missing or stale state.
- The UI must not infer hidden freshness or reclassify stale thresholds locally.
- Python can assist offline fixture generation or parity checks, but never as a live dependency.

### Context Semantics

- Slow context is explanatory and comparative, not causal proof of immediate market behavior.
- CME volume and open interest help frame institutional participation and positioning backdrop.
- ETF daily flow helps frame slower USA demand or supply pressure.
- In MVP, these inputs can influence operator interpretation but not hard realtime gating logic.

### ASCII Flow

```text
scheduled source fetches
   |                \
   |                 +--> CME volume / OI source(s)
   |                 +--> ETF daily flow source(s)
   v
slow-context ingestion boundary in Go services
   |
   v
services/normalizer or dedicated slow-context normalizer
   |
   +--> normalized slow-context records
   |         - symbol scope: BTC / ETH when applicable
   |         - source timestamp
   |         - ingest timestamp
   |         - cadence metadata
   |         - freshness state
   |
   v
service-owned storage + query surface
   |
   +--> current market-state queries remain independent
   |
   v
apps/web dashboard
   |
   +--> market-state panels (realtime, gating)
   +--> slow context panel (context only, slower cadence)
   |
   v
operator reads live state first, slow context second
```

## Ordered Delivery Notes

- Start with ingestion boundaries so later implementation does not hard-code a vendor or assume realtime semantics.
- Define normalized record shape and persistence before UI language so every surface shares the same freshness rules.
- Add the dashboard panel only after query semantics clearly guarantee non-blocking behavior.
- Test negative cases first-class: stale, missing, delayed, and partially available data.
- Preserve a clear product rule: slow context can explain `why now feels different`, but it cannot quietly decide `can I trade` in MVP.
