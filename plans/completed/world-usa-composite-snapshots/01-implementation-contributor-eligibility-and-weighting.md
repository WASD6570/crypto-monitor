# Implementation Contributor Eligibility And Weighting

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `libs/go`, `configs/*`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Build deterministic contributor selection and weighting for `WORLD` and `USA` composite groups for `BTC-USD` and `ETH-USD`.
- Keep all logic bounded to eligibility, penalties, normalization, clamping, and exclusion reasons.

## In Scope

- configured venue membership for `WORLD` and `USA`
- symbol and market-type eligibility gates
- timestamp trust and feed-health penalties used in contributor decisions
- quote-normalization allow/deny handling for WORLD contributors
- raw weight calculation, normalization, clamp application, and stable tie-breaking
- contributor-level provenance for include, penalize, cap, or exclude decisions

## Out Of Scope

- composite-to-composite divergence metrics
- 30s/2m/5m feature buckets
- regime state classification
- query handlers or dashboard read models

## Recommended Repo Breakdown

- `services/feature-engine`: orchestration that joins canonical venue inputs, feed health, and config into per-group contributor sets and final composite inputs.
- `libs/go`: pure helpers for eligibility evaluation, quote gate decisions, weight normalization, clamp ordering, and contributor summaries.
- `configs/*`: versioned membership lists, quote-proxy allowlists, penalty multipliers, and min/max clamp bounds.
- `tests/fixtures`, `tests/integration`, `tests/replay`: pinned fixture windows and deterministic replay checks.

## Contributor Group Defaults

- `WORLD`: Binance and Bybit spot/perp inputs approved by config for the symbol.
- `USA`: Coinbase and Kraken spot inputs approved by config for the symbol.
- Membership remains config-driven so later venue expansion does not require consumer contract redesign.
- The implementation should distinguish:
  - configured members
  - eligible members for the bucket
  - contributing members after penalties and exclusions

## Eligibility Pipeline

1. Start from configured group membership for symbol and market type.
2. Reject contributors with unusable canonical inputs for the bucket window.
3. Apply timestamp trust rules using `exchangeTs` first and `recvTs` fallback with explicit degraded status.
4. Apply feed-health gates for stale, gap-flagged, reconnect-loop, or otherwise degraded venues.
5. Apply quote-normalization allow/deny rules for WORLD contributors.
6. Produce contributor status as one of: included, penalized, clamped, or excluded with explicit reasons.

## Quote-Normalization Policy

- Preserve canonical symbol identity as `BTC-USD` and `ETH-USD` while respecting original quote context from prerequisites.
- Allow `USD` contributors directly.
- Allow `USDC` and `USDT` only through explicit config-backed proxy rules.
- If quote confidence degrades or the proxy is not approved for the config version, exclude the contributor instead of fabricating a normalized value.
- Quote decisions must be replay-pinned and deterministic; they cannot depend on live external lookups.

## Weighting And Clamp Policy

- Use bounded deterministic weighting rather than equal weighting.
- Recommended raw inputs:
  - recent liquidity or notional-quality proxy from canonical data already available to the feature engine
  - feed-health penalty multiplier
  - timestamp-degraded penalty multiplier
  - quote-confidence penalty multiplier for WORLD contributors
- Recommended execution order:
  1. compute raw quality scores for eligible contributors
  2. apply multiplicative penalties
  3. normalize to sum to `1.0`
  4. clamp each contributor to config-driven min/max bounds
  5. renormalize once using stable contributor ordering
- Record pre-clamp and post-clamp weights for replay validation and auditability.

## Conservative Trust Rules

- A stale, gap-flagged, or quote-untrusted contributor must go to zero contribution.
- A timestamp-degraded or partially unhealthy contributor may remain only at reduced influence when the input is still otherwise coherent.
- No single contributor may dominate while at least one healthy peer remains eligible.
- If all contributors are excluded, emit an unavailable composite input state instead of a synthetic fallback price.

## Determinism Notes

- Use stable contributor ordering whenever raw scores or penalties tie.
- Keep penalty composition and clamp order explicit in code and config.
- Avoid map-order dependence, wall-clock lookups, or mutable external references.
- Replay of the same fixture window and config snapshot must reproduce the same contributor statuses and final weights.

## Unit And Integration Test Expectations

- `go test ./libs/go/... -run 'Test(CompositeWeighting|StablecoinNormalization|CompositeClamping|ContributorEligibility)'`
- `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeConstruction|CompositeDegradedVenueHandling|CompositeAllContributorsExcluded)'`
- `go test ./tests/replay/... -run 'TestWorldUSACompositeDeterminism'`

## Summary

This module gives the next implementer one bounded Go slice: decide who can contribute, how much they can influence the snapshot, and why any venue was penalized, capped, or excluded. Later bucket, regime, and query work should consume these decisions, not recreate them.
