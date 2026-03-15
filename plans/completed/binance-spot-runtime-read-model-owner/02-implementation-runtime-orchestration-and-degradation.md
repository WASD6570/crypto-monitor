# Implementation: Runtime Orchestration And Degradation

## Requirements And Scope

- Compose the completed Spot supervisor and depth recovery owners into one command-owned runtime loop.
- Preserve explicit warm-up, reconnect, resync, refresh-due, snapshot-stale, and disconnected semantics in the read model.
- Avoid changing HTTP handlers, response schemas, or browser-facing routes in this slice.

## Target Repo Areas

- `cmd/market-state-api/live_provider.go`
- `cmd/market-state-api/*.go` for runtime loop helpers
- `services/venue-binance/*.go` only if one focused export is needed for orchestration or status inspection

## Implementation Notes

- Drive the Spot websocket supervisor as the single lifecycle owner for live stream state, including connect, subscribe, ping/pong, reconnect, and rollover transitions.
- Feed accepted Spot frames into the existing depth bootstrap and recovery surfaces so top-of-book and synchronized depth status stay aligned with the already-settled semantics.
- Treat runtime degradation as first-class read-model state. The owner should publish the latest trustworthy observation together with updated `FeedHealth` and `DepthStatus` when reconnects, sequence gaps, snapshot refresh due, or stale conditions occur.
- Keep warm-up honesty explicit: before a symbol has any trustworthy observation, snapshot reads should omit that symbol and let the provider assemble partial or unavailable responses.
- Record timing and state transitions in a way tests can drive deterministically with a fake clock or scripted timestamps instead of real sleeping loops.
- Keep shutdown and reconnect behavior bounded so the command can terminate cleanly even while a background connection or resync attempt is in flight.

## Unit Test Expectations

- Initial connect and subscribe posture transitions to a publishable observation only after accepted runtime data arrives.
- Gap-triggered resync updates degradation state immediately and returns to synchronized status only after replacement snapshot progression succeeds.
- Reconnect and rollover preserve prior read-model observations while connection state becomes degraded or stale in a machine-readable way.
- Stale snapshot or blocked refresh state remains visible in `FeedHealth` reasons and depth recovery status.
- Repeated scripted input sequences produce the same observation ordering and timestamps on repeated test runs.

## Summary

This module wires the sustained owner to real Binance runtime behavior instead of an on-demand snapshot call. It keeps every temporary failure mode explicit in read-model output so later provider cutover can remain a thin consumer-facing change.
