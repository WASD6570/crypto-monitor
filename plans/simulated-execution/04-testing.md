# Simulated Execution Testing

## Testing Goals

- prove deterministic simulation outputs for pinned fixtures and replay inputs
- prove spot long and perp long/short coverage, including leverage up to `x5`
- prove degraded-data behavior chooses fallback only when safe and refuses otherwise
- prove saved-run persistence and operator surfaces preserve auditability and explicit simulated labeling
- prove the feature cannot blur into live trading or credential handling

## Expected Report Artifact

- Write the implementation-phase test report to `plans/simulated-execution/testing-report.md`.

## Fixture And Environment Expectations

- Use deterministic replay fixtures with pinned alerts, outcome records, venue health snapshots, quote snapshots, and L2 snapshots where available.
- Keep review notionals fixed per preset so fill comparisons are reproducible.
- Prefer Go-owned tests and replay runners; any Python parity analysis stays optional and offline-only.

## High-Signal Validation Matrix

### 1. Deterministic replay smoke

- Goal: the same alert, outcome record, preset, and fixture set produces the same simulation output and confidence label twice.
- Command: `go test ./services/simulation-api/... ./tests/replay/...`
- Verify: identical entry/exit estimates, net result, reason codes, confidence label, and version references.

### 2. Mode coverage smoke

- Goal: `spot-long`, `perp-long`, and `perp-short` each produce valid outputs on healthy fixtures.
- Command: `go test ./services/simulation-api/... -run TestSimulationModeCoverage`
- Verify: mode-specific side handling is correct, spot stays `x1`, perps accept `x1/x2/x3/x5`.

### 3. L2 walk vs fallback vs refusal

- Goal: healthy L2 uses depth walk, partially degraded L2 uses deterministic fallback with lowered confidence, and critically degraded L2 refuses.
- Command: `go test ./services/simulation-api/... -run TestSimulationL2DecisionLadder`
- Verify: `slippageMethod`, confidence label, and refusal reason codes match fixture expectations.

### 4. Fees, latency, and leverage math

- Goal: cost inputs visibly affect net result and `x5` remains supported but conservatively labeled when assumptions weaken.
- Command: `go test ./services/simulation-api/... -run TestSimulationEconomics`
- Verify: net result decomposition is stable, latency shifts entry time correctly, unsupported leverage is rejected.

### 5. Persistence and retrieval smoke

- Goal: successful, low-confidence, and refused runs are all stored append-only and retrievable by alert.
- Command: `go test ./services/simulation-api/... ./tests/integration/... -run TestSimulationPersistence`
- Verify: no overwrites, all audit fields present, saved runs list matches inserted records.

### 6. Operator surface smoke

- Goal: the web surface renders simulation as hypothetical, shows confidence, and exposes refused reasons clearly.
- Command: `pnpm --dir apps/web test -- --run SimulatedExecution`
- Verify: explicit `SIMULATED` labeling, no live-trading wording, side-by-side preset comparison renders correctly.

## Negative Cases

- missing `alertId` or missing `outcomeRecordId` -> request rejected with stable validation error
- unsupported mode such as `spot-short` -> request rejected
- unsupported leverage such as `x10` -> request rejected
- no trusted venue/instrument mapping -> run refused, not auto-remapped to an arbitrary pair
- missing fee table for selected venue/mode -> run refused
- critically degraded L2 with no safe quote fallback -> run refused
- timestamp ordering beyond safe audit tolerance -> run refused or clearly marked `LOW_CONFIDENCE` per pinned rule
- attempt to create a run with exchange credentials or live-order fields in payload -> request rejected and audited
- duplicate saved-run request with same identifiers and preset -> either idempotent same result or distinct append-only rerun per final contract, but behavior must be explicit and tested

## Determinism And Audit Checks

- replay the same day/symbol fixture twice and compare serialized run payloads excluding append-only record IDs and write timestamps
- confirm every run stores `configVersion`, `algorithmVersion`, schema version, and timestamp-source provenance
- confirm fallback runs include confidence reason codes and that healthy runs do not inherit fallback labels accidentally
- confirm refused runs still persist enough audit metadata to explain why no result was produced

## Cross-Feature Checks

- verify simulation reuses outcome-evaluation identifiers and does not recompute target/invalidation ordering independently
- verify baseline alerts can be simulated with the same presets and confidence vocabulary as production alerts
- verify alert review queries can join alert, outcome, and simulation records within the operating-default query target posture

## Exit Criteria For Implementation

- deterministic replay checks pass for healthy and degraded fixtures
- all supported modes and leverage options through `x5` pass fixture coverage
- fallback vs refusal behavior is stable and auditable
- saved runs and refused runs are append-only, queryable, and clearly labeled as simulated
- no endpoint, contract, or UI path implies live trading or exchange credential use
