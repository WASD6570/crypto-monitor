# Implementation: Setup Definitions And Signal Gating

## Module Requirements And Scope

- Define a service-owned setup registry for `A`, `B`, and `C` inside Go alert-generation code.
- Specify how 30s nomination, 2m validation, and 5m market-state permissioning interact for each setup family.
- Keep the plan at deterministic rule-shape level; do not lock final production thresholds without replay evidence.
- Record separate decision layers for setup qualification, market-state gating, and future risk-state gating.

## Target Repo Areas

- `services/alert-engine`
- `services/feature-engine`
- `services/regime-engine`
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Setup Registry

- Represent each setup family as explicit config plus deterministic evaluator code, not ad hoc condition trees spread across handlers.
- Keep setup metadata stable across services: `setupKey`, family description, enabled symbols, eligible sides if later needed, default severity ceiling, and required evidence horizons.
- Prefer a registry structure that makes later replay reports show which evaluator and config bundle decided the alert.

### 2. Horizon Responsibilities

- 30s layer should only nominate a setup candidate and produce evidence fields, never final urgency by itself.
- 2m layer should confirm persistence, reject obvious fakeouts, and consume fragmentation and degraded-input penalties.
- 5m layer should apply symbol state plus global ceiling from the regime engine.
- Use a monotonic decision ladder: `candidate -> validated -> permitted -> emitted`.

### 3. Setup Family Expectations

- `A` should plan for rapid-move opportunity detection with strict 2m confirmation to avoid naive-breakout behavior.
- `B` should plan for reversal or reclaim conditions where confirmation should explicitly inspect failed continuation and absorption-like evidence from service-owned features.
- `C` should plan as a more context-heavy family that may start as `INFO` or `WATCH` under the MVP until replay evidence proves stronger actionability.
- Do not define final indicator formulas here; require them to come from feature-engine outputs or alert-engine derived values that remain deterministic and replay-safe.

### 4. Market-State Gating

- Consume the 5m symbol state and global ceiling exactly as emitted by Initiative 1.
- Safe default mapping:
  - symbol `TRADEABLE` + global `TRADEABLE` can allow full setup severity ceiling
  - any `WATCH` state caps severity at `WATCH`
  - any `NO-OPERATE` state caps severity at `INFO` or full suppression depending on explanation value
- Require the alert decision record to include both raw state inputs and the applied gate result.

### 5. Future Risk-State Hook

- Reserve a second gate stage after market-state permissioning and before final emission.
- Risk-state integration should not change setup detection inputs; it only caps or suppresses emission severity.
- Persist gate placeholders now so later implementation can add `riskStateBeforeEmission`, `riskCapApplied`, and `riskBlockReason` without breaking consumers.

### 6. Explainability Rules

- Every candidate and final alert should map to a bounded set of reason codes, not free-form live prose.
- Reason codes should distinguish setup evidence, validator evidence, state caps, and degraded-data suppression.
- Preserve both positive reasons and blocking reasons so later UI and outcome review can explain why an event did or did not reach the user.

## Data And Config Notes

- Keep setup enablement, evidence requirements, and horizon-specific thresholds in versioned config.
- Separate config namespaces for setup logic and gating policy so later tuning can compare them independently.
- Avoid config that lets clients or delivery surfaces reinterpret severity or permissioning.

## Unit And Integration Test Expectations

- unit tests for registry loading, per-setup gating tables, and decision transitions
- integration tests for 30s candidate rejected by 2m validator
- integration tests for 2m-valid setup capped by symbol/global 5m state
- replay tests proving the same event sequence emits the same decision ladder under the same config

## Summary

This module defines the alert decision skeleton: explicit setup families, strict 30s/2m/5m responsibilities, market-state gating now, and a clean insertion point for later risk-state gating. The next implementation slice should assume these decision stages exist before adding dedupe and alert-volume hygiene.
