# Implementation: Bootstrap Owner And Buffer State

## Module Requirements

- Add one bounded Spot depth bootstrap owner that coordinates snapshot fetch, buffered delta retention, and synchronized startup state for BTC/ETH.
- Keep ownership in `services/venue-binance`; do not move bootstrap orchestration into `libs/go/ingestion` or `services/normalizer`.
- Reuse the completed Spot supervisor as the data-frame source rather than creating a second independent websocket loop.
- Define explicit bootstrap state transitions such as idle, buffering, snapshot-requested, synchronized, and bootstrap-failed only if those states materially improve testability and handoff clarity.

## Target Repo Areas

- `services/venue-binance`
- `configs/*/ingestion.v1.json` only if the bootstrap owner needs an existing config field surfaced more explicitly

## Key Decisions

- Accept supervisor-fed `depthUpdate` frames as buffered native inputs before synchronization.
- Parse REST snapshots through `ParseOrderBookSnapshot(...)` and websocket deltas through `ParseOrderBookDelta(...)` so this feature does not fork parser behavior.
- Keep bootstrap-owner outputs narrow: synchronized snapshot message, ordered accepted buffered deltas, and explicit bootstrap failure reasons for the later recovery slice.
- Avoid introducing refresh or recurring resync orchestration here; only the initial bootstrap boundary belongs in this module.

## Unit Test Expectations

- buffered deltas are retained in arrival order before synchronization
- snapshot bootstrap refuses zero or malformed snapshot state
- startup state can surface when synchronization has not yet been achieved
- successful bootstrap emits one parsed snapshot plus the ordered accepted delta list needed by the next module

## Summary

This module establishes the only startup owner for Spot depth synchronization so the later acceptance logic does not leak into generic supervisor code or future recovery policy.
