# Testing Report

## Scope

- Feature: `binance-runtime-status-endpoint-and-ops-handoff`
- Date: 2026-03-15
- Outcome: `GET /api/runtime-status` now exposes the bounded Binance runtime-health surface, `/healthz` stays process-only, and the operator docs describe the new contract.

## Commands

| Command | Result | Notes |
|---|---|---|
| `/usr/local/go/bin/go test ./services/market-state-api ./cmd/market-state-api` | pass | Covered the additive route, command wiring, warm-up/degraded route proof, and existing API routes together |
| `/usr/local/go/bin/go test ./services/market-state-api ./cmd/market-state-api -count=2` | pass | Repeated accepted-input runs stayed green and preserved deterministic runtime-status assertions |
| `/usr/local/go/bin/go test -race ./cmd/market-state-api` | pass | Checked concurrency safety for command-owned runtime snapshot reads through the new route |
| `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinanceCurrentState\|TestIngestionBinance.*Runtime'` | pass | Confirmed current-state behavior stayed intact beside the runtime-status addition |

## Evidence

- Added public runtime-status response types, route handling, deterministic symbol normalization, and unsupported-provider handling in `services/market-state-api/api.go`.
- Added route-focused coverage in `services/market-state-api/api_test.go` for deterministic ordering, unsupported providers, healthy empty reasons, and invalid payload rejection.
- Wired `cmd/market-state-api` into the optional runtime-status seam in `cmd/market-state-api/live_provider.go` and projected the existing internal snapshot in `cmd/market-state-api/runtime_health.go`.
- Added command-level route proof in `cmd/market-state-api/main_test.go` for warm-up, degraded runtime visibility, `/healthz` separation, and fixed-symbol config enforcement.
- Updated `docs/runbooks/ingestion-feed-health-ops.md`, `docs/runbooks/degraded-feed-investigation.md`, and `services/market-state-api/README.md` to document `/api/runtime-status` as the bounded operator surface.

## Review

- Fresh-context code review: no blocking issues after follow-up fixes for response normalization, docs alignment, and config-drift fail-fast behavior.
- Fresh-context Go review: no blocking issues after follow-up fixes.

## Follow-On

- Next recommended implementation step: run `feature-implementing` for `plans/binance-usdm-influence-policy-and-signal/`.
- Keep `binance-usdm-output-application-and-replay-proof` blocked until the first USD-M child settles the influence contract.
