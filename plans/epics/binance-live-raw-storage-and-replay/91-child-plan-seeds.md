# Child Plan Seeds: Binance Live Raw Storage And Replay

## `binance-live-raw-append-and-feed-health-provenance`

- Outcome: one bounded slice wires the completed Binance Spot and USD-M canonical outputs into shared raw append entries with stable partition keys, duplicate identity, connection/session provenance, and degraded-feed retention for raw storage.
- Primary repo areas: `libs/go/ingestion`, `services/venue-binance`, `tests/integration`
- Dependencies: completed Binance Spot and USD-M implementation archives, existing raw append builders in `libs/go/ingestion/raw_event_log.go`, and current feed-health/source-ID rules already fixed in the initiative.
- Validation shape: targeted Go tests for Binance raw append entry construction across trade, top-of-book, order-book, funding, mark-index, open-interest, liquidation, and feed-health families; integration checks that degraded facts and partition routing remain stable.
- Why it stands alone: replay semantics depend on raw append identity, partitioning, and degraded-feed provenance being stable first.

## `binance-live-replay-binance-family-determinism`

- Outcome: one bounded slice extends replay acceptance for Binance raw partitions so repeated runs, duplicate inputs, and retained degraded evidence produce deterministic ordered outputs across the completed Binance live family set.
- Primary repo areas: `services/replay-engine`, `tests/replay`, `tests/integration`
- Dependencies: `binance-live-raw-append-and-feed-health-provenance`, shared replay manifest/runtime primitives already in `services/replay-engine`, and the completed Binance fixture corpus plus any raw append fixtures created by the prior slice.
- Validation shape: replay-engine tests for manifest resolution and ordered Binance partition loading, deterministic repeated-run checks, duplicate-input/idempotency checks, and replay assertions that degraded feed-health and timestamp evidence survive unchanged.
- Why it stands alone: replay acceptance is a separate audit boundary once raw append semantics are fixed, and it should not reopen the raw entry contract work from the first child slice.

## Validation Note

- Do not create a separate smoke-only or integration-only child feature for this epic.
- Attach any live or deterministic replay verification directly to the owning raw append or replay implementation slice.
