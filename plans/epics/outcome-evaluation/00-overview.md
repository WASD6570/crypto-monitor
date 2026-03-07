# Outcome Evaluation Overview

## Ordered Implementation Plan

1. Build deterministic outcome-state evaluation for target, invalidation, timeout, and horizon handling
2. Add net-viability boundary fields and cost-model inputs without crossing into simulated execution
3. Define append-only storage and consumer contracts for outcome records, review queries, and baseline comparison
4. Validate replay determinism, version pinning, negative cases, and review-loop usability

## Problem Statement

Every emitted alert needs an objective, replayable answer for what happened next. The system should tell the user whether the idea reached target, failed by invalidation, or simply timed out, and it should do so in a way that survives config changes, replay, and later comparison against baselines.

## Role In Initiative 2

- This is slice 4 of `crypto-market-copilot-alerting-and-evaluation`.
- It closes the first half of the initiative's trust loop: alert generation and delivery create the claim, outcome evaluation records the first objective result.
- It gives later slices a stable source of truth:
  - `simulated-execution` consumes the same alert and market path but adds fill, latency, fee, and slippage assumptions.
  - `operator-feedback-and-notes` and `replay-and-analytics-ui` read the stored outcome record instead of recomputing it.
  - `baseline-comparison-and-tuning` compares production and baseline alerts on the same outcome vocabulary.

## Role In The Review Loop

1. Alert fires with setup, direction, thresholds, and config version.
2. Outcome evaluator watches trusted post-alert market data across 30s, 2m, and 5m horizons.
3. The evaluator writes immutable horizon results plus shared review metrics such as MAE, MFE, time-to-hit, and regime attribution.
4. Review surfaces combine the alert payload, delivery history, outcome record, later simulation record, and operator feedback.
5. Tuning compares outcome quality and net-viability slices against baselines by config version instead of reinterpreting history.

## First-User Outcome

The user should be able to open a recent alert and quickly see:

- whether target, invalidation, or timeout happened first
- how favorable and adverse excursion evolved before close
- how long it took to hit the first decisive condition
- which market regime framed the result
- whether the alert still looked viable under conservative cost assumptions

## In Scope

- pure outcome evaluation for emitted alerts using trusted service-side data
- horizon-specific results for 30s, 2m, and 5m windows
- deterministic ordering rules for target, invalidation, and timeout
- MAE, MFE, time-to-hit, time-to-invalidation, and timeout duration fields
- regime attribution at open and outcome-close boundaries, plus coarse path attribution during the horizon
- conservative net-viability fields attached to outcome records
- config, algorithm, and schema version pinning for replay and review
- storage and consumer contracts for later UI, review, and baseline comparison

## Out Of Scope

- simulated fills, leverage, liquidation, portfolio sizing, or venue-specific execution logic
- live trading, real orders, exchange credentials, or execution auth
- auto-tuning or threshold mutation from outcome data alone
- new alert-generation logic beyond the fields outcome evaluation requires as inputs
- client-side recomputation of outcome decisions

## Safe Defaults For Ambiguous Areas

- Default evaluation basis: use the same trusted composite or canonical market-price source referenced by the alert payload; do not switch sources mid-horizon.
- Default tie policy: if target and invalidation are both first observed in the same ordered event, resolve conservatively in favor of invalidation.
- Default timeout policy: timeout is evaluated only after checking target and invalidation through the horizon end boundary.
- Default horizon policy: 30s, 2m, and 5m outcomes are evaluated independently from the same open event and may disagree.
- Default regime attribution: store regime at alert-open, regime at horizon close, and whether fragmentation/degradation flags were present during the evaluated window.
- Default net-viability posture: if cost inputs or trusted price observations are missing, keep raw outcome fields but mark net viability as `UNKNOWN` instead of inferring profit.

## Requirements

- Keep live outcome evaluation in Go-owned service logic; Python may inspect fixtures offline but cannot be required at runtime.
- Preserve deterministic replay for the same raw inputs, alert payload, config snapshot, and code version.
- Attach `configVersion`, `algorithmVersion`, and contract/schema version references to every stored outcome record.
- Keep outcome evaluation distinct from simulated execution while sharing identifiers and price-path seams cleanly.
- Ensure outcome records are directly comparable across production alerts and baseline alerts by symbol, setup family, horizon, regime, and version context.
- Store outcomes immutably; later replays append corrected records with provenance rather than overwriting history silently.

## Target Repo Areas

- `services/outcome-engine`
- `libs/go`
- `schemas/json/outcomes`
- `schemas/json/alerts`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`
- `tests/parity` for optional offline verification only

## Design Overview

### Boundary: Outcome Evaluation vs Simulated Execution

- Outcome evaluation answers market-truth questions: what levels were reached first, how far price traveled, how long it took, and what regime framed the move.
- Simulated execution answers execution-truth questions: what fills would likely occur after latency, fees, slippage, and instrument rules.
- The outcome record should expose enough shared inputs for simulation later:
  - alert open timestamp and evaluation source
  - path metrics by horizon
  - gross move and excursion fields
  - pinned config and regime context
- The outcome record should not decide final PnL, leverage policy, or fill realism beyond conservative viability fields.

### Relationship To Baselines

- Production alerts and baseline alerts should emit the same minimum evaluation inputs so outcome logic stays identical across both.
- Outcome storage should support slices by `setupFamily`, `baselineId`, `configVersion`, horizon, and regime.
- Baselines should be compared on outcome vocabulary first, then on later simulation results when that slice exists.

### Replay And Versioning

- Replays must read the preserved alert payload, preserved price-path inputs, and preserved config snapshot used at the original evaluation time.
- Outcome ordering must not depend on wall clock, mutable current defaults, or nondeterministic iteration.
- If replay produces a correction, persist a new outcome record linked to the original alert and replay manifest, with a reason and supersession reference.

## ASCII Flow

```text
emitted alert
- alertId
- setup/direction
- target + invalidation
- config/version refs
        |
        v
trusted post-alert market path
- canonical/composite price observations
- event ordering metadata
- regime + fragmentation snapshots
        |
        v
outcome evaluator
- 30s / 2m / 5m windows
- target vs invalidation ordering
- timeout after horizon end
- MAE / MFE / time-to-hit
        |
        +-----------------------------+
        |                             |
        v                             v
pure outcome record           net-viability boundary fields
- horizon result              - conservative fee/slippage view
- decisive timestamp          - gross vs net move summary
- regime attribution          - `POSITIVE` / `NEGATIVE` / `UNKNOWN`
        |                             |
        +-------------+---------------+
                      |
                      v
append-only outcome storage
- config/algorithm/schema versions
- replay provenance
- baseline comparability keys
                      |
          +-----------+------------+
          |                        |
          v                        v
 review UI + notes         later simulated execution
 baseline comparison       reuses path + IDs, adds fill logic
```
