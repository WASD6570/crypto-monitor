# Testing Plan: Binance Spot WS Runtime Supervisor

Expected output artifact: `plans/binance-spot-ws-runtime-supervisor/testing-report.md`

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestSpotWebsocketSupervisor|TestSpotSubscriptionState|TestSpotFeedHealth' -v`
- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestRuntimeEvaluateLoopState|TestRuntimeEvaluateAdapterInput' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestBinanceSpotSupervisor' -v`

## High-Signal Smoke Matrix

| Case | Setup | Expected outcome | Evidence |
|---|---|---|---|
| Healthy startup | Open scripted Spot socket, send subscribe ack, then fresh control/data traffic | supervisor reaches connected/subscribed state and emits `HEALTHY` feed-health | targeted `go test` output |
| Message stale path | Leave socket open but stop traffic past `messageStaleAfterMs` | feed-health becomes `STALE` with `message-stale`; no parser logic required | targeted `go test` output |
| Reconnect loop degradation | Force repeated disconnects until reconnect threshold is crossed | feed-health becomes `DEGRADED` with `connection-not-ready` and `reconnect-loop` | targeted `go test` output |
| Resubscribe after reconnect | Disconnect after first healthy subscribe, then reopen socket | desired stream set is resubscribed exactly once and supervisor returns to healthy after fresh traffic | targeted `go test` output |
| Proactive 24h rollover | Advance synthetic time to rollover deadline while socket is healthy | supervisor triggers intentional reconnect before venue cutoff, preserves desired subscriptions, and resumes healthy state | targeted `go test` output |
| Feed-health normalization seam | Feed one supervisor-generated health message into `services/normalizer` | canonical feed-health output preserves Wave 1 timestamps, provenance, and source-record ID | targeted `go test` output |

## Verification Checklist

- Spot supervisor filters runtime health to `trades` and `top-of-book` only, so snapshot freshness is not required.
- No trade parser, `bookTicker` parser, or depth bootstrap behavior is introduced in this feature.
- Reconnect backoff, reconnect-loop thresholds, and connect-rate limits come from config-backed runtime defaults.
- Feed-health remains machine-readable and test-asserted, not log-only.
- Supervisor tests run with deterministic time and scripted websocket behavior, without live Binance access.

## Handoff Notes For Feature Implementing

- Create the implementation under `services/venue-binance` first, with narrow unit tests before adding the integration smoke.
- Reuse existing `Runtime` and `AdapterLoop` helpers instead of creating a second health-policy stack.
- Stop if implementation starts requiring payload parsing or depth bootstrap semantics; those belong to later child features.
