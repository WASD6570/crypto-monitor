# Data Contracts And Query Surfaces

## Module Goal

Define the minimum service surfaces and frontend data boundaries required for `apps/web` to render the visibility dashboard without moving market logic into the client.

## Target Repo Areas

- `apps/web/src/api`
- `apps/web/src/features`
- `libs/ts` for shared client-side contract helpers if needed
- service query surfaces owned by the implementation features that precede this dashboard

## Requirements

- Depend on service-owned current-state outputs for BTC and ETH.
- Keep all market logic, regime computation, fragmentation logic, and degraded-state derivation in services.
- Surface freshness, timestamps, config version, and degradation reasons directly from upstream payloads.
- Support a fast current-state read path that can meet the under-2-second dashboard target.
- Support stale handling when one or more query surfaces fall behind, error, or return partial data.
- Avoid inventing concrete API schemas beyond the minimum payload shapes needed for planning.

## Query Surface Inventory

The dashboard should plan around four logical query surfaces. These may map to separate endpoints, a consolidated snapshot endpoint, or a mix of HTTP plus streaming updates. The feature implementation may choose transport, but the logical contract boundaries should remain stable.

### 1. Dashboard Snapshot Surface

Purpose: power the initial render and periodic resync.

Must provide:

- dashboard generation timestamp
- current per-symbol state for BTC and ETH
- last-update timestamps per symbol section
- global regime ceiling if present
- active config or ruleset version reference
- a top-level completeness indicator so partial snapshots are explicit

Usage rules:

- this is the preferred first request on page load
- if available, it should include enough data to render all top-level panels without waterfalls
- if the service cannot provide a fully consolidated snapshot, the client may parallelize section requests but must still expose partial completeness honestly

### 2. Symbol Detail Surface

Purpose: provide focused overview and microstructure detail for the currently selected symbol.

Should provide:

- service-owned overview metrics for the symbol
- microstructure metrics already normalized for UI consumption
- WORLD vs USA comparison values or labels needed for the detail view
- symbol-specific freshness, degraded reasons, and provenance metadata

Safe default:

- one query per active symbol view, cached per symbol with short-lived freshness metadata

### 3. Derivatives Context Surface

Purpose: provide offshore/perp context without coupling the UI to raw perp events.

Should provide:

- current derivatives-context summary for the selected symbol
- timestamps and cadence markers for each derivatives block
- explicit availability or unsupported indicators when data is not present for the current environment
- degraded reasons if upstream perp inputs are impaired

Safe default:

- load after snapshot success or in parallel when infrastructure allows, but do not block symbol overview from rendering

### 4. Feed Health And Regime Surface

Purpose: explain trust and degradation.

Should provide:

- per-venue health rows for relevant venues
- symbol-level health rollup used by regime or market-quality logic
- degraded reason codes and human-readable summaries from services
- freshness ages based on `recvTs` semantics
- explicit unavailable state when a health source is down

Safe default:

- poll or stream often enough that health changes reach the UI quickly, but always prefer bounded server-side aggregation over client fan-out to many venue-specific calls

## Minimal Payload Shape Expectations

These are planning-level expectations, not frozen schema definitions.

### Shared Metadata Expectations

Each logical surface should expose enough metadata for the UI to remain honest:

- `asOf` or equivalent response timestamp
- per-block freshness or last-success timestamp
- config or state version reference when relevant
- degraded and stale markers from services
- optional explanatory reason list intended for operator reading

### Symbol State Expectations

For BTC and ETH the UI depends on service-owned fields equivalent to:

- symbol identity
- current state label
- current state reason summary
- market-quality or fragmentation summary
- WORLD and USA comparison context
- last-updated times for state computation

### Microstructure Expectations

The UI expects a bounded set of display-ready metrics, not raw books or trade streams. Examples include service-computed spread quality, imbalance, short-horizon pressure, and venue agreement indicators.

### Derivatives Expectations

The UI expects context blocks already reduced into operator-facing summaries. It should not compute perp basis, funding state, leverage stress, or spot-perp mismatch locally.

### Feed Health Expectations

The UI expects venue health rows already classified into a small, shared severity vocabulary. It should not infer health from missing websocket messages on its own.

## Freshness, Staleness, And Loading Defaults

### Freshness Rules

- use service-reported timestamps first
- display relative age and exact timestamp where space allows
- if a block age exceeds the service-provided freshness expectation or the client cannot refresh it, label the block `stale`

### Stale-State Handling

- keep last known good data visible if it remains within a tolerable operator window
- overlay stale messaging with explicit age and trust reduction
- if the stale window becomes severe or a required surface is unavailable, degrade the whole panel to `unavailable` rather than implying the old state is still current

### Loading Rules

- initial snapshot loading may block interaction only until a minimum viable current-state shell renders
- symbol detail, derivatives context, and health sections may load independently after the shell appears
- no spinner-only full-screen state after the initial load unless all required query surfaces fail

## Error And Degradation Messaging Rules

- degraded-feed messaging must name the affected venue or surface when services provide that information
- partial data must state what is missing, not just that something failed
- if a service returns timestamp-degraded or fallback-to-`recvTs` markers, the UI must show a visible but compact trust note
- error copy must never instruct the operator to assume a neutral market state

## Performance And Caching Defaults

- optimize for one fast current-state snapshot request on entry
- avoid client fan-out to many per-venue endpoints if a server-side aggregation surface exists
- cache recent symbol payloads briefly to support quick BTC/ETH switching without jank
- route-split any heavy visualization code so the first current-state render stays within the initiative's trust-first goal

## Contract Risks To Resolve During Implementation

- whether one consolidated dashboard snapshot is sufficient for all panels or whether the health surface needs its own cadence
- whether regime explanation text is emitted directly by services or assembled from service-supplied reason fragments in the UI
- whether transport uses polling only or polling plus push updates; either approach must preserve the same logical query surfaces

## Unit And Integration Test Expectations

- client decoders or adapters reject malformed or incomplete critical fields cleanly
- stale markers propagate to panel-level state instead of silently dropping timestamps
- partial snapshot responses render visible partial-state warnings
- fixture payloads with timestamp fallback markers produce trust messaging without changing displayed state labels

## Summary

This module defines the UI data boundary: services publish trusted current-state, microstructure, derivatives, and health outputs; the Vite SPA consumes them through a small set of logical surfaces, preserves freshness and degradation metadata, and refuses to invent market meaning locally.
