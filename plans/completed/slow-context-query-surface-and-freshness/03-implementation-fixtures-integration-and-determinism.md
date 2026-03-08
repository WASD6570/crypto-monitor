# Fixtures, Integration, And Determinism

## Module Requirements And Scope

Target repo areas:

- `tests/fixtures/slow-context`
- `services/slow-context`
- `tests/integration`
- optional `tests/replay`

This module defines the deterministic fixture inventory and focused validation for freshness thresholds, unavailable behavior, and current-state isolation.

In scope:

- pinned-clock slow-context fixtures for fresh, delayed, stale, and unavailable cases
- focused integration tests for current-state success when slow-context lookup fails
- deterministic validation of correction-aware latest reads
- optional replay checks only if the implementation introduces replay-visible persistence now

Out of scope:

- dashboard browser tests
- parity checks unless this slice actually mirrors logic outside Go
- broad service end-to-end orchestration beyond the slow-context query seam

## Fixture Inventory

- fresh CME slow-context record
- delayed CME slow-context record just beyond the expected publish window
- stale CME slow-context record beyond 36 hours
- fresh ETF slow-context record
- stale ETF slow-context record beyond 48 hours
- unavailable slow-context query case with no trusted accepted record
- lookup-failure case where slow-context storage/query errors while current-state data remains healthy
- corrected same-as-of fixture showing revision-aware latest selection

## Validation Guidance

- Prefer focused Go package and integration tests over a repo-wide suite.
- Use pinned clocks for every freshness threshold boundary.
- Keep fixture names date-pinned and source-family specific so the later dashboard panel can reuse them directly.
- If current-state integration is added here, assert both the preserved market-state sections and the explicit slow-context degraded/unavailable block in the same test.

## Unit And Integration Test Expectations

- freshness classification is deterministic for CME and ETF threshold transitions
- unavailable responses are explicit when no trusted record exists
- current-state queries still succeed when slow-context lookup fails
- corrected same-as-of records return the latest accepted revision deterministically
- optional replay tests, if needed, prove pinned-clock reproducibility rather than generic storage coverage

## Summary

This module closes the query/freshness slice with deterministic evidence. Later UI planning can trust that the service-owned slow-context seam already has stable freshness, unavailable, and failure-isolation behavior.
