# Implementation: Spot-Driven Query Assembly

## Module Requirements

- Build one service-owned live assembler that converts accepted Binance Spot state into `SymbolCurrentStateQuery` and `GlobalCurrentStateQuery` inputs for `BTC-USD` and `ETH-USD`.
- Reuse existing composite, bucket, and regime builders instead of inventing a market-state-specific algorithm.
- Keep `usa` explicit as unavailable or partial when no live USA contributor exists.
- Keep slow context non-blocking and preserve machine-readable degradation from feed health, timestamp fallback, and depth recovery posture.

## Target Repo Areas

- `services/market-state-api`
- `services/feature-engine`
- `services/regime-engine`
- `libs/go/features` only if a tiny shared helper extraction is required

## Key Decisions

- Use accepted Binance Spot top-of-book as the primary live price input for the first cutover.
- Read supervisor and depth-recovery state from the completed Binance runtime seams instead of re-parsing raw frames in `market-state-api`.
- Drive `world` from Binance Spot contributor inputs and allow the existing composite builder to surface `usa` unavailability honestly.
- Reuse the bucket processor and regime evaluator to build 30s/2m/5m bucket summaries and symbol/global regime snapshots from the live observation stream.

## Data And Algorithm Notes

- The assembler should own one bounded in-memory read model for the two supported symbols, including:
  - latest accepted top-of-book input per symbol
  - latest feed-health and depth-recovery posture per symbol
  - bucket processor state for 30s, 2m, and 5m windows
  - latest symbol and global regime snapshots needed for query responses
- Prefer accepted exchange timestamps from upstream canonical inputs and preserve explicit timestamp fallback when Binance top-of-book lacks a trustworthy exchange time.
- Use the existing `feature-engine` bucket pipeline in order:
  - build world and usa composite snapshots for the symbol and bucket time
  - observe the combined world/usa snapshot into the bucket processor
  - advance buckets to the current query time
  - feed 5m buckets into `regime-engine.Observe`
  - assemble symbol and global current-state responses through the existing query builders
- When live data is insufficient, return honest unavailable or degraded sections rather than reusing deterministic fixture data.

## Unit Test Expectations

- `BTC-USD` and `ETH-USD` produce stable symbol query outputs from pinned Binance Spot inputs
- `usa` appears unavailable or partial rather than silently populated with deterministic values
- degraded depth or feed-health posture changes composite, bucket, or regime availability predictably
- repeated query assembly over the same inputs produces identical symbol/global responses
- missing top-of-book input yields explicit unavailable behavior instead of partial fake success

## Contract / Fixture / Replay Notes

- Preserve the existing current-state response contract and reserved history/audit seam from `libs/go/features/market_state_current.go`.
- Reuse completed Binance Spot fixture vocabulary and accepted-input semantics; do not invent a second naming scheme for the same events.
- If helper extraction is needed for bucket or regime assembly, keep it shared and minimal so replay-sensitive tests can reuse it later without duplicating logic.

## Summary

This module is the actual live read-model slice: it turns accepted Binance Spot state into honest current-state and regime inputs while keeping the existing API contract intact for the later provider cutover.
