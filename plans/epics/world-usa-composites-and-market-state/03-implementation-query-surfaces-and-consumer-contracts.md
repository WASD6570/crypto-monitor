# Implementation Query Surfaces And Consumer Contracts

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `services/regime-engine`, `schemas/json/features`, `configs/*`, `tests/fixtures`, `tests/integration`, `apps/web` as a consuming surface only
- Define the versioned outputs and query surfaces that expose composite, feature, and regime state to dashboards and later alert consumers.
- Ensure consumers can explain current state without recomputing service-owned logic.

## In Scope

- versioned payloads for current composite and regime state
- query/read models for current state and recent bucket history
- provenance, degradation, and config/version fields required for auditability
- consumer guidance for UI and alert-engine usage

## Out Of Scope

- dashboard component implementation
- authentication or delivery-channel design beyond existing service conventions
- mutation endpoints for changing live config

## Contract Recommendations

- Define a small set of explicit read models under `schemas/json/features`:
  - `composite_snapshot_v1`
  - `market_quality_snapshot_v1`
  - `market_regime_snapshot_v1`
  - `market_state_query_response_v1`
- Keep payloads additive and versioned.
- Do not force consumers to join raw venue rows to understand current state; provide both summary and provenance in the response.

## Required Response Content

- current WORLD and USA composite snapshots for the symbol
- current divergence and fragmentation summaries
- current 30s, 2m, and 5m feature summaries relevant to operator trust
- current symbol regime and current global regime ceiling
- degraded reasons, excluded venues, and feed-health-derived penalties
- `configVersion`, `algorithmVersion`, `schemaVersion`, and source bucket timestamps
- recent-history window sufficient for dashboard sparklines or regime transition context without client recomputation

## Query Surface Recommendations

- current symbol state endpoint or RPC:
  - example shape: `GET /api/market-state/{symbol}`
- recent state history endpoint or RPC:
  - example shape: `GET /api/market-state/{symbol}/history?window=2h`
- global ceiling endpoint or included payload section:
  - example shape: `GET /api/market-state/global`
- Replay/audit callers should be able to retrieve a state artifact by symbol, bucket, and version context for diffing.

## Consumer Rules

- `apps/web` renders service-owned outputs and may derive display-only formatting such as colors, labels, and sorting.
- `apps/web` must not recalculate composite weights, market-quality scores, fragmentation buckets, or `TRADEABLE/WATCH/NO-OPERATE` states.
- later alert/risk services may consume the same regime outputs as inputs, but must preserve version and provenance references in downstream records.
- optional Python research may read these contracts offline for analysis or parity checks, never as a live dependency.

## Degraded And Negative Cases To Expose

- unavailable composite because all contributors were excluded
- partial composite with reduced coverage but still useful as `WATCH`
- timestamp-degraded bucket source
- replay-corrected historical state that differs from originally emitted live state
- config version change producing different thresholds across adjacent historical windows

## Versioning And Compatibility

- Query responses should include schema version and contract family version references.
- If response semantics change materially, add a new versioned schema rather than silently reinterpreting fields.
- Historical reads must preserve the original version context used to emit the state.
- Avoid embedding UI-only wording in contracts; use machine-readable codes plus concise human-readable reasons where needed.

## Unit And Integration Test Expectations

- schema validation for all new read models
- fixture-backed tests that ensure degraded reasons and provenance fields are present when expected
- integration tests confirming current-state queries return service-computed values without client-side joins
- replay/audit tests verifying historical reads remain tied to original config and algorithm version metadata

## Summary

This module defines how the rest of the product reads the new market-state layer. The critical rule is that services publish complete, versioned, audit-friendly outputs so the UI and future alerting layers can consume trustworthy state without rebuilding logic or losing provenance.
