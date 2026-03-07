# Binance Adapter Loop Runtime

## Ordered Implementation Plan

1. Finish the mutable loop-state surface around the existing Binance runtime helpers.
2. Add bounded transition helpers for gaps, reconnects, resyncs, and snapshot recovery history.
3. Add a tiny in-memory runtime loop harness that turns state into shared feed-health decisions.
4. Validate healthy, degraded, stale, and recovery transitions with synthetic time.

## Problem Statement

`services/venue-binance` already has parser boundaries and deterministic health-building blocks, but it still lacks the small state-machine layer that a future live loop can call directly.

## Requirements

- Keep the live path in Go under `services/venue-binance`.
- Reuse the existing shared vocabulary in `libs/go/ingestion`.
- Treat connection state, sequence gaps, reconnect counts, resync counts, snapshot recovery, and clock offset as first-class loop inputs.
- Keep transitions deterministic under synthetic time.
- Avoid live websocket or REST dependencies in this slice.

## Out Of Scope

- Real Binance network clients.
- Full subscription orchestration.
- Multi-venue abstractions.

## Design Notes

- Keep `AdapterLoopState` as the mutable state holder.
- Add only small explicit mutators and one thin harness/driver entry point.
- Preserve a clear split between state mutation, shared health snapshot building, and final decision derivation.

## Target Repo Areas

- `services/venue-binance`
- `libs/go/ingestion`

## ASCII Flow

```text
loop events
  - message recv
  - snapshot recv
  - gap detected
  - reconnect
  - resync
  - snapshot recovery
        |
        v
AdapterLoopState mutators
        |
        v
EvaluateLoopState(now)
        |
        v
shared FeedHealthStatus
```
