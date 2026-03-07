# Implementation Module: Ingestion Smoke Matrix And Runbooks

## Requirements And Scope

- Add focused integration coverage for:
  - venue happy path
  - sequence gap degradation
  - stale feed transition
  - bounded reconnect / snapshot recovery behavior
- Write a metric inventory covering message lag, reconnect count, resync count, sequence gaps, and snapshot recovery pressure.
- Write a degraded-feed runbook and alert-condition matrix.

## Target Repo Areas

- `tests/integration`
- `docs/runbooks`
- `services/venue-*`
- `services/normalizer`

## Key Decisions

- Validation should stay fixture-driven and reproducible.
- Runbook language must match the shared degradation reasons and feed states.
- Alert conditions belong to ops visibility, not business-alert logic.

## Unit / Integration Test Expectations

- Health-state transitions are asserted from emitted outputs, not inferred from logs.
- Retry behavior is deterministic under synthetic time.
- A future agent can rerun the same smoke matrix without needing live credentials.

## Contract / Fixture / Replay Impacts

- Smoke outputs should be reusable by later replay/storage features.
- Runbooks should reference the same canonical state names future dashboards will show.

## Summary

This module finishes the operability side of ingestion by turning the runtime semantics into repeatable tests and operator-facing documentation.
