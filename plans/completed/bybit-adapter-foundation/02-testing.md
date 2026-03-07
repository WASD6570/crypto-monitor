# Testing Plan: Bybit Adapter Foundation

Expected output artifact: `plans/completed/bybit-adapter-foundation/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Trade parse | Native Bybit trade fixture | canonical trade output | `go test` output |
| Book parse | Native Bybit book fixture | canonical book or deterministic degraded result | `go test` output |
| Runtime health | Synthetic loop inputs | healthy/degraded/stale decision | `go test` output |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-bybit/... ./libs/go/...`

## Verification Checklist

- Bybit stays venue-local in `services/venue-bybit`.
- Shared normalization and health vocabulary match Binance behavior.
