# Implementation: USD-M Runtime Status Visibility

## Requirements And Scope

- Add USD-M status to the existing runtime-status route without changing `/healthz` or current-state response shapes.
- Reuse existing health policy from `venuebinance.USDMRuntime` and `venuebinance.USDMOpenInterestPoller`.
- Keep the response additive and optional for future non-Binance providers.
- Keep readiness semantics unchanged: Spot publishability controls `READY` versus `NOT_READY`; USD-M health explains derivatives context quality.

## Target Files

- `cmd/market-state-api/usdm_influence_owner.go`
- `cmd/market-state-api/runtime_health.go`
- `cmd/market-state-api/live_provider.go`
- `cmd/market-state-api/runtime_health_test.go`
- `cmd/market-state-api/main_test.go`
- `services/market-state-api/api.go`
- `services/market-state-api/api_test.go`
- `services/market-state-api/README.md`

## Implementation Notes

- Add a small command-owned USD-M runtime snapshot method on `binanceUSDMInfluenceOwner`.
- Build the snapshot under existing locks so websocket state, open-interest poll state, and feed-health inputs are read consistently enough for an operator status payload.
- Use `USDMRuntime.State()` and `USDMRuntime.HealthStatus(now)` for websocket status.
- Use `USDMOpenInterestPoller.State()` and `USDMOpenInterestPoller.HealthStatus(sourceSymbol, now)` for per-symbol open-interest status.
- Extend `binanceRuntimeHealthSymbolSnapshot` with an optional USD-M status structure.
- Extend `marketstateapi.RuntimeStatusSymbolResponse` with an additive `USDMStatus` pointer field using the JSON name `usdmStatus` and `omitempty`.
- Keep `normalizeRuntimeStatusResponse` focused on symbol ordering and supported symbols; it should not require `usdmStatus` from all providers.

## Suggested Response Shape

```go
type RuntimeStatusUSDMStatusResponse struct {
    Websocket                 RuntimeStatusFeedHealthResponse `json:"websocket"`
    OpenInterest              RuntimeStatusFeedHealthResponse `json:"openInterest"`
    ConnectionState           ingestion.ConnectionState       `json:"connectionState"`
    ConsecutiveReconnects     int                             `json:"consecutiveReconnects"`
    LastMarkPriceAt           *time.Time                      `json:"lastMarkPriceAt"`
    LastOpenInterestAt        *time.Time                      `json:"lastOpenInterestAt"`
    NextOpenInterestPollAt    *time.Time                      `json:"nextOpenInterestPollAt"`
    OpenInterestRateLimitUntil *time.Time                      `json:"openInterestRateLimitUntil"`
}
```

The exact field names can be adjusted during implementation if the existing API package style suggests a smaller name, but the route must stay additive and machine-readable.

## Unit Test Expectations

- `services/market-state-api` tests prove `usdmStatus` serializes when supplied and the existing handler still rejects missing, duplicate, or unsupported runtime-status symbols.
- `cmd/market-state-api` tests prove warm-up returns two symbols with existing Spot status plus USD-M status.
- `cmd/market-state-api` tests prove USD-M websocket reconnect/stale and open-interest rate-limit states appear in the additive status without changing Spot readiness.
- Repeated runtime-status calls at the same injected time are deeply equal except for fields intentionally driven by test input.

## Contract And Compatibility Impact

- No shared JSON schema currently owns `/api/runtime-status`; do not add one in this feature unless implementation discovers an existing consumer contract that requires it.
- If a schema or generated TypeScript contract is introduced despite the default, update the schema, Go producer, TypeScript consumer validation, and contract tests in the same slice.
- Existing runtime-status consumers must continue to work because all old fields remain unchanged.

## Summary For Next Agent

Start by adding command-owned USD-M snapshot plumbing and the additive `usdmStatus` response field, then prove serialization and deterministic symbol ordering before adding broader failure-sequence tests.
