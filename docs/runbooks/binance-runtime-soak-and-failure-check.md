# Binance Runtime Soak And Failure Check

## Purpose

Use this runbook to validate the Binance Spot plus USD-M live runtime boundary before adding enriched Binance indicators or alerting behavior.

Required checks are local and deterministic. They use Go tests, local HTTP/websocket fixtures, and injected clocks; they do not call public Binance, Docker, Python, browser code, private endpoints, or exchange credentials.

## Required Local Checks

Run these commands from the repository root:

```sh
go test ./services/venue-binance -run 'Test(Runtime|Spot|USDM)'
go test ./cmd/market-state-api ./services/market-state-api
go test -race ./cmd/market-state-api
go test ./tests/integration -run 'TestIngestionBinance(CurrentState|SpotDepthRecovery|USDM)|TestBinanceSpotSupervisor|TestIngestionRetrySafetyStaysBounded|TestIngestionRunbookAlignmentUsesSharedHealthVocabulary'
go test ./tests/replay -run 'TestReplayBinance|TestMarketStateCurrentReplayDeterminism|TestReplayBinanceMarketStateDeterminism'
go test ./...
make contracts-validate
make fixtures-validate && CONTRACT_FIXTURES=1 make replay-smoke
```

Record pass/fail output, skipped checks, and residual risks in the active feature `testing-report.md`.

## Runtime Status Reading

Query `GET /api/runtime-status` for operator runtime health. The symbol scope is fixed to `BTC-USD` and `ETH-USD`, in that order.

Use these fields first:

- `readiness`: Spot publishability; `NOT_READY` means warm-up, while `READY` means consumer current-state reads can be available even when health is degraded.
- `feedHealth.state`: canonical Spot health, one of `HEALTHY`, `DEGRADED`, or `STALE`.
- `feedHealth.reasons`: canonical Spot reasons such as `connection-not-ready`, `message-stale`, `snapshot-stale`, `sequence-gap`, `reconnect-loop`, `resync-loop`, `rate-limit`, or `clock-degraded`.
- `depthStatus`: Spot depth recovery, sequence-gap, cooldown, retry, and rate-limit posture.
- `usdmStatus.websocket`: additive USD-M websocket health for mark-price, funding, and liquidation streams.
- `usdmStatus.openInterest`: additive USD-M REST open-interest poll health and rate-limit posture.
- `usdmStatus.lastMarkPriceAt`, `usdmStatus.lastOpenInterestAt`, `usdmStatus.nextOpenInterestPollAt`, and `usdmStatus.openInterestRateLimitUntil`: timestamps to record when USD-M context is stale, reconnecting, or rate-limited.

`usdmStatus` explains derivatives context quality. It does not decide Spot `readiness`.

## Expected Failure-Path Evidence

Record the exact route fields, not paraphrased labels:

- Warm-up: `readiness=NOT_READY`; `/healthz` still returns only process health.
- Readable degradation: `readiness=READY` with `feedHealth.state=DEGRADED` or `STALE` and explicit reasons.
- Spot reconnect: `connectionState` is not connected and `feedHealth.reasons` includes `connection-not-ready`.
- Spot stale: `feedHealth.state=STALE` with `message-stale` or `snapshot-stale`.
- Depth sequence gap: `depthStatus.sequenceGapDetected=true` or a visible recovery state.
- Depth rate limit: `depthStatus.state=rate-limit-blocked` and `feedHealth.reasons` includes `rate-limit`.
- USD-M websocket degraded or stale: `usdmStatus.websocket.state=DEGRADED` or `STALE` with canonical reasons.
- USD-M open-interest rate limit or stale: `usdmStatus.openInterest.reasons` includes `rate-limit` or `message-stale`.
- Shutdown: provider close tests pass and no runtime-health semantics are moved into `/healthz`.

## Optional Live Boundary

Public Binance Spot depth validation is optional and must be explicitly enabled:

```sh
BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive
```

This uses public read endpoints only, creates no exchange mutations, and requires no credentials. If public network access is unavailable, record the skip reason and keep the required local deterministic checks as the pass gate.

## Optional Compose Proof

When Docker is available, run:

```sh
make compose-smoke
```

This checks `/api/runtime-status`, `/api/market-state/global`, and `/healthz` through the checked-in Compose stack. If Docker is unavailable, record that reason and do not weaken the required local matrix.

## Boundaries

- `/healthz` is process health only.
- `/api/runtime-status` is the bounded operator runtime-health surface.
- `/api/market-state/global` and `/api/market-state/:symbol` are consumer current-state reads.
- The browser must not talk to Binance directly or compute canonical market state.
- Python must not be required for live runtime operation.
