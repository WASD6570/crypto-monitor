# Child Plan Seeds: Binance Spot Trades And Top Of Book Runtime

## `binance-spot-ws-runtime-supervisor`

- Outcome: one bounded Spot websocket supervisor exists on paper for BTC/ETH `trade` and `bookTicker`, including connect, subscribe, ping/pong, proactive rollover, bounded reconnect, resubscribe, and adapter-scoped feed-health behavior.
- Primary repo areas: `services/venue-binance`, `configs/local/ingestion.v1.json`, `tests/integration`
- Dependencies: Wave 1 contract seam decisions; existing feed-health vocabulary from `plans/completed/market-ingestion-and-feed-health/`
- Validation shape: targeted runtime lifecycle tests for heartbeat timeout, reconnect backoff thresholds, reconnect-loop degradation, and resubscribe-on-reconnect behavior
- Why it stands alone: both stream-specific handoff features depend on the same lifecycle owner, and separating it keeps runtime recovery concerns from being hidden inside a parser feature

## `binance-spot-trade-canonical-handoff`

- Outcome: Spot `trade` payload handling is planned as a bounded adapter-to-normalizer path that preserves Wave 1 identity, timestamp, and provenance rules while integrating with runtime health expectations.
- Primary repo areas: `services/venue-binance`, `services/normalizer`, `tests/fixtures/events/binance`, `tests/integration`
- Dependencies: `binance-spot-ws-runtime-supervisor`; Wave 1 source-ID and timestamp rules
- Validation shape: fixture-backed parser checks, timestamp-degradation checks, and direct validation that accepted trades emit canonical outputs without dropping provenance fields
- Why it stands alone: the trade stream has its own payload shape, timing semantics, and duplicate-sensitivity, so it can be planned and reviewed independently from top-of-book handling

## `binance-spot-top-of-book-canonical-handoff`

- Outcome: Spot `bookTicker` payload handling is planned as a bounded adapter-to-normalizer path that emits canonical top-of-book events with explicit provenance and degradation semantics, without bringing in depth sequencing.
- Primary repo areas: `services/venue-binance`, `services/normalizer`, `tests/fixtures/events/binance`, `tests/integration`
- Dependencies: `binance-spot-ws-runtime-supervisor`; Wave 1 source-ID and timestamp rules
- Validation shape: fixture-backed top-of-book normalization checks, stale-message/feed-health interaction coverage, and direct validation that top-of-book outputs stay distinct from later depth work
- Why it stands alone: `bookTicker` is the first live best-bid/best-ask path, but it should remain separable from both trade parsing and later depth bootstrap decisions

## Validation Note

- Do not create a separate smoke-only or integration-only child feature for this epic.
- After the supervisor, trade handoff, and top-of-book handoff slices are implemented, run combined Spot runtime validation directly against the real Binance API and record the result in the current handoff or implementation report.
