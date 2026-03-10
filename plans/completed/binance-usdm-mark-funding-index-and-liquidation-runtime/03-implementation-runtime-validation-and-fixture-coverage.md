# Implementation Module: Runtime Validation And Fixture Coverage

## Scope

- Add the high-signal fixture corpus and targeted integration checks needed to prove the USD-M websocket runtime, canonical mapping, and degradation behavior are correct.
- Keep this validation slice focused on websocket runtime behavior for `markPrice@1s` and `forceOrder`; do not fold in REST `openInterest` coverage.

## Target Repo Areas

- `tests/fixtures/events/binance/BTC-USD/*.fixture.v1.json`
- `tests/fixtures/events/binance/ETH-USD/*.fixture.v1.json`
- `tests/fixtures/manifest.v1.json`
- `tests/integration/ingestion_smoke_test.go`
- `tests/integration/binance_usdm_runtime_test.go`
- `docs/runbooks/ingestion-feed-health-ops.md`

## Requirements

- Extend the Binance fixture corpus with real USD-M websocket-shaped raw payloads and expected canonical outputs for funding, mark/index, and liquidation.
- Cover both happy and degraded timestamp paths, plus at least one reconnect/resubscribe or stale-health runtime case.
- Keep fixture IDs, categories, and target schema names consistent with the existing corpus conventions.
- Add targeted integration tests that use the new fixtures with `services/venue-binance` and `services/normalizer` together.
- Keep validation high-signal and deterministic; avoid broad compose or end-to-end work that belongs to later slices.

## Key Decisions

- Prefer a small Binance-specific fixture set over broad synthetic dumps. Each fixture should explain one operator-relevant behavior.
- Use fixture-backed integration tests to prove both canonical output correctness and health visibility; do not rely on parser-only tests for degradation semantics.
- Add one stale-health test driven by elapsed time after the last `markPrice@1s` update and one non-stale test showing that a lack of `forceOrder` traffic alone does not degrade health.
- If runbook terminology changes are needed, update only the operator notes that mention USD-M websocket degradation so they stay aligned with the shared health vocabulary.

## Unit Test Expectations

- Fixture manifest validation includes the new Binance USD-M fixtures.
- Integration smoke proves happy-path mark/funding and liquidation normalization against expected canonical outputs.
- Integration smoke proves timestamp-degraded `markPrice@1s` still emits canonical events with degraded timestamp status.
- Runtime smoke proves reconnect or resubscribe degradation remains machine-visible through feed-health output.
- Integration smoke proves `forceOrder` inactivity does not create false stale feed-health.

## Contract / Fixture / Replay Impacts

- Fixture corpus expands, but contract families should remain unchanged.
- The added fixture matrix becomes the seed input for the later `binance-usdm-context-sensor-fixtures-and-integration` child feature; keep naming and categories stable for that follow-on work.
- Replay implications are limited to deterministic canonical IDs and degraded reasons; full replay-engine coverage remains a later feature.

## Summary

This module proves the USD-M websocket runtime is trustworthy with a compact Binance fixture set and targeted integration smoke that covers happy flow, degraded timestamps, reconnect behavior, and the deliberate non-stale treatment of sparse liquidation traffic.
