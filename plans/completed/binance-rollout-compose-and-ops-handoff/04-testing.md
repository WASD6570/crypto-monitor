# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Compose rendering | `docker compose config` | Prove `docker-compose.yml` remains valid after the rollout-handoff changes | Rendered compose output succeeds and shows one `market-state-api` startup posture |
| Guardrail regression | `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'` | Confirm docs or compose-facing changes do not drift away from the settled config-path and override guardrails | The focused provider tests stay green |
| Repeatable rollout smoke | `make compose-smoke` | Prove the checked-in compose stack starts, keeps `/healthz` process-only, and exposes `/api/runtime-status` plus `/api/market-state/global` through the same-origin web path | The smoke helper brings the stack up, reaches the expected routes, accepts warm-up posture, and tears the stack down cleanly |
| Operator handoff verification | Follow the exact sequence documented in `docs/runbooks/` for the new rollout runbook | Confirm the written handoff matches the runnable compose posture and the existing degraded-feed investigation path | The runbook steps are accurate, route names match reality, and degradation follow-up points to the existing runbooks without contradiction |

## Verification Checklist

- Compose and docs describe one prod-like startup posture everywhere.
- `/healthz` remains the compose/process-health gate only.
- `/api/runtime-status` remains the bounded operator runtime-health route for `BTC-USD` and `ETH-USD`.
- `GET /api/market-state/global` and `GET /api/market-state/:symbol` remain consumer read routes reachable through the same-origin web path.
- Warm-up is documented as expected and distinct from degraded runtime.
- Override variables remain explicit, paired, and non-default; no new environment-selection behavior appears.

## Reporting

- Record validation evidence in `plans/binance-rollout-compose-and-ops-handoff/testing-report.md` while the feature is active.
- After implementation and a passing `feature-testing` run, move the full directory to `plans/completed/binance-rollout-compose-and-ops-handoff/` so the testing report archives with the rest of the feature history.
