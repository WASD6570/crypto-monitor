# Implementation: Current-State API Proof

## Requirements And Scope

- Prove the sustained runtime owner feeds the existing symbol and global current-state endpoints through the real API/provider path.
- Preserve startup honesty, partial availability, unsupported-symbol behavior, and machine-readable degradation semantics.
- Keep contract validation focused on the existing HTTP and provider surfaces; do not expand scope into new operator endpoints or `/healthz` redesign.

## Target Repo Areas

- `services/market-state-api/*.go`
- `tests/integration/*.go`
- `tests/replay/*.go`
- `cmd/market-state-api/*_test.go` when command-backed API seams need focused coverage

## Implementation Notes

- Add or refresh focused integration coverage that exercises the command/provider path with the sustained runtime owner instead of a polling reader.
- Keep symbol and global assertions shape-oriented: both tracked symbols remain addressable, unsupported symbols still fail the same way, and partial/degraded payloads remain machine-readable.
- Preserve deterministic current-state proof by asserting stable output ordering and stable degradation metadata for identical accepted input sequences.
- If a command-backed handler test or fixture-backed API harness is needed, keep it narrow and reuse the existing Binance test conventions.

## Unit Test Expectations

- Integration tests prove the global endpoint and both symbol endpoints continue to assemble through `NewLiveSpotProvider(...)` with the runtime owner underneath.
- Negative tests prove unsupported symbols and runtime-unavailable startup states still return the expected contract posture.
- Replay tests prove the current Binance determinism selector is up to date and actually executes the intended proof.

## Contract / Replay Notes

- No route or payload shape changes are planned.
- If any new additive metadata is needed to preserve honesty, it must remain backward-compatible and machine-readable.
- Determinism coverage should use the current replay test names so the plan and implementation do not drift apart.

## Summary

This module is the API-proof step for the cutover. It makes the consumer-facing boundary the place where the sustained runtime is validated, without reopening the response contract.
