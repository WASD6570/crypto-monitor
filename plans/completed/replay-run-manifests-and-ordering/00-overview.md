# Replay Run Manifests And Ordering

## Ordered Implementation Plan

1. Define replay run manifest, preserved snapshot references, and versioned output contracts.
2. Define deterministic ordering, runtime modes, and replay execution guardrails for Go replay paths.
3. Add targeted deterministic replay, missing-snapshot, and contract validation coverage.

## Problem Statement

The raw event boundary now preserves immutable canonical history, but the platform still lacks a deterministic replay run contract that explains exactly what inputs were loaded, how ordering was decided, and what outputs a replay run produced.

This feature makes replay runs auditable and repeatable by pinning manifests, snapshot references, runtime modes, and ordering rules before later backfill checkpointing or operator apply workflows are introduced.

## Bounded Scope

- replay run manifest fields and lifecycle for inspect, rebuild, and compare flows
- preserved snapshot references for raw partitions, config, contract versions, and build provenance
- one documented deterministic ordering model for replayed canonical events
- replay runtime mode semantics and default safety posture
- replay result and comparison output contracts that downstream audit slices can consume later
- deterministic validation for repeated runs, tie-break ordering, and missing-snapshot failure paths

## Out Of Scope

- backfill request orchestration, checkpoint resume, and operator audit workflow state machines
- publish/apply operational controls beyond naming the future contract seam
- storage-vendor-specific schema, table, or migration design
- UI controls, dashboard replay surfaces, or operator consoles
- downstream business logic for features, regimes, alerts, or outcomes
- any Python dependency in the live replay path

## Requirements

- Build directly on `plans/completed/raw-event-log-boundary/` and preserve its partition manifest continuity assumptions.
- Reuse the canonical timestamp and degraded-marker semantics from `plans/completed/canonical-contracts-and-fixtures/` and `plans/completed/market-ingestion-and-feed-health/`.
- Keep replay execution in Go service and shared Go helper paths, primarily around `services/replay-engine` and `libs/go`.
- Preserve storage-engine neutrality by defining logical partition references and snapshot metadata instead of concrete database or object-store layouts.
- Keep replay independent from live venue connections and mutable current config.
- Preserve the operating defaults from `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`:
  - `exchangeTs` stays primary when plausible
  - `recvTs` stays stored and auditable
  - late events stay persisted and replay-corrected rather than silently mutated
  - one-symbol, one-day replay should remain feasible within the 10 minute local/dev expectation

## Target Repo Areas

- `services/replay-engine`
- `libs/go`
- `schemas/json/replay`
- `configs/*`
- `tests/replay`
- `tests/integration`
- optional offline parity fixtures in `tests/parity` only if ordering parity needs proof later

## Module Breakdown

### 1. Run Manifest And Snapshot Contracts

- Define the replay request, run manifest, preserved snapshot references, and result artifacts.
- Keep contract boundaries explicit so later backfill and audit slices can reuse them without redefining replay identity.

### 2. Deterministic Ordering And Runtime Modes

- Define the single ordering algorithm, tie-break precedence, runtime mode behavior, and deterministic execution expectations.
- Keep apply-side orchestration out of scope while preserving a clean seam for a later slice.

## Design Details

### Replay Run Identity

- A replay run should have one stable run ID plus immutable references to:
  - requested scope
  - resolved raw partition manifest set
  - config snapshot ID or digest
  - schema/contract version set
  - code/build provenance
  - selected runtime mode
- A retry of the same request should either reuse the original run manifest safely or emit a linked retry record that preserves identical input references.

### Snapshot Preservation

- Snapshot references should point to immutable material, not mutable defaults:
  - raw partition manifest entries from the completed raw boundary
  - replay runtime config snapshot for watermark, skew, ordering, and mode rules
  - replay/output schema versions under `schemas/json/replay`
  - build provenance such as git SHA or release identifier
- Missing or mismatched snapshot references should fail replay before ordered execution starts.

### Output Contract Direction

- Prefer versioned replay contracts that distinguish:
  - run manifest: what was requested and resolved
  - run result: status, counters, digests, and artifact refs
  - compare summary: deterministic diff metadata for non-apply review
- Keep output contracts audit-friendly and machine-readable so later backfill or operator slices consume structured facts instead of scraping logs.

### Live vs Research Boundary

- Go owns live-safe replay manifest loading, ordering, and result generation.
- Python may inspect manifests or reproduce ordering offline later, but this plan must not require Python for replay execution.

## Acceptance Criteria

- Another agent can implement the replay manifest and ordering slice without reopening the parent epic.
- The plan names concrete repo areas for contracts, Go runtime code, config snapshot handling, and tests.
- Validation commands cover deterministic double-run behavior, equal-key ordering, missing-snapshot failure, and mode-specific output expectations.
- The plan stays bounded to replay manifests, preserved snapshots, deterministic ordering, runtime modes, and output contracts.
- Backfill checkpointing and apply-side operational workflows remain explicitly deferred.

## ASCII Flow

```text
raw partition manifests + replay request
                |
                v
replay run resolver
  - scope
  - runtime mode
  - config snapshot ref
  - contract version refs
  - build provenance
                |
                v
run manifest
  - stable run id
  - partition set
  - snapshot digests
  - ordering policy id
                |
                v
deterministic replay runtime
  - event-time ordering
  - tie-break rules
  - inspect / rebuild / compare
                |
                +----> run result contract
                +----> compare summary contract
                +----> isolated rebuild artifacts
```
