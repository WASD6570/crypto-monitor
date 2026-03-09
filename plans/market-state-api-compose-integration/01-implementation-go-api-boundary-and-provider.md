# Implementation: Go API Boundary And Provider

## Requirements And Scope

- Add one explicit Go service boundary at `services/market-state-api/`.
- Serve `GET /api/market-state/global`, `GET /api/market-state/:symbol`, and a minimal health endpoint for Compose.
- Keep temporary deterministic state ownership inside Go only.
- Reuse existing Go current-state assembly surfaces instead of serving opaque hand-written JSON blobs where practical.

## Target Repo Areas

- `services/market-state-api`
- `services/feature-engine`
- `services/regime-engine`
- `services/slow-context`
- `libs/go` only if a very small shared helper is clearly justified

## Implementation Notes

- Create a narrow provider interface inside `services/market-state-api`:
  - `CurrentGlobalState(ctx)`
  - `CurrentSymbolState(ctx, symbol)`
- Back the first implementation with deterministic Go-owned fixture/state builders that assemble through:
  - `featureengine.Service.QueryCurrentStateWithSlowContext` for symbol payloads
  - `regimeengine.Service.QueryCurrentGlobalState` for global payloads
- Prefer package-local deterministic builders or API-owned fixture helpers over frontend-derived scenario data.
- Keep supported symbols explicit (`BTC-USD`, `ETH-USD`) and return structured 404 responses for unsupported symbols.
- Add minimal CORS only if needed for non-compose host dev; Compose should prefer same-origin proxying so CORS is not the default path.
- Add `cmd/market-state-api` or equivalent explicit entrypoint rather than relying on tests as the service runner.

## Testing Expectations

- `"/usr/local/go/bin/go" test ./services/market-state-api/...`
- handler tests cover happy path, unsupported symbol, and provider failure
- responses preserve the existing current-state schemas and slow-context shape already consumed by `apps/web`

## Summary

This step creates the backend boundary the dashboard should have been reading all along: a Go-owned API that can start with deterministic data now and swap in real live state later without moving the web trust boundary.
