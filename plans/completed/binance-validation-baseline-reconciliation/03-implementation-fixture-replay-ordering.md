# Implementation: Fixture Replay Ordering

## Requirements And Scope

- Make `CONTRACT_FIXTURES=1 make replay-smoke` pass by reconciling replay seed expectations with referenced fixture canonical records.
- Preserve deterministic replay ordering for the same fixture manifest and seed inputs.
- Do not change `scripts/dev/replay_smoke.py` unless implementation proves the script no longer reflects the intended replay contract.

## Target Files

- `tests/replay/seeds/btc-normal-microstructure-window.seed.v1.json`
- `tests/replay/seeds/btc-fragmented-world-usa-window.seed.v1.json`
- `tests/replay/seeds/eth-degraded-feed-window.seed.v1.json`
- `tests/fixtures/events/**/*fixture.v1.json` as the source of canonical `sourceRecordId` values
- `tests/fixtures/manifest.v1.json`
- `tests/replay/manifest.v1.json`
- `scripts/dev/replay_smoke.py` only if fixture materialization semantics are wrong

## Current Drift

The replay smoke script materializes ordering from `fixtureRefs` by reading each referenced fixture's `expectedCanonical[].sourceRecordId`. At least the BTC normal seed expects `ticker:2026-03-06T12:00:01.000Z`, while the referenced Coinbase top-book fixture canonical output currently contains `ticker:2026-03-06T12:00:01Z`.

## Implementation Steps

1. Materialize each replay seed's expected source-record ordering from `tests/fixtures/manifest.v1.json` and referenced fixture files.
2. Update seed `expectedDeterminism.orderedSourceRecordIds` entries to match exact fixture canonical values.
3. Keep `expectedDeterminism.eventCount` equal to the materialized canonical event count for each seed.
4. If multiple seeds fail after the first fix, reconcile all affected seed expectations in the same slice.
5. Run `CONTRACT_FIXTURES=1 make replay-smoke` twice or run it once after confirming the script computes the same checksum twice internally.
6. Run `make fixtures-validate` to ensure fixture and seed JSON remains structurally valid.

## Replay And Determinism Notes

- The deterministic ordering source is the fixture manifest order plus each fixture's canonical output order.
- Timestamp spelling must match canonical source record IDs exactly; do not normalize only in replay smoke unless that is a deliberate contract change.
- Fixture replay smoke is offline validation and must not become a live runtime dependency.

## Unit Test Expectations

- Replay smoke fails if seed ordering differs from materialized fixture ordering.
- Replay smoke fails if event counts differ.
- Replay smoke computes a stable deterministic checksum for repeated materialization of the same seed.

## Next-Agent Summary

The likely minimal patch is to update affected replay seed expected source-record IDs to match fixture canonical IDs exactly, especially timestamp strings normalized to `Z`.
