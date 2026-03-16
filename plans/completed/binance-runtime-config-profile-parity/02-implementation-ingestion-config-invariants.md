# Implementation Module 2: Ingestion Config Invariants

## Scope

- Strengthen `libs/go/ingestion` coverage so the repo validates all checked-in environment profiles instead of only the current local happy path.
- Add the smallest negative regression coverage needed to keep the Binance open-interest requirements from silently drifting again.

## Target Repo Areas

- `libs/go/ingestion/config.go`
- `libs/go/ingestion/config_test.go`

## Requirements

- Add table-driven coverage for `configs/local/ingestion.v1.json`, `configs/dev/ingestion.v1.json`, and `configs/prod/ingestion.v1.json`.
- Assert the environment label, fixed symbol set, and Binance runtime invariants for each profile.
- Pin the chosen Binance open-interest defaults and the monotonic environment-pressure relationships in tests.
- Preserve `VenueRuntimeSource.RuntimeConfig(...)` as the single validation authority; avoid spreading config policy into unrelated packages.
- Add a focused negative regression that proves an open-interest stream is rejected when its poll interval or per-minute limit is missing or zero.

## Key Decisions

- Prefer real-file loading tests for the checked-in profiles so the repo fails fast if a committed JSON profile regresses.
- Keep negative cases small and synthetic so they target only the open-interest rule instead of duplicating full environment fixtures.
- Preserve current validation error semantics unless a clearer error message is required to distinguish missing interval versus missing per-minute limit.

## Unit Test Expectations

- All three checked-in environment files parse successfully.
- Binance runtime config from each file includes explicit positive open-interest poll settings.
- The environment ladder assertions fail if a later edit makes `dev` or `prod` more aggressive than the profile above it.
- Invalid open-interest profile combinations still fail with explicit validation errors.

## Contract / Fixture / Replay Impacts

- No schema or replay changes are expected.
- This module hardens the checked-in config contract that later rollout features and tests will rely on.

## Summary

This module turns the checked-in environment profiles into a tested contract: real files must load, Binance open-interest settings must remain explicit, and the repo must reject any future attempt to configure the stream without the required polling fields.
