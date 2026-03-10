# Refinement Map: Binance Live Contract Seams And Fixtures

## Current Status

This epic is newly materialized from `initiatives/crypto-market-copilot-binance-live-market-data/03-handoff.md` and is still too broad for direct `feature-planning` or `feature-implementing`.

Completed prerequisite coverage already exists in:

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`

Those completed slices already provide:

- shared event families under `schemas/json/events`
- canonical BTC/ETH symbol vocabulary with preserved quote context
- strict timestamp fallback semantics and degraded markers
- feed-health vocabulary for stale, reconnect-loop, resync-loop, and sequence-gap states
- deterministic fixture conventions and manifest-based fixture indexing

## Existing Partial Coverage In The Repo

- `services/venue-binance/README.md` already names the intended Spot and USD-M MVP stream inventory.
- `configs/local/ingestion.v1.json` already reserves Binance runtime defaults for reconnects, snapshots, and health thresholds.
- `tests/fixtures/events/binance/` already includes starter fixtures for Spot trade, Spot sequence-gap handling, timestamp degradation, and mark/index.

## What This Epic Still Needs

- a Binance-specific identity policy for canonical symbols, `marketType`, `quoteCurrency`, and `sourceSymbol`
- stable source-record ID rules across the selected Spot and USD-M stream families
- stream-specific exchange-time selection rules, including how REST-polled open interest is timestamped and degraded
- missing Binance fixture coverage for top-of-book, funding/open-interest/liquidation, and duplicate or replay-sensitive cases
- validation and runbook alignment so later runtime epics inherit one stable contract seam

## What Should Not Be Re-Done

- shared contract-family creation already finished in `plans/completed/canonical-contracts-and-fixtures/`
- generic feed-health vocabulary already finished in `plans/completed/market-ingestion-and-feed-health/`
- runtime connection handling, depth recovery logic, and API cutover work that belong to later Binance live epics

## Refinement Waves

### Wave 1

- `binance-live-identity-and-time-policy`
- Why first: every later Binance runtime slice depends on one stable answer for canonical identity, source-record IDs, and stream-specific timestamp rules.

### Wave 2

- `binance-live-fixture-corpus-expansion`
- Why next: fixture coverage should be derived from the wave-1 rules instead of inventing its own interpretation of identity and time semantics.

### Wave 3

- `binance-contract-validation-and-runbook-alignment`
- Why last: consumer and runbook proof should validate the already-chosen semantics and fixture matrix instead of leading the contract decisions.

## Notes For Future Planning

- Keep canonical symbols asset-centric as `BTC-USD` and `ETH-USD`; do not introduce a second downstream-facing symbol namespace in this epic.
- Preserve `sourceSymbol`, `quoteCurrency`, and `marketType` explicitly so Spot and USD-M stay distinguishable without contract guesswork.
- Treat REST-polled `openInterest` as part of the same contract seam as WS streams; do not leave its timestamp/freshness semantics implicit.
- Later runtime epics should inherit these rules rather than reopen them inside adapter implementation work.
