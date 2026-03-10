# Binance Live Contract Seams And Fixtures

## Ordered Refinement Plan

1. Lock Binance live identity, market-type, timestamp, and source-record conventions
2. Expand the Binance fixture corpus to cover the chosen Spot and USD-M streams plus degraded paths
3. Prove schema, fixture, and runbook alignment so later runtime epics can consume one stable contract seam

## Problem Statement

The Binance live initiative cannot safely move into runtime work until one explicit contract seam exists for canonical symbol mapping, stream-specific time selection, stable source-record IDs, and deterministic fixture coverage.

Shared event families already exist, but this epic must decide how live Binance Spot and USD-M behavior fits those families without forcing later runtime slices to improvise.

## Requirements

- Reuse the shared event families already defined under `schemas/json/events` unless this epic proves a concrete Binance-specific gap.
- Preserve the asset-centric canonical symbol vocabulary already used downstream: `BTC-USD` and `ETH-USD`.
- Preserve `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` on canonical outputs so Spot and USD-M remain distinguishable.
- Lock source-record ID rules for the first live Binance slice across:
  - Spot `trade`
  - Spot `bookTicker`
  - Spot order-book snapshot and delta handling
  - USD-M `markPrice@1s`
  - USD-M `forceOrder`
  - USD-M REST `openInterest`
  - adapter-scoped `feed-health`
- Lock stream-specific time semantics:
  - choose the correct Binance exchange timestamp field per stream
  - preserve `recvTs` everywhere
  - use strict degraded fallback when exchange time is missing, invalid, or implausible
- Expand deterministic fixtures for happy, degraded, duplicate, stale, and replay-sensitive Binance cases needed by later runtime epics.
- Align runbooks and consumer checks with the chosen semantics so later planners do not reinterpret the rules.

## Out Of Scope

- websocket connection manager implementation
- order-book buffering and snapshot recovery implementation
- raw storage or replay engine changes
- `services/market-state-api` live cutover
- frontend behavior changes

## Already Covered By Earlier Work

- `plans/completed/canonical-contracts-and-fixtures/` already defines the shared contract families, canonical BTC/ETH symbol vocabulary, timestamp fallback model, and deterministic fixture conventions.
- `plans/completed/market-ingestion-and-feed-health/` already defines feed-health vocabulary, normalizer handoff expectations, and the rule that sequence gaps and stale data must degrade explicitly.
- `services/venue-binance/README.md` and `configs/local/ingestion.v1.json` already reserve the intended MVP stream inventory for Spot and USD-M.
- Existing Binance fixtures already cover a subset of the target surface, including trade, mark/index, timestamp degradation, and sequence-gap cases.

## What This Epic Still Needs

- one written rule for canonical symbol mapping across Spot and USD-M, including quote-currency preservation
- one written rule for source-record identity patterns across the selected Binance stream families
- one written rule for stream-specific exchange-time precedence and degraded fallback behavior
- a Binance-specific fixture matrix that covers the selected live streams and the operator-relevant edge cases
- targeted validation and runbook notes that later runtime epics can inherit without reopening contract debates

## Design Notes

### Identity And Provenance

- Canonical symbols stay asset-centric so downstream current-state and feature consumers do not need a second symbol namespace for Binance USD-M.
- Instrument identity still stays explicit via `sourceSymbol`, `quoteCurrency`, and `marketType`.
- Source-record IDs should be stable enough for raw append and replay but should not expose runtime-only sequencing assumptions that belong to the adapter internals.

### Time And Freshness

- Trades, book events, mark/index updates, liquidations, and polled open-interest snapshots should each name the exact Binance field used as exchange time.
- `recvTs` is mandatory on every accepted message for freshness, skew checks, and auditability.
- REST-polled sensors need the same explicit treatment as WS events; poll time should never be silently substituted for exchange time without a degraded reason.

### Fixture Strategy

- Prefer small paired fixtures with raw inputs and expected canonical outputs.
- Cover normal and degraded cases for the streams actually chosen in the initiative, not speculative future streams.
- Include at least one duplicate or re-delivery-sensitive case so later replay work can validate stable source identities.

## Target Repo Areas

- `schemas/json/events`
- `libs/go/contracts`
- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`
- `tests/integration`
- `docs/runbooks`

## Validation Shape

- schema and fixture-manifest validation
- targeted integration tests for timestamp fallback and stable source-record identities
- fixture-backed checks that Binance Spot and USD-M canonicalization preserve provenance and degraded reasons
