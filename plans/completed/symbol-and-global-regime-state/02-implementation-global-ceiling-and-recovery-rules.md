# Implementation Global Ceiling And Recovery Rules

## Module Requirements And Scope

- Target repo areas: `services/regime-engine`, `libs/go/features`, `configs/local`, `configs/dev`, `configs/prod`, `schemas/json/features` only if a global regime seam schema is reserved now, `tests/fixtures`, `tests/integration`, `tests/replay`
- Consume emitted symbol regime snapshots and the underlying 5m bucket provenance needed to explain cross-symbol ceilings.
- Emit a global regime ceiling that caps symbol state per operating defaults while preserving symbol-specific differences when shared conditions do not justify a full global downgrade.
- Keep the module storage-neutral and query-neutral; it defines runtime policy and output seams only.

## In Scope

- global state classification as `TRADEABLE`, `WATCH`, or `NO-OPERATE`
- ceiling application rules across `BTC-USD` and `ETH-USD`
- shared downgrade triggers versus symbol-local degradation
- global recovery hysteresis and anti-flap handling
- explicit global reason outputs and transition provenance
- deterministic config for global thresholds and transition persistence

## Out Of Scope

- adding new symbol-level feature math or recomputing bucket metrics
- query contracts or history/audit retrieval design
- asset-universe expansion beyond `BTC-USD` and `ETH-USD`
- any consumer-specific business logic for alerts, risk sizing, or UI wording

## Target Structure

- `libs/go/features/regime.go` extensions for global evaluators, or `libs/go/features/regime_global.go` if implementers split the helper cleanly
- `libs/go/features/regime_test.go` or `libs/go/features/regime_global_test.go`
- `services/regime-engine/service.go` global evaluation and state capping flow
- `services/regime-engine/service_test.go` additions for cross-symbol transitions
- `configs/*/regime-engine.market-state.v1.json`
- shared fixture sequences under `tests/fixtures/world_usa_regime/`

## Global Ceiling Inputs

- current symbol regime snapshots for `BTC-USD` and `ETH-USD`
- symbol reason families and trigger metrics
- bucket-derived shared trust markers such as severe fragmentation, widespread timestamp-trust loss, or simultaneous coverage collapse
- config-driven persistence counters for global downgrade and recovery

## Global State Semantics

- `TRADEABLE`: no shared cross-symbol trust failure is active and no global watch/no-operate ceiling is required.
- `WATCH`: shared conditions reduce confidence across the market, but a full hard-stop is not yet justified.
- `NO-OPERATE`: shared trust loss is severe enough that all symbols should be capped to informational-only posture.

Global state is a ceiling, not a replacement for symbol state:

- if global is `NO-OPERATE`, every symbol is capped to `NO-OPERATE`
- if global is `WATCH`, no symbol may exceed `WATCH`
- if global is `TRADEABLE`, symbols retain their own emitted state

## Shared Downgrade Rules

Recommended conservative triggers for a global downgrade:

- both symbols are `NO-OPERATE` because of shared venue-group or trust-fabric failures
- both symbols show severe fragmentation or unavailable-side conditions in the same or consecutive closed 5m windows
- timestamp-trust loss or coverage collapse affects the shared market-quality picture broadly enough that symbol-local separation is no longer trustworthy

Recommended non-global cases to preserve:

- one symbol degrades because of symbol-specific fragmentation while the other remains healthy
- one symbol is `WATCH` for incomplete recovery while the other remains `TRADEABLE`
- localized data-quality issues that do not affect the shared WORLD/USA interpretation across both symbols

## Recovery And Anti-Flap Rules

- Global downgrade may happen immediately on a qualifying severe shared failure.
- Global recovery should require more persistence than a single clean window after a shared failure.
- Recovery should step conservatively: `NO-OPERATE -> WATCH -> TRADEABLE` unless config explicitly allows direct recovery and the evidence is still deterministic.
- When symbol states improve before the global ceiling does, keep the ceiling active until the required shared recovery windows are satisfied.

## Explicit Output Shape

Each global output should include:

- `state`
- `reasons`: ordered shared reason codes
- `primaryReason`
- `affectedSymbols`
- `appliedCeilingToSymbols`
- `previousState`
- `transitionKind`
- `effectiveBucketEnd`
- `configVersion`, `algorithmVersion`, and schema/version seam

The service may also emit derived symbol-effective state after ceiling application, but this plan treats that as a runtime seam, not a query contract.

## Fixture Guidance

- both symbols clean and aligned
- both symbols degraded by shared fragmentation across the same window
- one symbol healthy and one symbol degraded for symbol-local reasons only
- shared failure followed by partial but insufficient recovery
- threshold-edge sequence that would oscillate without global hysteresis

## Unit Test Expectations

- `TestGlobalCeilingRules` for `TRADEABLE`, `WATCH`, and `NO-OPERATE` ceilings
- `TestGlobalCeilingDoesNotHideSymbolSpecificDifferences` for one-healthy one-degraded cases
- `TestGlobalRecoveryRequiresPersistence` for staged recovery
- `TestGlobalTransitionReasonsAreDeterministic` for stable reason ordering and edge handling

## Integration And Replay Expectations

- integration tests should prove the regime-engine applies the global ceiling after symbol classification, not before
- integration tests should cover both-symbol shared failure, one-symbol-local failure, and shared recovery sequences
- replay tests should compare full global outputs, applied ceilings, and reason ordering across repeated identical runs
- replay tests should confirm config-version changes alter ceiling behavior only when the pinned config changes

## Downstream Seams Only

- Later `market-state-current-query-contracts` may expose the global state and ceiling-applied symbol view.
- Later history/audit work may retrieve emitted global transitions with their reason codes and provenance.
- This module must not design those downstream envelopes now.

## Summary

This module owns the market-wide ceiling that turns symbol decisions into one authoritative cross-symbol trust cap. It keeps global downgrades fast, recoveries conservative, and shared-failure reasoning explicit without dragging query or storage concerns into scope.
