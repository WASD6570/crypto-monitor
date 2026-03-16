# Implementation Module 2: Live Provider And Regime Assembly Wiring

## Scope

- Wire the active market-state assembly path to evaluate and apply the settled USD-M signal during symbol/global current-state construction.
- Reuse the existing venue-side input owner and Binance runtime pieces instead of inventing new semantics in the HTTP layer.
- Cover the smallest command/provider seams needed to feed live USD-M context into `services/market-state-api`.

## Target Repo Areas

- `services/market-state-api`
- `services/regime-engine`
- `services/feature-engine`
- `services/venue-binance`
- `cmd/market-state-api`

## Requirements

- Keep the current Spot assembly path intact for the baseline inputs already used by `/api/market-state/*`.
- Add a narrow read-only USD-M seam so the provider can obtain the latest accepted influence input or signal for `BTC-USD` and `ETH-USD`.
- Reuse the existing `USDMInfluenceInputOwner` and `EvaluateUSDMInfluence` logic rather than reimplementing influence policy in `cmd/market-state-api` or the HTTP handler.
- Apply the adjusted symbol regimes before global regime evaluation so global output reflects the same bounded semantics.
- Keep `/healthz`, runtime-status policy, and unrelated spot recovery work out of scope.

## Key Decisions

- Prefer exposing a snapshot reader from the command/runtime side rather than teaching the HTTP package to own Binance USD-M polling or websocket lifecycle.
- Keep USD-M signal evaluation inside Go-owned market-state assembly so deterministic provider fixtures and live providers can share the same helper path.
- If the deterministic provider cannot reuse the live reader seam directly, add a fixture-friendly helper that feeds pinned USD-M inputs through the same application logic.
- Keep symbol ordering deterministic everywhere: `BTC-USD` first, `ETH-USD` second.

## Unit Test Expectations

- Live provider tests cover both unchanged auxiliary/no-context behavior and degrade-cap application on current-state responses.
- Global regime output reflects capped symbol snapshots when both symbols share the same applied cap condition.
- Command/provider seams reject missing or unsupported USD-M inputs cleanly rather than silently mutating output.
- Repeated bundle reads with the same accepted inputs produce identical symbol/global responses.

## Contract / Fixture / Replay Impacts

- No new public route should be added.
- Existing live-provider integration fixtures need paired Spot plus USD-M inputs so the new consumer-facing behavior can be exercised intentionally.
- Any helper added here must preserve deterministic timestamp handling and repeated-run stability.

## Summary

This module makes the previously internal USD-M signal real in the live current-state path while keeping transport ownership, policy evaluation, and HTTP serialization in their existing bounded homes.
