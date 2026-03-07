# Implementation Module: Normalizer Handoff Surface

## Requirements And Scope

- Define the minimal Go input types or service functions the normalizer accepts from venue services.
- Route shared trade/book/feed-health outputs through `services/normalizer` without changing the shared normalization semantics.
- Add deterministic tests that prove the handoff preserves timestamps and degraded markers.

## Target Repo Areas

- `services/normalizer`
- `libs/go/ingestion`

## Key Decisions

- Do not duplicate normalization rules inside the service layer.
- Keep the service API explicit and small.
- Treat feed-health outputs as equal citizens to market event outputs.

## Unit Test Expectations

- Trade handoff preserves canonical trade output.
- Book handoff preserves canonical book output.
- Feed-health handoff preserves degradation reasons and state.

## Contract / Fixture / Replay Impacts

- This slice stabilizes the handoff later replay and storage layers will consume.
- No new contract family should be invented unless the existing shared schemas are insufficient.

## Summary

This module turns the shared normalization library into an explicit service boundary so later ingestion consumers do not need to call library internals directly.
