# Refinement Map

## Already Done

- `plans/completed/binance-spot-ws-runtime-supervisor/` settled machine-readable Spot supervisor health inputs, reconnect posture, and runtime timestamps
- `plans/completed/binance-spot-runtime-read-model-owner/` settled the sustained command-owned Spot runtime and preserved read-model degradation semantics
- `plans/completed/binance-market-state-live-reader-cutover/` cut the stable current-state API over to the sustained runtime while keeping `/healthz` process-only
- `plans/completed/binance-runtime-health-snapshot-owner/` settled the internal command-owned runtime-health snapshot contract and deterministic per-symbol snapshot reads
- `docs/runbooks/ingestion-feed-health-ops.md` and `docs/runbooks/degraded-feed-investigation.md` already define the shared health vocabulary and operator meaning this epic must reuse

## Remaining Work

- expose that snapshot through an additive status surface without breaking `/api/market-state/*` or changing `/healthz` semantics
- document the operator investigation path and validate warm-up, reconnect, stale, rate-limit, and recovery behavior through real API checks

## Overlap And Non-Goals

- do not redesign the current-state payloads or dashboard shell; current-state honesty stays intact, but operator visibility should not depend on reading those payloads alone
- do not settle USD-M current-state or regime influence; that remains the separate Wave 2 seed
- do not bundle environment rollout defaults, config hardening, or long-run soak validation into this epic
- do not replace shared health vocabulary with Binance-specific aliases in docs, logs, or status fields

## Refinement Waves

### Wave 1

- `binance-runtime-health-snapshot-owner`
- Why first: the repo lacked one explicit operator-facing runtime-health contract that downstream handlers and docs could rely on

### Wave 2

- `binance-runtime-status-endpoint-and-ops-handoff`
- Why later: the additive endpoint and runbook handoff should follow the settled snapshot contract rather than inventing a surface before the status model exists

### Direct Post-Implementation Checks

- verify `/healthz` still reports process health independently of market-data freshness while the new status surface shows runtime state
- verify current-state routes stay stable and operator docs use the shared `HEALTHY` / `DEGRADED` / `STALE` vocabulary with canonical reasons
