# Dashboard Fixture Smoke Matrix

## Ordered Implementation Plan

1. Consolidate one deterministic dashboard scenario catalog and shared mock/helper seam so healthy, degraded, stale, partial, and unavailable current-state cases stay named consistently across component and browser tests.
2. Expand dashboard unit and component smoke coverage so each supported scenario proves route honesty, panel-level fallback behavior, and selection state without re-deriving service logic in the UI.
3. Expand desktop and mobile Playwright smoke coverage around the shared scenario matrix and record final results in `plans/completed/dashboard-fixture-smoke-matrix/testing-report.md`.

## Problem Statement

`plans/completed/dashboard-negative-state-mobile-a11y/` makes the dashboard honest and accessible for the current set of degraded/mobile cases, but the remaining dashboard-core gap is test organization and breadth. Current fixtures and mocked responses prove key paths, yet the matrix is still scattered across individual tests and helper snippets. The last bounded child slice should make the supported dashboard states explicit, deterministic, and reusable so later work can extend the route safely without losing confidence in trust warnings, route-backed selection, or panel availability behavior.

## Bounded Scope

- deterministic dashboard scenario catalog for healthy, degraded, stale, partial, and unavailable current-state reads
- shared helper seams for Vitest and Playwright dashboard response mocking
- bounded unit/component/browser smoke matrix covering route rail, summary cards, focused panels, and route-backed symbol/section state
- final dashboard-core testing report for the current matrix

## Out Of Scope

- backend, schema, replay, or service contract changes
- new dashboard information architecture, new panel types, or additional live data sources
- broad visual redesign or new accessibility behavior beyond proving the completed behavior is locked in
- end-user alerting, historical reads, or slow-context surfaces
- exhaustive combinatorial coverage of every theoretical dashboard permutation

## Requirements

- Stay inside `apps/web` and the active plan directory; keep the Vite SPA read-only and service-trusting.
- Reuse the completed seams from `plans/completed/dashboard-query-adapters-and-trust-state/`, `plans/completed/dashboard-detail-panels-and-symbol-switching/`, and `plans/completed/dashboard-negative-state-mobile-a11y/`.
- Keep one obvious source for scenario names and mocked-response setup so component and browser tests do not drift.
- Cover at least these route-relevant states: healthy baseline, degraded timestamp-trust reduction, stale last-known-good refresh failure, partial input loss, and unavailable surface fallback.
- Prove both desktop and mobile route behavior without exploding into a brittle Cartesian-product matrix.
- Keep service-owned trust vocabulary unchanged: `loading`, `ready`, `stale`, `degraded`, `unavailable`.
- Do not add client-side market-state, derivatives, or trust recomputation while building test fixtures.

## Target Repo Areas

- `apps/web/src/features/dashboard-state`
- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/test`
- `apps/web/tests/e2e`
- `apps/web/playwright.config.ts`

## Module Breakdown

### 1. Scenario Catalog And Shared Test Helpers

- Normalize one named dashboard scenario catalog for mocked current-state responses and any shell-facing fixture helpers that still matter.
- Add shared helper utilities so Vitest and Playwright both consume the same scenario naming and timestamp-freshening behavior.

### 2. Unit And Component Smoke Matrix

- Expand mapper and shell tests to iterate through the supported scenarios without duplicating ad hoc setup.
- Keep assertions high-signal: trust copy, warning hierarchy, panel fallback honesty, and route-backed selection state.

### 3. Browser Smoke Matrix And Reporting

- Extend the current dashboard browser spec around the shared scenarios for desktop and mobile.
- Record the final validation results in `plans/completed/dashboard-fixture-smoke-matrix/testing-report.md`.

## Design Details

### Scenario Strategy

- Prefer a small, named scenario list over open-ended inline fixture edits.
- Each scenario should describe what changed in service-owned payload terms, not UI-only inventions.
- Timestamp handling should stay deterministic and shared so stale/fresh checks do not drift between test layers.

### Coverage Strategy

- Unit tests should prove mapper reduction and shell presentation for each scenario class.
- Component tests should prove the user can still understand trust, selection, and fallback behavior.
- Browser tests should prove the composed route still works under desktop and mobile layouts with mocked current-state responses.

### Operator-Honesty Focus

- Healthy state should remain readable and not be over-warned.
- Degraded, stale, partial, and unavailable paths should each preserve visible text explaining what is reduced.
- Unavailable surfaces should stay explicit without blanking unaffected neighboring content.

## Acceptance Criteria

- Another agent can implement the final dashboard-core smoke-matrix work without reopening the parent epic.
- The plan identifies one shared scenario seam for both Vitest and Playwright coverage.
- The plan keeps scope bounded to current dashboard route behavior and avoids backend or contract work.
- The validation shape proves the final route-level matrix across desktop and mobile with deterministic scenario inputs.

## ASCII Flow

```text
named dashboard scenarios
  (healthy / degraded / stale / partial / unavailable)
                |
                v
shared test helpers
  - response cloning
  - timestamp freshening
  - route/mock setup
                |
     +----------+----------+
     |                     |
     v                     v
Vitest mapper/shell     Playwright route smoke
checks                  desktop + mobile
     |                     |
     +----------+----------+
                v
deterministic dashboard-core proof
for trust honesty and route behavior
```

## Live-Path Boundary

- Services remain the source of truth for dashboard trust, freshness, and availability semantics.
- This feature is test-only and frontend-only; it must not change live service behavior or introduce Python runtime dependencies.
