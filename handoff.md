# Handoff

## Current Goal

Continue the visibility flow by refining the remaining `slow-context-panel` epic now that dashboard-core implementation is archived.

## Completed Plan Archive

- `plans/completed/canonical-contracts-and-fixtures/`
- `plans/completed/market-ingestion-and-feed-health/`
- `plans/completed/binance-adapter-loop-runtime/`
- `plans/completed/bybit-adapter-foundation/`
- `plans/completed/coinbase-adapter-foundation/`
- `plans/completed/kraken-l2-adapter-foundation/`
- `plans/completed/normalizer-feed-health-handoff/`
- `plans/completed/ingestion-ops-validation-and-runbooks/`
- `plans/completed/raw-event-log-boundary/`
- `plans/completed/dashboard-shell-and-summary-strip/`
- `plans/completed/replay-run-manifests-and-ordering/`
- `plans/completed/world-usa-composite-snapshots/`
- `plans/completed/backfill-checkpoints-and-audit-trail/`
- `plans/completed/market-quality-and-divergence-buckets/`
- `plans/completed/replay-retention-and-safety-validation/`
- `plans/completed/raw-storage-and-replay-foundation/`
- `plans/completed/symbol-and-global-regime-state/`
- `plans/completed/market-state-current-query-contracts/`
- `plans/completed/dashboard-query-adapters-and-trust-state/`
- `plans/completed/market-state-history-and-audit-reads/`
- `plans/completed/dashboard-detail-panels-and-symbol-switching/`
- `plans/completed/dashboard-negative-state-mobile-a11y/`
- `plans/completed/dashboard-fixture-smoke-matrix/`

## Current Recommended Next Step

- Run `program-refining` for `plans/epics/slow-context-panel/`.
- Treat `visibility-dashboard-core` as completed and archived through its child plans.

## Notes For Future Agents

- Broad unfinished work stays under `plans/epics/`.
- Active implementation-ready child plans stay under `plans/`.
- Implemented and validated plans move to `plans/completed/`.
- Use epics as refinement inputs and archived plans as read-only prerequisite context when a later slice depends on prior decisions.
- `plans/completed/raw-event-log-boundary/` is the first implemented child slice from the completed replay/storage epic in `plans/completed/raw-storage-and-replay-foundation/`.
- `plans/completed/raw-storage-and-replay-foundation/` now records the finished replay/storage epic and its completed child slices.
- `plans/completed/replay-run-manifests-and-ordering/` now has replay manifest, deterministic ordering, and runtime-mode implementation plus passing targeted Go validation.
- `plans/epics/visibility-dashboard-core/` now has refinement artifacts in `90-refinement-map.md`, `91-child-plan-seeds.md`, and `92-refinement-handoff.md`.
- `plans/completed/dashboard-shell-and-summary-strip/` now has route-shell implementation plus passing unit, build, and Playwright validation.
- `plans/epics/world-usa-composites-and-market-state/` now has refinement artifacts in `90-refinement-map.md`, `91-child-plan-seeds.md`, and `92-refinement-handoff.md`.
- `plans/completed/world-usa-composite-snapshots/` now has Go-owned WORLD/USA composite snapshot implementation plus passing unit, service, integration, and replay validation.
- `plans/completed/backfill-checkpoints-and-audit-trail/` now has bounded replay request, checkpoint, audit, and apply-gate implementation plus passing targeted Go validation.
- `plans/completed/market-quality-and-divergence-buckets/` now has deterministic bucket, divergence, fragmentation, and market-quality implementation plus passing unit, service, integration, and replay validation.
- `plans/completed/replay-retention-and-safety-validation/` now has integrated replay retention, continuity, and side-effect-safety validation plus passing Go replay/integration checks.
- `plans/completed/symbol-and-global-regime-state/` now has service-owned symbol/global regime classification plus passing unit, service, integration, and replay validation.
- `plans/completed/market-state-current-query-contracts/` now has versioned current-state schemas and service query surfaces plus passing schema, service, integration, and replay validation.
- `plans/completed/dashboard-query-adapters-and-trust-state/` now has adapter-backed dashboard state, trust/fallback rendering, and passing unit/build/browser validation.
- `plans/completed/market-state-history-and-audit-reads/` now has version-pinned historical reads, replay-corrected audit provenance, and passing targeted Go validation.
- `plans/completed/dashboard-detail-panels-and-symbol-switching/` now has focused overview, microstructure, derivatives-gap, and health/regime panels plus passing unit, build, and desktop browser validation.
- `plans/completed/dashboard-negative-state-mobile-a11y/` now has explicit route/panel warning hierarchy, mobile-safe layout hardening, keyboard-safe navigation, and passing unit/build/desktop/mobile validation.
- `plans/completed/dashboard-fixture-smoke-matrix/` now has the shared scenario catalog, mapper/shell matrix, and desktop/mobile Playwright smoke coverage for the final dashboard-core slice.
- `plans/epics/visibility-dashboard-core/` now has all child features implemented and validated.
- Browser validation passed by temporarily downloading Ubuntu runtime libraries locally and exporting `LD_LIBRARY_PATH`; no system package installation was required.
- The ingestion wave is complete; do not restart it unless the user explicitly asks for follow-up work.

## Environment Reminder

- Use `"/usr/local/go/bin/go"` and `"/usr/local/go/bin/gofmt"` explicitly because `go` is not reliably on PATH in this shell.
