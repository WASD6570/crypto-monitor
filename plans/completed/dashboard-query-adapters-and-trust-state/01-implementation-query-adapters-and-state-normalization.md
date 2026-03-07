# Implementation: Query Adapters And State Normalization

## Module Goal

Create the read-only `apps/web` data layer that talks to the service-owned dashboard current-state surfaces, validates required fields, and reduces payloads into a small normalized state model the shell can trust.

## Target Repo Areas

- `apps/web/src/api`
- `apps/web/src/features/dashboard-state`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/pages/dashboard`
- `apps/web/src/hooks` only if a reusable polling or page-visibility helper is justified

## Module Scope

- transport wrappers for dashboard snapshot and related current-state reads
- response guards/decoders for required fields and completeness metadata
- feature-local cache and refresh bookkeeping
- normalized state and view-model mapping used by the dashboard page
- adapter test fixtures and mocked-response helpers

## Out Of Scope

- shell copy polish beyond required trust messaging
- panel-specific layout changes beyond consuming the normalized state
- historical, analytics, or audit retrieval
- service contract changes or schema ownership

## Requirements

- Treat `plans/completed/market-state-current-query-contracts/` as the source of expected logical surfaces and response semantics.
- Prefer a feature-local folder such as `apps/web/src/features/dashboard-state/` over spreading behavior across `src/state` unless a reusable pattern clearly emerges.
- Keep adapters tolerant of partial-surface failure while still rejecting malformed critical fields.
- Preserve service timestamps, `configVersion`, `algorithmVersion`, completeness indicators, and degraded reason lists when present.
- Ensure BTC/ETH detail reads can be cached briefly per symbol so switching does not blank the shell unnecessarily.
- Keep stale timers deterministic and testable; avoid tying logic directly to uncontrolled wall-clock reads where a clock helper or injected `now` can keep tests stable.

## Recommended Folder Shape

- `apps/web/src/api/dashboard/`
  - `dashboardClient.ts`: low-level read-only fetchers per logical surface
  - `dashboardContracts.ts`: TypeScript interfaces for the consumed service payload subset
  - `dashboardDecoders.ts`: guards/parsers for critical fields and completeness markers
- `apps/web/src/features/dashboard-state/`
  - `dashboardQueryState.ts`: normalized state types and trust enums
  - `dashboardStateMapper.ts`: service payload -> shell-facing state reduction
  - `useDashboardData.ts`: page-level orchestration hook for load/refresh/cache behavior
  - `dashboardStateFixtures.ts`: mocked adapter payloads for tests only if the existing shell fixtures are not sufficient
- `apps/web/src/features/dashboard-shell/model/`
  - extend or replace shell model types only where needed so `DashboardShell` receives normalized state rather than hard-coded fixture data

## Adapter Boundary Decisions

### Logical Surfaces To Support

1. dashboard snapshot for initial route load
2. focused symbol detail surface
3. derivatives context surface
4. feed health and regime surface

Implementation can collapse some surfaces behind one client call if upstream does, but the internal adapter API should keep these logical seams visible so panel-level fallbacks stay honest.

### Decoder Rules

- required identity fields: symbol, state label, timestamps, and any required completeness/trust metadata
- optional blocks: derivatives, health, and explanatory notes may be absent if the payload marks them unavailable
- malformed critical payloads should resolve to `unavailable` for the affected surface and log or expose a structured error path for tests
- decoders should never synthesize missing state labels, timestamps, or regime reasons

### Normalized State Shape

The normalized layer should expose:

- global dashboard trust state and rail metadata
- per-symbol summary state for BTC and ETH
- per-section trust state for `overview`, `microstructure`, `derivatives`, and `health`
- explicit completeness markers and missing-surface notes
- per-symbol cache entries with last-success timestamps and in-flight status

Prefer a discriminated union or small typed object family over loosely optional blobs so stale and unavailable transitions are testable.

## Refresh And Cache Behavior

- On page entry, request the snapshot first and render the shell as soon as the rail and summary strip can be populated safely.
- Start dependent detail reads after snapshot success, or in parallel if upstream serves them independently and the shell can still preserve partial completeness.
- When the user switches symbols, reuse fresh cached detail state immediately and trigger background refresh if the cache is near staleness.
- If refresh fails after a last-known-good payload exists, move the affected surface to `stale` before eventually escalating to `unavailable` once the severe stale threshold is crossed.
- Keep retry triggers bounded to the current route session; do not introduce app-wide remote-state machinery for this slice.

## Unit Test Expectations

- decoder accepts healthy and degraded payloads with service-owned timestamp fallback notes
- decoder rejects malformed critical fields without inventing defaults
- mapper turns partial snapshot completeness into visible panel or rail warnings
- stale propagation preserves last-known-good data for the bounded grace window
- symbol cache reuse keeps BTC/ETH switching populated while a refresh is in flight
- unavailable detail/derivatives/health surfaces do not wipe unaffected summary data

## Suggested Validation Commands

```bash
pnpm --dir apps/web test -- --runInBand
pnpm --dir apps/web build
```

## Summary

This module defines the dashboard client data seam in `apps/web`: read-only adapters consume service-owned current-state payloads, decoders preserve trust metadata, and a small feature-local normalization layer produces deterministic shell-facing state without re-deriving market logic.
