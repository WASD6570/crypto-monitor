# Query Surface And Current-State Isolation

## Module Requirements And Scope

Target repo areas:

- `services/slow-context`
- `services/feature-engine`
- optional `schemas/json/features` only if a non-Go consumer contract must be introduced now

This module defines how services expose slow context and how the current-state path includes it without letting slow-context failure break BTC/ETH market-state delivery.

In scope:

- one service-owned slow-context response block or endpoint
- integration seam for current-state reads to include slow context optionally
- explicit unavailable/error-scoped response behavior
- compatibility with existing current-state availability/provenance conventions where helpful

Out of scope:

- dashboard composition and UI copy details
- turning slow context into a gate for realtime symbol/global state
- historical audit/replay retrieval beyond the bounded latest/current read seam

## Planning Guidance

### Response Shape

- The response should include enough fields for the later dashboard panel without client reconstruction:
  - latest value and unit
  - previous comparable value when directly available
  - `asOfTs`, `publishedTs`, `ingestTs`
  - absolute age or age basis
  - expected cadence label
  - freshness state
  - availability state
  - revision/correction marker
  - operator-safe message key or message text
- Reuse existing `available` / `degraded` / `partial` / `unavailable` availability vocabulary when it helps consistency, but keep freshness separate from availability.

### Current-State Integration Rule

- Slow-context lookup failure must degrade only the slow-context block.
- The main current-state response for BTC/ETH must still succeed when slow context is missing, stale, delayed, or unavailable.
- If slow context is integrated into current-state responses, it should appear as an explicit nested block, never as silently omitted fields.
- If implementation needs a dedicated endpoint first, keep the assembler reusable so later inline current-state integration does not fork semantics.

### Contract Discipline

- Add JSON schemas only if this slice must expose the block to a non-Go consumer immediately.
- If schemas are added, keep them obviously scoped under `schemas/json/features` and validate touched consumers in Go.
- If schemas are not added yet, document the Go-owned response seam clearly enough that the later dashboard slice can consume it without re-deciding field semantics.

## Unit And Integration Test Expectations

- query assembly returns explicit unavailable payloads instead of omitting the block
- lookup failure leaves the core current-state response successful
- slow-context availability/freshness fields remain stable under deterministic fixtures
- current-state integration does not alter existing regime/composite availability semantics

## Summary

This module defines the service-owned way consumers read slow context and ensures that slow-context problems stay isolated. The next module can validate the seam through deterministic fixtures and focused integration tests.
