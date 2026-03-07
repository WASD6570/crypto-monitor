# Implementation Composite Construction And Weighting

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `libs/go`, `configs/*`, `schemas/json/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Build service-owned WORLD and USA composite inputs for `BTC-USD` and `ETH-USD` only.
- Define venue membership, eligibility gates, stablecoin normalization policy, weighting, and clamping behavior.
- Preserve enough provenance for downstream services and UI to explain why venues were included, penalized, capped, or excluded.

## In Scope

- venue group membership for WORLD and USA
- quote-normalized composite input selection
- liquidity-quality weighting and per-venue clamp policy
- degraded feed penalties and venue exclusion rules
- composite-level health and provenance fields

## Out Of Scope

- 30s/2m/5m feature math beyond fields needed from composite snapshots
- final regime classification
- UI rendering or chart aggregation logic

## Recommended Service Boundary

- `services/feature-engine` owns live composite assembly from canonical event and feed-health streams.
- `libs/go` may hold pure deterministic helpers for weighting, clamping, normalization checks, and contribution summaries.
- `schemas/json/features` should define versioned payloads for:
  - per-venue composite input snapshot
  - WORLD composite snapshot
  - USA composite snapshot
  - composite provenance and degradation reasons

## Venue Membership Defaults

- WORLD: Binance and Bybit spot/perp inputs approved by config for the symbol.
- USA: Coinbase and Kraken spot inputs approved by config for the symbol.
- Membership must be config-driven so future venues can be added without changing consumer contracts.
- Composite payloads should distinguish configured members, currently eligible members, and contributing members.

## Stablecoin Normalization Recommendations

- Inputs carry canonical symbol plus original quote context.
- Before weighting, map approved WORLD quotes into USD-equivalent values using explicit config for trusted proxy relationships.
- Safe MVP default:
  - `USD` requires no proxy adjustment.
  - `USDC` may be treated as near-par if configured as trusted and healthy.
  - `USDT` may be treated as a permitted proxy only when explicitly enabled and not currently degraded by quote-confidence checks.
- If a quote proxy is disallowed or degraded, exclude that venue from the affected composite and record the reason.
- Do not let the UI or offline research code decide live quote-proxy inclusion.

## Weighting Policy Recommendations

- Use a deterministic bounded weight formula instead of equal weighting.
- Recommended input factors:
  - recent executable or observed notional turnover proxy
  - spread or order-book quality proxy when available from canonical features
  - feed-health penalty multiplier
  - timestamp degradation penalty multiplier
  - quote-normalization confidence multiplier for WORLD stablecoin venues
- Safe default weighting flow:
  1. compute raw quality weights from configured liquidity inputs
  2. apply degradation penalties
  3. normalize weights to sum to 1 among eligible contributors
  4. clamp each venue to configured min/max contribution bounds
  5. renormalize once after clamping using stable venue order for determinism

## Clamping Policy Recommendations

- Clamp to prevent one venue from dominating composite state during local surges or degraded peer conditions.
- Keep clamp thresholds config-driven by venue group and market type.
- Safe default posture:
  - no venue should reach 100 percent contribution while at least one healthy peer exists
  - a degraded venue may remain present at reduced weight if still timely and internally consistent
  - a stale, gap-flagged, or quote-untrusted venue should move to zero contribution
- Record both pre-clamp and post-clamp weights for audit and replay comparison.

## Composite Output Fields

- `symbol`
- `bucketTs`
- `compositeGroup` as `WORLD` or `USA`
- `priceBasis` and `quoteNormalizationMode`
- `contributors[]` with venue, market type, raw weight, final weight, penalties, include/exclude reason
- `compositePrice`
- `compositeReturn` or equivalent short-horizon delta input
- `healthScore` or equivalent bounded quality field
- `coverageRatio` for configured vs contributing venues
- `degraded` flag plus explicit reasons
- `configVersion`, `algorithmVersion`, `schemaVersion`

## Negative And Edge Cases

- one configured venue missing entirely for a bucket
- one venue stale while another is healthy and moving sharply
- timestamp-degraded events forcing fallback to `recvTs`
- quote-normalization confidence loss for `USDT` or `USDC`
- transient single-venue spike that would dominate without clamping
- all contributors excluded, producing an unavailable composite instead of a fabricated number

## Determinism Notes

- Use stable contributor ordering when equal scores occur.
- Keep all weight and clamp thresholds in the replayed config snapshot.
- Avoid weight formulas that depend on wall-clock state or mutable external reference data.
- Persist enough provenance to recompute the same composite during replay and verify exact contributor decisions.

## Unit And Integration Test Expectations

- unit tests for quote-normalization allow/deny decisions
- unit tests for weight normalization and clamp application order
- unit tests for degraded penalty composition and exclusion rules
- integration tests for mixed healthy/degraded venue sets producing stable composite outputs
- replay tests confirming the same raw fixture window reproduces identical contributor sets and weights

## Summary

This module defines how WORLD and USA composites are constructed, which venues count, and how they are safely weighted without pushing business logic into the UI. The critical implementation details are explicit quote normalization, bounded contribution clamping, degraded-input penalties, and replay-safe provenance for every contributor decision.
