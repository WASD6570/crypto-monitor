# Testing

## Testing Goals

- prove WORLD and USA composites are deterministic for the same inputs and config snapshot
- prove degraded feeds, timestamp fallback, and quote-normalization failures reduce trust instead of silently preserving `TRADEABLE`
- prove fragmented regimes are visible and force conservative 5m state transitions
- prove consumer contracts expose enough information for UI and replay/audit consumers without recomputation

## Output Artifact

- Write the implementation-phase test report to `plans/world-usa-composites-and-market-state/testing-report.md`.

## Test Matrix

### 1. Composite Construction And Weighting

- Purpose: verify venue membership, weighting, clamping, and stablecoin normalization behavior.
- Fixtures:
  - healthy WORLD and USA contributors for BTC and ETH
  - one WORLD venue stale
  - one WORLD quote proxy disallowed or degraded
  - one USA venue gap-flagged
- Validation commands:
  - `go test ./libs/go/... -run 'Test(CompositeWeighting|StablecoinNormalization|CompositeClamping)'`
  - `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeConstruction|CompositeDegradedVenueHandling)'`
- Verify:
  - contributors, weights, and clamp outputs are deterministic
  - excluded venues carry explicit reasons
  - all-contributors-excluded produces unavailable composite state, not synthetic values

### 2. Feature Buckets And Divergence

- Purpose: verify 30s, 2m, and 5m bucket assignment and divergence metrics.
- Fixtures:
  - aligned WORLD and USA move
  - fragmented directional disagreement over several 30s buckets
  - late events inside and outside watermark thresholds
- Validation commands:
  - `go test ./services/feature-engine/... -run 'Test(BucketAssignment|DivergenceMetrics|LateEventHandling)'`
  - `go test ./tests/integration/... -run 'TestWorldUSABucketReplayWindow'`
- Verify:
  - bucket assignment uses `exchangeTs` first and `recvTs` fallback with degraded markers
  - late events beyond watermark do not silently rewrite already-emitted live state
  - divergence outputs stay stable across repeated runs

### 3. Regime Classification

- Purpose: verify `TRADEABLE/WATCH/NO-OPERATE` classification and ceiling rules.
- Fixtures:
  - clean aligned market
  - moderate fragmentation with partial coverage loss
  - severe fragmentation with degraded feed health
  - one-symbol healthy and one-symbol degraded case
- Validation commands:
  - `go test ./services/regime-engine/... -run 'Test(RegimeClassification|FragmentedMarketDowngrade|GlobalCeilingRules)'`
  - `go test ./tests/integration/... -run 'TestWorldUSAMarketStateTransitions'`
- Verify:
  - trust degrades quickly on critical health loss
  - recovery requires stable evidence over multiple buckets
  - global state caps symbol state per operating defaults

### 4. Consumer Contracts And Query Surfaces

- Purpose: verify read models are complete and versioned.
- Fixtures:
  - healthy current-state payload
  - degraded payload with excluded venues and timestamp fallback
  - replay-corrected historical payload
- Validation commands:
  - `go test ./services/feature-engine/... -run 'Test(MarketStateQueryResponse|CompositeSnapshotSchema)'`
  - `go test ./services/regime-engine/... -run 'Test(MarketRegimeSnapshotSchema|HistoricalStateVersionContext)'`
- Verify:
  - response contains provenance, degraded reasons, and version fields
  - UI can render current state without recalculating market logic
  - historical reads preserve original config and algorithm version references

### 5. Replay And Determinism

- Purpose: prove the feature is replay-safe and audit-friendly.
- Fixtures:
  - pinned single-day BTC fixture window
  - pinned single-day ETH fixture window
  - mixed degraded and late-event window
- Validation commands:
  - `go test ./tests/replay/... -run 'TestWorldUSAReplayDeterminism'`
  - `go test ./tests/replay/... -run 'TestWorldUSALateEventReplayCorrection'`
  - `go test ./tests/integration/... -run 'TestWorldUSAConfigVersionPinnedReplay'`
- Verify:
  - repeated runs with identical inputs and config emit identical composites, features, and regimes
  - changing config version intentionally changes outputs while preserving auditability
  - replay outputs include counts and reasons for degraded contributors and regime transitions

## Required Negative Cases

- no healthy WORLD contributors
- no healthy USA contributors
- mixed timestamp trust across venues
- quote proxy enabled in one config version and disabled in another
- regime threshold edge causing possible flapping
- partial history query where one bucket is unavailable

## Determinism And Parity Notes

- Go is the live source of truth for all tests.
- Optional parity checks may be added under `tests/parity`, but they must consume pinned fixtures and compare against Go-produced expected outputs.
- Do not rely on nondeterministic time, randomized maps, or live venue calls in any validation path.

## Exit Criteria For Implementation

- targeted unit, integration, and replay commands pass
- degraded and fragmented negative cases are covered by fixtures
- repeated deterministic replay runs match exactly for the same config snapshot
- consumer contracts expose enough provenance for dashboard and alert consumers without client recomputation
