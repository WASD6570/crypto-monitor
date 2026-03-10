# Implementation: Fixtures And Binance Replay Proof

## Module Requirements

- Add deterministic replay proof for representative Binance raw entries produced by the completed raw-append slice.
- Reuse the completed Binance fixture corpus and raw append seams where practical instead of inventing parallel replay-only payload vocabularies.
- Cover both happy-path replay acceptance and degradation-sensitive cases that materially affect replay audit evidence.

## Target Repo Areas

- `tests/replay`
- `tests/integration`
- `services/replay-engine`
- `services/venue-binance`

## Key Decisions

- Prefer replay fixtures assembled from settled Binance raw append outputs over handwritten replay-only entries when practical.
- Keep end-to-end proof focused on replay facts that matter for auditability:
  - stable ordered IDs across repeated runs
  - stable duplicate counters
  - preserved degraded timestamp and feed-health evidence
  - distinct replay treatment for USD-M websocket vs REST surfaces
- Attach the replay proof directly to this feature instead of creating a validation-only child feature.

## Integration Expectations

- one replay proof covers Spot websocket-originated families using shared venue partitions
- one replay proof covers Spot depth degraded feed-health using dedicated feed-health partitions and retained degraded linkage
- one replay proof covers USD-M mixed websocket and REST families using distinct source identities and stream-family partitions
- one duplicate-input proof covers deterministic replay counters and digest behavior for repeated Binance raw entries

## Unit Test Expectations

- fixture-backed replay helpers can load representative Binance raw entries without contract drift
- replay assertions preserve ordered IDs, duplicate counters, and degraded evidence across inspect, rebuild, and compare modes
- deterministic repeated runs do not drift in ordered IDs, compare classification, or replay digests

## Contract / Fixture / Replay Notes

- Reuse the completed raw-append archive and touched replay helpers as source context; do not reopen the raw entry shape.
- If new replay fixtures are added, keep them deterministic and scoped to the completed Binance family set only.

## Summary

This module closes the replay audit gap by proving that settled Binance raw entries can be replayed deterministically with unchanged degradation and duplicate evidence before market-state cutover planning continues.
