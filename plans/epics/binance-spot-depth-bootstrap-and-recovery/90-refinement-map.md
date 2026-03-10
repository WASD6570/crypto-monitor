# Refinement Map: Binance Spot Depth Bootstrap And Recovery

## Current Status

This Wave 3 epic is newly materialized from the initiative handoff and still needs decomposition before `feature-planning`.

## What Is Already Covered

- `plans/completed/binance-spot-ws-runtime-supervisor/` already fixes the Spot websocket lifecycle owner for BTC/ETH Spot streams and feed-health posture outside depth-specific sequence recovery
- `plans/completed/binance-spot-trade-canonical-handoff/` and `plans/completed/binance-spot-top-of-book-canonical-handoff/` already lock the non-depth Spot handoff seams that this epic must not reopen
- `services/venue-binance/orderbook.go` and `services/venue-binance/orderbook_test.go` already parse Binance depth REST snapshots and websocket deltas into the shared ingestion order-book model
- `services/venue-binance/runtime.go` and `services/venue-binance/runtime_test.go` already provide snapshot recovery cooldown, rate-limit, reconnect-loop, resync-loop, and staleness evaluation primitives
- `configs/local/ingestion.v1.json` already carries snapshot-recovery, snapshot-refresh, stale-message, and resync-loop defaults for Binance
- `tests/fixtures/events/binance/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json` already seeds the expected degraded feed-health outcome for a depth sequence gap

## What Remains

- define one bounded depth bootstrap slice that combines REST snapshot bootstrap with buffered `depth@100ms` startup alignment under the existing Spot runtime boundary
- define one bounded recovery slice that handles resync triggers, cooldown and rate-limit behavior, snapshot refresh cadence, and feed-health degradation for gaps or stale snapshot state
- keep accepted depth outputs compatible with existing shared order-book normalization and explicit provenance rules
- reserve direct live Binance validation for after implementation instead of creating a validation-only child feature

## What This Epic Must Not Absorb

- trade or top-of-book parsing already completed in prior Spot child features
- raw append or replay rollout mechanics from the later replay epic
- market-state API wiring or frontend/dashboard cutover work
- generalized multi-venue sequencer redesign beyond what concrete Binance depth behavior requires

## Refinement Waves

### Wave 3A

- `binance-spot-depth-bootstrap-and-buffering`
- Why first: buffered startup alignment is the minimum reliable depth path and defines the sequence boundary the later recovery slice must honor.

### Wave 3B

- `binance-spot-depth-resync-and-snapshot-health`
- Why next: resync loops, snapshot refresh cadence, and degraded health semantics depend on the bootstrap slice settling how depth state starts and what counts as accepted progression.

## Notes For Future Planning

- extend the completed Spot supervisor only as needed for depth stream ownership; do not duplicate connection lifecycle logic already settled for Spot
- keep Binance depth acceptance explicit about `U`/`u` alignment and snapshot `lastUpdateId` handling
- attach fixtures and deterministic tests to the owning implementation slices instead of creating a separate validation-only child feature
- keep direct live Binance validation as a post-implementation check for the completed depth path
