# Testing Plan: Ingestion Ops Validation And Runbooks

Expected output artifact: `plans/ingestion-ops-validation-and-runbooks/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Adapter happy path | Deterministic fixture-driven venue inputs | canonical events plus healthy feed state | integration report |
| Gap path | Inject book gap | degraded/resync output | integration report |
| Stale path | Advance time beyond freshness threshold | stale output | integration report |
| Retry safety | Trigger reconnect and snapshot recovery pressure | bounded retry behavior | integration report |
| Runbook alignment | Compare docs language to emitted states | vocabulary matches shared health model | review checklist |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance/... ./services/venue-bybit/... ./services/venue-coinbase/... ./services/venue-kraken/... ./services/normalizer/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./tests/integration -run Ingestion`

## Verification Checklist

- Health-state transitions are visible from deterministic test output.
- Operator docs use the same state and reason names as the code.
- Validation stays local and replay-friendly.
