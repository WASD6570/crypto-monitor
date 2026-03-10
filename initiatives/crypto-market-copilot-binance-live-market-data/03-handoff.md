# Binance Live Market Data Handoff

## Epic Queue

1. `plans/epics/binance-live-contract-seams-and-fixtures/`
2. `plans/epics/binance-spot-trades-and-top-of-book-runtime/`
3. `plans/epics/binance-usdm-context-sensors/`
4. `plans/epics/binance-spot-depth-bootstrap-and-recovery/`
5. `plans/epics/binance-live-raw-storage-and-replay/`
6. `plans/epics/binance-live-market-state-api-cutover/`

## Planning Waves

### Wave 1

- `binance-live-contract-seams-and-fixtures`
- Why now: every later slice depends on one stable answer for canonical symbols, timestamp resolution, source-record identity, and fixture vocabulary.

### Wave 2

- `binance-spot-trades-and-top-of-book-runtime`
- `binance-usdm-context-sensors`
- Why parallel: both consume the same contract decisions but touch different live surfaces and do not need to redefine the same recovery path.

### Wave 3

- `binance-spot-depth-bootstrap-and-recovery`
- Why later: depth recovery inherits Spot runtime behavior but adds sequence, snapshot, and resync risk that should not block initial live trade/top-of-book and USD-M context planning.

### Wave 4

- `binance-live-raw-storage-and-replay`
- Why later: replay and audit acceptance should lock in after the live Spot and USD-M semantics are stable enough to avoid duplicate identity drift.

### Wave 5

- `binance-live-market-state-api-cutover`
- Why later: consumer cutover should happen only after live ingestion, resync behavior, and replay-safe identities are already understood.

## Epic Seeds

### `plans/epics/binance-live-contract-seams-and-fixtures/`

- Problem statement: live Binance work will drift quickly unless symbol naming, timestamp selection, source-record IDs, and fixture expectations are decided before runtime slices branch.
- In scope: confirm canonical `BTC-USD` and `ETH-USD` mapping rules, preserve `sourceSymbol`, define stream-specific exchange-time selection, define source-ID patterns for Spot and USD-M, add deterministic live-fixture coverage, note any schema gaps.
- Out of scope: full live connection managers or API cutover.
- Target repo areas: `schemas/json/events`, `libs/go/contracts`, `tests/fixtures/events/binance`, `tests/integration`, `docs/runbooks`
- Contract/fixture/parity/replay implications: this slice is the prerequisite for replay-safe identity and timestamp behavior across all later epics.
- Likely validation shape: schema validation, fixture decoding, and targeted integration checks that timestamps and IDs stay stable across duplicate inputs.

### `plans/epics/binance-spot-trades-and-top-of-book-runtime/`

- Problem statement: the current stack needs a real Spot market-data path before any live dashboard state can be trusted.
- In scope: Spot WS connection loop, ping/pong handling, proactive 24h reconnect, bounded resubscribe behavior, `trade` parsing, `bookTicker` parsing, feed-health updates, symbol normalization, canonical handoff.
- Out of scope: depth snapshot/bootstrap sequencing and market-state API cutover.
- Target repo areas: `services/venue-binance`, `services/normalizer`, `configs/*`, `tests/integration`
- Contract/fixture/parity/replay implications: must follow wave-1 timestamp and source-ID rules and emit visible degraded reasons.
- Likely validation shape: targeted runtime tests for reconnect/backoff behavior plus normalization/integration fixtures for Spot trade and top-of-book events.

### `plans/epics/binance-usdm-context-sensors/`

- Problem statement: the product needs live derivatives context from Binance USD-M, but the venue surface mixes WS streams and REST polling.
- In scope: `markPrice@1s`, funding/index extraction, `forceOrder`, REST `openInterest` polling cadence, symbol normalization, canonical event handoff, feed-health behavior for mixed WS/REST sensors.
- Out of scope: private futures endpoints, order entry, and multi-venue derivatives rollout.
- Target repo areas: `services/venue-binance`, `services/normalizer`, `schemas/json/events`, `configs/*`, `tests/integration`
- Contract/fixture/parity/replay implications: must preserve market type and source symbols while keeping canonical symbols aligned with existing feature consumers.
- Likely validation shape: parser fixtures, polling freshness tests, and direct mixed WS plus REST validation against the implemented service boundary.

### `plans/epics/binance-spot-depth-bootstrap-and-recovery/`

- Problem statement: Spot order-book integrity depends on exact snapshot bootstrap and gap detection semantics, not just a connected socket.
- In scope: `/api/v3/depth` bootstrap, `depth@100ms` buffering, sequence acceptance rules, bounded resync loops, snapshot refresh cadence, explicit degradation on gaps and drift.
- Out of scope: redesigning the shared sequencer contract beyond what live Binance requires.
- Target repo areas: `services/venue-binance`, `libs/go/ingestion`, `configs/*`, `tests/integration`, `tests/fixtures/events/binance`
- Contract/fixture/parity/replay implications: order-book gaps and resync triggers must remain replay-visible and feed-health-visible.
- Likely validation shape: deterministic gap/resync tests, snapshot cooldown/rate-limit checks, and direct validation for snapshot plus delta recovery.

### `plans/epics/binance-live-raw-storage-and-replay/`

- Problem statement: live ingestion is not trustworthy if accepted Spot and USD-M inputs cannot be appended and replayed with stable identities.
- In scope: raw append integration, partitioning and source-ID validation, replay acceptance for live Binance event families, degraded feed-health retention, deterministic fixture-backed replay checks.
- Out of scope: broad historical migration or unrelated replay-engine redesign.
- Target repo areas: `libs/go/ingestion`, `services/replay-engine`, `tests/replay`, `tests/integration`
- Contract/fixture/parity/replay implications: this slice is the audit gate for later live market-state and alerting work.
- Likely validation shape: repeated replay runs with identical outputs and targeted duplicate-input/idempotency checks.

### `plans/epics/binance-live-market-state-api-cutover/`

- Problem statement: the dashboard already has a stable Go API boundary, but it still needs live Binance-backed data behind that boundary.
- In scope: replace deterministic provider behavior, wire live current-state sourcing, preserve same-origin `/api` consumer behavior, update compose/local validation coverage, document operator-visible limits and degradation behavior.
- Out of scope: frontend-owned venue logic, broad UI redesign, or a second API boundary.
- Target repo areas: `services/market-state-api`, `services/feature-engine`, `services/regime-engine`, `cmd/market-state-api`, `docker-compose.yml`, `apps/web/tests/e2e`
- Contract/fixture/parity/replay implications: must preserve current consumer response shape while switching the backing source from deterministic to live.
- Likely validation shape: targeted Go tests, compose validation, and browser/API checks against the live-backed API path.

## Open Questions That Still Matter

- Should the first live API cutover expose only Spot-driven current-state inputs, with USD-M context remaining auxiliary, or should derivatives context influence the same current-state path immediately?
- What polling cadence and freshness ceiling should `openInterest` use in `local`, `dev`, and `prod` without creating false staleness or unnecessary REST pressure?
- Should Spot and USD-M each use a single combined connection per environment for the tracked symbols, or should refinement keep separate connections per stream family for simpler failure isolation?
- Do any existing event schemas need additional required fields for live Binance semantics, or can source-specific richness stay outside the canonical contract for now?
