# Binance Live Raw Storage And Replay

## Epic Summary

Materialize the Binance raw append and replay audit boundary so accepted Spot and USD-M live inputs can be persisted with stable provenance, partitioned deterministically, and replayed through the existing replay engine without output drift.

## In Scope

- raw append integration for accepted Binance Spot and USD-M canonical events and feed-health outputs
- stable partitioning, source-ID, sequence-ID, and degraded-feed provenance for Binance live families
- replay-engine acceptance of Binance raw partitions and event families already normalized elsewhere in the initiative
- deterministic replay checks for repeated runs and duplicate-input/idempotency behavior
- fixture-backed verification that retained degraded feed-health and timestamp facts survive raw storage and replay

## Out Of Scope

- broad historical migration or backfill planning
- replay-engine redesign unrelated to concrete Binance live families
- market-state API cutover and frontend/dashboard behavior
- reopening already-completed Spot or USD-M parsing and canonical handoff logic

## Target Repo Areas

- `libs/go/ingestion`
- `services/replay-engine`
- `tests/replay`
- `tests/integration`

## Validation Shape

- targeted Go tests for raw append entry construction, duplicate audit, partition routing, and degraded-feed provenance
- replay-engine tests that resolve Binance partitions and prove identical ordered output across repeated runs
- deterministic fixture-backed integration for Spot and USD-M accepted inputs flowing into raw append then replay boundaries
- direct duplicate-input and idempotency checks instead of a validation-only child feature

## Major Constraints

- preserve stable canonical symbols (`BTC-USD`, `ETH-USD`) with explicit `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- retain explicit degraded feed-health evidence and timestamp fallback facts in raw entries so replay remains auditable
- do not make Python part of the live or replay runtime path
- reuse the existing raw append and replay contracts rather than inventing a Binance-only storage format
- keep this epic attached to the existing initiative ordering: raw/replay must settle before live market-state API cutover
