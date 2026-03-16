# Testing Report

## Outcome

- Completed the internal USD-M influence contract and evaluator slice without changing `/api/market-state/*` or `/healthz`.
- Added a deterministic venue-side evaluator input seam, bounded feature-engine evaluator logic, replay proof, and focused regression coverage.

## Files Delivered

- `libs/go/features/usdm_influence.go`
- `libs/go/features/usdm_influence_test.go`
- `services/feature-engine/service.go`
- `services/feature-engine/usdm_influence.go`
- `services/feature-engine/usdm_influence_test.go`
- `services/venue-binance/usdm_influence_inputs.go`
- `services/venue-binance/usdm_influence_inputs_test.go`
- `tests/replay/binance_usdm_influence_replay_test.go`
- `tests/integration/binance_usdm_influence_test.go`

## Validation Commands

- `/usr/local/go/bin/go test ./libs/go/features ./services/feature-engine ./services/venue-binance ./services/market-state-api ./cmd/market-state-api ./tests/replay ./tests/integration`
- `/usr/local/go/bin/go test ./libs/go/features ./services/feature-engine ./services/venue-binance ./tests/replay -count=2`

## What Passed

- Contract validation covers fixed symbol ordering, schema/version fields, and signal primary-reason consistency.
- Venue input seam preserves websocket vs REST freshness semantics, ignores older out-of-order events, and keeps deterministic snapshot ordering.
- Feature-engine evaluator covers `NO_CONTEXT`, `DEGRADED_CONTEXT`, `AUXILIARY`, and bounded `DEGRADE_CAP` posture without mutating current-state outputs.
- Replay and focused integration checks prove repeated identical USD-M inputs yield identical signal outputs while current Spot-only API behavior remains unchanged.

## Review Evidence

- Fresh-context code review: no blocking findings after clarifying that this child stops short of consumer-facing application.
- Fresh-context Go review: no blocking findings after fixing out-of-order overwrite handling, unknown liquidation-age gating, and non-finite numeric rejection.

## Follow-On

- The next child is `binance-usdm-output-application-and-replay-proof`, which can now plan against the settled internal signal contract instead of rediscovering semantics.
