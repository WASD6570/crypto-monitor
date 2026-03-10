# Implementation: Operator Docs

## Requirements And Scope

- Update operator-facing docs to reflect that the command and Compose path now use the live Spot-backed provider rather than deterministic bundles.
- Document the first-cutover limits clearly: Spot-driven current state, explicit `usa` unavailability, optional slow context, and machine-readable degradation.
- Keep docs scoped to the current service boundary and local workflow; do not write broad rollout or incident-response runbooks in this slice.

## Target Repo Areas

- `services/market-state-api/README.md`
- `README.md`

## Implementation Notes

- Replace deterministic wording with live-backed wording, but keep the contract/stability message intact.
- Call out operator-visible startup behavior:
  - `/healthz` reflects process readiness only
  - symbol/global payloads may be partial or unavailable during warm-up
  - `usa` remains explicit rather than synthesized
  - degradation reasons remain in payloads
- Document any new command env var or config-path expectation introduced by runtime wiring.
- Keep local startup instructions rooted in `docker compose up --build` and the existing dashboard URL.

## Testing Expectations

- Doc updates align with the implemented command behavior and Compose validation results.
- The documented local commands and endpoints match the testing matrix in `04-testing.md`.

## Summary

This module updates the operator story so the repo no longer promises deterministic local state where the command now serves live Spot-backed current state. The docs should help later agents and operators understand warm-up limits, explicit unavailable sections, and where to validate the boundary.
