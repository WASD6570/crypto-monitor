# Implementation: Fixtures And Determinism Proof

## Module Requirements

- Add deterministic proof that the new live query assembler produces stable symbol and global current-state responses from pinned Binance Spot inputs.
- Cover healthy, degraded, and unavailable paths without requiring live Binance network access.
- Keep verification attached to this feature instead of creating a validation-only follow-on slice.

## Target Repo Areas

- `services/market-state-api`
- `tests/integration`
- `tests/replay`
- `tests/fixtures/events/binance` only if one additional deterministic fixture is truly needed

## Key Decisions

- Reuse the completed Binance Spot fixture corpus and current-state contract assertions wherever possible.
- Add focused test helpers near the owning tests instead of creating a broad cross-repo fixture harness.
- Keep repeated-run determinism checks in Go so the live read model remains replay-safe and auditable.

## Data And Algorithm Notes

- Cover at least these cases:
  - healthy Spot top-of-book input for both symbols with explicit `usa` unavailability
  - degraded timestamp or feed-health input that lowers availability without changing the contract shape
  - depth unsynchronized or gap-visible posture that remains machine-readable in current-state output
  - repeated identical input runs yielding identical current-state symbol/global responses
- Integration proof should validate assembled responses, not just intermediate package structs.
- Replay-style proof should assert that stable accepted inputs yield stable query outputs and version fields across repeated runs.

## Unit Test Expectations

- `services/market-state-api` tests cover live-source construction and symbol/global response assembly
- integration tests verify current-state contract stability for Binance-backed symbol/global responses
- replay-style tests prove repeated runs stay byte-for-byte or structurally identical for pinned inputs
- no test depends on live websocket or REST access

## Contract / Fixture / Replay Notes

- Preserve existing schema family versions and response field names.
- If new fixtures are added, keep them deterministic and scoped to this feature's live assembly needs.
- Any determinism drift found here should block implementation until the source seam or bucket/regime assembly is corrected.

## Summary

This module proves the live query assembly is stable enough to hand to the later provider-cutover slice without relying on manual browser checks or live exchange traffic.
