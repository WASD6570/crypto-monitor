# Fixtures And Targeted Validation

## Module Requirements And Scope

Target repo areas:

- `tests/fixtures/slow-context`
- `services/slow-context`

This module defines the deterministic fixture inventory and high-signal tests needed to prove the source boundary behaves safely before later query or UI work starts.

In scope:

- representative CME and ETF source fixtures
- pinned-clock test helpers for publication-window and delay behavior
- targeted Go tests for parsing, idempotency, correction handling, and source-health isolation
- minimal reporting expectations for later implementation handoff

Out of scope:

- broad end-to-end dashboard tests
- replay parity work unless implementation later puts the source boundary on replay-visible paths
- schema or frontend validation

## Fixture Inventory

- CME published fixture with new session/daily value
- ETF published fixture with new daily value
- repeated same-as-of fixture for unchanged polling
- corrected same-as-of fixture with updated value or revision marker
- not-yet-published fixture inside the expected window
- delayed-publication fixture beyond the expected window
- transient source failure fixture or mock response

## Validation Guidance

- Prefer focused Go package tests over a repo-wide suite.
- Use a pinned clock for publish-window boundaries so delay classification stays deterministic.
- Keep fixture names date-pinned and source-family specific so future query/persistence work can reuse them directly.
- If implementation adds metrics/log fields, validate them through stable structured outputs rather than string-only log assertions.

## Unit Test Expectations

- source fixtures parse consistently into classified poll results
- delayed-publication state is deterministic under a pinned clock
- repeated unchanged polling does not produce duplicate accepted outputs
- corrected same-as-of fixture preserves explicit revision/correction handling
- slow-source parse/fetch failure remains isolated from realtime feed-health semantics

## Summary

This module closes the feature with deterministic evidence. Later child planning can build storage/query semantics on top of a proven source-boundary fixture set instead of inventing new edge-case data.
