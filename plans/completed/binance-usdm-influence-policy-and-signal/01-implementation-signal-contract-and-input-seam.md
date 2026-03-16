# Implementation Module 1: Signal Contract And Input Seam

## Scope

- Define the internal USD-M influence signal shape and the evaluator input bundle it consumes.
- Cover deterministic symbol ordering, no-context rules, degraded-context rules, and per-input freshness semantics.
- Exclude current-state/regime output application.

## Target Repo Areas

- `libs/go/features`
- `services/venue-binance`
- focused unit tests in the same areas

## Requirements

- Introduce one internal signal contract that can represent auxiliary, bounded degrade-cap, no-context, and degraded-context posture.
- Keep the contract per-symbol for `BTC-USD` and `ETH-USD` only.
- Preserve provenance and freshness inputs from funding, mark/index, liquidation, and open-interest context rather than collapsing them into one opaque boolean.
- Keep field semantics deterministic and stable for repeated identical input bundles.
- Avoid any API-facing schema change in this module.

## Key Decisions

- Place reusable signal types in `libs/go/features` so later application work can consume them without moving core types again.
- Add the smallest Go-owned evaluator-input seam near the existing USD-M runtime surfaces in `services/venue-binance` rather than teaching downstream services to query raw venue state directly.
- Make no-context and degraded-context first-class states instead of encoding them as missing data or implied reason strings.
- Include explicit reason codes, trigger metrics, and input timestamps so later replay and application work stay auditable.

## Unit Test Expectations

- The signal contract validates required symbols, version fields, and deterministic ordering rules.
- Identical evaluator inputs produce identical signal objects.
- Missing all USD-M context produces the planned no-context posture without mutating current Spot-only behavior.
- Mixed fresh and degraded inputs preserve the planned degraded-context semantics and reason ordering.

## Contract / Fixture / Replay Impacts

- This module adds an internal Go contract only; no public schema change is expected.
- If helper fixtures are needed, keep them local to Go tests and aligned with the existing Binance USD-M fixture vocabulary.
- Field naming should be stable enough for replay assertions in the next modules.

## Summary

This module settles the exact internal language for USD-M influence so later evaluator and replay work can depend on one explicit contract instead of rediscovering semantics from scattered venue inputs.
