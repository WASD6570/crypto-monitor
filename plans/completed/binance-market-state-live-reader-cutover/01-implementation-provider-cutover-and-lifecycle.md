# Implementation: Provider Cutover And Lifecycle

## Requirements And Scope

- Replace the remaining temporary reader/bootstrap assumptions in `cmd/market-state-api` with the sustained runtime owner from the prerequisite feature.
- Keep provider construction, startup, and shutdown explicit and bounded in the command process.
- Preserve the stable `services/market-state-api` provider contract and avoid route or schema changes.
- Keep the implementation limited to the command/provider seam; deeper venue-runtime semantics stay owned by `services/venue-binance` and the prerequisite runtime-owner slice.

## Target Repo Areas

- `cmd/market-state-api/*.go`
- `cmd/market-state-api/*_test.go`
- `services/market-state-api/*.go` only if small additive lifecycle or provider wiring support is required

## Implementation Notes

- Treat `newProvider()` / `newProviderWithOptions(...)` as the main cutover seam.
- Ensure provider startup always starts the runtime owner before serving requests and cleanly stops it on provider/server shutdown.
- Remove any lingering polling-oriented naming, config, or fallback logic that no longer matches the sustained runtime source.
- Keep failure modes explicit: configuration, runtime-owner startup, and provider assembly errors should remain actionable and should not silently fall back to deterministic or ad hoc polling behavior.
- Preserve the current `Close(context.Context) error` shutdown handoff so `main.go` can stop the runtime owner without needing service-specific logic in the handler layer.

## Unit Test Expectations

- Provider construction proves the command uses the runtime owner path and rejects missing config/websocket prerequisites clearly.
- Shutdown tests prove the runtime owner is stopped on provider close and server shutdown without leaking goroutines or hanging indefinitely.
- Command tests prove no per-request REST polling fallback remains when the runtime owner exists but has not published data yet.

## Contract / Replay Notes

- No shared schema change is expected.
- Replay-sensitive behavior stays at the read-model output seam: repeated accepted inputs must still produce stable provider-visible snapshots.

## Summary

This module finishes the command-side cutover so the live provider is truly backed by the sustained runtime owner rather than a temporary local seam. The key outcome is lifecycle-safe startup and shutdown with no hidden fallback path.
