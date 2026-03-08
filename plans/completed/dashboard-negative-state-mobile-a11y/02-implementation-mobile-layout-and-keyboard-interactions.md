# Implementation Step 2: Mobile Layout And Keyboard Interactions

## Requirements And Scope

- Keep the existing dashboard route usable on smaller screens and through keyboard-only navigation.
- Strengthen, rather than replace, the current symbol and section route-state controls.
- Ensure negative-state warnings remain visible in the mobile reading order.

## Target Repo Areas

- `apps/web/src/features/dashboard-shell/components`
- `apps/web/src/pages/dashboard`
- `apps/web/src/styles.css`

## Implementation Notes

- Review the current summary-card stack and section button layout at mobile widths to ensure:
  - tap targets remain at least 44px
  - focused symbol remains obvious
  - active section remains obvious
  - warning text is still visible before deep scrolling
- Add keyboard-visible focus treatment where needed for symbol and section buttons.
- Consider semantic improvements such as `aria-current`, `aria-describedby`, or related labeling only where they clarify the current route state.
- Preserve the current URL-driven model: changing symbol or section should still update `symbol` and `section` query params without introducing local-only navigation state.
- Avoid heavy motion or layout shifts when switching symbols or sections, especially when negative states are active.

## Component And Interaction Test Expectations

- Extend component tests to verify:
  - keyboard symbol switching updates the focused heading while both summaries remain visible
  - section navigation exposes the active section semantically and visually
  - mobile-specific warning or summary content is not hidden behind hover-only or desktop-only layout assumptions

## Summary

This step makes the existing dashboard navigation resilient on mobile and keyboard paths without changing the established page structure or trust boundary.
