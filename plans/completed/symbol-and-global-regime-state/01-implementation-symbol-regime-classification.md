# Implementation Symbol Regime Classification

## Module Requirements And Scope

- Target repo areas: `services/regime-engine`, `libs/go/features`, `configs/local`, `configs/dev`, `configs/prod`, `schemas/json/features` only if a regime seam schema is reserved now, `tests/fixtures`, `tests/integration`, `tests/replay`
- Consume completed 5m bucket summaries from the feature-engine layer without recomputing divergence, fragmentation, coverage, or market-quality math.
- Emit symbol-level 5m state for `BTC-USD` and `ETH-USD` as `TRADEABLE`, `WATCH`, or `NO-OPERATE` with explicit reason outputs and transition metadata.
- Keep the module focused on symbol evaluation only; global ceiling behavior belongs in the next module.

## In Scope

- symbol-level classification inputs and threshold evaluation
- downgrade and recovery hysteresis for a single symbol across consecutive closed 5m windows
- explicit trigger-reason families and transition metadata
- unavailable or incomplete bucket handling
- deterministic tie-break rules for threshold edges
- config-versioned thresholds and persistence windows for symbol state

## Out Of Scope

- cross-symbol global ceiling policy
- current-state query response shapes or route design
- persistence, audit-read retrieval, or replay control-plane behavior
- alert-engine entry logic or risk-engine position policy

## Target Structure

- `libs/go/features/regime.go`
- `libs/go/features/regime_test.go`
- `services/regime-engine/service.go`
- `services/regime-engine/service_test.go`
- `configs/*/regime-engine.market-state.v1.json` or an equally explicit regime-engine config file
- fixture windows under `tests/fixtures/world_usa_regime/`

## Classification Inputs

Read only completed 5m bucket summaries and related provenance seams such as:

- `fragmentation.severity`
- `fragmentation.persistenceCount`
- `marketQuality.combinedTrustCap`
- `marketQuality.downgradedReasons`
- `world.coverageRatio` and `usa.coverageRatio`
- `timestampTrust` fallback and cap flags
- bucket completeness markers such as `missingBucketCount` and `closedBucketCount`
- composite unavailable markers coming from the bucket layer

This module may combine those inputs deterministically, but it must not reopen upstream formulas.

## Recommended Internal Types

- `RegimeState`: `TRADEABLE`, `WATCH`, `NO-OPERATE`
- `RegimeTriggerReason`: explicit reason families such as `fragmentation-severe`, `coverage-low`, `timestamp-trust-loss`, `composite-unavailable`, `market-quality-cap`, `late-window-incomplete`
- `RegimeDecisionInput`: symbol, bucket window metadata, bucket-derived severity fields, config/version refs
- `RegimeTransition`: prior state, next state, downgrade or recovery path, reason set, effective bucket end
- `SymbolRegimeConfig`: thresholds, downgrade precedence, recovery persistence counts, inclusive/exclusive edge rules
- `SymbolRegimeSnapshot`: emitted symbol state plus reasons, trigger metrics, and provenance

## State Semantics

- `TRADEABLE`: 5m summaries show acceptable coverage, aligned or low-fragmentation conditions, no critical timestamp-trust loss, and no severe market-quality cap.
- `WATCH`: moderate fragmentation, partial degradation, incomplete confidence, or recovery-in-progress conditions where the tape is still informative but not fully trustworthy.
- `NO-OPERATE`: severe fragmentation, unavailable composite state, critical trust loss, or insufficient coverage/completeness for a reliable market decision.
- Safe default posture: when multiple conditions disagree, choose the more conservative state.

## Decision Order

1. Reject obviously unusable inputs first: unavailable composite side, missing critical 5m completeness, or explicit severe trust-cap markers.
2. Evaluate hard-stop conditions that force `NO-OPERATE`.
3. Evaluate downgrade-to-`WATCH` conditions for moderate fragmentation, coverage asymmetry, or timestamp-trust degradation.
4. Allow `TRADEABLE` only when no hard-stop or watch-cap condition is present and recovery persistence rules are satisfied.
5. Record the ordered reason set and the specific bucket-derived metrics that triggered the final state.

## Downgrade And Recovery Posture

- Downgrade should happen on the first qualifying closed 5m window for severe trust loss.
- Recovery should require consecutive healthy closed 5m windows, with stricter persistence for `NO-OPERATE -> WATCH` and `WATCH -> TRADEABLE` than for downgrade.
- Hysteresis must be implemented as explicit counts of closed windows, not elapsed wall time.
- Equality edges must be stable and documented once in config so replay and live paths match exactly.

## Explicit Reason Outputs

Each symbol output should include:

- `state`
- `reasons`: ordered, machine-readable reason codes
- `primaryReason`
- `triggerMetrics`: minimal numeric or enum evidence tied to the reasons
- `previousState`
- `transitionKind`: `enter`, `hold`, `downgrade`, `recover`, or similarly explicit values
- `effectiveBucketEnd`
- `configVersion`, `algorithmVersion`, and schema/version seam

Reason codes should be boring and bounded. Prefer a small stable vocabulary over free-form strings.

## Fixture Guidance

- clean aligned 5m window for `BTC-USD`
- moderate fragmentation with partial coverage loss for `ETH-USD`
- severe fragmentation plus degraded timestamp trust
- composite unavailable on one side
- threshold-edge case that would flap without hysteresis
- recovery sequence requiring multiple healthy consecutive windows

## Unit Test Expectations

- `TestRegimeClassification` for clean, mixed, and hard-stop symbol states
- `TestRegimeDowngradePrecedence` for multiple simultaneous downgrade reasons
- `TestRegimeRecoveryRequiresPersistence` for conservative recovery windows
- `TestRegimeThresholdEdgesAreDeterministic` for inclusive/exclusive edge handling
- `TestRegimeReasonsIncludeTriggerMetrics` for explainable outputs

## Integration And Replay Expectations

- integration tests should prove `services/regime-engine` reads bucket summaries as inputs rather than recomputing upstream math
- integration tests should cover both `BTC-USD` and `ETH-USD` state transitions on the same bucket-family seam
- replay tests should run the same pinned window twice and compare full symbol regime outputs including ordered reasons and provenance
- replay tests should show config-version changes alter outputs intentionally while preserving explainability

## Downstream Seams Only

- Later `market-state-current-query-contracts` may package symbol regime snapshots as read models.
- Later alert, risk, and dashboard consumers should read emitted symbol state and reasons directly.
- No consumer in this slice should require client-side or downstream service recomputation of regime logic.

## Summary

This module owns the conservative symbol trust gate. It turns completed 5m bucket evidence into one deterministic symbol state with explicit reasons, fast downgrades, and slower replay-safe recovery.
