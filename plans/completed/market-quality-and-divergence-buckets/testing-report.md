# Testing Report: Market Quality And Divergence Buckets

## Outcome

- Passed deterministic 30s assignment and 2m/5m rollup coverage, including missing-slot propagation.
- Passed divergence, fragmentation, coverage, timestamp-trust, and market-quality summary coverage without adding regime classification.
- Passed integration and replay checks for deterministic bucket emission and config-version provenance.

## Commands

1. `/usr/local/go/bin/go test ./libs/go/... -run 'Test(BucketAssignment|BucketAssignmentRecvTsFallback|BucketRollupFrom30sClosures|DivergenceMetrics|FragmentationSeverity|CoverageAndTimestampTrustSummaries|MarketQualityCaps|UnavailableCompositeSummary|Missing30sBucketPropagation)'`
   - Result: passed
2. `/usr/local/go/bin/go test ./services/feature-engine/... -run 'Test(BucketAssignment|Missing30sBucketPropagation|LateEventHandling|DivergenceMetrics|FragmentationOutputs|MarketQualitySummary|TimestampTrustPropagation)'`
   - Result: passed
3. `/usr/local/go/bin/go test ./tests/integration/... -run 'TestWorldUSABucketReplayWindow|TestWorldUSABucketSummarySeams'`
   - Result: passed
4. `/usr/local/go/bin/go test ./tests/replay/... -run 'TestWorldUSABucketDeterminism|TestWorldUSABucketConfigVersionPinning'`
   - Result: passed

## Notes

- Rollups are derived only from closed 30s buckets and fill skipped intervals with explicit missing placeholders.
- After-watermark events stay visible through disposition/late markers and do not rewrite already-emitted live outputs.
- `market-quality.v1` stays config-versioned across `configs/local`, `configs/dev`, and `configs/prod`, while replay pinning is exposed on emitted bucket windows.
