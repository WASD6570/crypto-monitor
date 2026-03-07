# Slow Feed Ingestion Boundaries

## Module Requirements And Scope

Target repo areas:

- `services/*` for scheduled fetchers, adapter boundaries, and operator-visible health state
- `tests/fixtures` for deterministic source payload fixtures when implementation starts
- optional `schemas/json/features` only if a shared slow-context query contract is required later

This module defines how slower institutional context enters the system without polluting realtime ingestion assumptions.

In scope:

- adapter boundaries for CME volume and open interest context
- adapter boundaries for ETF daily flow context
- scheduled fetch cadence guidance and publish-window expectations
- source timestamp capture, ingest timestamp capture, and adapter health semantics
- idempotent ingest expectations for repeated polling of unchanged daily values

Out of scope:

- provider-specific lock-in to one vendor or auth model
- concrete storage schema design beyond what normalization needs
- client/UI rendering details
- using slow-source health to hard-fail realtime market-state endpoints

## Planning Guidance

### Source Abstraction

- Treat CME and ETF data as source families with pluggable adapters, not permanent vendor names.
- Design adapter inputs around normalized acquisition responsibilities:
  - fetch latest published point(s)
  - report source as-of timestamp
  - report ingest timestamp
  - expose source identifiers sufficient for dedupe
  - surface transient fetch failures separately from `no new publication yet`
- Keep source-specific parsing inside the adapter boundary so downstream services see only normalized candidate records.

### Cadence Model

- CME volume/open interest should be modeled as session-based or daily published context, not assumed intraday tick streams.
- ETF daily flow should be modeled as daily published context, often lagging the cash-session narrative.
- Adapters must classify three states before normalization:
  - `published_new_value`
  - `published_same_value_or_same_asof`
  - `not_yet_published`
- Repeated polling during expected publish windows must not create duplicate records or false freshness bumps.

### Health And Failure Boundaries

- Keep slow-source health separate from venue feed-health used for realtime gate calculations.
- Surface health categories such as:
  - `healthy`
  - `delayed_publication`
  - `source_unavailable`
  - `parse_failed`
- `delayed_publication` means the source has not produced a fresh point inside the expected window, not that the market-state engine is unhealthy.
- Adapter failures should be visible to operators and logs, but they must not block current-state endpoint delivery.

### Idempotency And Ordering

- Deduplicate by stable source family, instrument/fund identifier, publication date or session, and metric family.
- Preserve both source-published time and fetch-received time.
- If a source republishes corrected values for the same as-of date, store a corrected record or revision marker explicitly rather than silently overwriting the prior point.
- Any correction handling should remain replay-auditable.

## Implementation Notes For Later Agents

- Prefer a dedicated slow-context ingestion service or module inside an existing service only if ownership stays obvious.
- Use scheduled fetches appropriate for slow data instead of forcing them through websocket-oriented ingestion stacks.
- Keep credentials, rate limits, and source-specific constraints configurable without making the plan depend on one provider.
- Expose adapter metrics for last successful fetch, last new publication, delayed count, and failure count.

## Unit Test Expectations

- adapter parsing succeeds on representative CME and ETF fixtures
- repeated polling of the same published value is idempotent
- `not_yet_published` does not create a false new record
- corrected publication for the same as-of key is explicit and auditable
- delayed publication classification occurs when expected publish window passes without a new point
- source failure and parse failure remain isolated from realtime venue-health status

## Summary

This module should hand the next agent a source-agnostic, scheduled-ingestion boundary for slow CME and ETF context. The critical rule is separation: slow adapters may be delayed or unavailable without becoming a hidden hard dependency for live market-state gating.
