# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Runtime snapshot unit and command tests | `go test ./cmd/market-state-api ./services/venue-binance` | Prove snapshot mapping, lifecycle wiring, and deterministic status propagation across startup, reconnect, stale, recovery, and rate-limit paths | Targeted tests for the new snapshot owner pass without changing the current API contract |
| Deterministic repeated-input proof | `go test ./cmd/market-state-api ./services/venue-binance -count=2` | Catch unstable symbol ordering or nondeterministic snapshot field behavior | Repeated runs stay green with no flaky runtime-health assertions |
| Regression guard for existing API path | `go test ./cmd/market-state-api` | Confirm the internal snapshot seam does not break the sustained Spot runtime owner or current market-state provider wiring | Existing command-level market-state tests still pass |

## Verification Checklist

- Snapshot entries exist for both `BTC-USD` and `ETH-USD`, even before observations become publishable.
- Readiness is distinct from feed-health degradation.
- Shared health states and reasons remain unchanged from the existing runbooks.
- No new public route or `/healthz` behavior change is introduced in this slice.
- Symbol ordering and field semantics remain deterministic across repeated accepted runtime inputs.

## Reporting

- Write results to `plans/binance-runtime-health-snapshot-owner/testing-report.md` while the feature remains active.
- When implementation and validation finish, move the full directory to `plans/completed/binance-runtime-health-snapshot-owner/`.
