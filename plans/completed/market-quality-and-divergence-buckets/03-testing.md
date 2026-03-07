# Testing

## Testing Goals

- prove deterministic bucket assignment and rollup behavior for 30s, 2m, and 5m families
- prove divergence, fragmentation, coverage, timestamp-trust, and market-quality summaries are stable for identical inputs and config
- prove late events after watermark do not silently mutate already-emitted live outputs
- prove downstream seams are sufficient for later regime/query work without embedding classification logic here

## Output Artifact

- Write the implementation-phase test report to `plans/completed/market-quality-and-divergence-buckets/testing-report.md`.

## Test Matrix

### 1. Bucket Assignment And Window Closure

- Purpose: verify UTC assignment, timestamp-source choice, and canonical 30s closure behavior.
- Fixtures:
  - clean `exchangeTs`-ordered BTC window
  - mixed `exchangeTs` and `recvTs` fallback ETH window
  - sparse interval with one missing 30s slot
- Validation commands:
  - `go test ./libs/go/... -run 'Test(BucketAssignment|BucketAssignmentRecvTsFallback|BucketRollupFrom30sClosures)'`
  - `go test ./services/feature-engine/... -run 'Test(BucketAssignment|Missing30sBucketPropagation)'`
- Verify:
  - chosen bucket source is explicit and deterministic
  - 2m and 5m windows are derived from closed 30s buckets only
  - incomplete windows remain visible with missing-count metadata

### 2. Late And Out-Of-Order Events

- Purpose: verify within-watermark handling and after-watermark replay handoff behavior.
- Fixtures:
  - one event arriving within the 30s watermark
  - one event arriving after the 30s watermark
  - one event with implausible `exchangeTs` forcing `recvTs` fallback
- Validation commands:
  - `go test ./services/feature-engine/... -run 'TestLateEventHandling'`
  - `go test ./tests/integration/... -run 'TestWorldUSABucketReplayWindow'`
- Verify:
  - within-watermark events affect the still-open bucket deterministically
  - after-watermark events are marked late and do not rewrite emitted live outputs
  - fallback timestamp handling propagates degraded markers into the bucket summary

### 3. Divergence And Fragmentation Summaries

- Purpose: verify aligned, mixed, and fragmented market states are reflected without final regime classes.
- Fixtures:
  - aligned WORLD and USA move across a full 5m window
  - sustained directional disagreement over several 30s buckets
  - one-side unavailable composite case
- Validation commands:
  - `go test ./libs/go/... -run 'Test(DivergenceMetrics|FragmentationSeverity|UnavailableCompositeSummary)'`
  - `go test ./services/feature-engine/... -run 'Test(DivergenceMetrics|FragmentationOutputs)'`
- Verify:
  - divergence stays multi-part and explainable
  - fragmentation severity and reason codes are deterministic
  - unavailable-side behavior emits explicit markers instead of synthetic values

### 4. Coverage Timestamp Trust And Market Quality

- Purpose: verify trust caps respond conservatively to degraded coverage and timestamp issues.
- Fixtures:
  - healthy coverage on both sides
  - USA missing one configured peer
  - WORLD fallback-heavy timestamps with otherwise aligned prices
  - concentration-heavy composite with degraded health score
- Validation commands:
  - `go test ./libs/go/... -run 'Test(CoverageAndTimestampTrustSummaries|MarketQualityCaps)'`
  - `go test ./services/feature-engine/... -run 'Test(MarketQualitySummary|TimestampTrustPropagation)'`
- Verify:
  - coverage asymmetry is surfaced explicitly
  - timestamp degradation can cap quality even when divergence is small
  - 5m market-quality outputs stop short of regime classification

### 5. Replay Determinism And Config Pinning

- Purpose: prove the bucket layer is replay-safe and version-auditable.
- Fixtures:
  - pinned single-day BTC window
  - pinned single-day ETH window
  - mixed degraded and late-event fixture window
- Validation commands:
  - `go test ./tests/replay/... -run 'TestWorldUSABucketDeterminism'`
  - `go test ./tests/replay/... -run 'TestWorldUSABucketConfigVersionPinning'`
  - `go test ./tests/integration/... -run 'TestWorldUSABucketSummarySeams'`
- Verify:
  - repeated runs with identical inputs and config emit identical bucket outputs
  - changing config version intentionally changes outputs while preserving explainable version fields
  - replay preserves late-event and timestamp-trust markers in the emitted summaries

## Required Negative Cases

- WORLD available while USA unavailable
- USA available while WORLD unavailable
- both sides available but one side fallback-heavy on timestamps
- fragmented prices with clean coverage
- clean prices with degraded coverage and concentration
- missing 30s slot inside an otherwise valid 5m rollup

## Determinism Notes

- Go remains the live source of truth.
- Use pinned fixtures only; no live venue calls, wall-clock dependence, or nondeterministic iteration in validation paths.
- Replay assertions should compare full bucket outputs, including reason codes and version metadata.

## Exit Criteria For Implementation

- targeted unit, integration, and replay commands pass
- bucket outputs are deterministic across repeated identical runs
- degraded and late-event cases remain visible and conservative
- later regime/query slices can read bucket summaries without adding hidden recomputation
