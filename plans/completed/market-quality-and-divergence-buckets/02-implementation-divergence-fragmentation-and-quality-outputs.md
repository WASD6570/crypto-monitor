# Implementation Divergence Fragmentation And Quality Outputs

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `libs/go/features`, `configs/local`, `configs/dev`, `configs/prod`, `schemas/json/features` only if a bucket seam schema is reserved now, `tests/fixtures`, `tests/integration`, `tests/replay`
- Transform closed WORLD and USA bucket inputs into deterministic summaries for divergence, fragmentation, coverage, timestamp trust, and market quality.
- Expose service-owned bucket outputs that later regime and query slices can read without recomputing feature math.
- Keep final classification and query contract design explicitly out of scope.

## In Scope

- divergence metric families for WORLD versus USA bucket comparisons
- fragmentation severity summaries and reason codes
- coverage summaries and missing-contributor context derived from composite snapshot seams
- timestamp trust summaries built from bucket-source and fallback counts
- market-quality summaries that conservatively cap trust without naming final state classes
- config-versioned thresholds and severity cutoffs for the bucket layer

## Out Of Scope

- mapping bucket summaries to `TRADEABLE`, `WATCH`, or `NO-OPERATE`
- global ceiling logic
- query route design, pagination, or current-state response envelopes
- UI labels, colors, copy, or client-side transforms

## Target Structure

- `libs/go/features/market_quality.go`
- `libs/go/features/market_quality_test.go`
- `services/feature-engine/service.go` extensions for bucket summary emission
- `services/feature-engine/service_test.go` additions for summary seams
- optional `schemas/json/features/world-usa-market-quality-bucket.v1.schema.json`
- fixture windows under `tests/fixtures/world_usa_buckets/`

## Recommended Output Shape

One service-owned bucket output per symbol, family, and bucket end time with these top-level groups:

- `window`: family, start/end timestamps, closed bucket counts, config/algorithm version
- `world`: composite availability, coverage ratio, health score, max contributor weight, timestamp fallback count
- `usa`: composite availability, coverage ratio, health score, max contributor weight, timestamp fallback count
- `divergence`: price-distance, directional disagreement, participation gap, leader-churn summary, reason codes
- `fragmentation`: severity, persistence count, primary causes, unavailable-side markers
- `timestampTrust`: bucket source mix, fallback ratios, degraded counts, trust cap flag
- `marketQuality`: world quality cap, usa quality cap, combined trust cap, downgraded reasons, replay provenance seam

## Divergence Recommendations

- Keep divergence as a multi-part summary, not one opaque score.
- Minimum deterministic metrics:
  - `priceDistanceBps`: normalized WORLD versus USA composite distance
  - `directionAgreement`: aligned, mixed, or opposed movement over the active family window
  - `participationGap`: difference between WORLD and USA coverage ratios
  - `leaderChurnGap`: whether top-weight contributors changed materially across recent 30s closures
- If one side is unavailable, emit explicit unavailable markers instead of synthesizing a distance value.

## Fragmentation Recommendations

- Fragmentation should summarize whether the tape is interpretable, not whether it is merely volatile.
- Suggested severity ladder:
  - `low`: composites aligned and coverage stable
  - `moderate`: disagreement, missing peers, or contributor churn is visible but not persistent enough for a hard stop
  - `severe`: sustained disagreement, asymmetric degradation, or unavailable composite state
- Suggested reason families:
  - `price-divergence`
  - `directional-disagreement`
  - `coverage-asymmetry`
  - `timestamp-trust-loss`
  - `contributor-churn`
  - `composite-unavailable`

## Coverage And Timestamp Trust

- Coverage summary should stay close to the completed composite seam:
  - configured contributors
  - eligible contributors
  - contributing contributors
  - coverage ratio by side
  - missing configured peer markers
- Timestamp trust summary should include:
  - fallback contributor counts by side
  - fallback ratio for the bucket family
  - whether `recvTs` fallback touched only one side or both sides
  - whether timestamp degradation alone should cap market quality

## Market-Quality Summary

- Market quality is a conservative trust summary for downstream readers.
- Minimum inputs:
  - WORLD composite health score
  - USA composite health score
  - coverage asymmetry
  - timestamp-trust degradation
  - contributor concentration after clamping
  - fragmentation severity
- Safe default posture:
  - any severe fragmentation or unavailable side should cap quality sharply
  - critical timestamp-trust loss or low completeness should cap quality even when prices look aligned
  - 5m quality summary is the seam later regime work reads; it must not already encode the final class name

## Config Guidance

- Keep thresholds boring and explicit.
- Minimum config sections:
  - price-distance thresholds by family
  - contributor-churn thresholds by family
  - coverage asymmetry thresholds
  - timestamp-trust cap thresholds
  - fragmentation severity cutoffs
  - market-quality cap rules and tie-break order
- Config must produce deterministic edges for equality cases; define inclusive or exclusive comparisons once.

## Unit Test Expectations

- `TestDivergenceMetrics` for aligned, mixed, and opposed WORLD versus USA moves
- `TestFragmentationSeverity` for low, moderate, and severe cases
- `TestCoverageAndTimestampTrustSummaries` for asymmetric missing peers and fallback-heavy buckets
- `TestMarketQualityCaps` for fragmentation, concentration, and timestamp-driven downgrades
- `TestUnavailableCompositeSummary` for one-side and two-side unavailable cases

## Integration And Replay Expectations

- integration tests should prove summary seams are emitted from prior composite snapshots without reusing raw venue logic directly
- integration tests should cover clean market, fragmented market, and mixed coverage cases for both `BTC-USD` and `ETH-USD`
- replay tests should prove identical fixture windows emit byte-for-byte equal bucket summaries for the same config snapshot
- replay tests should prove a config version change alters outputs intentionally while preserving version fields and explainable reason codes

## Downstream Seams Only

- Later `symbol-and-global-regime-state` should read 5m bucket summaries such as fragmentation severity, market-quality caps, coverage completeness, and timestamp-trust flags.
- Later `market-state-current-query-contracts` should package emitted bucket summaries as-is, plus version metadata, instead of recomputing any component metric.

## Summary

This module defines the bucket outputs that the rest of the epic depends on. It must stay conservative, explainable, and classification-free: later readers get explicit divergence and quality evidence, not hidden regime decisions.
