# Implementation: Fixtures, Integration, And Replay Proof

## Requirements And Scope

- Target repo areas: `tests/fixtures/events/binance`, `tests/integration`, `tests/replay`, and helper code only where needed.
- Use accepted public Binance Spot depth/top-of-book inputs; do not use mocks or fixture-backed code in the live runtime path.
- Prove deterministic feature output from the same raw inputs and config.
- Keep replay proof local and Go-owned.

## Fixture Expectations

- Reuse existing depth bootstrap and recovery fixtures where practical:
  - `tests/fixtures/events/binance/BTC-USD/happy-native-depth-bootstrap-usdt.fixture.v1.json`
  - `tests/fixtures/events/binance/BTC-USD/happy-native-depth-resync-usdt.fixture.v1.json`
  - existing edge cooldown, rate-limit, snapshot-stale, and sequence-gap fixtures
- Add a dedicated multi-level depth-liquidity fixture when existing fixtures are too shallow to prove depth notional and slippage proxy.
- Include at least one ETH fixture or replay input so symbol ordering and source-symbol validation are not BTC-only.
- Any new fixture should remain an events fixture unless a formal `schemas/json/features` contract is intentionally added in this child.

## Integration Coverage

- Add a Binance Spot depth-liquidity integration test that parses accepted snapshot/delta/top-of-book inputs, builds synchronized book state, evaluates depth-liquidity output, and asserts pinned metrics.
- Extend current-state integration/API tests to prove the Binance contributor liquidity weight uses the observed score instead of a fixed `100`.
- Cover degraded paths:
  - sequence-gap or resync state cannot improve score
  - snapshot stale remains visible and caps/excludes liquidity contribution
  - cooldown/rate-limit recovery block remains machine-readable and does not produce a healthy liquidity score
- Preserve existing depth canonical normalization assertions.

## Replay Proof

- Add a replay test, likely `tests/replay/binance_spot_depth_liquidity_test.go`, that feeds accepted raw snapshot/delta/top-of-book inputs through the same parser and feature evaluator path used by integration tests.
- Run the same replay input twice and compare complete output with `reflect.DeepEqual` or a stable digest.
- Assert stable symbol ordering, score, spread bps, notional, imbalance, slippage proxy, timestamp fallback, and degradation reasons.
- Include one repeated-input or stale-delta case to prove idempotent cache behavior and no duplicate side effects.

## Current-State Regression Proof

- Add or update tests proving current-state remains unavailable during warm-up and reconnect until publishable depth/top-of-book data exists.
- Add or update tests proving public response schemas remain unchanged while derived weight and quality values change.
- Ensure `MarketStateCurrentResponse.Composite.World.Contributors` shows the observed Binance raw/final weight for the happy path.
- Ensure degraded depth cannot increase `MarketQuality.CombinedTrustCap` compared with the healthy baseline for the same symbol/time window.

## Validation Inputs To Hand Off

- No credentials are required for required testing.
- Optional public Binance live validation remains gated by `BINANCE_LIVE_VALIDATION=1`.
- Optional Compose validation remains gated by Docker availability.

## Summary For Next Agent

- Add enough fixture depth to prove notional and slippage math, but keep live runtime free of fixture fallback.
- The key proof is deterministic replay plus current-state regression showing fixed liquidity weight is gone without public schema churn.
