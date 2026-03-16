# Binance Rollout Compose And Ops Handoff

## Ordered Implementation Plan

1. Make `docker-compose.yml` express one explicit prod-like `market-state-api` startup posture so the checked-in stack stops reading like a local-only special case.
2. Align `README.md`, `services/market-state-api/README.md`, and the relevant operator runbooks around the same startup posture, warm-up expectations, `/healthz`, `/api/runtime-status`, and the same-origin `/api/market-state/*` boundary.
3. Add one repeatable compose rollout proof for operators and developers so the handoff is runnable instead of prose-only.
4. Run the focused validation matrix and record evidence in `plans/binance-rollout-compose-and-ops-handoff/testing-report.md` while the feature remains active.

## Requirements

- Scope is limited to compose wiring, rollout documentation, operator runbooks, and the smallest repeatable validation helper needed for the Binance handoff.
- Keep the checked-in runtime posture prod-like everywhere; do not reintroduce environment-specific startup selection or different runtime personalities for `local`, `dev`, and `prod`.
- Preserve the settled route semantics: `/healthz` stays process health only, `/api/runtime-status` stays the operator runtime-health surface, and `GET /api/market-state/global` plus `GET /api/market-state/:symbol` stay consumer read endpoints.
- Keep `BTC-USD` and `ETH-USD` fixed.
- Keep the same-origin Go API boundary intact; `apps/web` remains a consumer of `/api/*` and does not talk to Binance directly.
- Preserve the existing loud-failure rule for unsafe Spot-only override combinations; override variables remain explicit expert/operator tools, not a new environment-selection layer.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Compose startup posture

- Prefer one explicit compose startup contract over a vague local-first default story.
- If compose needs to pin a config path, use one checked-in prod-like profile path consistently rather than creating a per-environment matrix.
- Keep the existing `market-state-api` healthcheck on `/healthz` and keep `web` dependent on that process-health gate.

### Documentation and handoff posture

- Root docs should explain the single supported startup posture at the repo level.
- `services/market-state-api/README.md` should remain the service-specific source for runtime routes and override guardrails.
- Add one rollout-focused runbook under `docs/runbooks/` for compose startup, warm-up interpretation, and operator verification, then cross-link existing feed-health runbooks instead of duplicating their degradation guidance.

### Repeatable rollout proof

- Prefer one repo-owned helper or tightly scoped command sequence that runs `docker compose config`, starts the stack, checks `/api/runtime-status` and `/api/market-state/*` through the same-origin web path, checks `/healthz` inside `market-state-api`, and tears the stack down cleanly.
- Treat warm-up as expected. The proof should verify route availability and contract posture even when `readiness=NOT_READY` appears briefly.

### Live-path boundary

- Runtime ownership stays in Go under `cmd/market-state-api` and `services/market-state-api`.
- This feature must not introduce browser-side Binance logic, Python runtime dependencies, or rollout automation that depends on infrastructure-specific secrets handling.

## ASCII Flow

```text
docker compose / repo startup docs
            |
            v
 market-state-api container
   - one prod-like config posture
   - /healthz for process health
            |
            +--------------------+
            |                    |
            v                    v
 web same-origin /api proxy   operator runbook
   - /api/runtime-status       - warm-up expectations
   - /api/market-state/*       - degraded investigation entry
            |                    |
            +---------+----------+
                      v
           repeatable rollout smoke proof
```

## Archive Intent

- Keep this feature active under `plans/binance-rollout-compose-and-ops-handoff/` while implementation and validation are in progress.
- After a passing `feature-testing` run, move the full directory and `testing-report.md` to `plans/completed/binance-rollout-compose-and-ops-handoff/`.
