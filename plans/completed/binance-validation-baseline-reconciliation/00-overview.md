# Binance Validation Baseline Reconciliation

## Ordered Implementation Plan

1. Reconcile Binance runtime tests with the current prod-like runtime config values in `configs/*/ingestion.v1.json`.
2. Reconcile replay run-result contract-family validation with the actual replay result schema and Go producer shape.
3. Reconcile replay seed deterministic ordering with fixture `expectedCanonical[].sourceRecordId` values.
4. Run the baseline validation matrix and update planning handoff state with the passing evidence or remaining blockers.

## Outcome

Restore the Binance validation baseline so `go test ./...`, contract-family validation, and deterministic fixture replay smoke pass before adding new Binance market-intelligence indicators.

## Requirements

- Keep the work limited to validation drift and planning state reconciliation.
- Do not add Spot trade-flow, Spot liquidity, USD-M enrichment, alerting, dashboard, or API indicator behavior in this feature.
- Preserve Go as the live runtime owner and keep Python scripts offline/dev validation only.
- Keep tracked symbols fixed to `BTC-USD` and `ETH-USD`.
- Keep `/healthz` process-only and `/api/runtime-status` as the operator runtime-health surface.
- Treat config, contracts, fixtures, and replay outputs as deterministic for the same inputs and code version.
- Avoid breaking replay run-result consumers; if a contract shape changes, update producer, schema, fixtures, and consumer validation in the same implementation slice.

## Current Drift To Reconcile

- `go test ./...` fails in `services/venue-binance/runtime_test.go` and `services/venue-binance/spot_depth_recovery_test.go` because stale expectations no longer match the current Binance runtime config values.
- `make contracts-validate` fails because `scripts/dev/validate_contract_families.py` expects replay run-result fields that do not match `schemas/json/replay/replay-run-result.v1.schema.json` and the Go replay contract shape.
- `CONTRACT_FIXTURES=1 make replay-smoke` fails because at least `tests/replay/seeds/btc-normal-microstructure-window.seed.v1.json` expects a source record ID timestamp spelling that does not match the referenced fixture canonical output.
- Docker is unavailable in the current WSL environment, so Compose proof remains a later operator or different-host validation, not part of this feature's required pass gate.

## Target Repo Areas

- `services/venue-binance/runtime_test.go`
- `services/venue-binance/spot_depth_recovery_test.go`
- `configs/local/ingestion.v1.json`
- `configs/dev/ingestion.v1.json`
- `configs/prod/ingestion.v1.json`
- `scripts/dev/validate_contract_families.py`
- `schemas/json/replay/replay-run-result.v1.schema.json`
- `libs/go/contracts/replay.go`
- `services/replay-engine/runtime.go`
- `tests/replay/seeds/*.seed.v1.json`
- `tests/fixtures/events/**/*fixture.v1.json`
- `scripts/dev/replay_smoke.py`
- `plans/STATE.md`
- `plans/epics/binance-market-intelligence-gap-closure/92-refinement-handoff.md`
- `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md`

## Data And Validation Flow

```text
configs/*/ingestion.v1.json
  -> ingestion.LoadEnvironmentConfig
  -> services/venue-binance Runtime
  -> runtime/depth recovery tests

schemas/json/replay/replay-run-result.v1.schema.json
  + libs/go/contracts.ReplayRunResult
  + services/replay-engine result producer
  + scripts/dev/validate_contract_families.py
  -> make contracts-validate

tests/fixtures/manifest.v1.json
  -> fixture expectedCanonical[].sourceRecordId
  -> tests/replay/seeds/*.seed.v1.json expectedDeterminism
  -> scripts/dev/replay_smoke.py
  -> CONTRACT_FIXTURES=1 make replay-smoke
```

## Design Notes

- Current checked-in Binance config values are `reconnectBackoffMinMs=1000`, `reconnectBackoffMaxMs=15000`, `reconnectLoopThreshold=3`, `resyncLoopThreshold=2`, and `snapshotCooldownMs=2000` in `configs/local/ingestion.v1.json`.
- If those config values are intentional, update tests to assert the config-driven behavior instead of restoring older expectations such as 5 second max reconnect backoff or 1 second snapshot cooldown.
- If a test needs a shorter cooldown to exercise recovery mechanics, use a test-local config override rather than changing repo-wide checked-in runtime config.
- The replay run-result schema currently matches the Go producer family around `runId`, `outputDigest`, `inputCounters`, `startedAt`, `finishedAt`, and `manifestDigest`; the validator's `id`, `seedId`, `symbol`, and `outputChecksum` expectations should be treated as suspected stale validation policy unless implementation discovers a real producer migration requirement.
- Replay seed ordering should match the exact canonical source record IDs emitted by referenced fixtures, including timestamp normalization such as `Z` versus `.000Z`.

## Acceptance

- Focused Binance runtime/depth recovery tests pass.
- `go test ./...` passes.
- `make contracts-validate` passes without weakening family validation for unrelated schemas.
- `CONTRACT_FIXTURES=1 make replay-smoke` passes and remains deterministic across repeated runs.
- Planning state identifies this feature as ready for testing or archived only after implementation evidence exists.
- No new live runtime dependency on Python, browser-side Binance access, or fixture-backed live behavior is introduced.
