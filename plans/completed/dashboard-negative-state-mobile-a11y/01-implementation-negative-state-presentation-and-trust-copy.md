# Implementation Step 1: Negative-State Presentation And Trust Copy

## Requirements And Scope

- Make stale, degraded, unavailable, and partial-data conditions render consistently across the status rail, summary cards, and focused panels.
- Keep the mapper authoritative for presentational state reduction so components stay simple and do not re-interpret service payloads.
- Preserve healthy-path readability while making trust reductions impossible to miss.

## Target Repo Areas

- `apps/web/src/features/dashboard-state`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/features/dashboard-shell/components`

## Implementation Notes

- Extend the shell-facing view model only as needed for warning hierarchy, such as:
  - route-level primary warning text
  - summary-card warning or fallback note
  - panel-level warning emphasis metadata
- Keep the trust vocabulary unchanged: `loading`, `ready`, `stale`, `degraded`, `unavailable`.
- Normalize partial-data cases into explicit copy and emphasis rather than leaving them to implicit color or chip state alone.
- Prefer repeated short warning text near the affected surface rather than one large generic warning far away from the issue.
- Preserve the existing dashboard visual direction; this step should harden honesty, not restyle the whole route.

## Unit Test Expectations

- Extend dashboard-state and shell tests to verify:
  - stale panels expose fallback wording consistently
  - degraded or partial summary cards keep warning text visible
  - unavailable panels keep honest messaging without blanking safe neighboring surfaces
  - trust cues are represented by text and not color alone

## Summary

This step defines one consistent negative-state presentation seam so later mobile and accessibility work builds on stable warning semantics instead of scattered per-component behavior.
