# Testing Report: binance-rollout-compose-and-ops-handoff

## Environment
- Target: checked-in Compose startup, same-origin web `/api/*` proxy, and rollout/operator docs for the Binance prod-like startup posture
- Date/time: 2026-03-16 00:21:09 UTC
- Commit/branch: `main` @ `6bad752`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| compose rendering | `docker compose config` | Compose stays valid and renders one explicit `market-state-api` startup posture | Passed; rendered `MARKET_STATE_API_CONFIG_PATH=/app/configs/prod/ingestion.v1.json` with no environment matrix | PASS |
| guardrail regression | `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'` | Provider config-path loading and override/runtime-status guardrails stay green | Passed (`ok`, cached) | PASS |
| repeatable rollout smoke | `make compose-smoke` | Checked-in stack starts, exposes same-origin runtime routes, keeps `/healthz` bounded, and tears down cleanly | Passed; helper observed `readiness=NOT_READY`, a readable `GET /api/market-state/global` payload, internal `/healthz`, and clean shutdown | PASS |
| operator handoff verification | Manual sequence from `docs/runbooks/binance-compose-rollout.md` | Runbook commands match actual stack posture and route behavior | Passed; `docker compose up --build -d`, `docker compose ps`, same-origin `curl` checks, and internal `/healthz` probe all matched the documented flow | PASS |

## Execution Evidence

### compose-rendering
- Command/Request: `docker compose config`
- Expected: rendered Compose output succeeds and shows one explicit prod-like startup posture for `market-state-api`.
- Actual: command passed and rendered `MARKET_STATE_API_CONFIG_PATH=/app/configs/prod/ingestion.v1.json` in the only `market-state-api` environment block.
- Verdict: PASS

### guardrail-regression
- Command/Request: `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'`
- Expected: provider config-path loading stays valid, partial Spot-only override combinations still fail loudly, and runtime-status symbol overrides stay rejected.
- Actual: targeted provider suite passed (`ok`, cached) with no failures.
- Verdict: PASS

### repeatable-rollout-smoke
- Command/Request: `make compose-smoke`
- Expected: helper renders Compose, starts the stack, reaches same-origin `GET /api/runtime-status` and `GET /api/market-state/global`, verifies internal `GET /healthz`, and tears down cleanly.
- Actual: helper passed end-to-end; runtime-status was reachable during warm-up with `readiness=NOT_READY`, current-state returned a readable payload, internal `/healthz` returned `{"status":"ok"}`, and cleanup ran successfully.
- Verdict: PASS

### operator-handoff-verification
- Command/Request: documented manual rollout sequence from `docs/runbooks/binance-compose-rollout.md`
- Expected: the runbook's startup, verification, warm-up interpretation, and cleanup commands match the checked-in stack behavior.
- Actual: `docker compose ps` showed `market-state-api` healthy and `web` bound to `4173`; same-origin `GET /api/runtime-status` returned fixed `BTC-USD` and `ETH-USD` symbol entries with `readiness=NOT_READY`; `GET /api/market-state/global` returned `200` with `schemaVersion="market_state_current_global_v1"`; internal `GET /healthz` returned `{"status":"ok"}`.
- Verdict: PASS

## Side-Effect Verification

### same-origin-runtime-and-current-state-surface
- Evidence: manual rollout responses captured during the `docs/runbooks/binance-compose-rollout.md` verification flow
- Expected state: `web` proxies same-origin operator and consumer routes, `runtime-status` stays bounded to fixed symbols, and current-state stays reachable during startup.
- Actual state: `GET /api/runtime-status` returned symbol-scoped runtime entries for `BTC-USD` and `ETH-USD`; `GET /api/market-state/global` returned a readable degraded/no-operate payload through the same-origin web path.
- Verdict: PASS

### process-health-gate-remains-bounded
- Evidence: `docker-compose.yml:12` and the live `docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz` result
- Expected state: Compose health gating depends on `/healthz` process health only, distinct from runtime readiness.
- Actual state: Compose waited for `market-state-api` to become healthy before starting `web`, and `/healthz` returned `{"status":"ok"}` while runtime-status was still allowed to report `readiness=NOT_READY` during warm-up.
- Verdict: PASS

### durable-state-sync
- Evidence: `plans/STATE.md`, `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md`, `plans/epics/binance-environment-config-and-rollout-hardening/92-refinement-handoff.md`
- Expected state: successful feature-testing updates durable planning docs and archives the completed child plan in the same pass.
- Actual state: state docs were updated and the feature was archived under `plans/completed/binance-rollout-compose-and-ops-handoff/` after the matrix passed.
- Verdict: PASS

## Blockers / Risks
- none

## Next Actions
1. Run `program-refining` for `binance-long-run-runtime-hardening`.
2. Use `plans/completed/binance-rollout-compose-and-ops-handoff/` as the settled prod-like startup and operator-handoff reference for the next wave.
