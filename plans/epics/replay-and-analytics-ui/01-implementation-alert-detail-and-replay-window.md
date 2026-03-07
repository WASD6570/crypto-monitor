# Alert Detail And Replay Window

## Module Requirements

- Target repo area: `apps/web`
- Primary responsibility: render one recent alert review surface that answers the why-fired question and exposes bounded replay evidence without client-side recomputation
- Required inputs: service-owned alert detail payload, outcome summary payload, replay-coverage metadata, optional delivery summary, optional feedback summary, optional simulation-presence metadata
- Validation focus: fixture-backed rendering of complete, partial, stale, and missing replay states

## Scope

- recent alert stream or list with compact rows/cards for severity, symbol, setup family, direction, trigger time, market-state ceiling, outcome status, freshness, and degraded markers
- alert detail header that prioritizes the explanation bundle: setup, severity, symbol, alert time, config/version provenance, market-state gate, risk-state gate placeholder, degraded flags, and reason codes
- outcome summary block that renders service-owned decisive result, horizon summaries, MAE/MFE style headline metrics, and timing fields
- replay window panel that shows a bounded pre-alert and post-alert evidence window, exact coverage boundaries, timestamp basis, and partial coverage notes
- delivery, feedback, and simulation presence summaries as secondary context panels

## Out Of Scope

- general historical chart exploration beyond the bounded review window
- recomputing thresholds, setup explanations, regime decisions, or outcome ordering in the browser
- editing operator feedback workflows beyond rendering existing summary status and links/seams
- implementing live streaming transport; polling vs stream selection may stay abstract so long as freshness semantics are explicit

## Reading Order

1. alert identity and current review status
2. why-fired explanation bundle
3. outcome summary
4. replay coverage status and open replay window action
5. secondary context: delivery, feedback, simulation presence, raw identifiers

## Query And Data Expectations

- Use one service-owned alert detail response per selected alert whenever practical so the client does not orchestrate business joins with inconsistent timestamps.
- The detail response should already include:
  - stable alert identifier and version metadata
  - setup family, direction, severity, symbol, event time, processing time, and selected horizon evidence
  - market-state decision and future-safe `riskStateDecision` field when available
  - outcome summary fields fit for UI review rather than raw event streams only
  - replay coverage metadata: available range, missing segments, degraded segments, and source timestamp basis
  - simulation summary presence fields or a stable pointer to fetch them separately
- If replay evidence is heavy, fetch it lazily when the replay panel opens; coverage metadata must still be present in the main detail payload.

## UI Structure

### Alert Stream

- Default to dense rows on desktop and stacked cards on mobile.
- Show only the smallest operator triage fields first.
- Preserve stable row height where possible so the list remains scan-friendly during refresh.

### Alert Detail Header

- Lead with why the user should trust or distrust the alert: severity, setup, symbol, market-state cap, degraded markers, and freshness.
- Show machine-readable reason codes alongside concise human labels supplied or approved by services, not invented in the client.
- Keep provenance visible: `configVersion`, `algorithmVersion`, timestamps, and source-of-truth markers.

### Outcome Summary

- Keep market-truth outcome ahead of simulation-truth context.
- Show horizon summaries in a fixed order: `30s`, `2m`, `5m`.
- If outcome is pending or unavailable, show explicit status text rather than placeholders that imply success or failure.

### Replay Window

- Bound the replay review to one alert-centered window with pre-alert, at-alert, and post-alert segments.
- Prefer event strips, summary tables, or compact markers over heavyweight charting by default.
- Allow an expandable evidence table only on demand.
- Mark all missing or degraded spans directly in the replay timeline or table; do not smooth or interpolate them visually.

## Safe Defaults And Edge Cases

- Empty alert stream: show review shell, last refresh time, and no-alert copy.
- Missing outcome: keep alert explanation visible and mark outcome `Pending` or `Unavailable` from the payload.
- Missing replay records: show `Replay unavailable` with the stated reason and preserve the rest of the page.
- Partial replay coverage: show exact known span plus `coverage incomplete` badge and reason text.
- Stale detail payload: keep data visible but freeze refresh-relative labels to the service timestamp and display age prominently.
- Unknown future `riskStateDecision`: render `Not yet applied` or equivalent neutral status instead of assuming approval.

## Unit And Component Test Expectations

- alert list renders mixed severities, degraded markers, and stale badges from fixture data
- alert detail preserves section order and hides optional panels until data or user action requires them
- replay panel renders full, partial, and unavailable coverage without crashing or fabricating segments
- outcome summary keeps service ordering and labels; client formatting does not alter decisive status
- mobile rendering keeps primary fields visible without horizontal scroll for core summary content

## Summary

This module defines the review path for one alert. It keeps the why-fired explanation, outcome summary, and replay coverage in a single service-driven detail view, with bounded replay evidence and conservative defaults when records are stale, partial, or missing.
