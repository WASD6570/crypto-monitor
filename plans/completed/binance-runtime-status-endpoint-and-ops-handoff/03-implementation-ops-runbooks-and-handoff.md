# Implementation Module 3: Ops Runbooks And Handoff

## Scope

- Update the operator runbooks and service README for the new runtime-status route.
- Explain how `/api/runtime-status`, `/healthz`, and `/api/market-state/*` differ.
- Exclude new UI work and non-Binance observability platform changes.

## Target Repo Areas

- `docs/runbooks/ingestion-feed-health-ops.md`
- `docs/runbooks/degraded-feed-investigation.md`
- `services/market-state-api/README.md`

## Requirements

- Document the runtime-status route as the bounded operator surface for warm-up, reconnect, stale, recovery, and rate-limit posture.
- Keep `/healthz` documented as process-health only.
- Reuse the shared `HEALTHY`, `DEGRADED`, and `STALE` vocabulary plus the canonical reason strings exactly.
- Show operators how to inspect readiness versus degradation without implying that the dashboard or client should compute runtime truth.
- Keep docs aligned with the actual response fields and the fixed `BTC-USD` / `ETH-USD` symbol scope.

## Key Decisions

- Update existing feed-health runbooks rather than creating a Binance-only duplicate runbook for the same vocabulary.
- Add one compact runtime-status example and triage sequence to the docs and README.
- Keep the runbook centered on machine-readable route output first and logs second.

## Unit Test Expectations

- No new code-level unit tests are required in this module, but the docs must match the implemented route path and response semantics.
- README and runbook examples should be validated by the same command/integration smoke in `04-testing.md`.

## Contract / Fixture / Replay Impacts

- No shared schema or replay artifact change is expected.
- Documentation must preserve the canonical reason names because later ops flows and tests rely on them.

## Summary

This module turns the endpoint into an operator-ready handoff so the new route becomes the clear debugging surface for runtime-health without weakening `/healthz` or current-state boundaries.
