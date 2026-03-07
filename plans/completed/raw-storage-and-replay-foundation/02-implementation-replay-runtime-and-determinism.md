# Implementation Module 2: Replay Runtime And Determinism

## Scope

Define the replay runtime contract, exact metadata and config snapshots to preserve, deterministic ordering behavior, runtime modes, and expected outputs for audit and rebuild flows.

## Target Repo Areas

- `services/*` replay runtime boundary added during implementation if needed
- `libs/go`
- `schemas/json/replay`
- `configs/*`
- `tests/replay`
- `tests/fixtures`
- `tests/integration`
- optional offline parity helpers in `libs/python` or `tests/parity`

## Requirements

- Replay must be deterministic for the same scoped inputs and config.
- Replay must preserve the operating-default timestamp, lateness, and runtime expectations.
- Replay must run without live venue dependencies.
- Replay must support dry-run inspection and isolated rebuild modes before any publish/apply path.
- Replay must expose enough metadata for later agents to compare outputs across runs.
- Python parity, if added, remains optional and offline-only.

## Replay Inputs To Preserve Exactly

Every replay run must reference immutable or snapshotted inputs for the requested scope. The implementation should preserve and load at least these categories:

- raw event input set:
  - partition manifest or equivalent partition list
  - partition checksums or content digest where available
  - event count and duplicate count summary
- schema and contract context:
  - event contract version(s)
  - replay manifest contract version(s)
  - any downstream output contract version(s) produced by the run
- runtime configuration snapshot:
  - symbol and venue scope
  - stream families included
  - watermark and late-event thresholds for 30s, 2m, and 5m windows
  - timestamp plausibility/skew thresholds
  - deduplication and ordering rules
  - degradation-handling policy toggles
  - target output mode: inspect, rebuild, compare, or apply
- code/build provenance:
  - git SHA or release identifier
  - build timestamp or artifact identifier
  - enabled feature flags that affect deterministic behavior
- invocation metadata:
  - requester identity or system actor
  - reason code
  - run start timestamp
  - source of the replay request, such as operator action or automated backfill recovery

Replay must not resolve these values from mutable current defaults once the run starts.

## Ordering And Determinism Rules

- Use a single documented ordering algorithm for live-consistent replay:
  1. chosen bucket timestamp source (`exchangeTs` when plausible, otherwise degraded `recvTs`)
  2. stable venue sequence or source order field when present
  3. stable canonical event ID
- Process equal-key ties in a deterministic stable sort; never rely on map iteration or worker race timing.
- Keep bucket assignment in UTC with event-time-first semantics.
- Preserve the original timestamp choice and degraded reason with each replayed record.
- For the same inputs, replay must produce identical bucket assignment, identical late-event marking, and identical output manifests.

## Runtime Modes

- `inspect`: read raw inputs, compute replay manifest, and report expected corrections without writing downstream state
- `rebuild`: write scoped rebuilt artifacts into isolated storage or namespaced outputs
- `compare`: run rebuild logic and emit deterministic diffs against current materialized outputs
- `apply`: publish or promote rebuilt outputs only after explicit gate checks and side-effect protections pass

Default mode should be `inspect` or `rebuild`, not `apply`.

## Output Expectations

Plan for replay outputs that are easy to diff and audit:

- replay run manifest
- rebuilt derived artifacts for the scoped window
- counts for input events, duplicates, late events, degraded timestamps, and skipped records
- applied watermark policy and config snapshot references
- comparison summary when existing outputs are present
- audit status showing whether the run was dry-run only, isolated rebuild, or promoted apply

## Performance And Runtime Guardrails

- A single-symbol single-day replay should target completion within 10 minutes on local/dev infrastructure under normal conditions.
- If replay exceeds that budget, optimize partition resolution, streaming reads, and observability before widening scope.
- Replay should stream partitions incrementally rather than require a full-memory load for one day of data.
- Any concurrency used for throughput must preserve deterministic merge order.

## Failure And Recovery Rules

- Contract mismatch, missing partitions, checksum drift, or missing config snapshots should fail the run clearly.
- Partial rebuild artifacts should remain isolated and identifiable; they must not be mistaken for promoted truth.
- Retrying the same replay request should either reuse the same run identity safely or emit a linked retry record with identical inputs.
- Replay logs should be useful, but the authoritative audit record should be structured output, not log scraping.

## Unit Test Expectations

- Double-run determinism test: same inputs twice, identical manifest and output digests.
- Ordering test: equal timestamp events with mixed sequence availability produce the same stable order every run.
- Watermark test: late events are marked consistently and do not silently mutate prior outputs in dry-run mode.
- Missing snapshot test: replay fails when required config or contract snapshot references are absent.
- Runtime mode test: `inspect` and `rebuild` produce no external side effects; `apply` requires explicit gate conditions.

## Contract / Fixture / Replay Impacts

- `schemas/json/replay` should reserve room for run manifests, config snapshot references, and comparison outputs.
- Fixtures should include out-of-order, duplicate, late, and timestamp-degraded windows.
- If parity tests are later added, they should compare shared deterministic ordering and bucket assignment only, not make Python a runtime dependency.

## Summary

This module makes replay trustworthy by freezing the exact inputs that shape a run and by defining one stable ordering model. The important outcome is deterministic reproduction, not a fast but opaque rebuild loop.
