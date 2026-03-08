# Slow-Context Contract And View Model

## Module Requirements And Scope

Target repo areas:

- `apps/web/src/api/dashboard/dashboardContracts.ts`
- `apps/web/src/api/dashboard/dashboardDecoders.ts`
- `apps/web/src/api/dashboard/dashboardClient.ts`
- `apps/web/src/features/dashboard-state/dashboardStateMapper.ts`
- `apps/web/src/features/dashboard-state/dashboardStateFixtures.ts`
- `apps/web/src/features/dashboard-shell/model/dashboardShellModel.ts`

This module defines how the completed service-owned slow-context seam enters the web app and becomes one explicit, shell-facing advisory panel model.

In scope:

- extending the symbol dashboard contract with one nested slow-context block
- decoding service-owned slow-context fields without local reinterpretation
- building one dedicated panel model for slow-context rows and panel-level helper copy
- preserving row-level isolation for mixed availability across CME and ETF metrics

Out of scope:

- JSX rendering, layout, or CSS details
- extra client polling loops or client-owned freshness logic
- route-level trust, summary-card, or warning changes outside the new panel model

## Planning Guidance

### Contract Shape

- Prefer one nested `slowContext` field on the symbol-scoped dashboard response so the focused symbol still loads through one obvious adapter path.
- Require the service block to stay explicit rather than optional-by-omission. If no metric is readable, the nested block should still decode into an explicit unavailable panel state.
- Preserve service-owned fields needed by the panel directly, including:
  - asset or symbol scope
  - metric family identifier
  - availability and freshness
  - value and unit
  - previous value when present
  - `asOfTs`, `publishedTs`, and `ingestTs`
  - cadence label or cadence key
  - revision or corrected-value marker when present
  - message key and/or service-safe helper text
  - error text for explicit unavailable rows when provided

### Shell-Facing Model

- Add a dedicated `slowContextPanel` field to the dashboard view model instead of forcing slow rows into the existing generic focused-panel metric list.
- The panel model should carry:
  - title and eyebrow copy
  - persistent `Context only` badge text
  - panel-level helper message
  - overall panel availability derived only from service row states
  - ordered metric rows for CME volume, CME open interest, and ETF daily flow
- Each row should carry enough presentational data to render without more derivation, such as:
  - display label
  - formatted value text supplied or assembled from raw value + unit only
  - freshness badge label
  - availability label
  - `as of` label
  - cadence label
  - previous-value label when present
  - revision note when present
  - helper note or error copy

### Mapping Rules

- Treat missing row data as explicit unavailable rows, not as a reason to hide the full panel.
- Derive panel-level status from row availability only for styling; do not feed it back into global dashboard trust.
- Keep panel helper text conservative:
  - prefer service-supplied row messages first
  - use a fixed panel-level note only for stable framing such as `These indicators update on a slower schedule than market-state feeds.`
- Keep timestamp formatting and small text assembly in the mapper, but do not recalculate age thresholds or message families.

## Unit Test Expectations

- decoder accepts the nested slow-context block and rejects malformed availability/freshness fields
- mapper returns three ordered row models for healthy data
- mapper returns mixed available and unavailable rows without changing summary-card trust state
- mapper preserves service-supplied delayed/stale/unavailable copy and revision visibility

## Summary

This module fixes the frontend contract seam for slow context and converts it into a dedicated advisory panel model. Later rendering work can stay purely presentational because row states, timestamps, cadence, and fallback copy are already explicit before JSX composition begins.
