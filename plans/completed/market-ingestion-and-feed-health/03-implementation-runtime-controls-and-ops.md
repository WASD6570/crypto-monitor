# Implementation Module 3: Runtime Controls And Operational Validation

## Scope

Plan the runtime safeguards, observability, and operational checks that keep ingestion safe and explainable over long-running operation.

## Target Repo Areas

- `services/venue-*`
- `services/normalizer`
- `libs/go`
- `configs/*`
- `docs/runbooks`
- `tests/integration`

## Requirements

- Define reconnect backoff and retry ceilings per venue.
- Define REST snapshot recovery behavior and rate-limit safety.
- Define message lag, staleness, reconnect-count, and resync-count metrics.
- Define internal ops alert conditions for feed degradation.
- Define runbook expectations for operators:
  - how to identify a stale venue
  - how to tell whether resync loops are transient or persistent
  - what later services should do with degraded inputs
- Ensure runtime controls are configuration-driven where thresholds may change by venue.

## Key Decisions To Lock

- Visibility into ingest health is part of product correctness, not just SRE hygiene.
- Config should tune thresholds, but the meaning of health states should stay stable across venues.
- Operators must be able to distinguish transport failure, sequencing failure, and timestamp-trust failure.

## Deliverables

- Ops metric inventory
- Internal alert condition matrix for feed health
- Runbook outline for degraded-feed investigation
- Config inventory for staleness, reconnect, and resync controls

## Unit Test Expectations

- Staleness thresholds should be integration-testable with synthetic time advancement.
- Retry policies should be deterministic in tests.
- Health-state transitions should be asserted from observable metrics or emitted health events.

## Contract / Fixture / Replay Impacts

- Feed-health events or payloads must integrate cleanly with future market-state logic.
- Runbooks should reference the same health vocabulary as contracts and logs.
- Replay or fixture tests should cover degraded operational scenarios, not just happy-path payloads.

## Summary

This module makes the ingestion system operable. It prevents the product from silently showing unreliable state during venue degradation and creates the operational trail needed for later no-operate decisions.
