# Slow Context Service Boundary And Adapters

## Module Requirements And Scope

Target repo areas:

- `services/slow-context`
- optional small shared helpers under `libs/go`
- `tests/fixtures/slow-context`

This module defines the live-path entry point for slower institutional context and the adapter contracts each source family must satisfy.

In scope:

- create the shallow `services/slow-context` boundary if it does not already exist
- define adapter interfaces for CME volume/open-interest families and ETF daily-flow families
- define the classified adapter result for published, repeated, and not-yet-published reads
- capture source identity and timing metadata needed for later normalization/query work

Out of scope:

- persistent storage or current-state query response design
- dashboard or API consumer rendering
- vendor-specific auth plumbing beyond config seams

## Planning Guidance

### Service Shape

- Prefer one dedicated service boundary instead of hiding scheduled slow-source logic in `services/normalizer` or venue websocket services.
- Keep the initial folder shallow and explicit:
  - `services/slow-context/adapters`
  - `services/slow-context/poller`
  - `services/slow-context/health`
  - `services/slow-context/cmd` only if the repo convention later needs a runnable entry point
- If a shared helper is needed, keep it narrow and data-oriented under `libs/go/slowcontext` rather than inventing a generic ingestion framework.

### Adapter Contract

- Each adapter should return one classified poll result with:
  - `sourceFamily`
  - `metricFamily`
  - `status`: `published_new_value`, `published_same_value_or_same_asof`, or `not_yet_published`
  - `sourceKey`
  - `asOfTs`
  - `publishedTs` when distinct
  - `ingestTs`
  - stable dedupe identity
  - parsed candidate value payload when a publication exists
  - source error metadata when parsing/fetching fails
- Keep source-specific parsing and vendor payload details inside the adapter.

### Family Coverage

- CME volume and CME open interest may share a source-family adapter if publication semantics stay explicit per metric family.
- ETF daily flow may use a separate source-family adapter because publication timing and identifiers are materially different.
- Do not assume BTC and ETH are always symmetric; preserve asset/instrument identity in the adapter result.

## Unit Test Expectations

- CME published fixture parses into a classified `published_new_value` result
- ETF published fixture parses into a classified `published_new_value` result
- repeated same-as-of fixture is classified as `published_same_value_or_same_asof`
- no-publication-yet fixture is classified as `not_yet_published`
- parse failure returns explicit source-family failure details without panicking

## Summary

This module gives later implementation one explicit Go home and one adapter vocabulary for slower institutional sources. The next module can build polling and health logic on top of these contracts without re-deciding where the live-path slow-context seam belongs.
