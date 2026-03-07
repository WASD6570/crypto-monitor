# Implementation Module 1: Adapter Boundaries And Feed Health

## Scope

Plan venue adapter responsibilities, stream coverage, and the common health-state outputs that downstream services rely on.

## Target Repo Areas

- `services/venue-binance`
- `services/venue-bybit`
- `services/venue-coinbase`
- `services/venue-kraken`
- `libs/go`
- `configs/*`

## Requirements

- Define one explicit adapter boundary per venue.
- Inventory the MVP stream types per venue:
  - trades
  - top of book or best bid/ask
  - order book snapshots and deltas where applicable
  - funding, open interest, mark/index, and liquidations for perp venues where available
- Define standard connection lifecycle behavior:
  - initial connect
  - auth if ever needed later, but not required for public MVP feeds
  - heartbeat handling
  - reconnect backoff
  - resubscribe behavior
- Define a shared feed-health output model across adapters.
- Define config surfaces for rate limits, staleness thresholds, reconnect thresholds, and snapshot refresh policies.

## Key Decisions To Lock

- Prefer explicit venue services over a generic multi-venue runtime abstraction.
- Keep shared resilience primitives in `libs/go` only when multiple adapters genuinely need them.
- Treat feed health as first-class output, not as a logging afterthought.
- Add `services/venue-kraken/` during implementation if the repo still lacks it.

## Deliverables

- Adapter responsibility matrix by venue and stream
- Shared health-state schema or equivalent contract usage plan
- Config matrix for reconnect, staleness, and rate-limit behavior
- Clear ownership split between venue adapter and normalizer

## Unit Test Expectations

- Connection-state transitions should be testable without live venues.
- Health-state outputs should reflect reconnect loops, staleness, and repeated resyncs.
- Config validation should reject impossible thresholds or missing stream definitions.

## Contract / Fixture / Replay Impacts

- Health outputs must align with canonical contracts and later regime inputs.
- Fixtures should include connection degradation scenarios alongside payload examples.
- Replay should be able to reconstruct when a feed was healthy or degraded.

## Summary

This module defines what each venue adapter owns and how feed trust is exposed. It gives later modules a stable operational vocabulary for healthy, degraded, and stale conditions.
