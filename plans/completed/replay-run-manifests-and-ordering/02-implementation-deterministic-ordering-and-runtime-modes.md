# Implementation Module 2: Deterministic Ordering And Runtime Modes

## Module Requirements

- Define one deterministic ordering algorithm for replayed canonical events that preserves the existing timestamp and degraded-marker semantics.
- Define bounded runtime modes for `inspect`, `rebuild`, and `compare` without expanding into apply-side orchestration.
- Ensure concurrency, partition streaming, and artifact writing never change replay order or output digests for identical inputs.
- Keep the implementation in Go service and shared helper paths.

## Target Repo Areas

- `services/replay-engine`
- `libs/go`
- `configs/*`
- `tests/replay`
- `tests/integration`
- optional `tests/parity` only if offline ordering proof is needed later

## Scope

- implement the ordered read model that consumes the frozen run manifest from module 1
- define tie-break behavior for equal timestamps and mixed sequence availability
- define runtime-mode semantics and artifact boundaries for inspect, rebuild, and compare
- define deterministic guardrails for streaming, batching, and isolated rebuild outputs

## Ordering Policy

Replay should use a single stable precedence chain:

1. persisted canonical bucket timestamp choice from raw storage (`exchangeTs` when plausible, else degraded `recvTs`)
2. stable venue sequence or source ordering field when present
3. stable canonical event ID as the final tie-breaker

Additional planning requirements:

- preserve the persisted timestamp-source decision from the raw boundary; replay must not recompute a new bucket source
- sort equal-key records with a deterministic stable sort, never map iteration or goroutine completion order
- keep bucket assignment in UTC and event-time-first semantics
- carry forward late-event and timestamp-degraded markers unchanged into replay outputs

## Mixed Sequence Availability

- Events from streams that expose sequence fields should use them before canonical event ID.
- Events from streams without a reliable sequence field should fall back directly from timestamp choice to canonical event ID.
- Mixed partitions must still merge into one deterministic total order using the same precedence chain for every run.

## Runtime Modes

### `inspect`

- load manifest and ordered inputs
- compute counters, watermarks, and expected correction metadata
- emit run result plus optional inspection artifact refs
- do not write rebuilt downstream state

### `rebuild`

- execute ordered replay and write isolated rebuilt artifacts into a run-scoped namespace
- emit run result with artifact refs and digests
- do not promote or publish outputs outside the isolated namespace

### `compare`

- execute the same ordered replay path as `rebuild`
- compare rebuilt artifacts against current materialized outputs or pinned expected outputs
- emit a deterministic comparison summary contract
- do not publish or mutate live outputs

Reserve `apply` for a later slice. This module may leave an enum seam or reserved validation branch, but it should not plan operator approvals, checkpoint resume, or external side effects.

## Deterministic Runtime Guardrails

- Stream partitions incrementally to stay within the one-symbol, one-day local/dev replay target.
- If concurrency is used for reading or transformation, merge outputs through one deterministic ordering stage before artifact emission.
- Use explicit artifact naming that includes run ID and mode so repeated runs can be compared cleanly.
- Do not depend on wall-clock time, live network calls, or mutable config lookups after manifest freeze.

## Negative Cases To Plan

- equal timestamps with partial sequence coverage
- duplicate raw entries that remain audit-visible but should not destabilize ordering
- timestamp-degraded events crossing a UTC day boundary
- missing output schema version or compare target reference
- replay request that attempts to use unsupported runtime mode

## Unit Test Expectations

- double-run determinism test: identical manifest twice yields identical order, counters, and output digests
- tie-break test: equal timestamp events with mixed sequence availability resolve to the same order every run
- degraded timestamp ordering test: persisted fallback to `recvTs` remains stable across reruns
- runtime mode test: `inspect` writes no rebuilt artifacts, `rebuild` writes isolated artifacts only, `compare` emits deterministic diff metadata only
- unsupported mode test: reserved future modes fail clearly without partial side effects

## Validation Commands

- `go test ./services/replay-engine/... -run 'TestReplayDeterministicDoubleRun|TestReplayStableOrderingWithMixedSequenceAvailability|TestReplayPreservesDegradedTimestampOrdering'`
- `go test ./services/replay-engine/... -run 'TestReplayInspectModeDoesNotWriteArtifacts|TestReplayRebuildModeWritesIsolatedArtifacts|TestReplayCompareModeEmitsDeterministicSummary|TestReplayRejectsUnsupportedMode'`
- `go test ./tests/integration -run 'TestReplayOneSymbolOneDayDeterministicManifestExecution'`

## Summary

This module makes replay behavior reproducible by locking one ordering model and three bounded runtime modes around the frozen manifest inputs. Later slices can build checkpointing or operator promotion on top of these semantics without redefining order, mode meaning, or result artifacts.
