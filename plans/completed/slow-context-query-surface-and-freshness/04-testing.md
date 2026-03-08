# Slow Context Query Surface And Freshness Testing

## Goal

Verify that slow CME and ETF context can be stored, classified, and queried with explicit freshness and unavailable semantics while leaving the core current-state response successful when slow-context lookup fails.

Expected report output path while this feature is active: `plans/slow-context-query-surface-and-freshness/testing-report.md`

## Validation Commands

```bash
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextFreshnessClassification
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextUnavailableState
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextLatestRevisionSelection
/usr/local/go/bin/go test ./services/feature-engine/... -run TestCurrentStateSucceedsWhenSlowContextFails
/usr/local/go/bin/go test ./services/feature-engine/... -run TestSlowContextResponseExplicitlyUnavailable
```

## Smoke Matrix

### 1. Freshness Classification Under Pinned Clock

- Purpose: prove CME and ETF slow context classify as `fresh`, `delayed`, `stale`, and `unavailable` deterministically.
- Inputs:
  - pinned-clock CME fixtures around the expected publish window and 36-hour boundary
  - pinned-clock ETF fixtures around the expected publish window and 48-hour boundary
- Verify:
  - threshold transitions are exact and repeatable
  - freshness metadata includes the classification basis

### 2. Explicit Unavailable Behavior

- Purpose: prove missing or failed slow-context lookup returns an explicit unavailable block.
- Inputs:
  - no trusted accepted record
  - store/query failure
- Verify:
  - slow-context response uses explicit unavailable fields
  - no silent omission creates client ambiguity

### 3. Correction-Aware Latest Selection

- Purpose: prove same-as-of revisions stay auditable and deterministic.
- Inputs:
  - baseline accepted record
  - corrected same-as-of record with newer revision
- Verify:
  - latest query returns the corrected revision
  - lineage/revision metadata remains visible

### 4. Current-State Isolation

- Purpose: prove slow-context failure does not break BTC/ETH market-state delivery.
- Inputs:
  - healthy current-state fixture data from prior visibility slices
  - unavailable or failing slow-context query seam
- Verify:
  - core current-state response still succeeds
  - slow-context block becomes unavailable or error-scoped only
  - symbol/global regime, composite, and bucket sections remain unchanged

## Required Negative Cases

- CME context delayed but not yet stale
- CME context stale beyond 36 hours
- ETF context stale beyond 48 hours
- no trusted slow-context record exists
- same-as-of corrected record supersedes an earlier accepted revision
- slow-context store/query failure while current-state market data remains healthy

## Determinism And Replay Notes

- Use pinned clocks for all freshness-threshold tests.
- If this slice adds replay-visible persistence earlier than planned, add a focused replay determinism check and record that scope expansion in the testing report.
- Do not add Python parity coverage unless implementation actually mirrors slow-context logic outside Go.

## Exit Criteria

- Freshness classification is deterministic for CME and ETF thresholds.
- Unavailable behavior is explicit and client-safe.
- Latest revision selection is correction-aware and auditable.
- Current-state delivery still succeeds when slow-context lookup fails.
- Results are written to `plans/slow-context-query-surface-and-freshness/testing-report.md` during implementation.
