# Implementation Step 2: Unit And Component Smoke Matrix

## Requirements And Scope

- Expand mapper and shell coverage around the shared scenario catalog.
- Prove the dashboard stays honest and readable across the supported route states without broadening into exhaustive UI permutation testing.

## Target Repo Areas

- `apps/web/src/features/dashboard-state`
- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/pages/dashboard`

## Implementation Notes

- Extend mapper coverage so each named scenario proves the expected route warning, summary warning, panel fallback, and overall trust state.
- Extend shell or page tests so both symbol cards stay visible, the focused symbol and active section remain semantically exposed, and warning text stays readable in degraded/stale/partial/unavailable paths.
- Keep assertions centered on trust honesty, selection state, and unaffected-surface stability.
- Do not duplicate browser-only assertions here; component tests should stay fast and deterministic.

## Unit Test Expectations

- Cover healthy baseline readability with no unnecessary warning escalation.
- Cover degraded timestamp-trust reduction with route and focused-symbol warning visibility.
- Cover stale last-known-good behavior after refresh failure.
- Cover partial input loss and unavailable panel/surface fallback without blanking safe neighboring content.
- Cover route-backed selection semantics for symbol and section controls where component tests can assert them directly.

## Summary

This step locks in the dashboard’s service-trusting warning semantics at the fast test layer before browser smoke adds viewport and reload proof.
