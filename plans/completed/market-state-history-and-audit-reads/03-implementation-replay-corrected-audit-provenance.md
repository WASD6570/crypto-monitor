# Implementation Replay Corrected Audit Provenance

## Module Requirements And Scope

- Target repo areas: `services/replay-engine`, `services/feature-engine`, `services/regime-engine`, `schemas/json/replay`, `libs/go`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Define the audit lineage and provenance rules that explain how a historical market-state answer was produced or corrected.
- Ensure replay correction semantics stay deterministic, conservative, and machine-readable.

## In Scope

- provenance assembly for composite snapshot ids, bucket artifact ids, regime artifact ids, replay manifest ids, and correction lineage
- audit reason codes that explain why a state is original, replay-corrected, superseded, or unavailable
- lineage rules for late events, corrected bucket membership, config-version changes, and algorithm-version changes
- service-owned audit response mapping that allows downstream investigations to explain a historical answer without reading raw event logs directly

## Out Of Scope

- rebuilding raw event diff viewers or operator-facing incident tooling
- changing replay execution semantics, retention policy, or manifest schema beyond what this response needs to reference
- cross-feature audit for alerting or outcome systems

## Provenance Rules

- Every history response should point to the exact replay or live-produced artifact chain used to answer the lookup.
- If a later replay supersedes a prior answer, the audit response should expose both the corrected authoritative lineage and the superseded lineage reference.
- Distinguish correction causes clearly:
  - `late_event_rebucket`
  - `replay_config_pin_change`
  - `algorithm_version_change`
  - `artifact_unavailable`
- When no correction exists, mark the response as `authoritative_original` and still include enough artifact ids to reproduce the answer.

## Audit Surface Expectations

- Prefer the audit response to embed the resolved lookup tuple, correction status, and provenance chain in one service-owned payload.
- Keep the lineage bounded: immediate authoritative answer plus directly superseded predecessor references are enough for this slice.
- Do not require callers to fetch raw manifests or bucket tables separately just to understand why a historical answer changed.

## Determinism And Safety Expectations

- The same replay manifest, bucket key, and version tuple must produce the same correction status and provenance ids on repeated runs.
- Correction lineage must never point across mismatched symbols or bucket families.
- If lineage is incomplete or inconsistent, fail closed with explicit unavailable audit status rather than emitting partial provenance.

## Test Expectations

- integration tests for audit responses on authoritative-original and replay-corrected cases
- replay tests covering late-event correction lineage and config-version-pinned provenance
- regression checks that audit responses stay aligned with the resolved history payload for the same lookup tuple

## Summary

This module makes historical reads explainable. The implementation should expose replay-corrected provenance as a bounded machine-readable lineage that tells operators which artifacts produced the authoritative answer and whether replay changed it.
