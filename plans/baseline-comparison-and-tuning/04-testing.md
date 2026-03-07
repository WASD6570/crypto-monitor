# Baseline Comparison And Tuning Testing

## Testing Goal

Prove that baseline comparison and tuning stay deterministic, replay-backed, version-pinned, and operationally safe without introducing live AI optimization.

## Output Artifact

- Write the implementation test report to `plans/baseline-comparison-and-tuning/testing-report.md`.

## Required Validation Commands

Implementation should make the following commands runnable:

- `go test ./services/alert-engine/... -run 'Test(BaselineRegistry|BaselineComparability|BaselineEpisodeConstruction)'`
- `go test ./services/replay-engine/... -run 'Test(TuningCandidateReplayBundle|TuningPromotionDecision|TuningRollbackDecision)'`
- `go test ./tests/integration/... -run 'Test(BaselineComparisonWorkflow|ConfigSnapshotActivation|TuningReportContracts)'`
- `go test ./tests/replay/... -run 'Test(BaselineComparisonReplayDeterminism|ConfigVersionPinnedReplay|RollingWindowRollbackChecks)'`
- `go test ./tests/parity/... -run 'TestOfflineBaselineParity'`

If implementation chooses different package paths, keep the commands equally narrow and feature-specific.

## Smoke Matrix

### 1. Baseline Definitions And Comparable Episodes

- Command: `go test ./services/alert-engine/... -run TestNaiveBreakoutBaseline`
- Verify: `naive-breakout` emits deterministic control alerts without 2m validation or 5m market-state gating.

- Command: `go test ./services/alert-engine/... -run TestNaiveVWAPReversionBaseline`
- Verify: `naive-vwap-reversion` emits deterministic VWAP-distance control alerts with versioned parameters.

- Command: `go test ./services/alert-engine/... -run TestSingleVenueTriggerBaseline`
- Verify: `single-venue-trigger` emits one-venue control alerts without composite confirmation.

- Command: `go test ./services/alert-engine/... -run TestComparableEpisodeConstruction`
- Verify: production-only, baseline-only, and shared episodes are all preserved in deterministic join output.

### 2. Outcome And Simulation Join Compatibility

- Command: `go test ./tests/integration/... -run TestBaselineAlertsShareOutcomeContract`
- Verify: baseline alerts and production alerts pass through the same outcome-evaluation contract.

- Command: `go test ./tests/integration/... -run TestBaselineSimulationComparability`
- Verify: when simulation coverage exists, baseline and production records remain comparable on net viability outputs.

- Command: `go test ./tests/integration/... -run TestComparableJoinKeysByRegimeAndHorizon`
- Verify: reports can slice by symbol, setup family, baseline family, regime, and horizon without client recomputation.

### 3. Config Snapshot Lifecycle And Promotion Gates

- Command: `go test ./services/replay-engine/... -run TestConfigSnapshotManifestValidation`
- Verify: candidate snapshots require parent version, report references, replay windows, and environment scope.

- Command: `go test ./services/replay-engine/... -run TestCandidatePromotionRequiresPinnedAndRollingWindows`
- Verify: promotion fails when pinned fixtures or the 14-day rolling window are missing.

- Command: `go test ./services/replay-engine/... -run TestCandidatePromotionThresholds`
- Verify: promotion succeeds only when precision and net viability are preserved or improved and fragmented-market false positives do not worsen.

- Command: `go test ./tests/integration/... -run TestSingleHumanApprovalWithPassedAutomation`
- Verify: one human approver can activate a snapshot only after automated evidence gates pass.

### 4. Reporting Outputs And Recommendation Rules

- Command: `go test ./services/replay-engine/... -run TestTuningReportIncludesRequiredBaselineSections`
- Verify: report bundles include separate sections for `naive-breakout`, `naive-vwap-reversion`, and `single-venue-trigger`.

- Command: `go test ./services/replay-engine/... -run TestTuningReportCarriesCountsAndPercents`
- Verify: aggregate deltas include both percentages and raw counts.

- Command: `go test ./services/replay-engine/... -run TestPromotionRecommendationClassification`
- Verify: report recommendation becomes `PROMOTE`, `HOLD`, or `REJECT` from deterministic thresholds only.

- Command: `go test ./tests/integration/... -run TestReportArtifactsReferencedFromSnapshotManifest`
- Verify: machine-readable and human-readable report artifacts are pinned from the candidate snapshot.

### 5. Rollback Behavior

- Command: `go test ./services/replay-engine/... -run TestRollbackWhenPrecisionDropsBeyondThreshold`
- Verify: active snapshot rolls back when rolling-window precision drops more than 3 absolute points.

- Command: `go test ./services/replay-engine/... -run TestRollbackWhenFragmentedFalsePositivesWorsen`
- Verify: rollback occurs when fragmented-market false positives exceed the allowed delta.

- Command: `go test ./tests/integration/... -run TestRollbackRestoresPriorKnownGoodSnapshot`
- Verify: rollback reactivates the preserved prior snapshot instead of hot-editing the failing one.

- Command: `go test ./tests/replay/... -run TestRollbackAuditTrailIsAppendOnly`
- Verify: rollback appends a new state transition and preserves the failed snapshot history.

### 6. Replay Determinism And Research Boundary

- Command: `go test ./tests/replay/... -run TestBaselineComparisonReplayDeterminism`
- Verify: identical replay inputs and snapshots regenerate byte-equivalent comparable episodes and reports.

- Command: `go test ./tests/replay/... -run TestReplayUsesPinnedConfigSnapshot`
- Verify: replay uses stored snapshots and not current defaults.

- Command: `go test ./tests/parity/... -run TestOfflineBaselineParity`
- Verify: optional offline Python helpers, if present, match Go-produced baseline fixtures without becoming a runtime dependency.

- Command: `go test ./tests/integration/... -run TestNoLiveAIAuthorityInPromotionPath`
- Verify: promotion path rejects unversioned AI or notebook output as a source of active config changes.

## Negative Cases That Must Exist

- missing `baselineId` on a baseline alert
- `naive-breakout` accidentally using 2m validator or 5m market-state gate
- `single-venue-trigger` accidentally reading composite confirmation fields
- production and baseline records sharing no comparable episode because bucket keys drift
- candidate snapshot missing parent active version or rollback pointer
- candidate report missing one required baseline section
- candidate improving precision only by suppressing nearly all alerts
- rolling window shorter than the declared default without explicit report annotation
- promotion attempted with failed automated replay determinism checks
- approval attempted by an unaudited identity or with no attached report artifact
- rollback attempted with no prior known-good snapshot reference
- offline Python suggestion attempting to activate config without a versioned snapshot

## Acceptance Checklist

- Required baselines are explicit, deterministic, and replayable.
- Comparable joins preserve both production-only and baseline-only misses.
- Candidate promotion is evidence-based and version-pinned.
- Reporting outputs include separate baseline deltas, counts, and recommendation state.
- Rollback is mechanical, append-only, and auditable.
- Replay stays deterministic for the same windows and snapshots.
- Python parity remains optional and offline-only.
- Live config authority stays separate from AI optimization.
