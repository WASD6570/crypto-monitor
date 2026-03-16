# Binance Runtime Config Profile Parity

## Ordered Implementation Plan

1. Align the checked-in Binance `rest` profile blocks in `configs/local/ingestion.v1.json`, `configs/dev/ingestion.v1.json`, and `configs/prod/ingestion.v1.json` so the existing perpetual `open-interest` stream has explicit, valid polling defaults in every environment without changing the tracked symbol set.
2. Expand `libs/go/ingestion` config coverage so `LoadEnvironmentConfig(...)` and `RuntimeConfigFor(ingestion.VenueBinance)` are exercised for `local`, `dev`, and `prod`, including environment-ladder assertions and negative cases for invalid open-interest settings.
3. Add narrow `cmd/market-state-api` consumption proof that the current live provider can accept each checked-in environment profile through `configPath` while preserving the already-settled Spot and USD-M override guardrails.
4. Run the focused validation matrix, write `plans/binance-runtime-config-profile-parity/testing-report.md`, and keep the plan active under `plans/binance-runtime-config-profile-parity/` until implementation and validation are complete.

## Requirements

- Scope is limited to checked-in Binance environment-profile parity plus the smallest proof that the current Go-owned live runtime can still consume those profiles.
- Keep Go as the live runtime path; Python remains offline-only.
- Keep the tracked symbols fixed to `BTC-USD` and `ETH-USD`.
- Do not add new environment-selection variables or change config-path precedence in `cmd/market-state-api`; that belongs to `binance-market-state-api-startup-defaults-and-override-guardrails`.
- Keep the Binance perpetual `open-interest` stream enabled in all three environment files; do not remove the stream to bypass validation.
- `dev` and `prod` must gain explicit positive `openInterestPollIntervalMs` and `openInterestPollsPerMinuteLimit` values, and those defaults must stay no more aggressive than `local`; `prod` must stay no more aggressive than `dev`.
- Keep non-Binance venue behavior unchanged unless a tiny shared test-helper change is required for coverage.
- Preserve `/healthz`, `/api/runtime-status`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol` behavior exactly; this feature is config-only and must not widen runtime or API semantics.
- Preserve deterministic config loading for the same checked-in inputs.

## Design Notes

### Profile alignment posture

- Use the existing `local` Binance profile as the baseline reference because it is the only profile that currently satisfies the live USD-M open-interest poller requirements.
- Make the `dev` and `prod` Binance `rest` blocks explicit rather than relying on missing-field behavior.
- Keep the environment pressure ladder conservative and monotonic: `local` may remain the fastest polling profile, `dev` must be equal or slower than `local`, and `prod` must be equal or slower than `dev`.
- Once the `dev` and `prod` numbers are chosen, pin them in tests so later edits cannot silently increase Binance pressure.

### Validation authority

- Keep `libs/go/ingestion/config.go` as the single authority for profile parsing and open-interest validation.
- Prefer table-driven tests that load the real checked-in files instead of duplicating profile data in Go fixtures.
- Add a focused negative regression that proves open-interest streams still fail validation when the interval or per-minute limit is missing or zero.

### Command consumer proof

- Use the existing `cmd/market-state-api/main_test.go` stub-server harness to prove the current provider can start from each checked-in profile without talking to real Binance.
- Keep this proof narrow: validate `configPath` consumption and the existing Spot versus USD-M override guardrails, but do not introduce the later startup-default selection contract in this feature.

### Live-path boundary

- Runtime ownership remains in Go under `cmd/market-state-api`, `services/venue-binance`, and `services/market-state-api`.
- This feature changes only checked-in config data plus Go validation/consumer tests; it must not introduce browser-side Binance logic or any Python dependency.

## ASCII Flow

```text
checked-in configs/{local,dev,prod}/ingestion.v1.json
              |
              v
libs/go/ingestion.LoadEnvironmentConfig(...)
              |
              v
RuntimeConfigFor(BINANCE)
  - fixed symbols
  - explicit open-interest polling defaults
  - conservative environment ladder
              |
              +--> libs/go/ingestion tests
              |
              +--> cmd/market-state-api provider tests with stub endpoints
                              |
                              v
                     same live runtime shape as Wave 2
```

## Archive Intent

- Keep this feature active under `plans/binance-runtime-config-profile-parity/` while implementation and validation are in progress.
- When complete, move the full directory and `testing-report.md` to `plans/completed/binance-runtime-config-profile-parity/`.
