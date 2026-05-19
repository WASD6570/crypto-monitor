# Binance Spot Trade Flow Feature Inputs Testing Report

Date: 2026-05-19
Status: passed and archived

## Tested Flows

- Internal Go-owned Binance Spot trade-flow aggregation for `BTC-USD` and `ETH-USD` accepted trade observations.
- Feature-engine wrapper observation and snapshot behavior.
- Sustained Spot runtime owner recording of accepted websocket `trade` frames into the internal read model.
- Integration handoff from Binance Spot supervisor/parser/normalizer through trade-flow input recording.
- Replay determinism for accepted raw Binance Spot trade inputs, including duplicate and timestamp-degraded evidence.
- Existing market-state and runtime response compatibility through regression coverage.

## Run-Specific Inputs

- No `test-helpers/` assets were present for this feature.
- Used checked-in Binance Spot trade fixtures, including `tests/fixtures/events/binance/BTC-USD/happy-native-trade-flow-usdt.fixture.v1.json`.
- Used the planned local deterministic validation matrix now archived at `plans/completed/binance-spot-trade-flow-feature-inputs/04-testing.md`.

## Command Results

| Check | Command | Result |
|---|---|---|
| Feature model and service tests | `go test ./libs/go/features ./services/feature-engine` | Passed |
| Binance runtime read-model tests | `go test ./cmd/market-state-api -run 'TestBinanceSpotRuntimeOwner.*TradeFlow|TestBinanceSpotRuntimeOwner.*Publishes|TestBinanceSpotRuntimeOwner.*Reconnect'` | Passed |
| Binance trade integration tests | `go test ./tests/integration -run 'TestIngestionBinanceSpotTrade|TestIngestionBinanceCurrentState|TestBinanceSpotSupervisor'` | Passed |
| Replay determinism | `go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism|TestReplayBinanceSpotTradeFlow'` | Passed |
| Full Go regression | `go test ./...` | Passed |
| Contract validation | `make contracts-validate` | Passed: contract family manifests and docs validate successfully |
| Fixture validation | `make fixtures-validate` | Passed: fixture corpus and replay seeds validate successfully |
| Replay smoke | `CONTRACT_FIXTURES=1 make replay-smoke` | Passed: replay smoke checks passed |
| Whitespace validation | `git diff --check` | Passed |

## Side-Effect Verification

- Trade-flow buckets are generated only from accepted Binance Spot trade inputs for the configured tracked symbols.
- Duplicate source record IDs are covered by unit, runtime, integration, and replay checks and do not double-count aggregate metrics.
- Timestamp fallback/degraded posture remains machine-readable in internal bucket output and is covered by runtime and replay checks.
- Repeated snapshot and replay checks prove deterministic output ordering and values for the same pinned inputs.
- Public `/api/market-state/*`, `/api/runtime-status`, and `/healthz` response shapes remain unchanged in this child; public indicator/API exposure remains deferred.
- No Python runtime dependency, browser-side Binance access, private Binance endpoint, or fixture-backed live runtime was introduced.

## Optional Checks Not Run

- `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive` was not run because the required matrix is deterministic/local and live Binance validation is optional.
- `make compose-smoke` was not run because Docker is not available in this WSL environment.

## State Updates

- Archived feature plan to `plans/completed/binance-spot-trade-flow-feature-inputs/` after the validation matrix passed.
- Updated `plans/STATE.md`, `plans/epics/binance-market-intelligence-gap-closure/92-refinement-handoff.md`, and `initiatives/crypto-market-copilot-binance-integration-completion/03-handoff.md` to point at the archived evidence and the next child planning step.

## Result

- `binance-spot-trade-flow-feature-inputs` passed feature-testing and is archived.
- Next recommended action: run `feature-planning` for `binance-spot-depth-liquidity-indicators` when ready to continue Binance market-intelligence closure.
