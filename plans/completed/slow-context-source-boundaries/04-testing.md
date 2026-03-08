# Slow Context Source Boundaries Testing

## Goal

Verify that slow CME and ETF source families can be polled and classified in Go with explicit publication states, isolated source health, and idempotent repeated polling before any storage/query or dashboard work is added.

Expected report output path in the archived feature directory: `plans/completed/slow-context-source-boundaries/testing-report.md`

## Validation Commands

```bash
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextAdapterParsesPublishedFixtures
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextRepeatedPollingIsIdempotent
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextDelayedPublicationClassification
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextCorrectionHandling
/usr/local/go/bin/go test ./services/slow-context/... -run TestSlowContextSourceFailuresStayIsolated
```

## Smoke Matrix

### 1. Published Fixture Parsing

- Purpose: prove CME and ETF fixtures parse into classified new-publication results.
- Inputs:
  - deterministic CME published fixture
  - deterministic ETF published fixture
- Verify:
  - source family and metric family are preserved
  - `asOfTs`, `publishedTs`, and `ingestTs` remain distinct
  - stable dedupe identity is emitted

### 2. Repeated Polling Idempotency

- Purpose: prove unchanged repeated polls are safe to retry.
- Inputs:
  - same-as-of repeated CME or ETF fixture
- Verify:
  - repeated polls classify as `published_same_value_or_same_asof`
  - no duplicate accepted output is emitted
  - source health does not drift toward a false fresh/new-publication state

### 3. Delayed Publication Classification

- Purpose: prove expected publish windows and delayed-publication state are deterministic.
- Inputs:
  - pinned-clock test cases before and after the expected publish window
  - `not_yet_published` fixture or mock response
- Verify:
  - inside-window results remain non-failing
  - crossing the expected window produces `delayed_publication`
  - the same fixture under the same clock produces the same classification repeatedly

### 4. Correction Handling

- Purpose: prove same-as-of republishes remain explicit and auditable.
- Inputs:
  - corrected publication fixture for an already-seen as-of key
- Verify:
  - correction metadata or revision marker is explicit
  - the original accepted publication identity is not silently erased

### 5. Source Failure Isolation

- Purpose: prove slow-source failures do not poison realtime feed-health semantics.
- Inputs:
  - transient fetch failure
  - parse failure fixture
- Verify:
  - slow-source health becomes `source_unavailable` or `parse_failed`
  - no realtime venue-health status is changed by this feature alone
  - retries remain possible without destructive cleanup

## Required Negative Cases

- no publication yet inside the expected window
- delayed publication after the expected window
- repeated unchanged polling of the same as-of value
- corrected publication for the same as-of key
- transient source outage
- parse failure on malformed source payload

## Determinism And Replay Notes

- Use pinned clocks for all publish-window tests.
- If implementation stores accepted source-boundary outputs in replay-visible paths earlier than planned, add a targeted replay determinism check and record that scope expansion explicitly in the testing report.
- Do not add Python parity checks for this feature unless a later offline helper actually mirrors Go classification logic.

## Exit Criteria

- CME and ETF adapters parse representative source fixtures successfully.
- Repeated polling is idempotent.
- Delayed-publication behavior is deterministic under a pinned clock.
- Correction handling is explicit and auditable.
- Slow-source failures stay isolated from realtime feed-health semantics.
- Results are written to `plans/completed/slow-context-source-boundaries/testing-report.md` during implementation.
