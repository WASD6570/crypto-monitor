# Implementation Module 3: Command Consumption Proof

## Scope

- Add narrow provider-level proof that `cmd/market-state-api` can consume each checked-in Binance environment profile via `configPath` without changing startup-default selection rules.
- Reuse the existing stub-server harness so the proof stays offline and deterministic.

## Target Repo Areas

- `cmd/market-state-api/main_test.go`

## Requirements

- Add a table-driven test that exercises `newProviderWithOptions(...)` with `local`, `dev`, and `prod` config paths against stubbed Spot and USD-M endpoints.
- Prove the provider starts and closes cleanly for each profile while preserving the fixed symbol set and current runtime-status assumptions.
- Keep the existing Spot-only override rejection and symbol-guardrail tests intact.
- Do not add environment-selection variables, fallback behavior, or new runtime wiring in this module.

## Key Decisions

- Treat this as consumer proof, not startup-contract design: the test should validate that the current provider accepts the checked-in profiles, not redefine how operators pick one.
- Reuse the existing websocket and REST harness patterns in `cmd/market-state-api/main_test.go` instead of introducing new test infrastructure.
- Keep the assertions structural and deterministic; the test should prove profile consumption, not pin live market-state values.

## Unit Test Expectations

- `newProviderWithOptions(...)` succeeds for each checked-in environment config path when given stub endpoints.
- The provider can be closed cleanly after startup for each profile.
- Existing override guardrail tests still pass unchanged.

## Contract / Fixture / Replay Impacts

- No API, schema, or replay contract changes are expected.
- The benefit is earlier detection if a checked-in profile stops being consumable by the live command owner.

## Summary

This module proves the checked-in environment files are not just syntactically valid; they remain acceptable inputs to the current Go-owned live provider without prematurely taking on the next feature's startup-default design work.
