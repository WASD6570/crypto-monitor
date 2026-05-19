# Binance Live Runtime Soak And Failure Hardening Testing Report

## Status

- Date: 2026-05-11
- Result: feature-testing passed
- Feature state after this report: archived
- Archive state: moved to `plans/completed/binance-live-runtime-soak-and-failure-hardening/`

## Implementation Evidence

- Added additive `usdmStatus` to `GET /api/runtime-status` without removing or renaming existing Spot fields.
- Preserved Spot-driven `readiness`; USD-M websocket and open-interest health are auxiliary derivatives-context visibility.
- Added deterministic command/API coverage for repeated runtime-status reads, USD-M reconnect/stale and open-interest rate-limit visibility, warm-up status, route serialization, invalid runtime-status symbol handling, and provider shutdown.
- Added focused venue coverage for open-interest rate-limit timestamps in poller state.
- Added operator runbook `docs/runbooks/binance-runtime-soak-and-failure-check.md` and updated existing runbooks to explain `usdmStatus`, optional live checks, optional Compose checks, and `/healthz` process-health boundaries.

## Feature-Testing Evidence

- Loaded `plans/STATE.md`, the active feature overview, the validation matrix, current handoff context, runtime-status handler code, and the runtime-soak runbook before execution.
- No helper assets were present under `plans/binance-live-runtime-soak-and-failure-hardening/test-helpers/`.
- Required local validation was rerun from the repository root; Go test output was cached where applicable and all commands exited successfully.
- Runtime-status scope remains bounded to `BTC-USD` and `ETH-USD`; `usdmStatus` is additive and does not become the Spot readiness source.
- Side-effect verification: no persisted database records were expected, no exchange mutations were possible, and optional network/container checks were not run in this environment.

## Required Validation

| Check | Command | Result |
|---|---|---|
| Venue runtime failure helpers | `go test ./services/venue-binance -run 'Test(Runtime|Spot|USDM)'` | Passed in feature-testing |
| Command and API runtime-status proof | `go test ./cmd/market-state-api ./services/market-state-api` | Passed in feature-testing |
| Command race proof | `go test -race ./cmd/market-state-api` | Passed in feature-testing |
| Binance integration regression | `go test ./tests/integration -run 'TestIngestionBinance(CurrentState|SpotDepthRecovery|USDM)|TestBinanceSpotSupervisor|TestIngestionRetrySafetyStaysBounded|TestIngestionRunbookAlignmentUsesSharedHealthVocabulary'` | Passed in feature-testing |
| Replay determinism regression | `go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism'` | Passed in feature-testing |
| Full Go baseline | `go test ./...` | Passed in feature-testing |
| Contract family baseline | `make contracts-validate` | Passed in feature-testing |
| Fixture and replay smoke baseline | `make fixtures-validate && CONTRACT_FIXTURES=1 make replay-smoke` | Passed in feature-testing |
| Whitespace validation | `git diff --check` | Passed after archive/state update |

The required matrix was rerun as one chained command during feature-testing, then `git diff --check` was run after archive/state updates:

```sh
go test ./services/venue-binance -run 'Test(Runtime|Spot|USDM)' && go test ./cmd/market-state-api ./services/market-state-api && go test -race ./cmd/market-state-api && go test ./tests/integration -run 'TestIngestionBinance(CurrentState|SpotDepthRecovery|USDM)|TestBinanceSpotSupervisor|TestIngestionRetrySafetyStaysBounded|TestIngestionRunbookAlignmentUsesSharedHealthVocabulary' && go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism' && go test ./... && make contracts-validate && make fixtures-validate && CONTRACT_FIXTURES=1 make replay-smoke
```

Observed summary:

```text
ok github.com/crypto-market-copilot/alerts/services/venue-binance
ok github.com/crypto-market-copilot/alerts/cmd/market-state-api
ok github.com/crypto-market-copilot/alerts/services/market-state-api
ok github.com/crypto-market-copilot/alerts/tests/integration
ok github.com/crypto-market-copilot/alerts/tests/replay
ok github.com/crypto-market-copilot/alerts/tests/parity
Contract family manifests and docs validate successfully.
Fixture corpus and replay seeds validate successfully.
Replay smoke checks passed.
```

## Optional Validation

| Check | Command | Result |
|---|---|---|
| Public Binance boundary | `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive` | Skipped; optional live public-network check was not explicitly enabled for this feature-testing pass |
| Compose smoke | `make compose-smoke` | Skipped; Docker unavailable in this WSL environment |

Docker availability check:

```text
The command 'docker' could not be found in this WSL 2 distro.
```

## Residual Risks

- Optional public Binance and Compose validation still need a network/Docker-capable operator environment if those proofs are desired before deployment.
- Enriched Spot trade-flow, Spot liquidity, USD-M derivatives indicator enrichment, alerting, and dashboard readiness remain out of scope for this archived runtime-confidence gate.

## Handoff

- Archived plan: `plans/completed/binance-live-runtime-soak-and-failure-hardening/`.
- Next action: run feature-planning for the next Binance indicator child from `plans/epics/binance-market-intelligence-gap-closure/` when prioritized.
- Do not start alerting or dashboard readiness work until service-owned enriched indicator boundaries are planned, implemented, tested, and archived.
