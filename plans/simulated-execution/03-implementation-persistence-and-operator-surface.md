# Implementation 03: Persistence And Operator Surface

## Module Requirements And Scope

- Define append-only persistence for saved simulation runs and refused-run audit records.
- Expose operator-facing retrieval and comparison surfaces without moving simulation logic into the client.
- Ensure every run is explicitly separated from live trading and clearly auditable by alert, outcome, config, and assumption provenance.
- Support saved runs as first-class review artifacts for later notes, analytics, and baseline comparison.

## Target Repo Areas

- `services/simulation-api`
- `apps/web`
- `schemas/json/simulation`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `configs/*`
- `tests/integration`

## Persistence Plan

- Store every simulation attempt as append-only: successful run, low-confidence fallback run, or refused run.
- Persist immutable identifiers for `simulationRunId`, `alertId`, `outcomeRecordId`, preset/version identifiers, requester identity or system actor, and request timestamp.
- Persist the full assumption envelope so future review does not depend on current defaults.
- Persist derived outputs separately from raw assumptions so audit tools can distinguish chosen settings from computed results.
- Keep schema room for replay provenance and supersession links, but do not overwrite historical records when presets evolve.

## Minimum Stored Fields

- request context: symbol, setup family, alert direction, simulation mode, leverage, venue, notional preset
- timing context: alert emission time, latency preset, timestamp source, degraded timestamp flags
- market-data confidence: L2 health, quote health, venue substitution flags, degradation reason codes
- economic assumptions: fee preset, slippage method, funding inclusion or omission, config version
- outputs: entry estimate, exit estimate, gross move, fees, slippage, net estimate, confidence label, refusal reason if any
- provenance: algorithm version, schema version, replay/live source manifest, saved-run label or operator note reference

## Operator Surface Plan

- `apps/web` should render simulation as a clearly labeled review tool inside alert drill-down or related saved-run views.
- The UI should allow choosing a bounded preset set rather than arbitrary free-form execution math.
- Saved runs should display side-by-side comparisons across `spot-long`, `perp-long`, and `perp-short` where applicable, with leverage and confidence visible at a glance.
- Refused runs should be visible with actionable explanation text derived from stable reason codes, not silently hidden.
- Every surface must visibly state that results are simulated and non-executable.

## API And Query Expectations

- Provide service endpoints or query methods to create a saved run, list saved runs for an alert, fetch a single run by ID, and compare runs by mode/preset.
- Authorization must remain server-side; the client cannot mark a run trusted by itself.
- Query surfaces should support filters by symbol, setup family, mode, leverage, confidence label, and config version for later analytics reuse.
- Keep current-state query targets in mind from operating defaults; alert drill-down simulation retrieval should stay in the same under-5-second review posture as other alert details.

## Storage And Retention Expectations

- Follow operating defaults: alerts, outcomes, simulations, and decision logs remain hot for at least 180 days and retained for at least 2 years in cold storage.
- Treat refused runs and low-confidence runs as audit-relevant records, not disposable noise.
- Include indexes or equivalent query structure around `alertId`, `outcomeRecordId`, symbol, mode, confidence label, and config version.

## Separation From Live Trading

- No persistence field should imply broker connectivity, API key storage, order intent, or account balances.
- UI actions must be named around `save simulation`, `rerun preset`, or `compare assumptions`, never `place trade` or similar.
- Service boundaries should ensure that simulation endpoints cannot reuse future live-trading credentials even if such a feature is introduced elsewhere.

## Unit-Test Expectations

- persistence tests for append-only writes and retrieval by `alertId` and `simulationRunId`
- schema/contract tests for successful, low-confidence, and refused run payloads
- UI rendering tests for explicit `SIMULATED` labeling and refused-run explanation display
- authz tests proving unauthorized clients cannot create or fetch another operator's private saved-run metadata if such scoping exists
- retention/query tests or migration checks that preserve audit fields and version references

## Summary

This module turns simulation into a durable review artifact: append-only saved runs with full audit context, retrieval surfaces in `apps/web`, and strong visual and contractual separation from any live-trading concept. Later notes, analytics, and baseline comparisons should reuse these stored records directly.
