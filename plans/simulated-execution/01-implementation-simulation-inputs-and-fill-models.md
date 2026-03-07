# Implementation 01: Simulation Inputs And Fill Models

## Module Requirements And Scope

- Define the service-owned simulation request shape and internal execution-input bundle.
- Cover `spot-long`, `perp-long`, and `perp-short` modes with safe defaults for symbol, venue, and leverage selection.
- Reuse emitted alert fields and outcome-evaluation path references instead of introducing separate market-truth logic.
- Specify entry and exit fill-model boundaries, including when to use L2 depth walk, top-of-book fallback, or outright refusal.
- Preserve deterministic inputs where possible and record when nondirect assumptions lower confidence.

## Target Repo Areas

- `services/simulation-api`
- `services/outcome-engine`
- `libs/go`
- `schemas/json/simulation`
- `schemas/json/outcomes`
- `schemas/json/alerts`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Inputs And Dependencies

- Required source identifiers: `alertId`, `outcomeRecordId`, symbol, setup family, alert direction, alert timestamp, config version, and algorithm version.
- Required market-path references: trusted post-alert price path, regime/degradation markers, and timestamp-source provenance already pinned by outcome evaluation.
- Required simulation selectors: mode, side, leverage option for perps, venue preset or safe default, latency preset, fee preset, and slippage preset.
- Optional but high-value inputs: L2 snapshots, top-of-book quotes, and venue health snapshots around entry and exit windows.

## Safe Instrument And Venue Defaults

- Spot defaults to the canonical spot instrument already associated with the alert symbol, with long-only support.
- Perps default to the canonical perp instrument for the same symbol when the alert source already supports that mapping.
- If both spot and perp venues are healthy, choose the venue already designated as the primary trusted venue for the symbol/mode in config.
- If the preferred venue is degraded but another trusted venue remains healthy, allow substitution only when the run stores the substitution reason and confidence drops.
- If no trusted venue/instrument mapping exists, refuse the run rather than inventing a synthetic pair.

## Fill-Model Plan

### Entry Trigger

- Simulation starts from alert emission, not signal nomination, because the user can only react after emission.
- Entry clock begins at stored alert-emission processing time plus the selected latency preset.
- If timestamp provenance is degraded, continue only if ordering remains auditable; otherwise refuse and record `TIMESTAMP_UNTRUSTED`.

### Entry Price Selection

- Preferred method: L2 depth walk against the selected venue's order book at or immediately after simulated order-start time.
- Fallback method: top-of-book quote plus configured impact curve when L2 is degraded but quote health is still acceptable.
- Last-resort refusal: if neither L2 nor trustworthy quote data exists, do not synthesize a fill from unrelated venues or midpoints.

### Exit Price Selection

- Use the same venue and mode as entry unless an explicit config-defined failover rule applies.
- Exit conditions should align to stored outcome boundaries where possible: target, invalidation, timeout, or configured horizon close.
- Exit fill modeling uses the same L2/fallback/refusal ladder as entry, with independent confidence degradation if exit-side data is weaker.

### Partial Fills And Simplifications

- MVP should support one deterministic fill result per leg rather than an unbounded stream of partial fills.
- If the order size or notional assumption would force meaningful partial-fill modeling, cap size to a fixed review notional preset or refuse unsupported size classes.
- Keep review presets simple enough that replay fixtures can prove deterministic outcomes.

## Direction And Leverage Rules

- `spot-long`: buy then sell; no margin borrowing, no synthetic short, leverage fixed at `x1`.
- `perp-long`: buy/open long then close long; support `x1`, `x2`, `x3`, `x5`.
- `perp-short`: sell/open short then buy-to-close; support `x1`, `x2`, `x3`, `x5`.
- For perps, store liquidation-distance and maintenance-margin assumptions only as informational guardrails unless deterministic venue rules are available in config.
- If a requested leverage or side is unsupported for the chosen instrument, refuse with a stable reason code.

## Determinism Rules

- Reuse pinned outcome-evaluation path references so both features observe the same event order.
- Pin all preset values by versioned config identifiers instead of free-form UI inputs.
- Sort venue candidates deterministically before applying health-based selection.
- Do not derive random slippage, random latency, or Monte Carlo behavior in MVP.

## Unit-Test Expectations

- request validation tests for missing `alertId`, `outcomeRecordId`, unsupported mode, and unsupported leverage
- deterministic venue-selection tests where preferred venue is healthy, substituted, or unavailable
- L2 depth-walk fixture tests for spot long, perp long, and perp short entry/exit calculations
- refusal tests for missing trusted quotes, critically degraded L2, or ambiguous instrument mapping
- replay tests that confirm identical simulation outputs for identical fixtures and config versions

## Summary

This module defines the trusted input seam for simulation and the fill-model ladder from L2 depth walk to explicit refusal. It keeps spot long and perp long/short coverage bounded, preserves deterministic venue selection, and records every fallback that weakens realism for later confidence labeling.
