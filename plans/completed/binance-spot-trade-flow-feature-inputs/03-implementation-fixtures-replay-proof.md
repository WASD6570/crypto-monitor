# Implementation: Fixtures And Replay Proof

## Scope

- Target areas: `tests/fixtures`, `tests/integration`, `tests/replay`, `services/replay-engine` only if replay needs a minimal feature-output seam, and `schemas/json/features` only if a formal additive feature contract is required in this child.
- Prove that the new internal trade-flow feature inputs can be rebuilt deterministically from accepted Binance Spot trade records.

## Fixture Requirements

- Reuse existing Binance native trade fixtures where possible:
  - `tests/fixtures/events/binance/BTC-USD/happy-native-trade-usdt.fixture.v1.json`
  - `tests/fixtures/events/binance/ETH-USD/edge-native-timestamp-degraded-trade-usdt.fixture.v1.json`
- Add only the smallest new fixture data needed for multi-trade aggregation, duplicate suppression, and buy/sell imbalance.
- Keep fixture symbols limited to `BTC-USD` and `ETH-USD`.
- Fixture validation must stay deterministic through `make fixtures-validate`.

## Integration Proof

- Extend or add Binance integration tests that feed accepted Spot trade frames through the supervisor/parser/normalizer path into the feature model.
- Verify aggregate metrics, source identity dedupe, timestamp fallback counts, and stable symbol ordering.
- Include a negative case for unsupported source symbol or invalid trade numeric fields if that boundary is implemented in the runtime owner.
- Existing Spot depth and runtime current-state tests must continue passing; trade-flow input availability must not make unpublishable top-of-book/depth state appear ready.

## Replay Proof

- Add a replay test that starts from accepted raw Binance `market-trade` entries and rebuilds trade-flow buckets twice, comparing full bucket output or a stable digest.
- Include duplicate raw trade evidence so replay proves duplicate suppression is deterministic.
- Include timestamp-degraded trade evidence so replay proves bucket assignment and degraded counts are stable.
- Preserve existing Binance family determinism tests for raw partition ordering and do not weaken current replay checks.

## Contract Impact

- Preferred posture: no public JSON schema change in this child.
- If implementation creates a durable feature artifact under `schemas/json/features`, it must be additive, versioned, included in `schemas/json/features/family.v1.json`, backed by fixtures, and validated with `make contracts-validate` plus `make fixtures-validate`.
- Do not add dashboard decoders or frontend view models in this child.

## Summary For Next Agent

Use fixtures to prove behavior that unit tests cannot cover: supervisor handoff, replay determinism, duplicates, timestamp degradation, and current-state compatibility. Keep schema changes optional and only add them if the implementation truly emits a persisted/shared feature artifact.
