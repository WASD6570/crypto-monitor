# Testing

## Validation Matrix

| Check | Command | Goal | Expected Evidence |
|---|---|---|---|
| Venue runtime failure helpers | `go test ./services/venue-binance -run 'Test(Runtime|Spot|USDM)'` | Prove reconnect, stale, rate-limit, depth recovery, USD-M websocket, and open-interest health helpers remain deterministic | Focused venue tests pass, including any new failure-path tests added by this feature |
| Command and API runtime-status proof | `go test ./cmd/market-state-api ./services/market-state-api` | Prove additive `usdmStatus`, warm-up, degradation, current-state preservation, and route serialization work together | Command and API tests pass with deterministic `BTC-USD`, `ETH-USD` ordering |
| Command race proof | `go test -race ./cmd/market-state-api` | Catch unsafe runtime-status reads, provider close races, or owner snapshot races | Race detector passes for command-owned runtime tests |
| Binance integration regression | `go test ./tests/integration -run 'TestIngestionBinance(CurrentState|SpotDepthRecovery|USDM)|TestBinanceSpotSupervisor|TestIngestionRetrySafetyStaysBounded|TestIngestionRunbookAlignmentUsesSharedHealthVocabulary'` | Confirm current-state, feed-health vocabulary, retry safety, Spot recovery, and USD-M integration behavior stay aligned | Focused integration tests pass |
| Replay determinism regression | `go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism'` | Confirm replay/current-state semantics remain deterministic for pinned Binance inputs | Focused replay tests pass |
| Full Go baseline | `go test ./...` | Prove the feature did not break unrelated Go packages | Full Go suite passes |
| Contract family baseline | `make contracts-validate` | Prove shared contract manifests remain valid if any contract-adjacent files changed | Contract family validation passes |
| Fixture and replay smoke baseline | `make fixtures-validate && CONTRACT_FIXTURES=1 make replay-smoke` | Prove fixture corpus and replay smoke remain deterministic | Fixture validation and replay smoke pass |
| Optional public Binance boundary | `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive` | Prove the public live Spot depth boundary still works without mocks when network access is available | Passes or is recorded as skipped/unavailable with reason |
| Optional Compose smoke | `make compose-smoke` | Prove checked-in Compose exposes `/api/runtime-status`, `/api/market-state/global`, and `/healthz` correctly when Docker is available | Passes or is recorded as skipped/unavailable with Docker reason |

## Required Inputs And Environment

- No exchange credentials are required.
- Required local validation should not depend on public Binance responses.
- Optional live validation requires public network access and `BINANCE_LIVE_VALIDATION=1`.
- Optional Compose validation requires Docker and the checked-in Compose stack.
- Keep `MARKET_STATE_API_CONFIG_PATH` defaults or explicit checked-in config paths; do not use private or secret config files.

## Verification Checklist

- Runtime-status still returns exactly `BTC-USD` and `ETH-USD` in deterministic order.
- Existing runtime-status Spot fields remain backward-compatible.
- Additive `usdmStatus` exposes websocket and open-interest health without becoming the Spot readiness source.
- Warm-up, readable degradation, stale state, depth recovery, rate limit, and shutdown outcomes are explicitly asserted.
- Current-state outputs preserve existing Spot price semantics and conservative USD-M cap behavior.
- `/healthz` still returns process health only and does not include runtime freshness.
- No browser-side Binance access, Python live-runtime dependency, private endpoint, or fixture-backed live runtime is introduced.

## Side Effects To Check

- No persisted database records are expected.
- No exchange mutations are possible because only public read endpoints are used.
- Optional live and Compose checks may open network connections and local containers only.
- Provider shutdown tests should close local websocket/REST resources and should not leave background goroutines hanging.

## Reporting

- Record implementation and validation evidence in `plans/binance-live-runtime-soak-and-failure-hardening/testing-report.md` while this feature remains active.
- Include command output summaries, pass/fail status, skipped optional checks with reasons, and residual risks.
- After implementation and a passing `feature-testing` run, move the full directory to `plans/completed/binance-live-runtime-soak-and-failure-hardening/`.
