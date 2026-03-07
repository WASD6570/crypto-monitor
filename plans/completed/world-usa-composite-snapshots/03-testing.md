# Testing

## Testing Goals

- prove contributor eligibility, weighting, and clamping are deterministic for the same inputs and config snapshot
- prove degraded feed, timestamp-fallback, and quote-normalization failures reduce trust or exclude contributors instead of preserving synthetic confidence
- prove WORLD and USA snapshots emit stable provenance, degraded reasons, and unavailable states for downstream service consumers

## Output Artifact

- Write the implementation-phase test report to `plans/completed/world-usa-composite-snapshots/testing-report.md`.

## Test Matrix

### 1. Contributor Eligibility And Quote Gating

- Purpose: verify group membership, timestamp trust, feed-health gates, and quote-proxy allow/deny behavior.
- Fixtures:
  - healthy WORLD and USA contributors for `BTC-USD`
  - one WORLD contributor with timestamp fallback to `recvTs`
  - one WORLD contributor with disallowed or degraded `USDT`/`USDC` proxy
  - one USA contributor with stale or gap-flagged health
- Validation commands:
  - `go test ./libs/go/... -run 'Test(ContributorEligibility|StablecoinNormalization)'`
  - `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeConstruction|CompositeDegradedVenueHandling)'`
- Verify:
  - eligibility status is explicit and deterministic
  - quote-proxy denial excludes the contributor with a recorded reason
  - timestamp fallback carries degraded state forward

### 2. Weighting And Clamping

- Purpose: verify raw score calculation, penalty composition, normalization, clamp order, and post-clamp renormalization.
- Fixtures:
  - balanced healthy contributors
  - one high-liquidity contributor that would dominate without clamping
  - mixed healthy and degraded contributors
- Validation commands:
  - `go test ./libs/go/... -run 'Test(CompositeWeighting|CompositeClamping)'`
  - `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeConstruction|CompositeAllContributorsExcluded)'`
- Verify:
  - weight sums are deterministic and normalized
  - clamping caps oversized contributors when healthy peers exist
  - all-contributors-excluded produces unavailable state, not fabricated values

### 3. Snapshot Output And Degradation Seams

- Purpose: verify emitted snapshot fields are complete enough for later service-side consumers without re-deriving contributor logic.
- Fixtures:
  - healthy snapshot
  - degraded snapshot with excluded contributors and timestamp fallback
  - unavailable snapshot with zero valid contributors
- Validation commands:
  - `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeSnapshotShape|CompositeUnavailableState)'`
  - `go test ./tests/integration/... -run 'TestWorldUSACompositeSnapshotSeams'`
- Verify:
  - snapshot contains contributor provenance, coverage counts, version fields, and degrade reason codes
  - degraded and unavailable states are distinct
  - downstream seam fields are present without introducing bucket, regime, or query-contract logic

### 4. Replay Determinism

- Purpose: prove the same raw fixture window and config snapshot reproduces identical contributor decisions and snapshot outputs.
- Fixtures:
  - pinned BTC single-day window
  - pinned ETH single-day window
  - mixed degraded and quote-confidence-loss window
- Validation commands:
  - `go test ./tests/replay/... -run 'TestWorldUSACompositeDeterminism'`
  - `go test ./tests/replay/... -run 'TestWorldUSAReplayTimestampFallback'`
- Verify:
  - repeated runs match contributor sets, pre/post-clamp weights, and snapshot flags exactly
  - timestamp-fallback scenarios stay stable under replay
  - config-version changes only alter outputs when the pinned config changes

## Required Negative Cases

- no healthy WORLD contributors
- no healthy USA contributors
- mixed timestamp trust across contributors
- quote proxy enabled in one config version and disabled in another
- one venue missing while another is sharply moving
- equal-score tie case that relies on stable contributor ordering

## Exit Criteria For Implementation

- targeted Go unit, integration, and replay commands pass
- degraded and unavailable snapshot paths are covered by deterministic fixtures
- repeated replay runs for the same input/config window produce identical snapshot outputs
- the resulting test report is written to `plans/completed/world-usa-composite-snapshots/testing-report.md`
