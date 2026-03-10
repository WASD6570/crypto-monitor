# Testing Report: Binance Live Current State Query Assembly

## Result

- Status: passed
- Scope: live query source seam, Spot-driven current-state assembly, integration coverage, and replay-style determinism proof

## Commands

| Command | Purpose | Result |
|---|---|---|
| `"/usr/local/go/bin/go" test ./services/market-state-api -run 'Test(Deterministic|Live|Handler|CurrentState).*' -v` | validate provider seam, live Spot source assembly, unsupported-symbol handling, and handler behavior | Passed |
| `"/usr/local/go/bin/go" test ./tests/integration -run 'Test(MarketStateCurrent|IngestionBinance.*CurrentState).*' -v` | validate Binance-backed current-state symbol/global responses and degradation visibility | Passed |
| `"/usr/local/go/bin/go" test ./tests/replay -run 'Test(Replay.*MarketState|MarketStateCurrentReplay).*' -v` | validate repeated-run determinism for pinned current-state inputs | Passed |

## Notes

- `services/market-state-api` now supports a replaceable Spot live provider without changing the existing HTTP handler contract.
- The first live assembly posture keeps `usa` explicit as unavailable while preserving the current response schema and reserved history seam.
- Degraded feed-health and unsynchronized depth posture remain machine-readable in assembled current-state output.
- Validation used pinned deterministic inputs only; no live Binance network access was required.
