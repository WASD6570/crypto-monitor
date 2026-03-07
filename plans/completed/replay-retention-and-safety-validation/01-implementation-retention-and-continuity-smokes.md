# Implementation Module 1: Retention And Continuity Smokes

## Requirements And Scope

- Prove integrated replay works across hot and cold raw retention tiers without changing logical scope resolution, canonical event identity, or deterministic ordering inputs.
- Keep the module bounded to validation harnesses, deterministic fixtures, and minimal Go hooks needed to expose continuity evidence.
- Reuse manifest continuity, snapshot preservation, and replay ordering behavior from completed prerequisite slices rather than redefining them.
- Cover late-event and degraded timestamp fixtures only where they materially prove continuity through retention boundaries.

## Target Repo Areas

- `services/replay-engine`
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`
- `docs/runbooks`

## Implementation Notes

### Harness Responsibilities

- Add a deterministic smoke harness that can load a scoped symbol/day fixture set in a hot-layout simulation, a cold-layout simulation, and a staged-restore path when the harness needs to mimic cold recovery.
- Keep retention movement modeled as logical manifest/reference transitions, not vendor storage operations.
- Capture evidence that the resolved partition set, manifest digests, ordered event IDs, late/degraded counters, and replay result digests remain stable across tiers.

### Fixture Shape

- Use one primary symbol/day fixture family with:
  - normal in-order events
  - a degraded timestamp event that falls back from implausible `exchangeTs` to `recvTs`
  - at least one late event that remains persisted and visible to replay
  - stable source identity fields for duplicate-safe ordering assertions
- Keep fixtures small enough for fast local replay while still forcing hot/cold continuity checks.

### Validation Hooks

- Expose replay evidence through Go-visible result summaries instead of log scraping.
- Prefer result digests, ordered ID lists, and manifest references that can be asserted directly from tests.
- Add restore-state markers only if needed to prove a cold read required explicit staging before replay execution.

### Deterministic Assertions

- Same symbol/day scope must produce the same ordered canonical event IDs before and after tier movement.
- Manifest continuity must preserve logical partition identity even if the underlying tier reference changes.
- Replay must continue to load preserved config, contract, and build references instead of mutable defaults.
- Late-event and degraded-event counters must remain stable across hot, cold, and restored reads.

## Unit And Integration Test Expectations

- `go test ./tests/integration -run 'TestReplayRetentionContinuityAcrossHotAndCold|TestReplayColdRestoreIsExplicitAndDeterministic'`
- `go test ./tests/replay/... -run 'TestReplayDeterminismAcrossRetentionTiers|TestReplayRetentionPreservesLateAndDegradedEvidence'`
- `go test ./services/replay-engine/... -run 'TestReplayScopeResolutionDoesNotGuessTierPaths|TestReplayRetentionUsesPreservedSnapshots'`

## Contract / Fixture / Replay Impacts

- No new core replay contract family should be introduced unless a tiny evidence field is required for deterministic smoke assertions.
- Fixture updates should stay under `tests/fixtures` and remain storage-neutral.
- Runbook notes may document how the smoke harness simulates hot, cold, and restore states without naming a concrete storage vendor.

## Summary

This module proves the raw retention seam is safe for replay consumers. The key result is deterministic, tier-agnostic replay evidence for one logical scope without reopening raw persistence or replay-ordering design.
