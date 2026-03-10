# Implementation Module: USD-M Websocket Runtime And Subscription Shape

## Scope

- Add the bounded Binance USD-M websocket runtime that owns subscription setup, reconnect/backoff, resubscribe, and per-symbol runtime state for `markPrice@1s` and `forceOrder`.
- Exclude REST `openInterest`, Spot stream handling, and any shared multi-venue supervisor abstraction.

## Target Repo Areas

- `services/venue-binance/README.md`
- `services/venue-binance/usdm_runtime.go`
- `services/venue-binance/usdm_runtime_test.go`
- `services/venue-binance/usdm_mark_price.go`
- `services/venue-binance/usdm_force_order.go`
- `configs/local/ingestion.v1.json`

## Requirements

- Reuse Binance runtime config bounds already defined in `configs/local/ingestion.v1.json` for heartbeat timeout, bounded reconnect backoff, reconnect-loop threshold, and resubscribe-on-reconnect behavior.
- Build the active USD-M stream set from configured canonical symbols so BTC/ETH support stays declarative.
- Keep the runtime Go-only and local to `services/venue-binance`.
- Treat `markPrice@1s` and `forceOrder` as websocket-only stream families in this slice.
- Surface websocket-specific health transitions for connection state, stale `markPrice@1s`, reconnect loops, and resubscribe failure.

## Key Decisions

- Use one dedicated USD-M websocket runtime for this feature rather than sharing Spot runtime machinery; this keeps parallel implementation with the Spot websocket supervisor safe and avoids premature cross-market abstraction.
- Keep a small explicit subscription registry that maps canonical symbol -> source symbol -> stream names. The registry should emit deterministic subscription payloads and deterministic resubscribe order.
- Track message freshness from `markPrice@1s` only. `forceOrder` updates should refresh connection activity when present, but the absence of force-order messages must not be treated as stale.
- Emit adapter-scoped feed-health per tracked symbol so downstream consumers can see that a symbol's USD-M context is degraded even when the websocket connection is shared.
- Reuse the existing runtime health helpers where possible, but add a USD-M-specific loop state if the current order-book-oriented snapshot fields are not a clean fit.

## Unit Test Expectations

- Subscription registry expands configured BTC/ETH symbols into the exact four stream names and keeps a stable ordering.
- Runtime reconnect delay and reconnect-loop detection stay within configured bounds.
- Resubscribe on reconnect sends the full active USD-M stream set exactly once per reconnect attempt.
- `markPrice@1s` staleness degrades after the configured freshness window.
- Empty `forceOrder` periods do not trigger stale status by themselves.
- Connection loss or reconnecting state degrades all tracked symbols consistently.

## Contract / Fixture / Replay Impacts

- No new schema family is expected.
- Fixture work for this module should add raw websocket examples that mirror actual Binance `markPrice@1s` and `forceOrder` payload shapes, including reconnect and stale-timer test inputs.
- Replay-sensitive behavior is limited to stable subscription ordering and deterministic health transitions; raw append and replay implementation remain out of scope for this feature.

## Summary

This module gives `services/venue-binance` one explicit USD-M websocket runtime surface that can subscribe, reconnect, resubscribe, and evaluate health for mark/index, funding, and liquidation traffic without mixing in REST or Spot responsibilities.
