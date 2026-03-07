# Implementation: Baseline Definitions And Evaluation Joins

## Module Requirements And Scope

- Define the deterministic baseline families required by program success criteria.
- Make baseline alerts comparable to production alerts without forcing identical internal logic.
- Specify the join model that lets outcomes, simulations, and reports compare production and baseline behavior on the same windows.
- Keep joins stable under replay and config version changes.

## Target Repo Areas

- `services/alert-engine`
- `services/outcome-engine`
- `services/replay-engine`
- `libs/go`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `schemas/json/replay`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Baseline Registry

- Represent baselines as explicit registry entries beside production setup families, not as ad hoc report-only calculations.
- Each baseline should carry stable metadata: `baselineId`, mapped production comparison family, enabled symbols, evaluation source, and versioned parameter bundle.
- Keep baseline outputs service-owned so outcome evaluation reads them exactly like production alerts.

### 2. Baseline Family Expectations

- `naive-breakout` should trigger on a simple reference-level break with no 2m validator and no 5m regime gate.
- `naive-vwap-reversion` should trigger on fixed-distance composite VWAP deviation with direct reversion assumption and no richer context filters.
- `single-venue-trigger` should trigger from one designated venue feed with no WORLD vs USA composite confirmation and no fragmentation-aware suppression.
- Baselines should remain intentionally naive, but still use trusted canonical event ordering and versioned parameters so replay stays fair.

### 3. Comparable Alert Payload Minimums

- Production and baseline alerts should both provide: `alertId`, `symbol`, `direction`, `openedAt`, target and invalidation fields, regime context at open, `configVersion`, `algorithmVersion`, and evaluation source identifiers.
- Baseline-specific fields should be additive, such as `baselineId` and baseline parameter references.
- Client surfaces and reports should not infer missing fields differently for baseline versus production records.

### 4. Evaluation Join Model

- Join on comparable dimensions rather than exact event identity.
- Recommended join keys:
  - `symbol`
  - `direction`
  - production `setupFamily` mapped to a baseline comparison family
  - `baselineId` when present
  - `openBucket` using the replay-pinned bucket policy
  - `horizon`
  - `evaluationSource`
  - regime slices such as open regime, fragmentation flag, and degraded flag
  - `configVersion` for both production and baseline snapshots
- When multiple alerts land in one bucket, preserve one-to-many linkage and aggregate at the episode level instead of forcing a lossy one-to-one match.

### 5. Comparable Episode Rules

- Define a comparable episode as the smallest shared evaluation window in which production and baseline alerts can be assessed on the same market path.
- Episode construction should be deterministic and based on pinned bucket windows or cluster windows, not human labels.
- If production emits no alert but baseline does, keep the baseline-only episode for false-positive accounting.
- If production emits and baseline does not, keep the production-only episode so improvement is not hidden.

### 6. Outcome And Simulation Join Expectations

- Outcome evaluation should consume baseline alerts without special-case logic beyond `baselineId` metadata.
- Simulation should reuse the same alert-open and price-path seams so baseline and production net viability remain comparable when simulation data exists.
- Report aggregates should permit outcome-only comparison even when simulation is unavailable for some windows.

### 7. Determinism Notes

- Baseline evaluators must use the same canonical event ordering, timestamp rules, and replay manifests as production alerts.
- Joins must never depend on mutable current defaults, wall-clock query time, or nondeterministic record ordering.
- A replay with identical inputs and snapshots must regenerate the same comparable episodes and aggregate counts.

## Unit And Integration Test Expectations

- unit tests for baseline registry loading and parameter version resolution
- unit tests for each baseline evaluator on clean positive and negative fixtures
- integration tests for production-only, baseline-only, and shared comparable episodes
- replay tests proving stable comparable-episode construction under pinned snapshots
- contract tests proving baseline alerts pass through the same outcome evaluator interface as production alerts

## Summary

This module defines the honest comparison surface for tuning: explicit naive baselines, production-compatible alert payloads, and deterministic joins that preserve both wins and misses. The next module should assume these comparable records exist before introducing candidate snapshots and graduation rules.
