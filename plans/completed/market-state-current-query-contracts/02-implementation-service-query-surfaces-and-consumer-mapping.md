# Implementation Service Query Surfaces And Consumer Mapping

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `services/regime-engine`, `libs/go`, `tests/fixtures`, `tests/integration`, `tests/replay`, `apps/web` as a consuming surface only
- Define the service-owned query surfaces that assemble the current-state contracts from upstream composite, bucket, and regime outputs.
- Map the read-only consumer rules for dashboard and later service consumers so market-state trust stays on the server side.

## In Scope

- symbol current-state query surface for `BTC-USD` and `ETH-USD`
- global current-state query surface for the global ceiling and capped symbol summary view
- assembly rules that package composite snapshots, bucket summaries, and regime outputs into one versioned response
- bounded recent-context assembly from already-closed windows only
- consumer mapping for dashboard adapters and later service-side readers
- integration and replay seams that verify the assembled responses stay deterministic and version-pinned

## Out Of Scope

- arbitrary historical queries, audit lookup endpoints, or replay diff APIs
- storage-layer indexing, retention, or persistence design
- dashboard route/component implementation or client caching strategy beyond trust guidance
- new business logic for weighting, divergence, quality, or regime state calculation

## Service Surface Recommendations

- Keep current-state assembly service-owned in Go, with `services/feature-engine` providing composite and bucket summaries and `services/regime-engine` providing symbol/global regime outputs.
- Recommended read surfaces:
  - symbol current state: `GET /api/market-state/{symbol}` or equivalent RPC
  - global current state: `GET /api/market-state/global` or equivalent RPC
- Prefer a single symbol response payload that already includes the currently effective global ceiling so consumers do not need to fan out and merge two requests just to render trust state.
- Keep the global endpoint for overview screens and service consumers that need the ceiling independently.

## Assembly Rules

- Read only the latest closed composite, bucket, and regime artifacts; do not expose in-flight or partially closed bucket state.
- Carry through the exact version metadata attached upstream rather than recomputing it at the query layer.
- Package degraded/unavailable reasons directly from upstream outputs and add only transport-level completeness markers where needed.
- Bound recent context to a fixed response section assembled from the latest closed windows; no user-controlled date-range or cursor semantics belong here.
- If upstream state is unavailable, return explicit unavailable sections with reason metadata rather than omitting fields or fabricating fallback values.

## Consumer Mapping

### `apps/web`

- Reads the current-state response as the source of truth for summary strip, trust state, divergence cards, and recent sparkline context.
- May derive presentation-only labels, colors, and ordering.
- Must not recompute composite weights, bucket severities, regime state, or global ceiling application.

### Later Service Consumers

- Alert/risk or operator-support services read the same current-state contracts to obtain authoritative state and provenance.
- Downstream records should preserve response version metadata and upstream reason codes when they persist or relay state.
- Consumers may cache briefly for read performance, but they must treat service responses as invalid once `asOf` or version context changes.

## History/Audit Seam To Preserve

- Keep request/response semantics limited to current state plus bounded recent context.
- Reserve explicit extension points through bucket identifiers, source timestamps, and version metadata so `market-state-history-and-audit-reads` can later add retrieval by bucket or replay context.
- Do not introduce `/history`, pagination, or audit-detail variants in this slice, even if the current response reuses fields those later reads will need.

## Validation Expectations

- integration tests for symbol current-state assembly in healthy and degraded cases
- integration tests for global-state assembly and ceiling propagation into symbol responses
- fixture-backed consumer smoke checks showing one current-state payload is sufficient for dashboard adapters
- replay tests confirming repeated pinned runs return identical current-state responses and version metadata

## Summary

This module defines how services assemble and expose current-state read models. The critical boundary is that services package authoritative state once, consumers read it directly, and later history/audit work extends the family without turning this slice into a general-purpose retrieval API.
