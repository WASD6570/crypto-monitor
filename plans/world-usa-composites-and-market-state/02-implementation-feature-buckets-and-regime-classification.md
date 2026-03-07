# Implementation Feature Buckets And Regime Classification

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `services/regime-engine`, `libs/go`, `configs/*`, `schemas/json/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Transform composite snapshots and per-venue health inputs into 30s, 2m, and 5m feature buckets.
- Classify 5m symbol and global market state as `TRADEABLE`, `WATCH`, or `NO-OPERATE`.
- Make fragmented regime handling, degraded inputs, and determinism explicit.

## In Scope

- feature bucket definitions and bucket-assignment behavior
- divergence and fragmentation metrics between WORLD and USA
- market-quality metrics that consume composite and feed-health outputs
- 5m regime classification and downgrade/ceiling rules

## Out Of Scope

- alert entry logic or trade setup confirmation
- UI explanation copy and rendering details
- slow-context rules that would make CME or ETF feeds hard dependencies for realtime state

## Bucket Strategy

- 30s bucket: fast tape quality and short-horizon disagreement detection.
- 2m bucket: short confirmation layer for persistence versus noise.
- 5m bucket: primary regime gate for operator trust and later alert gating.
- Bucket assignment follows operating defaults: event time first, `recvTs` fallback with degraded marker, explicit late-event handling.

## Recommended Feature Families

### 30s Features

- composite return or directional delta for WORLD and USA
- cross-composite spread or basis distance
- freshness and coverage deltas versus prior bucket
- micro-fragmentation indicators such as sign disagreement, abrupt contributor churn, and spread expansion

### 2m Features

- smoothed WORLD and USA directional agreement
- persistence of fragmentation versus one-bucket noise
- market-quality stabilization or deterioration trend
- count and duration of degraded contributors

### 5m Features

- sustained divergence magnitude and duration
- composite market-quality summary
- venue coverage ratio summary
- fragmentation severity bucket
- regime-ready confidence score or equivalent bounded summary used only inside service-owned logic

## Divergence Metrics Recommendations

- Define divergence as a family of metrics, not a single scalar.
- Minimum recommended metrics:
  - price divergence: normalized distance between WORLD and USA composite prices
  - directional divergence: disagreement in return sign and magnitude over 30s and 2m windows
  - participation divergence: difference in contributor coverage and health between groups
  - leader-churn divergence: instability in top-weight venue composition
- Keep the final regime engine free to combine these metrics via config instead of hard-coding one opaque formula in contracts.

## Market-Quality Metrics Recommendations

- Market quality should answer whether the operator can trust the apparent tape, not whether volatility is high or low.
- Minimum recommended inputs:
  - feed freshness and gap state
  - timestamp trust
  - venue coverage ratio
  - concentration after clamping
  - spread/order-book quality proxies when available
  - stability of composite contributors over recent buckets
- Safe default: any critical degradation in freshness, sequence integrity, or coverage should cap quality and prevent `TRADEABLE`.

## Fragmented Regime Handling

- Fragmented state means the venue groups disagree enough that the operator should distrust clean directional interpretation.
- Recommended classification ladder:
  - low fragmentation: composites aligned and well-covered
  - moderate fragmentation: disagreement or missing coverage pushes regime toward `WATCH`
  - severe fragmentation: sustained disagreement, degraded coverage, or unstable leadership pushes regime to `NO-OPERATE`
- Fragmentation should be durable over more than one 30s bucket before forcing a state flip, unless feed-health failure is critical.

## 5m Regime Classification Recommendations

- Symbol regime is the primary gate for BTC and ETH.
- Global regime is a ceiling over symbol regimes per operating defaults.
- Safe default state semantics:
  - `TRADEABLE`: composites sufficiently aligned, coverage acceptable, no critical degradation, quality above configured threshold
  - `WATCH`: mixed or uncertain conditions, moderate fragmentation, or partial degradation where context is still useful
  - `NO-OPERATE`: severe fragmentation, critical feed degradation, insufficient coverage, or unavailable trustworthy composite state
- Recommended transition posture:
  - degrade quickly on critical trust loss
  - recover more conservatively after persistence across multiple buckets
  - record explicit transition reasons and triggering metrics

## Global Ceiling Rules

- If both symbols degrade severely because shared venue groups are unhealthy, global state becomes `NO-OPERATE`.
- If one symbol is healthy and the other is degraded, keep symbol states separate unless shared market-quality inputs justify a global downgrade.
- Slow USA context may later inform explanation, but should not be required to emit a realtime 5m state in MVP.

## Negative And Edge Cases

- WORLD available but USA unavailable
- USA clean but WORLD fragmented by perp/spot disagreement
- live late events that land after watermark and only correct on replay
- oscillation around a threshold causing regime flapping
- missing 30s buckets inside a 5m window
- replay run using older config snapshot and producing intentionally different results than current config

## Determinism Notes

- Bucket closure, late-event handling, and state transitions must be deterministic for the same event order and config snapshot.
- Any smoothing or persistence rule must use explicit bounded windows, not wall-clock timers.
- Tie-breakers for threshold edges should be documented and stable.
- Replay should preserve both the emitted state and the reasons so differences are explainable rather than guessed.

## Unit And Integration Test Expectations

- unit tests for bucket assignment with `exchangeTs` and `recvTs` fallback
- unit tests for divergence metric calculations on aligned and fragmented fixtures
- unit tests for regime downgrade and recovery hysteresis
- integration tests for degraded feed inputs propagating to market-quality and regime state
- replay tests for late-event correction and deterministic 5m state reproduction

## Summary

This module defines the feature buckets, divergence and market-quality metrics, and the final 5m tradeability gate. The next agent should preserve the bias toward conservative trust handling: fragmented or degraded conditions must lower confidence quickly, while recovery requires stable evidence over multiple deterministic buckets.
