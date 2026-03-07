# Testing Plan: Binance Adapter Loop Runtime

Expected output artifact: `plans/completed/binance-adapter-loop-runtime/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Healthy loop | Fresh message/snapshot with connected state | `HEALTHY` decision | `go test` output |
| Gap degradation | Mark sequence gap while stream remains fresh | `DEGRADED` with `sequence-gap` | `go test` output |
| Stale transition | Advance synthetic time past message threshold | `STALE` with stale reason preserved | `go test` output |
| Recovery reset | Clear gap and reset reconnect/resync counters | decision returns to healthy/degraded as appropriate | `go test` output |
| Snapshot recovery window | Record several recoveries across the minute boundary | old entries are ignored or pruned deterministically | `go test` output |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'TestAdapterLoopState|TestRuntimeEvaluateLoopState' -v`

## Verification Checklist

- Loop-state helpers remove the need for direct field mutation in tests.
- Decision outputs still use shared `ingestion.FeedHealthStatus`.
- No live venue dependency is introduced.
