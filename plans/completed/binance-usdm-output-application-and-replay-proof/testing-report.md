# Testing Report

## Outcome

- Completed the consumer-facing USD-M application slice with one bounded `DEGRADE_CAP` watch-cap, additive provenance on symbol/global current-state responses, and shared live/deterministic assembly helpers.
- Wired the command-owned live path to evaluate settled USD-M inputs through Go-only services while keeping `/healthz` unchanged and preserving spot fallback when USD-M reads fail.

## Files Delivered

- `libs/go/features/regime.go`
- `libs/go/features/market_state_current.go`
- `libs/go/features/usdm_influence_application.go`
- `libs/go/features/usdm_influence_test.go`
- `schemas/json/features/market-state-current-symbol.v1.schema.json`
- `services/feature-engine/current_state_test.go`
- `services/market-state-api/live_spot_provider.go`
- `services/market-state-api/api.go`
- `services/market-state-api/usdm_influence.go`
- `services/market-state-api/api_test.go`
- `cmd/market-state-api/live_provider.go`
- `cmd/market-state-api/usdm_influence_owner.go`
- `cmd/market-state-api/main_test.go`
- `tests/integration/binance_current_state_test.go`
- `tests/integration/binance_usdm_influence_test.go`
- `tests/replay/binance_usdm_influence_replay_test.go`

## Validation Commands

- `/usr/local/go/bin/go test ./libs/go/features ./services/feature-engine ./services/regime-engine ./services/market-state-api`
- `/usr/local/go/bin/go test ./services/venue-binance ./cmd/market-state-api ./services/market-state-api`
- `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinance(CurrentState|USDM)'`
- `/usr/local/go/bin/go test ./tests/replay -run 'TestReplayBinanceUSDM|TestReplayBinanceCurrentState' -count=2`
- `/usr/local/go/bin/go test -race ./cmd/market-state-api`

## What Passed

- Application helpers cap only `DEGRADE_CAP` signals to `WATCH`, preserve auxiliary/no-context outputs, and expose deterministic machine-readable provenance.
- Global current-state summaries now carry per-symbol USD-M provenance so consumers can distinguish spot-derived and USD-M-capped watch posture without route churn.
- Live and deterministic provider wiring both evaluate the settled signal set, use dedicated USD-M futures endpoints by default, keep custom spot endpoint overrides explicit, and fall back to spot-only output when USD-M reads fail.
- Focused integration and replay coverage proves bounded cap behavior, unchanged auxiliary behavior, stable repeated outputs, and race-free command-owned live seams.

## Review Evidence

- Fresh-context code review: no findings after fixing live USD-M endpoint routing, global provenance coverage, graceful fallback on USD-M read failures, and request-cancellation handling.
- Fresh-context Go review: no findings after moving open-interest polling off request reads, guarding live runtime state for race safety, and requiring explicit USD-M overrides when spot endpoints are customized.

## Follow-On

- The next initiative seed is `binance-environment-config-and-rollout-hardening`, which can now move into refinement against the settled runtime-health and USD-M market-state behavior.
