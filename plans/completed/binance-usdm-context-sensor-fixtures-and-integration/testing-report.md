# Testing Report: binance-usdm-context-sensor-fixtures-and-integration

## Environment
- Target: local Go test environment
- Date/time: 2026-03-09
- Commit/branch: working tree on `main`

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|
| Mixed-surface happy path | normalize websocket-derived funding/mark/liquidation plus REST open-interest fixtures together | canonical events preserve provenance across both acquisition modes | Integration tests passed | PASS |
| Distinct health visibility | websocket stays fresh while REST polling goes stale for the same symbol | separate canonical feed-health outputs remain machine-readable | Integration tests passed | PASS |
| Mixed degradation visibility | websocket reconnect pressure and REST rate limiting occur in the same validation run | degradation reasons remain distinct instead of collapsing into one generic sensor state | Integration tests passed | PASS |
| Replay-sensitive identity | repeat websocket and REST normalization with raw writing enabled | stable `sourceRecordId` values and duplicate facts per stream family | Integration and raw-boundary tests passed | PASS |

## Execution Evidence
### runbook-alignment
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionRunbookAlignmentUsesSharedHealthVocabulary' -v`
- Expected: runbooks reference the shared health vocabulary plus the settled `rate-limit` reason and mixed-surface Binance USD-M checks
- Actual: targeted runbook alignment test passed
- Verdict: PASS

### mixed-surface-integration
- Command/Request: `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM' -v`
- Expected: Binance USD-M integration covers fixture-backed happy path, timestamp degradation, mixed websocket/REST health visibility, and replay-sensitive identity stability
- Actual: all targeted integration tests passed
- Verdict: PASS

### normalizer-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OpenInterest|Funding|MarkIndex|Liquidation)' -v`
- Expected: service-owned normalization remains compatible while the new mixed-surface proof uses it
- Actual: all targeted tests passed
- Verdict: PASS

### raw-boundary-smoke
- Command/Request: `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestRawWriteBoundaryPersistsOpenInterestStreamFamily' -v`
- Expected: REST open-interest raw append entries keep their own stream family and partition routing during replay-sensitive proof
- Actual: targeted raw-boundary test passed
- Verdict: PASS

## Side-Effect Verification
### fixture-corpus
- Evidence: `tests/fixtures/events/binance/ETH-USD/happy-open-interest-usdt.fixture.v1.json` and `tests/fixtures/manifest.v1.json`
- Expected state: Binance USD-M fixture corpus covers the mixed-surface REST path for both tracked symbols without removing prior websocket coverage
- Actual state: ETH happy open-interest fixture and manifest entry were added and exercised by integration tests
- Verdict: PASS

### runbook-vocabulary
- Evidence: `docs/runbooks/ingestion-feed-health-ops.md`, `docs/runbooks/degraded-feed-investigation.md`, `docs/runbooks/binance-usdm-context-sensors.md`
- Expected state: operator docs preserve `HEALTHY`, `DEGRADED`, `STALE`, and the shared degradation reasons including `rate-limit`
- Actual state: runbooks and alignment test now include the settled mixed-surface vocabulary
- Verdict: PASS

### mixed-surface-semantics
- Evidence: `TestIngestionBinanceUSDMMixedSurfaceDistinctHealthVisibility`, `TestIngestionBinanceUSDMMixedSurfaceDegradedReasonsStayDistinct`
- Expected state: websocket and REST health signals stay separate and machine-visible for the same symbol
- Actual state: tests passed with websocket `runtime:binance-usdm-ws:` and REST `runtime:binance-usdm-open-interest:` source IDs remaining distinct
- Verdict: PASS

### replay-sensitive-identity
- Evidence: `TestIngestionBinanceUSDMMixedSurfaceDuplicateSourceIdentitiesStayStable`
- Expected state: repeated websocket funding and REST open-interest normalization keep stable `sourceRecordId` values and duplicate facts within their own stream families
- Actual state: tests passed with stable identities and duplicate audit occurrence `2` for both `funding-rate` and `open-interest`
- Verdict: PASS

## Blockers / Risks
- No blocker found in the current smoke matrix.
- The next remaining Binance live work should shift from proof coverage toward whichever live adapter entrypoint consumes these completed USD-M slices.

## Next Actions
1. Wire the completed USD-M websocket and REST slices into the eventual live Binance adapter entrypoint or raw/replay path that will consume them.
2. If desired, follow with the Spot trade/top-of-book canonical handoff or later raw-storage/replay work that depends on these settled Binance semantics.
