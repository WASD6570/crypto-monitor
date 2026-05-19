# Binance Spot Trade Flow Feature Inputs

## Ordered Implementation Plan

1. Define one internal Go-owned Spot trade-flow input contract for accepted Binance `trade` prints, scoped to `BTC-USD` and `ETH-USD`, without changing public market-state API output in this child.
2. Add deterministic trade-flow aggregation in `libs/go/features` and expose it through `services/feature-engine` so later indicator/API work can consume stable per-symbol buckets instead of raw websocket frames.
3. Wire the sustained Binance Spot runtime owner to normalize accepted trade frames into feature observations and include deterministic trade-flow snapshots in the command-to-service read model.
4. Add fixture, integration, and replay proof that accepted trade inputs stay ordered, duplicate-safe, timestamp-aware, and deterministic across repeated runs.
5. Run the validation matrix, write `plans/binance-spot-trade-flow-feature-inputs/testing-report.md`, then move the full directory to `plans/completed/binance-spot-trade-flow-feature-inputs/` after implementation and validation finish.

## Requirements

- Scope is limited to internal feature inputs derived from Binance Spot `trade` frames; this child does not add alerting, dashboard rendering, private endpoints, or order/account state.
- Keep tracked symbols fixed to `BTC-USD` and `ETH-USD` and source symbols fixed to `BTCUSDT` and `ETHUSDT` through existing runtime config bindings.
- Reuse the archived Spot trade canonical handoff, raw append, replay, and sustained runtime owner evidence instead of creating a second Spot lifecycle owner.
- Preserve event time versus processing time: bucket by resolved canonical trade event time, carry timestamp fallback/degraded counts, and reject or mark late events deterministically.
- Treat duplicates by stable `sourceRecordId`/trade ID so duplicate accepted frames cannot double-count trade-flow notional.
- Keep Go as the live runtime path; Python remains offline-only.
- Do not change `/healthz`, `/api/runtime-status`, `/api/market-state/global`, or `/api/market-state/:symbol` response schemas in this child unless implementation finds a blocking internal handoff need; any public contract change must be additive and explicitly deferred to the later API/dashboard readiness child.
- Keep degraded feed, stale, reconnect, and timestamp posture machine-readable in the internal trade-flow buckets so later consumers do not infer quality from logs.

## Design Notes

### Feature boundary

- The completed `binance-spot-trade-canonical-handoff` slice proves native Binance Spot trade parsing and canonical normalization; this feature consumes that seam and promotes it into feature-engine inputs.
- The sustained Spot runtime currently parses accepted trade frames and discards them; implementation should record them as bounded trade-flow observations on the same runtime/read-model path that already owns top-of-book and depth state.
- Public API exposure and dashboard rendering are intentionally deferred until `binance-market-indicator-api-and-dashboard-readiness` after Spot depth and USD-M indicator boundaries settle.

### Planned internal output

- Add a deterministic per-symbol trade-flow bucket model in `libs/go/features`, likely named around `SpotTradeFlowObservation` and `SpotTradeFlowBucket`.
- Required first metrics: trade count, buy/sell trade count, buy/sell notional, net aggressor notional, total notional, first/last price, VWAP, price-change bps, timestamp fallback count, duplicate count, and feed-health/degraded reasons.
- Bucket intervals should align with existing current-state bucket families (`30s`, `2m`, `5m`) so later market-state and alert plans can join trade flow with current market-quality buckets without inventing a second time model.
- Algorithm/config versions must be explicit and stable, for example `feature-engine.binance-spot-trade-flow.v1` and `binance-spot-trade-flow-buckets.v1`.

### Live-read boundary

- Extend only internal Go structs and interfaces as needed, such as `SpotCurrentStateSnapshot`, with a deterministic trade-flow section.
- Keep snapshots sorted by supported symbol order and bucket end time.
- Snapshot reads must be safe during warm-up and reconnect: absence of trade-flow buckets should mean no accepted trades yet, not fixture fallback or client-side computation.
- Runtime shutdown and repeated snapshot reads must remain idempotent.

### Compatibility posture

- Existing current-state responses should remain byte-shape compatible unless the feature explicitly adds internal-only fields that are not serialized publicly.
- Shared JSON schemas under `schemas/json/features` should be touched only if the implementation chooses to persist or expose the new feature family as a formal contract in this child; otherwise leave schema work to the later API/dashboard readiness child.
- Replay proof must use accepted raw Binance `market-trade` inputs and verify stable trade-flow output ordering for the same pinned inputs.

## ASCII Flow

```text
Binance Spot websocket runtime
  - accepted trade frame
  - existing supervisor and source-symbol binding
          |
          v
services/venue-binance
  ParseTradeFrame
          |
          v
libs/go/ingestion
  NormalizeTradeMessage
  - sourceRecordId = trade:<id>
  - canonical event time
  - timestamp status
          |
          v
libs/go/features + services/feature-engine
  Spot trade-flow observation processor
  - dedupe by sourceRecordId
  - bucket by canonical event time
  - aggregate buy/sell notional and price-action inputs
  - carry feed/timestamp quality
          |
          v
internal read model / replay proof
  - deterministic per-symbol trade-flow buckets
  - no public API or dashboard exposure in this child
```

## Acceptance Criteria

- Accepted Binance Spot trade frames for `BTC-USD` and `ETH-USD` produce deterministic internal trade-flow buckets through Go-owned feature code.
- Duplicate Spot trades do not double-count trade count, notional, VWAP, or price-action metrics.
- Timestamp fallback and degraded feed posture are represented in bucket output and covered by tests.
- Repeated fixture and replay runs produce identical bucket ordering and values.
- Existing market-state API schemas and responses remain unchanged in this child.
- The feature remains replayable from accepted raw trade inputs and does not require Python, browser-side Binance access, or mocks in live runtime.

## Archive Intent

- Keep this feature active under `plans/binance-spot-trade-flow-feature-inputs/` while implementation and validation are in progress.
- When implementation and feature-testing pass, move the full directory and `testing-report.md` to `plans/completed/binance-spot-trade-flow-feature-inputs/`.
