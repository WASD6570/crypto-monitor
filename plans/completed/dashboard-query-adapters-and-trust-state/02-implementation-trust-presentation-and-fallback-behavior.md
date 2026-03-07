# Implementation: Trust Presentation And Fallback Behavior

## Module Goal

Wire the adapter-backed normalized state into the existing dashboard route and shell so trust, staleness, degradation, and partial availability remain visible without expanding the dashboard into later detail/history work.

## Target Repo Areas

- `apps/web/src/pages/dashboard/DashboardPage.tsx`
- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-shell/hooks`
- `apps/web/src/app/App.tsx` only if route-level retry or not-found handling needs a narrow seam
- `apps/web/tests/e2e`

## Module Scope

- replace fixture-as-default page wiring with normalized adapter-backed state
- keep fixture or mocked injection seams for tests
- present global and per-section trust states in the shell
- map partial/unavailable surfaces into existing shell placeholders and notes
- add retry/fallback behavior appropriate for read-only current-state monitoring

## Out Of Scope

- richer detail cards, tables, charts, or symbol drill-down panels
- mobile navigation redesign beyond preserving visible trust signals in the current shell
- separate accessibility polish tasks beyond maintaining clear text labels and stable navigation
- history and analytics entry points

## Requirements

- Preserve the current `/dashboard` route shape, symbol query params, and section query params from the completed shell plan.
- Keep `DashboardShell` presentational: it receives normalized state, user interactions, and trust-note props, but does not own fetching or market interpretation.
- Surface trust at both global and local scope: status rail, summary cards, and section slots should all reflect the normalized state where relevant.
- Make fallback behavior explicit: stale content stays visible with age/trust notes; unavailable sections explain what is missing; partial snapshot states name the missing surface or symbol.
- Avoid full-page spinner lock after first successful render; later failures should degrade only the affected areas unless the snapshot surface itself becomes unavailable.
- Reserve later seams for detailed panel implementations by keeping slot props shallow and state vocabulary stable.

## Consumer Wiring Plan

### `DashboardPage`

- Replace direct import of `healthyDashboardFixture` as the runtime default with a dashboard data hook such as `useDashboardData()`.
- Keep an override prop for tests so existing unit and browser harnesses can still inject deterministic data when needed.
- Centralize retry actions and focused symbol changes here so the shell remains easy to test.

### `DashboardShell`

- Update props from `fixture` toward a shell-facing normalized data object, or introduce an adapter prop layer that can temporarily support both until tests migrate.
- Keep current summary strip and section slot layout unchanged.
- Show compact trust notes for:
  - global degraded reasons
  - timestamp fallback or degraded-source markers
  - partial completeness warnings
  - section-level unavailable reasons

### Fallback Mapping Rules

- snapshot unavailable before first success -> full-page unavailable state with retry and last-success timestamp if cached
- derivatives unavailable with overview ready -> derivatives section shows `unavailable`; overview and summary remain live
- feed health unavailable -> keep summary visible but elevate a dashboard trust warning because interpretability is impaired
- symbol detail missing for one symbol -> inactive symbol summary can still render if provided; focused detail slot for that symbol becomes `unavailable`
- stale snapshot after first success -> preserve last-known-good rail and summary data with explicit age and reduced-trust copy

## Presentation Guardrails

- Reuse the bounded vocabulary `loading`, `ready`, `stale`, `degraded`, `unavailable` everywhere; do not invent extra intermediate UI-only states.
- Keep text direct and operator-facing, for example `Coinbase stale; USA confirmation weakened`, rather than abstract transport error copy.
- Never show a healthy-looking panel when service completeness or health metadata says trust is reduced.
- Keep symbol switching instant-feeling, but do not hide stale or degraded notes during the switch.

## Test Expectations

- dashboard page renders adapter-backed data through the existing shell without losing both summary cards
- partial responses keep unaffected sections visible and explicitly mark missing ones
- stale refresh failures preserve prior data and show age/trust reduction
- global trust notes remain visible after switching between BTC and ETH
- retry path restores a previously unavailable snapshot when mocked responses recover

## Suggested Validation Commands

```bash
pnpm --dir apps/web test -- --runInBand
pnpm --dir apps/web exec playwright test tests/e2e/dashboard-query-adapters-and-trust-state.spec.ts --project=chromium
```

## Summary

This module keeps the existing dashboard shell intact while swapping in normalized adapter-backed state. The page owns fetching and retry behavior, the shell stays presentational, and every fallback path stays explicit so later detail and negative-state work can build on one stable trust vocabulary.
