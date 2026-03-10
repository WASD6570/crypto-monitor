# Implementation: Runtime Wiring

## Requirements And Scope

- Replace the default `market-state-api` command provider with the live Spot-backed provider built in the previous feature.
- Add one concrete runtime path that satisfies `marketstateapi.SpotCurrentStateReader` for the command entrypoint.
- Keep the HTTP server, route contract, and supported-symbol list unchanged.
- Keep the first live cutover Spot-only for current-state inputs; do not add USD-M weighting, schema changes, or alternate symbols.
- Preserve honest startup behavior: before enough Spot observations exist, symbol/global responses may remain unavailable or partial instead of silently falling back to deterministic bundles.

## Target Repo Areas

- `cmd/market-state-api/main.go`
- `cmd/market-state-api/*.go` for new bootstrap or reader helpers
- `cmd/market-state-api/*_test.go`
- `services/market-state-api/api.go` only if one tiny export/helper change is required for command wiring

## Implementation Notes

- Introduce a focused command-local constructor such as `newProvider(...)` or `newLiveSpotProvider(...)` so `main()` stays small and testable.
- Default the command to live Spot mode. If an override path is needed for tests or rollback, keep it explicit and narrow; do not make deterministic mode the default runtime again.
- Keep configuration local-first. A small environment override for the config path is acceptable if it removes hard-coded repo assumptions, but avoid introducing a broad multi-environment bootstrap system in this slice.
- Build the concrete reader around process-owned Spot observation state that is populated from the already-settled Binance Spot runtime seams. Reuse existing `services/venue-binance` types instead of creating a second market-data model.
- Keep slow context optional during cutover. Passing `nil` for the slow-context reader is acceptable if that keeps the feature bounded and the browser fallback behavior remains stable.
- Fail fast on startup for invalid command wiring or unreadable required config, but do not fail the server merely because live market observations have not arrived yet.

## Testing Expectations

- Add command-level tests that cover provider construction, address/config defaults, and failure paths for invalid runtime bootstrap.
- Keep existing `services/market-state-api` handler/provider tests passing unchanged, proving that only the entrypoint default changed.
- Add at least one command or integration proof that the live-wired server returns the existing JSON shape and honest unavailable or partial state before/while live inputs warm up.

## Summary

This module makes the live Spot provider the real `cmd/market-state-api` runtime without changing the HTTP contract. The command owns the concrete reader/bootstrap seam, startup remains explicit and testable, and deterministic state stops being the default runtime source.
