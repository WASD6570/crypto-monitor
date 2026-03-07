# Consumer Adoption And Change Protocol

## Validation Workflow

1. Update the touched schema family manifests and concrete schema files under `schemas/json/`.
2. Update fixture examples and replay seeds under `tests/fixtures/` and `tests/replay/` for every touched family.
3. Run `make contracts-validate` to verify manifests, concrete schema files, and fixture or replay schema references.
4. Run `make fixtures-validate` to verify the deterministic fixture corpus and replay seed catalog.
5. Update the affected language consumers listed below before merging the contract change.

## Go Adoption Checklist

- Go services decode canonical payloads through shared packages under `libs/go`, not per-service handwritten structs.
- The first shared Go package should wrap version checks before payload decoding so malformed or unsupported schema versions fail early.
- Every Go service that consumes a touched family must add or update a targeted fixture-backed validation test.
- Go adoption is mandatory for touched families on the live path.

## TypeScript Adoption Checklist

- TypeScript consumers derive shared runtime validation or generated types from the canonical schemas under `schemas/json/` through `libs/ts`.
- `apps/web` and any TS tooling must reject unsupported `schemaVersion` values before rendering or transforming payloads.
- Every touched family used by TypeScript must add or update a fixture-backed validation test or smoke check.
- TypeScript adoption is mandatory for touched families used by the UI or TS tooling.

## Optional Python Parity

- Python usage is offline-only and lives under `libs/python` or `apps/research`.
- Python may validate or compare deterministic fixture interpretation, but no live service may depend on Python modules at runtime.
- When parity helpers are added, they must reuse `tests/fixtures/` and `tests/replay/` directly instead of maintaining a duplicate corpus.
- If Python parity is not implemented for a touched family, the implementation diff should say that explicitly.

## Breaking Change Protocol

### Additive Change

- A change is additive when the active major version and existing required fields stay valid for old consumers.
- Additive changes must update the touched schema file, manifests if needed, and at least one fixture that exercises the new field or branch.
- Additive changes must name the touched families and touched consumers in the implementation diff.

### Breaking Change

- A change is breaking when it removes or renames a field, changes field meaning, changes required-field behavior, or requires consumer logic changes to remain correct.
- Breaking changes require a new major schema filename, manifest update, fixture updates for both the old and new contract paths when rollout overlap exists, and explicit consumer adoption notes.
- Breaking changes may not land without naming the affected Go consumers, affected TypeScript consumers, replay impact, and whether optional Python parity was updated or intentionally deferred.

## Required Diff Checklist

Every contract change should name:

- touched schema families and versions
- touched fixture ids and replay seed ids
- touched Go consumers
- touched TypeScript consumers
- Python parity status: `updated`, `not-applicable`, or `deferred`
- whether the change is additive or breaking

## Starter Consumer Homes

- Go shared adoption notes: `libs/go/README.md`
- TypeScript shared adoption notes: `libs/ts/README.md`
- Python parity notes: `libs/python/README.md`
- Parity test home: `tests/parity/README.md`
