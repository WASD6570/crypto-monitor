# Testing Report: binance-runtime-config-profile-parity

## Environment
- Target: local repository validation for checked-in `local` / `dev` / `prod` ingestion profiles and `cmd/market-state-api` config-path consumption
- Date/time: 2026-03-15 20:50:25 UTC
- Commit/branch: `main` @ `6bad752`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| checked-in config parsing and invariants | `/usr/local/go/bin/go test ./libs/go/ingestion` | Real-file profile loading passes, Binance defaults stay explicit, negative open-interest cases still fail validation | Passed (`ok`, cached) | PASS |
| provider profile consumption | `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'` | Provider accepts `local` / `dev` / `prod` config paths and preserves existing guardrails | Passed (`ok`, cached) | PASS |
| focused Binance USD-M smoke | `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinanceUSDM'` | Spot + USD-M sensor path still works after profile alignment | Passed (`ok`, cached) | PASS |

## Execution Evidence

### checked-in-config-parsing-and-invariants
- Command/Request: `/usr/local/go/bin/go test ./libs/go/ingestion`
- Expected: `LoadEnvironmentConfig(...)` succeeds for checked-in profiles, the conservative Binance ladder stays pinned, and invalid open-interest combinations remain rejected.
- Actual: package passed; the active profile-loading regression and negative open-interest tests remain present in `libs/go/ingestion/config_test.go`.
- Verdict: PASS

### provider-profile-consumption
- Command/Request: `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'`
- Expected: `newProviderWithOptions(...)` starts with each checked-in config path using stubbed endpoints, and existing override/symbol guardrails still fail loudly.
- Actual: targeted provider suite passed; the table-driven environment-profile test and the two guardrail tests remain present in `cmd/market-state-api/main_test.go`.
- Verdict: PASS

### focused-binance-usdm-smoke
- Command/Request: `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinanceUSDM'`
- Expected: the local Binance Spot + USD-M path still passes after config-profile alignment.
- Actual: focused integration suite passed with no failures.
- Verdict: PASS

## Side-Effect Verification

### explicit-open-interest-defaults
- Evidence: `configs/dev/ingestion.v1.json:25`, `configs/dev/ingestion.v1.json:62`, `configs/prod/ingestion.v1.json:25`, `configs/prod/ingestion.v1.json:62`
- Expected state: checked-in `dev` and `prod` profiles carry explicit positive open-interest polling defaults for venues that still configure perpetual `open-interest`, with `prod` no more aggressive than `dev`.
- Actual state: `dev` pins `10000ms` / `24` and `prod` pins `15000ms` / `18` in both affected venue blocks.
- Verdict: PASS

### invariant-and-guardrail-tests-present
- Evidence: `libs/go/ingestion/config_test.go:57`, `libs/go/ingestion/config_test.go:149`, `cmd/market-state-api/main_test.go:841`, `cmd/market-state-api/main_test.go:909`, `cmd/market-state-api/main_test.go:938`
- Expected state: repo-owned tests exist for profile loading, negative open-interest validation, provider config-path consumption, and existing guardrails.
- Actual state: all expected targeted tests are present and backed the passing validation commands.
- Verdict: PASS

### durable-state-sync
- Evidence: `plans/STATE.md`, `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md`, `plans/epics/binance-environment-config-and-rollout-hardening/92-refinement-handoff.md`
- Expected state: testing pass updates durable planning docs so the next child is visible after archive.
- Actual state: updated in the same pass after the matrix succeeded.
- Verdict: PASS

## Blockers / Risks
- none

## Next Actions
1. Run `feature-planning` for `binance-market-state-api-startup-defaults-and-override-guardrails`.
2. Keep `binance-rollout-compose-and-ops-handoff` behind the startup-defaults child.
