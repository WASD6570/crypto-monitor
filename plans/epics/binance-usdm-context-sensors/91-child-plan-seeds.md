# Child Plan Seeds: Binance USD-M Context Sensors

## `binance-usdm-mark-funding-index-and-liquidation-runtime`

- Outcome: plan the bounded USD-M websocket runtime slice that consumes `markPrice@1s` and `forceOrder`, emits canonical `funding-rate`, `mark-index`, and `liquidation-print` events, and surfaces websocket-specific feed health.
- Primary repo areas: `services/venue-binance`, `services/normalizer`, `tests/fixtures/events/binance`, `tests/integration`
- Dependencies: inherited Wave 1 symbol, timestamp, and source-record rules; existing schemas for funding, mark/index, and liquidation; Binance local ingestion defaults
- Validation shape: parser fixtures for happy and degraded mark price plus force-order cases, targeted runtime tests for USD-M reconnect/staleness handling, normalization checks for provenance preservation
- Why it stands alone: it is the websocket-only part of the derivatives surface and delivers most of the live context value without being blocked on REST cadence decisions

## `binance-usdm-open-interest-rest-polling`

- Outcome: plan the bounded REST slice that polls Binance USD-M `openInterest`, maps it to `open-interest-snapshot`, and makes poll freshness, exchange time, degraded reasons, and rate-limit posture explicit.
- Primary repo areas: `services/venue-binance`, `configs/local/ingestion.v1.json`, `tests/fixtures/events/binance`, `tests/integration`
- Dependencies: inherited Wave 1 timestamp and provenance policy; existing `open-interest-snapshot` schema; venue REST limits and health thresholds from config
- Validation shape: polling freshness tests, degraded timestamp fixtures when REST data lacks trustworthy exchange time, and integration checks that polled samples retain distinct REST provenance from websocket sensors
- Why it stands alone: REST cadence and freshness semantics are the main rollout-sensitive part of the epic and can be planned independently from websocket connection behavior

## `binance-usdm-context-sensor-fixtures-and-integration`

- Outcome: plan the proving slice that expands the USD-M fixture corpus and targeted integration coverage so websocket and REST sensors can be validated together without contract drift.
- Primary repo areas: `tests/fixtures/events/binance`, `tests/integration`, `docs/runbooks`
- Dependencies: settled child-plan outputs from the websocket runtime slice and the REST polling slice; inherited shared schema ownership from Wave 1
- Validation shape: fixture-manifest updates, mixed-surface integration smoke for freshness/degradation visibility, and checks that duplicate or replay-sensitive cases preserve stable source identities
- Why it stands alone: it is a validation and operational-proof slice that should consume the other two child decisions rather than shaping them prematurely
