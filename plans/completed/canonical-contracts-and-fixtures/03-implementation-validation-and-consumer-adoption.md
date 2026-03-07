# Implementation Module 3: Validation And Consumer Adoption

## Scope

Plan how shared contracts and fixtures are validated and adopted by Go, TypeScript, and optional Python consumers without making Python part of the live path.

## Target Repo Areas

- `tests/parity`
- `libs/go`
- `libs/ts`
- `libs/python`
- `docs/specs`
- build or validation command wiring at the repo root

## Requirements

- Define the validation workflow for schema files and fixtures.
- Define how Go services validate or decode canonical payloads.
- Define how TypeScript UI or tooling validates or types canonical payloads.
- Define optional offline Python validation/parity usage without making it required for live operation.
- Define change-management rules for contract updates:
  - what counts as additive
  - what counts as breaking
  - which fixtures and consumers must be updated together

## Key Decisions To Lock

- Contract consumers should validate against generated or shared artifacts, not ad hoc handwritten assumptions in each service.
- Consumer validation in Go and TypeScript is mandatory for touched families.
- Python parity checks are optional but should reuse the same fixtures and expected outputs.
- Any contract change must name the affected families, fixture updates, and downstream consumers in the implementation diff.

## Deliverables

- Validation command plan for schema and fixture checks
- Consumer-adoption checklist for Go and TypeScript
- Optional parity checklist for Python
- Breaking-change protocol for future contract revisions

## Unit Test Expectations

- Go validation should reject malformed canonical payloads.
- TypeScript validation should reject malformed canonical payloads or incompatible versions.
- Parity checks, when added, should compare deterministic fixture outputs for shared logic only.

## Contract / Fixture / Replay Impacts

- This module prevents later services from silently diverging in their interpretation of canonical contracts.
- Replay and outcome correctness depend on consumer agreement about field meaning and version identity.

## Summary

This module ensures shared contracts are not just documented but enforced. It reduces long-term drift across services, UI consumers, and optional offline analysis tooling.
