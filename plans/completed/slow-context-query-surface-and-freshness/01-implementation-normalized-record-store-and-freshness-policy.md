# Normalized Record, Store, And Freshness Policy

## Module Requirements And Scope

Target repo areas:

- `services/slow-context`
- optional narrow shared helpers under `libs/go/features` or `libs/go/slowcontext`

This module defines the accepted slow-context record, the latest/history lookup seam needed for current queries, and the pinned-clock freshness policy for CME and ETF context.

In scope:

- accepted slow-context record shape after source-boundary parsing
- latest trusted value lookup plus correction-aware history expectations
- explicit freshness classification for `fresh`, `delayed`, `stale`, and `unavailable`
- threshold basis metadata and operator-safe message keys or texts

Out of scope:

- dashboard/UI rendering
- provider-specific fetch logic already owned by `plans/completed/slow-context-source-boundaries/`
- warehouse/migration design beyond the bounded store/query seam needed here

## Planning Guidance

### Accepted Record Shape

- Preserve these fields at minimum:
  - `contextFamily`: `cme_volume`, `cme_open_interest`, `etf_daily_flow`
  - `asset`
  - `sourceKey`
  - `asOfTs`
  - `publishedTs`
  - `ingestTs`
  - `expectedCadence`
  - `revision`
  - value + unit
  - previous accepted value or reference when useful for comparison/audit
- Do not collapse `asOfTs` and `ingestTs`; both are required for later operator messaging.

### Freshness Policy

- Compute freshness under a pinned clock inside services.
- Recommended source-family defaults:
  - CME volume/open interest: `delayed` after the expected next publication window, `stale` after 36 hours
  - ETF daily flow: `delayed` after the expected daily publication window, `stale` after 48 hours
- `unavailable` means no trusted accepted record exists or the lookup/store path cannot safely return last-known-good data.
- Response metadata should include the threshold basis used so later UI copy stays deterministic.

### Store And Correction Semantics

- Keep storage append-safe and correction-aware.
- Same-as-of corrected values must remain explicit via `revision` or equivalent lineage, not silent overwrite.
- The latest query path should read preaccepted records, not re-run raw source parsing on every request.
- If implementation uses an in-memory or interface-backed store first, keep the interface narrow enough that a later persistence layer can replace it without changing response semantics.

## Unit Test Expectations

- accepted records preserve source, as-of, published, and ingest timestamps distinctly
- freshness classification is deterministic at the exact CME and ETF threshold boundaries
- `unavailable` is explicit when no trusted record exists
- corrected same-as-of records remain auditable and do not erase prior identity silently
- latest trusted lookup prefers the newest accepted revision for a given as-of identity

## Summary

This module fixes the service-owned meaning of slow context before any endpoint or current-state integration is added. The next module can expose a query surface confidently because freshness and latest-value semantics are already pinned down.
