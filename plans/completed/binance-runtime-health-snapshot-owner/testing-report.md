# Testing Report

## Scope

- Feature: `binance-runtime-health-snapshot-owner`
- Date: 2026-03-15
- Outcome: internal runtime-health snapshot seam now exists in `cmd/market-state-api` and stays separate from `/healthz` plus `/api/market-state/*`

## Commands

| Command | Result | Notes |
|---|---|---|
| `/usr/local/go/bin/go test ./cmd/market-state-api ./services/venue-binance -count=2` | pass | Covered deterministic repeated-input behavior across the new snapshot seam and depth-recovery helpers |
| `/usr/local/go/bin/go test ./cmd/market-state-api` | pass | Confirmed command-level provider/runtime regression coverage stayed green |
| `/usr/local/go/bin/go test -race ./cmd/market-state-api` | pass | Checked command-level runtime-health snapshot reads for race issues |
| `/usr/local/go/bin/go test -race ./services/venue-binance` | pass | Checked new time-relative depth status helpers for race issues |

## Evidence

- Added `cmd/market-state-api/runtime_health.go` for the internal runtime-health snapshot contract and owner read path.
- Added `cmd/market-state-api/runtime_health_test.go` for startup, healthy, degraded, stale, rate-limit, and concurrent-read coverage.
- Updated `services/venue-binance/spot_depth_recovery.go` and `services/venue-binance/spot_depth_recovery_test.go` so time-relative depth status and feed-health reasons remain self-consistent at snapshot read time.
- Preserved `/healthz` and `/api/market-state/*` behavior; this slice does not add a public status endpoint.

## Review

- Fresh-context code review: no meaningful blocking issues after follow-up fixes.
- Fresh-context Go review: no meaningful blocking issues after follow-up fixes.

## Follow-On

- Next runtime-health step: run `feature-planning` for `binance-runtime-status-endpoint-and-ops-handoff`.
- Parallel-safe planning step: run `feature-planning` for `binance-usdm-influence-policy-and-signal`.
