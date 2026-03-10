# Implementation: Fixtures And Integration Proof

## Module Requirements

- Add deterministic Binance depth fixtures that prove startup alignment behavior for happy and failure-adjacent bootstrap cases.
- Add targeted integration coverage that proves the completed Spot runtime boundary can feed depth frames into the bootstrap owner and shared normalization path.
- Keep proof focused on initial bootstrap and buffering behavior; recurring resync, refresh, cooldown, and snapshot-staleness policy remain out of scope.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`
- `tests/integration`

## Key Decisions

- Add at least one happy bootstrap fixture showing snapshot plus bridging delta acceptance for Spot depth.
- Add at least one failure-adjacent fixture showing stale or non-bridging buffered deltas that leave startup unsynchronized without silently accepting bad state.
- Reuse or extend the existing Binance sequence-gap fixture only when it directly clarifies the startup boundary; keep later resync-loop proof for the next child feature.
- Follow the completed Spot handoff pattern by proving supervisor/runtime accepted frames flow into the new bootstrap owner and then into canonical normalization without live Binance dependency during implementation.

## Unit Test Expectations

- fixture-backed integration proves synchronized bootstrap emits canonical snapshot/delta outputs in deterministic order
- startup failure cases remain explicit and do not emit silently synchronized depth state
- accepted depth outputs stay distinct from top-of-book behavior and preserve existing order-book event semantics
- fixture and manifest updates stay aligned with the existing shared contract and parity validation path

## Summary

This module closes the proof gap between unit-tested bootstrap alignment logic and the full in-repo depth startup path before recovery policy is added.
