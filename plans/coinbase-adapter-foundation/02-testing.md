# Testing Plan: Coinbase Adapter Foundation

Expected output artifact: `plans/coinbase-adapter-foundation/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Trade parse | Native Coinbase trade fixture | canonical trade output | `go test` output |
| Book parse | Native Coinbase book fixture | canonical book output or explicit degraded path | `go test` output |
| Runtime health | Synthetic loop inputs | healthy/degraded/stale decision | `go test` output |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-coinbase/... ./libs/go/...`

## Verification Checklist

- Coinbase parser boundaries stay explicit.
- Shared health vocabulary matches other venues.
