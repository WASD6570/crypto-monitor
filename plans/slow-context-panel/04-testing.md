# Slow Context Panel Testing

## Goal

Verify that slow CME and ETF context can be ingested, normalized, and rendered as advisory dashboard context without becoming a hidden dependency for realtime market-state surfaces.

Expected report output path after implementation: `plans/slow-context-panel/testing-report.md`

## Test Matrix

### 1. Slow-Source Ingestion Fixtures

- Purpose: prove adapters and normalization accept representative CME and ETF payloads and classify publish states correctly.
- Inputs:
  - deterministic fixture for newly published CME volume/OI point
  - deterministic fixture for newly published ETF daily flow point
  - deterministic fixture for repeated polling with no new publication
  - deterministic fixture for corrected publication on the same as-of date
- Validation commands to wire during implementation:
  - `go test ./services/... -run TestSlowContextAdapterParsesPublishedFixtures`
  - `go test ./services/... -run TestSlowContextPollingIsIdempotent`
  - `go test ./services/... -run TestSlowContextCorrectionHandling`
- Verify:
  - normalized records preserve source and ingest timing
  - duplicate polling does not create duplicate accepted records
  - correction path is explicit and auditable

### 2. Freshness Classification Under Pinned Clock

- Purpose: prove `fresh`, `delayed`, `stale`, and `unavailable` states classify deterministically.
- Inputs:
  - pinned clock fixtures around expected publish windows and threshold boundaries
- Validation commands to wire during implementation:
  - `go test ./services/... -run TestSlowContextFreshnessClassification`
  - `go test ./services/... -run TestSlowContextUnavailableState`
- Verify:
  - CME thresholds and ETF thresholds are applied by source family
  - boundary transitions are stable under repeated runs
  - unavailable state is explicit when no trusted record exists

### 3. Query Isolation And Non-Blocking Current State

- Purpose: prove slow-context lookup failure does not break core market-state delivery.
- Inputs:
  - mocked slow-context store outage or query error
  - valid market-state response fixtures from prior visibility slices
- Validation commands to wire during implementation:
  - `go test ./services/... -run TestCurrentStateSucceedsWhenSlowContextFails`
  - `go test ./services/... -run TestSlowContextResponseExplicitlyUnavailable`
- Verify:
  - current-state endpoint still returns BTC and ETH market-state payloads
  - slow-context block becomes unavailable or error-scoped only
  - no field omission causes client ambiguity

### 4. Web Rendering And Operator Messaging

- Purpose: prove the dashboard shows slow context clearly without changing symbol-state semantics.
- Inputs:
  - mocked API responses for fresh, delayed, stale, unavailable, and partial slow-context payloads
- Validation commands to wire during implementation:
  - `pnpm --filter web test -- --run SlowContextPanel`
  - `pnpm --filter web test -- --run DashboardCurrentStateWithSlowContextFallback`
  - `pnpm --filter web exec vite build`
- Verify:
  - `Context only` badge is always present
  - as-of timestamp and cadence labels remain visible on desktop and mobile
  - stale, delayed, and unavailable messaging is isolated to the panel
  - summary market-state cards do not change when slow context changes status

## Required Negative Cases

- stale CME context with otherwise healthy realtime market state
- missing ETF context on first load
- delayed publication where the latest known value is inside tolerated age but beyond expected publish window
- partial availability where CME is present and ETF is unavailable
- slow-context query error while dashboard current-state query succeeds
- repeated polling of unchanged source payloads

## Replay And Determinism Notes

- If implementation stores slow context in replay-visible paths, add a pinned-clock replay smoke such as:
  - `go test ./services/... -run TestSlowContextReplayDeterminism`
- If any normalization helper is mirrored in Python for offline research, add parity coverage such as:
  - `go test ./tests/parity/... -run TestSlowContextParityFixtures`
- These are conditional follow-ups, not permission to put Python in the live runtime path.

## Exit Criteria

- Ingestion is idempotent for repeated slow-source polling.
- Freshness classification is deterministic under a pinned clock.
- Dashboard current-state delivery stays successful when slow context is stale, missing, delayed, or unavailable.
- Operator messaging makes the advisory-only semantics unmistakable.
- Generated testing notes and results are written to `plans/slow-context-panel/testing-report.md` during implementation validation.
