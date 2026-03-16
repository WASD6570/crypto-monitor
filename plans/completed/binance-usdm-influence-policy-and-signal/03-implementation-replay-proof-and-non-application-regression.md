# Implementation Module 3: Replay Proof And Non-Application Regression

## Scope

- Add deterministic replay and focused integration proof for the new internal signal.
- Prove current `/api/market-state/*` behavior remains unchanged in this child.
- Exclude the later application of the signal to current-state or regime outputs.

## Target Repo Areas

- `tests/replay`
- `tests/integration`
- any focused Go test helpers needed in `services/feature-engine` or `services/venue-binance`

## Requirements

- Prove identical pinned Spot plus USD-M inputs yield identical influence signal outputs.
- Prove absent USD-M context preserves current Spot-only external behavior.
- Keep replay proof explicit and repeatable for both `BTC-USD` and `ETH-USD`.
- Avoid writing consumer-facing assertions that assume the follow-on child already applies the signal.

## Key Decisions

- Extend existing Binance replay and current-state determinism coverage rather than inventing a separate replay harness for USD-M influence.
- Add focused integration regression that the current API contract stays unchanged while the internal signal lands.
- Keep proofs centered on deterministic evaluator output, no-context fallback, and unchanged external behavior.

## Unit Test Expectations

- Replay tests confirm identical influence outputs across repeated runs.
- Focused integration tests confirm current market-state API outputs remain backward-compatible in this child.
- Regression coverage exercises both symbols and both mixed-surface USD-M acquisition modes when relevant.

## Contract / Fixture / Replay Impacts

- Replay fixtures should stay pinned and deterministic.
- If new evaluator fixtures are needed, keep them narrow and aligned with existing Binance USD-M fixture vocabulary.
- No API schema or shared contract change should be required here.

## Summary

This module closes the first child with proof that the new signal is deterministic and real while the public market-state surface remains intentionally unchanged until the follow-on application slice.
