# Implementation: Live Query Source Seam

## Module Requirements

- Replace package-local deterministic bundle construction with one explicit live query-source seam inside `services/market-state-api`.
- Preserve the existing `Provider` interface and handler routes unchanged.
- Make the new seam injectable so deterministic tests can still run without live sockets.
- Keep supported symbols explicit as `BTC-USD` and `ETH-USD` with unchanged unsupported-symbol behavior.

## Target Repo Areas

- `services/market-state-api`
- `services/market-state-api/api.go`
- `services/market-state-api/api_test.go`

## Key Decisions

- Introduce a package-local source abstraction beneath `Provider`, for example a live query source that returns assembled symbol and global query inputs or fully assembled symbol/global responses.
- Keep `DeterministicProvider` as a test and fallback implementation of the new seam until the follow-on provider-cutover feature switches the command entrypoint.
- Extract hard-coded assembly helpers out of `api.go` into focused files or types so deterministic and live implementations can share validation and supported-symbol logic.
- Keep environment and command wiring out of scope here; construction should accept explicit config, clock, and dependency injection rather than reaching into process env.

## Data And Algorithm Notes

- Treat `currentStateBundle` as the boundary to replace. The live seam should make it obvious where symbol state, global state, and slow-context inputs come from.
- Preserve route-level behavior:
  - unsupported symbol -> `ErrUnsupportedSymbol`
  - provider/source failure -> `500`
  - handler JSON shape unchanged
- Reuse the current package-local composite and regime config builders where practical, but make them callable from both deterministic and live query sources.

## Unit Test Expectations

- handler tests still pass unchanged against the new seam
- deterministic provider still serves `BTC-USD` and `ETH-USD`
- unsupported symbol handling remains `404`
- source failure still maps to `500`
- live-source constructor rejects missing required dependencies cleanly

## Contract / Fixture / Replay Notes

- No current-state HTTP schema changes are allowed in this module.
- No replay manifest or raw append contract changes are expected here.
- Any new helper types should stay package-local unless another service clearly shares them.

## Summary

This module creates the replaceable source seam the current API is missing, so the next module can assemble real Binance-backed query inputs without changing the consumer-facing API surface.
