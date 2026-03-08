# Implementation Step 1: Scenario Catalog And Shared Test Helpers

## Requirements And Scope

- Create one obvious scenario catalog for dashboard test states so the same named cases can drive mapper, component, and browser smoke coverage.
- Reduce fixture and mock duplication without changing runtime dashboard behavior.
- Keep shared helper logic deterministic and limited to test setup.

## Target Repo Areas

- `apps/web/src/features/dashboard-state`
- `apps/web/src/test`
- `apps/web/tests/e2e`

## Implementation Notes

- Start from the existing `dashboardStateFixtures` response sets and current browser mocking helpers.
- Introduce a small named scenario layer such as healthy, degraded timestamp fallback, stale last-known-good, partial input loss, and unavailable surface fallback.
- Keep timestamp freshening and response cloning in one reusable helper seam instead of duplicating them across specs.
- Only keep shell-facing static fixtures where they still help component isolation; prefer shared service-shaped scenario data when practical.
- Avoid mixing production code and test-only helpers beyond minimal export seams that already fit `apps/web` conventions.

## Unit Test Expectations

- Add or update tests to verify the shared scenario helper returns deterministic inputs for repeated runs.
- Verify stale scenarios preserve last-known-good trust semantics instead of silently degrading to generic unavailable messaging.
- Verify partial and unavailable scenarios preserve service-owned labels and do not invent client-only fallback states.

## Summary

This step creates the reusable scenario seam the rest of the smoke matrix depends on, so later test expansion uses named, deterministic dashboard states instead of ad hoc fixture edits.
