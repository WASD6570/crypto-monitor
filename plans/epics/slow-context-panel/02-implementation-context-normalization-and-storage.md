# Context Normalization And Storage

## Module Requirements And Scope

Target repo areas:

- `services/*` for normalization logic, persistence, and query assembly
- `schemas/json/features` only if implementation needs a shared API contract for slow-context responses
- `tests/fixtures` and `tests/parity` when deterministic fixtures or cross-language checks are needed

This module defines the service-owned record shape, freshness semantics, and storage/query behavior for slower institutional context.

In scope:

- normalized record fields for CME volume, CME open interest, and ETF daily flow
- source-time vs ingest-time handling
- cadence metadata and freshness classification rules
- storage expectations for append-safe, auditable slow-context history
- query semantics that keep slow context advisory and non-blocking

Out of scope:

- frontend layout decisions
- business logic that promotes slow context into hard gating
- speculative warehouse or migration design beyond MVP needs

## Normalized Record Guidance

Each normalized slow-context point should preserve enough information for operator display and replay-safe auditing:

- `contextFamily`: `cme_volume`, `cme_open_interest`, or `etf_daily_flow`
- `asset`: `BTC` or `ETH` when applicable; allow a broader instrument key if source granularity differs
- `sourceKey`: source family or provider identifier
- `asOfTs`: when the source says the value is effective
- `publishedTs`: when the source made the value available, if distinct
- `ingestTs`: when the service fetched and accepted the value
- `expectedCadence`: human-readable cadence such as `daily` or `session`
- `freshnessState`: `fresh`, `delayed`, `stale`, or `unavailable`
- `value` plus unit metadata
- `revision` or correction marker when a source republishes an as-of point

Do not collapse `asOfTs` and `ingestTs`. The operator needs both timing dimensions to understand whether the context is naturally slow or operationally delayed.

## Freshness And Cadence Rules

- Freshness is computed in services using source family defaults, not in the UI.
- Suggested MVP defaults:
  - CME volume/open interest: `fresh` when within the expected next-publication window, `delayed` after that window, `stale` after 36 hours
  - ETF daily flow: `fresh` when current for the latest expected daily publication, `delayed` after the expected publish window, `stale` after 48 hours
- `unavailable` is reserved for no accepted record yet or query failure that prevents a trustworthy last-known value.
- Query responses should include both the current freshness state and the threshold basis used to classify it, so UI copy stays deterministic.

## Storage Guidance

- Use append-safe persistence that can retain historical slow-context points and corrections.
- Do not overwrite last-known values without preserving revision history when a source publishes corrections.
- Retention can follow derived-feature policy unless implementation discovers materially different cost characteristics.
- If current-state endpoints include slow context inline, they should read pre-normalized records rather than re-fetching external sources on request.

## Query Surface Guidance

- Provide a dedicated slow-context response block or endpoint owned by services.
- Include summary fields needed for the dashboard without forcing the client to reconstruct semantics:
  - latest value
  - previous comparable value when helpful
  - absolute age
  - expected cadence label
  - freshness state
  - operator-safe message key or message text
- Query failure for slow context should degrade only that block; the rest of the dashboard current-state response must still succeed whenever possible.
- If slow context is unavailable, return an explicit empty/unavailable state instead of omitting the field silently.

## Replay, Determinism, And Compatibility Notes

- Slow-context fixtures should be deterministic and date-pinned so repeated runs classify freshness the same way under a controlled clock.
- If replay later includes slow context, replay time must be explicit so freshness state remains reproducible.
- Any shared contract change should update consumer validation in Go and TypeScript together.

## Unit Test Expectations

- normalization preserves source, as-of, published, and ingest timestamps distinctly
- freshness classification is deterministic under a pinned clock
- stale, delayed, fresh, and unavailable states are emitted exactly at threshold boundaries
- correction records for the same as-of date remain auditable
- query assembly returns explicit unavailable payloads instead of silent omission
- current-state query path still returns market-state data when slow-context lookup fails

## Summary

This module gives later implementation a service-owned slow-context model with explicit cadence and age semantics. The key outcome is a query surface that explains slow institutional context clearly while staying independent from realtime gate calculations.
