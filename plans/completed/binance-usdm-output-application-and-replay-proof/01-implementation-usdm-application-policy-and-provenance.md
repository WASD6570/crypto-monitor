# Implementation Module 1: USD-M Application Policy And Provenance

## Scope

- Define the conservative application rule for the settled USD-M influence signal and the smallest additive provenance summary needed on current-state symbol responses.
- Cover shared feature/regime types, helper logic, and any current-state schema updates needed for the additive symbol-level metadata.
- Exclude live runtime ownership and command wiring.

## Target Repo Areas

- `libs/go/features`
- `schemas/json/features`
- `services/feature-engine`

## Requirements

- Add one explicit rule for mapping `USDMInfluenceSignalSet` into current-state/regime behavior.
- Only `DEGRADE_CAP` may mutate the result, and it may cap output to `WATCH` only.
- `AUXILIARY`, `NO_CONTEXT`, and `DEGRADED_CONTEXT` must leave the existing spot-derived symbol/global outcome unchanged.
- Add one optional symbol-level provenance summary that explains whether USD-M influence was evaluated and whether it changed the output.
- Keep the additive metadata small, deterministic, and backward-compatible.

## Key Decisions

- Add a dedicated regime reason for the applied cap instead of overloading existing spot-derived reason codes.
- Keep the full internal signal schema in `libs/go/features/usdm_influence.go`; expose a smaller public summary in current-state provenance rather than leaking every trigger metric into the route.
- Prefer application helpers that return adjusted symbol/global snapshots plus provenance so the live provider and deterministic fixtures can share one path.
- Update the current-state schemas only where the new optional provenance summary actually appears.

## Unit Test Expectations

- Auxiliary, no-context, and degraded-context signals preserve the existing spot-derived regime output.
- Degrade-cap signals lower a tradeable symbol to `WATCH` and attach the planned USD-M reason.
- Repeated application of the same signal to the same input is deterministic.
- Current-state JSON/schema tests accept the additive provenance summary without changing required fields for older clients.

## Contract / Fixture / Replay Impacts

- `market-state-current-symbol` and `market-state-current-response` schemas likely need one additive provenance field.
- `market-state-current-global` should stay structurally stable unless implementation proves one tiny additive global hint is required.
- Replay fixtures remain pinned to Go-owned current-state/regime outputs; no Python parity work is introduced.

## Summary

This module settles the exact public semantics of the first USD-M consumer-facing change: a conservative watch-cap plus explicit provenance, without turning derivatives context into an unbounded new scoring layer.
