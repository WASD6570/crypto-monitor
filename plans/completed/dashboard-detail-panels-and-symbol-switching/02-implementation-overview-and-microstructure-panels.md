# Implementation Step 2: Overview And Microstructure Panels

## Requirements And Scope

- Replace the overview and microstructure placeholders in the focused-symbol region with dense, readable panel components.
- Preserve the existing route shape, summary strip, and section navigation; this step fills the shell rather than redesigning it.
- Keep panel copy compact and operator-oriented so the dashboard remains scannable on desktop and mobile.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/features/dashboard-shell/model`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`

## Implementation Notes

- Introduce focused panel components under `apps/web/src/features/dashboard-shell/components` for overview and microstructure content.
- Prefer a small shared panel frame component only if it reduces duplication without hiding semantic differences between panels.
- Overview panel should surface:
  - effective state
  - symbol state and global ceiling state
  - WORLD and USA composite availability or freshness callouts
  - service-owned reasons relevant to the focused symbol
- Microstructure panel should surface:
  - latest bucket family and close time
  - missing bucket counts
  - recent-context completeness for 30s, 2m, and 5m families
  - trust-reducing notes from the focused payload when present
- Keep visual treatment consistent with the existing dashboard shell: compact chips, metric rows, restrained emphasis, and clear active-state hierarchy.
- On desktop, maintain the two-column detail grid already established by the shell. On mobile, keep section-nav-driven reading order and avoid side-by-side metric compression that drops readability.

## Component Test Expectations

- Extend `apps/web/src/features/dashboard-shell/components/DashboardShell.test.tsx` or nearby component tests to verify:
  - overview panel renders focused symbol data without removing the summary strip
  - microstructure panel renders the expected bucket and completeness metrics
  - switching the focused symbol updates overview and microstructure content while keeping both summary cards visible
  - loading, degraded, and unavailable panel states keep trust chips and fallback copy visible

## Summary

This step turns the first two detail slots into real operator panels while preserving the stable route and shell structure created by the earlier dashboard features.
