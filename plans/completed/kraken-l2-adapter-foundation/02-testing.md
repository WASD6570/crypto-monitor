# Testing Plan: Kraken L2 Adapter Foundation

Expected output artifact: `plans/completed/kraken-l2-adapter-foundation/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Trade parse | Native Kraken trade fixture | canonical trade output | `go test` output |
| L2 happy path | Snapshot/delta fixture | canonical book output | `go test` output |
| L2 gap path | Inject or replay sequence gap | degraded/resync output | `go test` output |
| Runtime health | Synthetic loop inputs | healthy/degraded/stale decision | `go test` output |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/venue-kraken/... ./libs/go/...`

## Verification Checklist

- Kraken gap handling is explicit and deterministic.
- Shared health outputs reflect L2 integrity loss.
