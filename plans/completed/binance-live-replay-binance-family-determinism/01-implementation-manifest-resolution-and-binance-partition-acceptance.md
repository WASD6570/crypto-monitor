# Implementation: Manifest Resolution And Binance Partition Acceptance

## Module Requirements

- Extend replay manifest planning and validation so the existing replay engine accepts the final Binance raw partition posture from the completed raw-append slice.
- Preserve manifest-provided storage state, location, entry counts, and continuity checksums without guessing tier paths or collapsing stream-family distinctions.
- Keep scope resolution deterministic for Binance Spot and USD-M families that mix shared venue partitions and dedicated stream-family partitions.

## Target Repo Areas

- `services/replay-engine`
- `services/replay-engine/manifest_lookup.go`
- `services/replay-engine/runtime.go`
- `services/replay-engine/manifest_lookup_test.go`
- `services/replay-engine/runtime_test.go`

## Key Decisions

- Reuse the existing manifest reader and `RawPartitionLookupScope` contract instead of introducing Binance-specific partition lookup APIs.
- Treat the raw append output from the previous child slice as canonical for Binance stream-family routing.
- Keep manifest builder sorting and validation shared across venues; add only the minimum shared replay changes needed for concrete Binance partition acceptance gaps.

## Data And Algorithm Notes

- Cover both partition shapes already settled upstream:
  - shared date-symbol-venue partitions for Spot `trades` and top-of-book
  - dedicated `streamFamily` partitions for Spot `order-book` and `feed-health`, plus USD-M `funding-rate`, `mark-index`, `open-interest`, and `liquidation`
- Verify manifest build and validation across replay scopes that request:
  - one Binance family
  - multiple Binance families for the same symbol
  - mixed shared and dedicated Binance partitions in the same time window
- Ensure resolved partition drift checks still catch mismatched checksum, location, entry count, or storage state for Binance records.

## Unit Test Expectations

- manifest build resolves Binance raw partitions in stable logical-partition order
- resolver accepts Binance partitions without fabricating tier-specific paths
- validation catches Binance manifest drift just as it does for existing replay tests
- mixed Spot and USD-M stream-family scopes remain deterministic when replay requests specify explicit family lists

## Contract / Fixture / Replay Notes

- Do not change `ingestion.RawPartitionManifestRecord` or `contracts.ReplayPartitionRef` unless a concrete Binance replay blocker is found.
- If helper extraction is needed for Binance partition fixtures, keep it local to replay tests unless another consumer clearly shares it.

## Summary

This module makes replay manifest resolution accept the settled Binance partition posture as-is, so later replay execution can focus on deterministic ordering rather than storage lookup ambiguity.
