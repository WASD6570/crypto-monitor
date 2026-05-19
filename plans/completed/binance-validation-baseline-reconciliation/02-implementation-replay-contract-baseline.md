# Implementation: Replay Contract Baseline

## Requirements And Scope

- Make `make contracts-validate` pass for the replay family without weakening unrelated contract-family checks.
- Preserve the actual replay run-result producer and consumer shape unless implementation deliberately updates all affected contract surfaces.
- Keep changes additive where possible and document any non-additive contract decision before implementation proceeds.

## Target Files

- `scripts/dev/validate_contract_families.py`
- `schemas/json/replay/replay-run-result.v1.schema.json`
- `libs/go/contracts/replay.go`
- `services/replay-engine/runtime.go`
- `docs/specs/canonical-contracts-and-fixtures/00-contract-families.md` only if the canonical inventory needs clarification

## Current Drift

`scripts/dev/validate_contract_families.py` currently expects `replay-run-result.v1.schema.json` to require `id`, `seedId`, `symbol`, and `outputChecksum`. The checked-in schema and Go replay contract instead use the existing result shape around `runId`, `status`, `inputCounters`, `outputDigest`, `startedAt`, `finishedAt`, and `manifestDigest`.

## Implementation Steps

1. Inspect `libs/go/contracts.ReplayRunResult` and `services/replay-engine` result construction before deciding whether the schema or validator is stale.
2. If the current Go/schema shape is canonical, update `scripts/dev/validate_contract_families.py` so its replay run-result required field policy matches `replay-run-result.v1.schema.json`.
3. If implementation proves `id`, `seedId`, `symbol`, and `outputChecksum` are the intended canonical result shape, update the schema, Go contract type, replay engine producer, and any fixtures or tests in the same slice.
4. Keep `replay-run-result.v1.schema.json` internally consistent: every required field must have a property definition, and every producer-required field should have a matching Go JSON tag.
5. Run `make contracts-validate` after the change.

## Contract And Compatibility Notes

- Adding new required schema fields is not additive for existing producers or stored artifacts.
- Renaming `runId` or `outputDigest` would require a compatibility plan and should not be done as a baseline-only shortcut.
- Validator policy should catch missing required fields for implemented schemas, but it should not demand fields that no current producer emits unless a coordinated contract migration is planned.

## Unit Test Expectations

- Contract-family validation fails on real missing schema keys, missing required properties, and schema-version drift.
- The replay run-result schema passes validation with the same required fields that Go producers can emit.
- Go replay-engine tests, if present for result construction, continue to pass after any contract alignment.

## Next-Agent Summary

The expected minimal fix is likely validator-policy alignment, not a replay result schema migration. Verify before editing because contract changes must not silently break replay consumers.
