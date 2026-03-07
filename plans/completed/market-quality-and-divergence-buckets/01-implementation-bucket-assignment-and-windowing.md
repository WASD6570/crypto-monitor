# Implementation Bucket Assignment And Windowing

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `libs/go/features`, `configs/local`, `configs/dev`, `configs/prod`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Consume completed WORLD and USA composite snapshots plus feed-health-derived degraded markers from the prior slice.
- Assign deterministic 30s buckets in UTC and derive 2m and 5m windows from closed 30s buckets.
- Preserve bucket source, timestamp trust, late-event disposition, and config provenance for every emitted bucket.

## In Scope

- 30s base bucket alignment and closure behavior
- 2m and 5m rollups built from previously closed 30s buckets
- timestamp-source selection and degraded fallback markers
- watermark handling for on-time, within-watermark, and too-late events
- bounded in-memory or service-local rollup state needed by the live path
- fixture expectations for missing buckets, sparse contributors, and replay reuse

## Out Of Scope

- metric formulas beyond what is required to carry deterministic window inputs forward
- final bucket output schemas for regime/query consumers beyond the needed internal seam names
- any storage or backfill correction mechanism for late events beyond marking the disposition and preserving replay compatibility

## Target Structure

- `libs/go/features/buckets.go`
- `libs/go/features/buckets_test.go`
- `services/feature-engine/service.go` extensions for bucket building APIs
- `services/feature-engine/service_test.go` additions for windowing behavior
- `configs/*/feature-engine.market-quality.v1.json` or a similarly explicit feature-engine config file
- focused fixture windows under `tests/fixtures/world_usa_buckets/`

## Bucket Model

- Use 30s as the canonical base interval for all symbols.
- Derive 2m windows from 4 contiguous closed 30s buckets.
- Derive 5m windows from 10 contiguous closed 30s buckets.
- Keep all windows aligned to UTC boundaries so replay and live code share the same bucket math.
- Prefer rollup-from-30s over direct multi-window recomputation so every family is traceable to one canonical base series.

## Recommended Internal Types

- `BucketFamily`: `30s`, `2m`, `5m`
- `BucketSource`: `exchangeTs`, `recvTs`
- `LateEventDisposition`: `on-time`, `within-watermark`, `after-watermark`
- `BucketAssignment`: symbol, family, bucket start/end, chosen source, degraded flag, lateness
- `BucketWindowState`: ordered 30s closures plus rollup readiness metadata
- `BucketConfig`: watermark seconds, minimum closed-source buckets, and rollup completeness thresholds

## Assignment Rules

1. Choose `exchangeTs` when present and plausible.
2. Fall back to `recvTs` when `exchangeTs` is missing, invalid, or outside the sane skew window.
3. Mark every fallback event as timestamp-degraded and carry the chosen source into bucket metadata.
4. Floor timestamps to the UTC bucket start for the target family.
5. Close 30s buckets only after the configured watermark expires.
6. Build 2m and 5m rollups only from closed 30s buckets in deterministic order.

## Watermark And Lateness Policy

- Preserve operating-default watermarks:
  - 30s buckets: 2s
  - 2m buckets: 5s
  - 5m buckets: 10s
- Live path behavior:
  - on-time events update the open 30s bucket normally
  - within-watermark late events may still affect the still-open bucket before closure
  - after-watermark late events are marked for replay correction and must not silently mutate already-emitted live bucket outputs
- Replay behavior:
  - reprocess the same fixture window in canonical order
  - reproduce the same bucket series for identical input order and config
  - surface intentional differences only when config version changes

## Sparse And Missing Data Handling

- If a composite snapshot is unavailable for a 30s interval, still emit the bucket with explicit unavailable or incomplete markers rather than skipping the time slot.
- Rollups should carry:
  - `closedBucketCount`
  - `expectedBucketCount`
  - `missingBucketCount`
  - `containsAfterWatermarkLateEvent`
- 2m and 5m rollups stay deterministic even when incomplete; later regime logic can decide how to degrade trust from that summary.

## Config Guidance

- Keep this feature's config separate from ingestion config so bucket behavior is replay-pinned without coupling to venue transport settings.
- Minimum config sections:
  - family intervals and watermark settings
  - acceptable timestamp skew window
  - minimum completeness thresholds for 2m and 5m outputs
  - any tie-break rules for bucket closure and rollup emission
- Config fields should version cleanly across `configs/local`, `configs/dev`, and `configs/prod`.

## Unit Test Expectations

- `TestBucketAssignment` for clean `exchangeTs` assignment across all three families
- `TestBucketAssignmentRecvTsFallback` for degraded timestamp fallback
- `TestBucketRollupFrom30sClosures` for deterministic 2m and 5m construction
- `TestLateEventHandling` for within-watermark versus after-watermark behavior
- `TestMissing30sBucketPropagation` for incomplete 2m and 5m rollups

## Integration And Replay Expectations

- integration coverage should build WORLD and USA snapshots first, then prove bucket windows are emitted in stable UTC order
- replay coverage should run the same pinned fixture window twice and compare full bucket outputs
- replay coverage should include a late event that is ignored live after watermark but accepted in a replay-correction path later

## Summary

This module establishes the deterministic time boundary for the feature. It owns timestamp choice, 30s canonical closure, and the 2m and 5m rollup chain so later divergence and quality metrics never need to guess which events belonged to which window.
