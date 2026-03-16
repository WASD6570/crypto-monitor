# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Feature and schema unit coverage | `/usr/local/go/bin/go test ./libs/go/features ./services/feature-engine ./services/regime-engine ./services/market-state-api` | Prove the application helper, regime reasoning, and additive current-state provenance are correct | Targeted unit/schema tests pass for unchanged and capped cases |
| Live provider and command wiring | `/usr/local/go/bin/go test ./services/venue-binance ./cmd/market-state-api ./services/market-state-api` | Prove the live provider can read USD-M context and apply it without new HTTP-surface drift | Provider/command tests pass for auxiliary and degrade-cap cases |
| Focused integration proof | `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinance(CurrentState|USDM)'` | Confirm consumer-facing Binance current-state behavior stays stable except for the planned bounded USD-M application | Integration tests show unchanged spot-only cases and explicit capped cases |
| Replay determinism | `/usr/local/go/bin/go test ./tests/replay -run 'TestReplayBinanceUSDM|TestReplayBinanceCurrentState' -count=2` | Prove repeated pinned Spot plus USD-M inputs yield identical current-state and regime outputs | Repeated replay runs stay green with stable assertions |
| Concurrency safety for command-owned live seams | `/usr/local/go/bin/go test -race ./cmd/market-state-api` | Catch races in any new runtime snapshot or USD-M owner wiring | Race run passes without data races |

## Verification Checklist

- `GET /api/market-state/:symbol` remains backward-compatible apart from the planned additive USD-M provenance summary.
- `GET /api/market-state/global` reflects the bounded cap only through the settled regime output and reason codes.
- `AUXILIARY`, `NO_CONTEXT`, and `DEGRADED_CONTEXT` preserve the prior spot-derived regime result.
- `DEGRADE_CAP` never escalates a symbol above spot-derived output and never downgrades past the planned `WATCH` ceiling.
- `/healthz` stays unchanged.
- Repeated accepted-input runs keep deterministic symbol ordering and identical results.

## Reporting

- Record implementation validation in `plans/binance-usdm-output-application-and-replay-proof/testing-report.md` while the feature is active.
- Once implementation and validation complete, move the full directory to `plans/completed/binance-usdm-output-application-and-replay-proof/` so the report archives with the rest of the feature history.
