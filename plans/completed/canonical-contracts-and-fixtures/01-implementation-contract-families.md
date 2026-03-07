# Implementation Module 1: Contract Families

## Scope

Define the shared contract surface and naming rules for canonical events, derived state, alerts, outcomes, replay payloads, and simulations.

## Target Repo Areas

- `schemas/json/events`
- `schemas/json/features`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `schemas/json/replay`
- `schemas/json/simulation`
- `docs/specs`

## Requirements

- Document the contract family purpose for each schema directory.
- Establish versioning rules for schema filenames and compatibility expectations.
- Define canonical identifiers for:
  - `symbol`
  - `venue`
  - `marketType`
  - `quoteCurrency`
  - optional composite identifiers such as `WORLD_SPOT_COMPOSITE` and `USA_SPOT_COMPOSITE`
- Define canonical timestamp fields and their meaning:
  - `exchangeTs`
  - `recvTs`
  - any degraded timestamp indicator needed for auditability
- Reserve explicit fields for downstream provenance where relevant:
  - `configVersion`
  - `regimeTags`
  - `feedHealthState` or equivalent degradation marker
  - replay provenance or source record references where appropriate

## Key Decisions To Lock

- Version contracts per file or family, not via implicit README-only rules.
- Keep event contracts distinct from derived-state contracts; do not overload one family for every downstream use case.
- Preserve source quote and market context even when symbol identity is canonicalized to `BTC-USD` and `ETH-USD`.
- Define contract boundaries for feed health and degradation explicitly rather than burying them in service-specific logs.

## Deliverables

- Contract family inventory document
- Versioning and naming standard
- Reserved field glossary for identity, time, provenance, and degradation
- Initial schema file list to be implemented in follow-up work

## Unit Test Expectations

- Schema validation should fail on missing required identity or timestamp fields.
- Schema validation should fail when an unknown contract version is referenced.
- Fixtures should validate against the correct schema family and version only.

## Contract / Fixture / Replay Impacts

- Every later slice depends on these field names and family boundaries.
- Replay payloads must remain compatible with event contracts without inventing duplicate meanings.
- Derived-state contracts must be traceable back to the event-time semantics defined here.

## Summary

This module locks the vocabulary of the platform. If it is vague, every later implementation will drift. If it is explicit and minimal, later services can implement confidently without redesigning shared payloads.
