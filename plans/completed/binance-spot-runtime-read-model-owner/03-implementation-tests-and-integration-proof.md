# Implementation: Tests And Integration Proof

## Requirements And Scope

- Add focused validation for the sustained runtime owner without expanding into browser smoke or broader environment rollout.
- Cover startup honesty, accepted progression, reconnect/resync carry-forward, and deterministic repeated-input behavior.
- Reuse existing Binance fixtures and integration helpers where possible.

## Target Repo Areas

- `cmd/market-state-api/*_test.go`
- `tests/integration/*.go`
- `tests/replay/*.go` if repeated-input determinism needs a focused proof at the read-model seam
- `tests/fixtures/events/binance/**` only if one new runtime script fixture is required

## Implementation Notes

- Replace the current command tests that assume per-request `/api/v3/depth` polling with tests that drive the sustained owner through scripted supervisor and recovery events.
- Keep one narrow integration harness that injects accepted Spot runtime inputs and verifies the provider-facing snapshot contains the expected symbol observations and degradation posture.
- Add a deterministic repeated-input proof that runs the same accepted sequence twice and asserts identical observation ordering, symbol presence, prices, timestamps, and degradation reasons.
- Prefer fixture-backed accepted frame scripts over live network behavior so this feature stays fast and reproducible.
- If a small fake transport or scripted runtime adapter is needed, keep it local to tests and mirror existing Binance integration conventions.

## Unit Test Expectations

- Command-level tests prove provider construction starts the sustained owner and surfaces configuration failures clearly.
- Integration tests prove one symbol can be available while the other remains warming up, preserving partial current-state behavior.
- Integration tests prove a sequence gap or reconnect changes machine-readable degradation without breaking snapshot reads.
- Repeated-input tests prove deterministic output for identical accepted runtime inputs.

## Summary

This module provides the proof that the new read-model owner is safe to consume before the later provider cutover slice. The tests focus on runtime semantics and determinism, not on changing the existing HTTP contract.
