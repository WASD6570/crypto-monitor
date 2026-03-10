# Binance USD-M Context Sensor Fixtures And Integration

## Ordered Implementation Plan

1. Expand the Binance USD-M fixture corpus and operator runbooks so websocket-derived and REST-derived context sensors are described together without reopening contract rules.
2. Add mixed-surface integration coverage that proves websocket and REST feed-health signals stay distinct, machine-visible, and symbol-safe when they coexist.
3. Add replay-sensitive identity checks across websocket and REST normalization paths so repeated inputs preserve stable source-record IDs and raw duplicate facts.
4. Record validation evidence in `plans/binance-usdm-context-sensor-fixtures-and-integration/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to fixture corpus, integration proof, and runbook updates for the already-implemented Binance USD-M websocket and REST slices.
- Do not reopen canonical symbol, provenance, timestamp, or feed-health vocabulary decisions from Wave 1 and the completed child features.
- Keep websocket-derived and REST-derived freshness semantics explicit and separate in both tests and docs.
- Prove stable source-record identity and raw duplicate behavior for at least one websocket sensor and one REST sensor.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Validation Boundary

- Consume `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/` and `plans/completed/binance-usdm-open-interest-rest-polling/` as settled prerequisite behavior.
- Keep new work focused on `tests/fixtures/events/binance`, `tests/integration`, and `docs/runbooks`.
- Prefer fixture-backed and synthetic-time integration tests over live network access.

### What This Feature Proves

- websocket sensors and REST polling can coexist without hiding which acquisition mode produced a degraded condition
- operator runbooks reflect the same vocabulary and reason set emitted by canonical feed-health events
- replay-sensitive inputs keep stable source-record IDs and stream-family partitioning across repeated normalization

## ASCII Flow

```text
Binance USD-M fixtures
  - markPrice@1s
  - forceOrder
  - /fapi/v1/openInterest
          |
          v
tests/integration
  - mixed-surface happy path
  - distinct health visibility
  - duplicate / replay-sensitive identity proof
          |
          +----> docs/runbooks
          |       - operator vocabulary
          |       - mixed-surface checks
          v
completed validation evidence
  - stable sourceRecordId
  - explicit websocket vs REST degradation
```
