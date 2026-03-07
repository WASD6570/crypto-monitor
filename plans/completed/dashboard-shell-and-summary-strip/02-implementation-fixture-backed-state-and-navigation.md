# Implementation Module 2: Fixture-Backed State And Navigation

## Scope

Implement the feature-local fixture source, symbol-focus URL state, and mobile section navigation that drive the shell without depending on unstable backend query surfaces.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/fixtures`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-shell/hooks`
- `apps/web/src/hooks`
- `apps/web/src/state`
- `apps/web/src/pages/dashboard`

## Module Requirements

- Define a small shell-facing fixture model that represents service-shaped dashboard metadata and both symbol summaries.
- Support at least these fixture scenarios:
  - healthy default snapshot
  - degraded feed / weakened trust snapshot
  - stale snapshot with last-known-good content
  - partial snapshot where one lower section is unavailable
- Drive default focus from fixture-backed route state with `BTC-USD` as the safe default.
- Persist symbol focus and mobile section selection in query params so reloads and copied URLs are stable.
- Reject or normalize unknown query values without crashing or showing empty screens.
- Keep state management feature-local unless a tiny shared query-param helper is clearly justified.
- Preserve the seam where later query adapters can replace fixture loading while reusing the same shell-facing view model.

## Implementation Decisions To Lock

- Keep fixture definitions close to the dashboard shell feature rather than in a global mock layer.
- Separate raw fixture objects from the shell view model so later adapter code can target the same view-model contract.
- Treat trust state as input, not derived state; the mapper may only translate fixture labels into display-friendly props.
- Use URL state for user intent (`symbol`, `section`), not for duplicating fixture payload data.
- Keep the shell usable when one section fixture is `unavailable`; do not block the whole route unless the top-level snapshot is unavailable.

## Recommended Data Shape

At minimum the shell fixture model should include:

- top-level dashboard metadata:
  - `asOf`
  - `oldestAgeSeconds` or equivalent precomputed age input
  - `configVersion`
  - top-level trust state
  - top-level degraded note list
- symbol summaries for `BTC-USD` and `ETH-USD`:
  - service-supplied state label
  - reason list
  - freshness label
  - last-updated timestamp
  - WORLD vs USA summary text
  - optional timestamp-degraded note
- per-slot placeholder states for `overview`, `microstructure`, `derivatives`, and `health`

## Navigation Rules

### Symbol Focus

- query param source of truth: `symbol`
- accepted values: `BTC-USD`, `ETH-USD`
- default: `BTC-USD`
- switching symbols updates the lower focused header and preserves the current global rail state

### Mobile Section Selection

- query param source of truth: `section`
- accepted values: `overview`, `microstructure`, `derivatives`, `health`
- default: `overview`
- invalid values fall back silently to the default while preserving the route

## Fixture Strategy

- Keep fixture names explicit and operator-meaningful, such as `healthyDashboardFixture` and `degradedUsaConfirmationFixture`.
- Include at least one fixture where timestamp fallback or stale age is part of the trust note, because the shell must surface those states before live adapters exist.
- Keep fixtures deterministic and static; no random timestamps or wall-clock generation inside tests.

## Unit Test Expectations

- query-param parsing defaults to `BTC-USD` and `overview` for missing or invalid values
- symbol switching updates the focused-shell state while keeping both summary cards visible
- degraded and stale fixtures produce visible shell trust notes without changing service-supplied state labels
- partial fixture keeps the shell usable and marks only the unavailable section slot as unavailable

## Acceptance Criteria

- Another agent can implement the fixture-backed shell state without inventing a wider app state architecture.
- The plan defines the fixture scenarios needed to unblock shell work before live query stabilization.
- Navigation is stable, URL-driven, and bounded to symbol focus plus mobile section recall.
- The view-model seam is ready for the later `dashboard-query-adapters-and-trust-state` child feature.

## Summary

This module keeps the shell moving without backend churn: deterministic fixtures provide service-shaped truth, query params hold user intent, and the shell-facing view model becomes the future handoff point from fixtures to real query adapters.
