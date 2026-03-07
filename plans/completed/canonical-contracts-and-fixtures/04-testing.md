# Testing Plan: Canonical Contracts And Fixtures

Expected output artifact: `plans/completed/canonical-contracts-and-fixtures/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Contract family validation | Validate schema families and versions | All declared schema files compile or validate cleanly | Schema validation command output |
| Fixture validation | Validate canonical fixtures against target schemas | Happy-path and degraded fixtures map to the correct schema family and version | Fixture validation command output |
| Replay seed determinism | Replay the same seed twice | Event ordering, timestamp handling, and expected outputs match exactly | Replay smoke diff or checksum evidence |
| Consumer validation | Validate Go and TS consumers against sample canonical payloads | Both runtimes accept valid payloads and reject malformed ones | Targeted test output |
| Optional parity | Compare shared deterministic fixture interpretation across Go and Python if parity helpers exist | Outputs match for the scoped shared logic | Parity test output |

## Required Commands

The implementing agent should provide these exact commands or their direct equivalents:

- `make contracts-validate`
- `make fixtures-validate`
- `make replay-smoke CONTRACT_FIXTURES=1`
- `go test ./tests/parity/...`
- `pnpm test contracts`
- optional parity command if Python validation is added: `pytest tests/parity -k contracts`

If these commands do not exist yet, implementing this feature must add them or replace them with equally explicit project-standard commands.

## Verification Checklist

- Contract families exist in the correct schema directories.
- Version naming is explicit and documented.
- Canonical symbol, venue, market type, quote context, and timestamp semantics are test-covered.
- Replay seeds include at least one degraded-path scenario.
- Go and TypeScript consumers validate touched payload families.
- Python parity remains optional and offline-only.

## Negative Cases

- Missing `recvTs` should fail validation when the contract requires auditability.
- Unsupported symbol or venue identifiers should fail validation.
- A payload with mismatched schema version and fixture metadata should fail validation.
- Replay seed files with non-deterministic ordering or missing provenance should fail smoke validation.

## Handoff Notes

- No credentials are required for this feature.
- Validation should run entirely from local files and deterministic fixtures.
- This feature is done only when later slices can consume one shared contract and fixture vocabulary without guessing.
