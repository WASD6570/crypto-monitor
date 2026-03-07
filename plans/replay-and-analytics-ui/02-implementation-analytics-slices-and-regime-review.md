# Analytics Slices And Regime Review

## Module Requirements

- Target repo area: `apps/web`
- Primary responsibility: present service-owned review aggregates that help the user understand where alerts work best, where to avoid trust, and how regime context changes outcomes
- Required inputs: service-owned aggregate slices for setup family, regime, severity, horizon, fragmentation, degraded-state cohorts, sample size, freshness, and optional saved-simulation summaries
- Validation focus: contract-driven rendering of aggregate cohorts, empty slices, stale aggregates, and missing simulation-summary seams

## Scope

- review slice controls for bounded windows such as recent 24h, 7d, or a service-approved comparable period
- regime-aware aggregate panels for `TRADEABLE`, `WATCH`, `NO-OPERATE`, fragmented vs non-fragmented, and degraded vs normal cohorts
- setup and horizon slices for quality and coverage review
- best-condition and avoid-condition summaries rendered from service-owned rankings or summary bundles
- simulation-summary seam that adds stored net-viability context to aggregates when available without changing the source outcome truth
- freshness, sample-size, and coverage messaging that keeps low-confidence aggregates from overclaiming significance

## Out Of Scope

- baseline comparison product work beyond reserving comparable keys and display seams
- client-side ranking, statistical significance claims, or homemade condition scoring
- ad hoc query builders that encourage slow, unconstrained analytics exploration
- simulation PnL derivation in the browser

## Analytics Questions To Answer

1. Which setups and horizons are producing better recent review outcomes?
2. Which regimes or degraded conditions are associated with weaker trust or poorer net viability?
3. Which condition bundles look best to pay attention to, and which should be treated cautiously or avoided?
4. How fresh and complete is the aggregate being shown right now?

## Data And Contract Expectations

- Aggregates should arrive pre-sliced or cheaply sliceable by service-provided dimensions rather than requiring the SPA to join raw alert records.
- Every aggregate payload should include:
  - window boundaries and generation timestamp
  - grouping keys such as setup family, regime, severity, horizon, fragmentation class, degraded class, and symbol scope when relevant
  - sample size and any minimum-sample warning markers
  - outcome-summary fields and optional simulation-summary fields kept distinct
  - freshness or lag markers when upstream records are delayed
- Best/avoid summaries should be delivered as service-owned summary bundles or ranked slices, with explanation text or codes derived server-side.

## UI Structure

### Slice Controls

- Keep filters few and high value: time window, symbol scope, setup family, and optionally severity or regime.
- Default to the smallest stable review window that yields meaningful recent context without slow queries.
- On mobile, use a single-column filter drawer or stacked controls.

### Regime Review Panels

- Show regime slices before generalized rankings so the user sees market-state context first.
- Keep `TRADEABLE`, `WATCH`, and `NO-OPERATE` visually comparable with aligned metrics and sample counts.
- Separate fragmented and degraded cohorts so the user can see whether poor performance clusters in low-trust conditions.

### Best And Avoid Summaries

- Render no more than a few top conditions and a few avoid conditions by default.
- Each item should include the condition label, sample size, freshness, and the specific summary fields supplied by the service.
- If sample size is below threshold, show the condition with a caution badge instead of hiding it silently.

### Simulation Summary Seam

- Keep simulation aggregate metrics visually separated from pure outcome aggregates.
- If simulation data is absent for the selected slice, show `No saved simulation summary for this view` and retain outcome analytics.
- When both exist, label them as `Outcome` and `Simulation` so the user does not confuse market truth with execution assumptions.

## Safe Defaults And Edge Cases

- Empty analytics window: show no-data messaging tied to current filters, not zeroed performance boxes.
- Stale aggregate: keep the last computed slice visible with clear age and lag note.
- Partial aggregate coverage: show a coverage note if some alerts in the time window are still pending outcomes or simulations.
- Low-sample best/avoid entry: display with caution state and sample count; never imply confidence from tiny cohorts.
- Missing simulation aggregate: render only the outcome-based panels with explicit absence text.

## Unit And Component Test Expectations

- aggregate tables/cards render aligned regime cohorts with sample sizes and freshness labels
- best/avoid summary panels preserve service ordering and caution states for low-sample rows
- outcome-only aggregates render correctly when simulation fields are absent
- stale and partial-coverage banners render independently from alert-detail freshness states
- filters update the request key without local recomputation of ranking or regime logic

## Summary

This module defines the review analytics layer for the Vite SPA. It keeps regime slices, best/avoid summaries, and simulation-summary seams firmly service-owned, with conservative handling for stale, sparse, or partially complete aggregates.
