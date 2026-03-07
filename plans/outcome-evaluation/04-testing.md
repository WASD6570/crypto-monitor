# Outcome Evaluation Testing

## Testing Goal

Prove that outcome evaluation is deterministic, horizon-aware, conservative under ambiguity, and directly usable by the Initiative 2 review loop.

## Output Artifact

- Write the implementation test report to `plans/outcome-evaluation/testing-report.md`.

## Required Validation Commands

Implementation should make the following commands runnable:

- `go test ./services/outcome-engine/...`
- `go test ./tests/integration/... -run OutcomeEvaluation`
- `go test ./tests/replay/... -run OutcomeEvaluationReplay`
- `go test ./tests/parity/... -run OutcomeEvaluationParity`

If the implementation chooses different package paths, keep command names equally narrow and feature-specific.

## Smoke Matrix

### 1. Pure Outcome Closure

- Command: `go test ./services/outcome-engine/... -run TestOutcomeTargetBeforeInvalidation`
- Verify: target-first path closes the horizon as `TARGET_HIT`, records decisive timestamp, MAE, MFE, and `timeToHitMs`.

- Command: `go test ./services/outcome-engine/... -run TestOutcomeInvalidationBeforeTarget`
- Verify: invalidation-first path closes as `INVALIDATED` and does not populate `timeToHitMs`.

- Command: `go test ./services/outcome-engine/... -run TestOutcomeTimeout`
- Verify: timeout occurs only after horizon end, with non-zero excursion metrics preserved.

### 2. Conservative Ambiguity Handling

- Command: `go test ./services/outcome-engine/... -run TestOutcomeSameEventTieResolvesToInvalidation`
- Verify: same-event target and invalidation conflict resolves to `INVALIDATED`.

- Command: `go test ./services/outcome-engine/... -run TestOutcomeMissingThresholdsRejected`
- Verify: evaluation fails fast when target or invalidation is missing.

- Command: `go test ./services/outcome-engine/... -run TestOutcomeDataGapMarksUndecided`
- Verify: trusted data gaps produce `UNDECIDED_DATA_GAP` instead of optimistic timeout or target classification.

### 3. Horizon Independence

- Command: `go test ./services/outcome-engine/... -run TestOutcomeHorizonDisagreement`
- Verify: a single alert can produce different valid results across `30s`, `2m`, and `5m`.

- Command: `go test ./services/outcome-engine/... -run TestOutcomePerHorizonTimeMetrics`
- Verify: each horizon stores the correct close time and duration fields without leaking longer-horizon state backward.

### 4. Net Viability Boundary

- Command: `go test ./services/outcome-engine/... -run TestNetViabilityPositiveAfterCosts`
- Verify: coarse cost deductions still leave a `POSITIVE` net viability result.

- Command: `go test ./services/outcome-engine/... -run TestNetViabilityTurnsNegative`
- Verify: gross favorable movement can become `NEGATIVE` after pinned costs.

- Command: `go test ./services/outcome-engine/... -run TestNetViabilityUnknownWhenCoverageMissing`
- Verify: missing delayed observations or trusted coverage lead to `UNKNOWN`, not forced positive/negative values.

### 5. Replay Determinism And Version Pinning

- Command: `go test ./tests/replay/... -run TestOutcomeReplayDeterminism`
- Verify: same raw inputs, alert payload, config snapshot, and code version produce byte-equivalent outcome artifacts.

- Command: `go test ./tests/replay/... -run TestOutcomeReplayUsesPinnedConfigVersion`
- Verify: replay uses the preserved evaluation and cost-model versions, not current defaults.

- Command: `go test ./tests/replay/... -run TestOutcomeReplayAppendsSupersedingRecord`
- Verify: corrected replay appends a new artifact with supersession linkage rather than mutating the original record.

### 6. Consumer Contract And Review Loop Coverage

- Command: `go test ./tests/integration/... -run TestOutcomeContractForReviewQuery`
- Verify: the review consumer can fetch the latest non-superseded record by `alertId` and read all required horizon fields.

- Command: `go test ./tests/integration/... -run TestOutcomeBaselineComparability`
- Verify: production and baseline alerts are sliceable on the same outcome dimensions.

- Command: `go test ./tests/integration/... -run TestOutcomeUnknownViabilityRemainsQueryable`
- Verify: unknown net-viability records remain valid contract payloads and do not disappear from aggregates.

## Negative Cases That Must Exist

- missing target
- missing invalidation
- malformed direction vs threshold combination
- same-event target/invalidation tie
- late event arriving after live watermark but included correctly during replay
- degraded timestamp fallback changing bucket source while remaining deterministic
- config-version mismatch between alert and evaluation request
- baseline record missing baseline identifier
- superseded record returned as latest by mistake

## Acceptance Checklist

- Outcome closure is deterministic for the same inputs.
- Timeout never wins before target/invalidation checks finish.
- Horizon results remain independent.
- MAE, MFE, and time metrics match fixtures.
- Regime attribution survives replay and query serialization.
- Net viability stays conservative and does not impersonate execution.
- Replay corrections append new artifacts with provenance.
- Baseline and production outputs remain directly comparable.
