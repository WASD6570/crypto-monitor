# Testing Plan: Market Ingestion And Feed Health

Expected output artifact: `plans/completed/market-ingestion-and-feed-health/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Venue adapter happy path | Start adapter against deterministic fixtures or test harness inputs | Trades, top-of-book, and book events emit canonical payloads with both timestamps | Integration test output |
| Gap recovery | Inject sequence gap into order book deltas | Feed enters degraded state, resyncs, and resumes with explicit health transition | Gap/resync smoke output |
| Stale feed detection | Pause stream beyond threshold | Health output marks feed stale and later clears on recovery | Staleness test output |
| Timestamp degradation | Provide missing or implausible `exchangeTs` input | Normalizer falls back safely and marks timestamp-degraded state | Targeted normalization test output |
| Rate-limit-safe recovery | Trigger repeated reconnect or snapshot recovery in test harness | Backoff and retry behavior remain bounded and observable | Retry-policy test output |
| Multi-venue normalization consistency | Run canonical normalization on equivalent BTC/ETH examples across venues | Symbol, venue, quote context, and market type normalize consistently | Fixture validation output |

## Required Commands

The implementing agent should provide these exact commands or direct equivalents:

- `go test ./services/venue-binance/...`
- `go test ./services/venue-bybit/...`
- `go test ./services/venue-coinbase/...`
- `go test ./services/venue-kraken/...`
- `go test ./services/normalizer/...`
- `go test ./tests/integration -run Ingestion`
- `make replay-smoke INGESTION_FIXTURES=1`

If the exact package layout changes during implementation, replace these commands with equally explicit package- or target-level commands.

## Verification Checklist

- All MVP venues emit canonical payloads using the shared contract rules.
- Order book bootstrap and delta handling are explicit and test-covered.
- Sequence gaps trigger degradation and resync rather than silent continuation.
- `exchangeTs` and `recvTs` are always preserved or degraded fallback is made explicit.
- Health outputs expose reconnects, staleness, and resync counts.
- Runtime controls remain configuration-driven and rate-limit-safe.

## Negative Cases

- Missing snapshot during bootstrap should fail into degraded state rather than produce guessed L2 state.
- Out-of-order or duplicated deltas should not silently corrupt the in-memory book.
- Repeated reconnect loops should surface as degraded feed health.
- Invalid symbol or market-type mapping should fail normalization validation.
- Timestamp fallback should be visible, not hidden.

## Handoff Notes

- No exchange credentials are required for the MVP public feeds in this feature plan.
- Validation should rely on deterministic fixtures, local harnesses, or controlled test adapters rather than ad hoc live-market testing alone.
- This feature is done only when downstream services can distinguish healthy, degraded, stale, and resyncing feeds without reading adapter-specific code.
