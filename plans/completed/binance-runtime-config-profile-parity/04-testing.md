# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Checked-in config parsing and invariant coverage | `/usr/local/go/bin/go test ./libs/go/ingestion` | Prove `local`, `dev`, and `prod` ingestion profiles all load and keep the Binance open-interest invariants explicit | Table-driven profile tests pass, including the monotonic environment ladder and negative open-interest regressions |
| Command-facing profile consumption | `/usr/local/go/bin/go test ./cmd/market-state-api -run 'TestNewProviderWithOptions(LoadsBinanceEnvironmentProfiles|RejectsSpotOverridesWithoutUSDMOverrides|RejectsRuntimeStatusSymbolOverrides)'` | Prove the current live provider can consume each checked-in profile and still preserve the settled guardrails | Provider tests pass for all config paths and the existing override/symbol failures stay intact |
| Focused Binance USD-M integration smoke | `/usr/local/go/bin/go test ./tests/integration -run 'TestIngestionBinanceUSDM'` | Confirm the local Binance Spot plus USD-M sensor path still works after the config-profile alignment | Focused integration coverage passes without changing current-state or runtime-status semantics |

## Verification Checklist

- `configs/local/ingestion.v1.json`, `configs/dev/ingestion.v1.json`, and `configs/prod/ingestion.v1.json` all remain valid `schemaVersion: v1` environment profiles.
- The Binance runtime config in all three profiles keeps `BTC-USD` and `ETH-USD` fixed.
- The Binance perpetual `open-interest` stream remains configured and is backed by explicit positive polling defaults in every environment.
- `dev` is not more aggressive than `local`, and `prod` is not more aggressive than `dev`, for the chosen Binance open-interest defaults.
- `cmd/market-state-api` can consume each checked-in config path without changing `/healthz`, `/api/runtime-status`, or current-state route behavior.
- Existing Spot-versus-USD-M override guardrails remain intact.

## Reporting

- Record implementation validation in `plans/binance-runtime-config-profile-parity/testing-report.md` while the feature is active.
- Once implementation and validation complete, move the full directory to `plans/completed/binance-runtime-config-profile-parity/` so the report archives with the rest of the feature history.
