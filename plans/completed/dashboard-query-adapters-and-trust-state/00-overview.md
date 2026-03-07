# Dashboard Query Adapters And Trust State

## Ordered Implementation Plan

1. Add read-only dashboard query adapters and response decoding seams in `apps/web` for the service-owned current-state surfaces.
2. Normalize service trust metadata into one bounded client state model that preserves `loading`, `ready`, `stale`, `degraded`, and `unavailable` without recomputing market logic.
3. Wire the dashboard shell and page consumers to the new adapter-backed state while keeping later detail/history work out of scope.
4. Validate mocked-response decoding, stale/degraded propagation, and build/browser smoke coverage; record results in `plans/completed/dashboard-query-adapters-and-trust-state/testing-report.md`.

## Problem Statement

`plans/completed/dashboard-shell-and-summary-strip/` established the route, summary strip, and shell placeholders, but the dashboard still reads from hard-coded fixtures. The next bounded slice is the client trust boundary: `apps/web` needs read-only query adapters and a small normalization layer that can consume service-owned current-state payloads, preserve freshness and degradation honesty, and hand later panel work a stable consumer seam.

## Bounded Scope

- read-only dashboard query adapters in `apps/web`
- response decoding for the service-owned current-state surfaces defined upstream
- feature-local normalization from service payloads into panel/shell trust state
- stale, degraded, partial, and unavailable presentation mapping for existing shell consumers
- cache and refresh seams needed for instant-feeling BTC/ETH switching inside the current shell
- mocked-response and fixture-backed validation for adapter behavior

## Out Of Scope

- detail-panel content expansion beyond the existing reserved shell slots
- negative-state polish, mobile accessibility follow-up, or copy refinement beyond required trust honesty
- analytics, history, audit, or drill-down UI
- client-side derivation of tradeability, divergence, fragmentation, market quality, or regime decisions
- backend route implementation, concrete schema authoring, or service contract redesign
- streaming transport adoption unless it can sit behind the same adapter interface as an optional later seam

## Requirements

- Keep the work bounded to `dashboard-query-adapters-and-trust-state`; later child plans own panel composition depth, negative-state polish, and integrated smoke expansion.
- Depend on service-owned surfaces from `plans/completed/market-state-current-query-contracts/` and preserve service authority for timestamps, completeness, degraded reasons, and provenance metadata.
- Keep `apps/web` a React + TypeScript + Vite SPA with no SSR, server components, or Python runtime dependency.
- Reuse the completed shell route and summary strip instead of reshaping the page.
- Support one fast dashboard snapshot read on entry, plus bounded symbol/detail surface refresh behavior that can keep BTC/ETH switching responsive.
- Preserve partial availability honestly: if one logical surface fails, unaffected shell areas stay readable and explicitly marked.
- Keep the client normalization layer presentational only: formatting, state reduction, cache freshness bookkeeping, and fallback labeling are allowed; market interpretation is not.

## Target Repo Areas

- `apps/web/src/api`
- `apps/web/src/features/dashboard-shell`
- `apps/web/src/features/dashboard-state` or similarly named feature-local query-state folder
- `apps/web/src/hooks` only for narrow reusable polling/visibility helpers if they are truly shared
- `apps/web/src/state` only if a tiny app-local cache store is simpler than feature-local state
- `apps/web/src/pages/dashboard`
- `apps/web/tests/e2e`

## Module Breakdown

### 1. Query Adapters And State Normalization

- Own the read-only transport wrappers, decoder/guard seams, normalized dashboard state model, cache behavior, and symbol/detail loading orchestration.
- Keep payload interpretation limited to service-supplied shape validation, trust metadata propagation, and panel-state reduction.

### 2. Trust Presentation And Fallback Behavior

- Own the shell-facing consumer seams, trust-note mapping, partial-surface fallbacks, and retry/unavailable behavior inside the existing dashboard page and shell components.
- Keep later detail rendering, expanded mobile behavior, and history exploration explicitly deferred.

## Dependency Notes

- This plan assumes `plans/completed/market-state-current-query-contracts/` defines the service-owned snapshot and current-state consumer surfaces before implementation starts.
- Completed context from `plans/completed/world-usa-composite-snapshots/`, `plans/completed/market-quality-and-divergence-buckets/`, and `plans/completed/symbol-and-global-regime-state/` matters only through service payload semantics; `apps/web` must not reopen those calculations.
- The existing shell from `plans/completed/dashboard-shell-and-summary-strip/` remains the consumer entry point and should keep its route shape, symbol query params, and section seams.

## Design Overview

### Client Boundary

- `apps/web` should consume one logical dashboard snapshot first and then optional focused-symbol surfaces for detail, derivatives, and health when those are available.
- Adapter modules should hide transport choice from the page layer so later polling or push updates can reuse the same normalized outputs.
- The shell-facing state model should be small and explicit enough that tests can assert panel trust transitions deterministically.

### Normalized Trust Model

- `loading`: no safe payload yet for the surface
- `ready`: current payload is complete enough and not trust-reduced
- `stale`: last-known-good payload is retained past refresh target but still within a temporary operator-safe window
- `degraded`: payload is current enough to show, but service metadata reduces confidence
- `unavailable`: required payload is missing, critically incomplete, or too stale to present as current

### Consumer Seam Rules

- `DashboardPage` should stop importing concrete fixtures as the default source once adapter-backed state exists; tests may still inject fixtures or mocked adapter results.
- `DashboardShell` should accept normalized view data plus trust/fallback metadata, not raw service payloads.
- Later detail-panel work should mount into the same shell slots and read the same normalized trust vocabulary instead of inventing panel-specific states.

### Caching And Refresh Defaults

- initial page load favors one snapshot request that can populate the rail and summary strip immediately
- focused symbol detail and derivatives/health surfaces may refresh independently after the initial shell render
- BTC/ETH switching should reuse short-lived cached detail payloads when still fresh enough, then reconcile with the next service response
- cache expiry and severe-stale cutoffs should follow service freshness metadata where present and otherwise default conservatively

## Acceptance Criteria

- Another agent can implement the feature without reopening the parent dashboard epic.
- Repo areas under `apps/web` are explicit for adapters, normalized state, shell consumers, and tests.
- The plan clearly separates read-only query/state work from later detail/history and negative-state polish.
- Validation commands are deterministic and target mocked-response decoding, trust-state propagation, and browser smoke behavior.

## ASCII Flow

```text
service-owned current-state surfaces
  - dashboard snapshot
  - symbol detail
  - derivatives context
  - feed health / regime
                |
                v
apps/web adapter layer
  - fetch or subscribe
  - decode required fields
  - preserve completeness and timestamps
  - expose typed success/failure results
                |
                v
feature-local normalization layer
  - panel trust state
  - stale/degraded/unavailable mapping
  - short-lived symbol cache
  - shell-facing view model
                |
                v
existing dashboard page + shell
  - status rail
  - BTC/ETH summary strip
  - reserved overview/microstructure/derivatives/health slots
                |
                v
later child features fill panel content
without changing the trust boundary
```

## Live-Path Boundary

- Services remain the source of truth for state labels, reason families, timestamps, completeness, degraded markers, and provenance.
- This feature stops in `apps/web` at read-only consumption, normalization, caching, and presentation mapping.
- Any need to change response fields, rollout sequencing, or service query assembly belongs back in `plans/completed/market-state-current-query-contracts/` or later backend work, not in this plan.
