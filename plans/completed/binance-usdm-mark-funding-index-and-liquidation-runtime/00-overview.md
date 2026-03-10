# Binance USD-M Mark Funding Index And Liquidation Runtime

## Ordered Implementation Plan

1. Build the USD-M websocket runtime and subscription registry for `markPrice@1s` and `forceOrder` across the configured BTC/ETH symbols.
2. Map `markPrice@1s` into canonical `funding-rate` and `mark-index` candidates and map `forceOrder` into canonical `liquidation-print` candidates without reopening Wave 1 identity or timestamp rules.
3. Emit websocket-specific feed-health outputs that distinguish connection loss, reconnect loops, parser rejection, and mark-price staleness from the expected sparsity of liquidation traffic.
4. Add fixture-backed unit and integration coverage for happy, degraded-timestamp, reconnect, and resubscribe cases.
5. Record validation evidence in `plans/binance-usdm-mark-funding-index-and-liquidation-runtime/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to the Go live runtime for Binance USD-M websocket `markPrice@1s` and `forceOrder`.
- Emit only canonical `funding-rate`, `mark-index`, and `liquidation-print` outputs plus adapter-scoped `feed-health` outputs needed to make degradation machine-visible.
- Do not include REST `openInterest` polling, cadence, or freshness logic; that belongs to `binance-usdm-open-interest-rest-polling`.
- Keep downstream symbols asset-centric as `BTC-USD` and `ETH-USD` while preserving `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` on every canonical output.
- Inherit Wave 1 contract seam decisions for source-record identity, timestamp fallback, and fixture vocabulary instead of redefining them here.
- Keep Go as the live runtime path; Python remains offline-only.
- Reuse Binance runtime defaults from `configs/local/ingestion.v1.json` unless implementation proves a bounded need within this feature.

## Design Notes

### Runtime Boundary

- Keep all websocket connection, subscription, parse, and feed-health behavior inside `services/venue-binance`.
- Keep canonical event building in shared Go ingestion helpers plus `services/normalizer`; do not push venue-specific parsing into downstream consumers.
- Treat this feature as a dedicated USD-M runtime slice that can be implemented in parallel with the Spot websocket supervisor without sharing a speculative multi-market supervisor abstraction.

### Subscription Shape

- Use one USD-M websocket runtime surface that subscribes only to the tracked-symbol stream set: `btcusdt@markPrice@1s`, `ethusdt@markPrice@1s`, `btcusdt@forceOrder`, and `ethusdt@forceOrder`.
- Keep subscription inventory explicit and deterministic from the configured canonical symbols instead of hard-coding symbol lists in parser code.
- Resubscribe the full active USD-M stream set after reconnect using the existing Binance reconnect/backoff limits.

### Health And Degraded Behavior

- `markPrice@1s` is the freshness-bearing stream for this feature because it should produce regular messages; `forceOrder` is event-driven and must not create false stale alarms when no liquidation occurs.
- Connection loss, reconnect loops, and resubscribe failures degrade all tracked USD-M symbols because they share one websocket runtime surface.
- Parser rejection, invalid symbol mapping, or timestamp degradation should remain explicit in tests and feed-health evidence; no silent drops beyond intentionally rejected unsupported symbols.
- Reuse the existing feed-health vocabulary from `plans/completed/market-ingestion-and-feed-health/`; do not invent a USD-M-only health state model.

### Contract And Timestamp Posture

- Reuse existing schemas for `funding-rate`, `mark-index`, and `liquidation-print`; schema edits are out of scope unless implementation proves a concrete blocker.
- Preserve both `exchangeTs` and `recvTs` on every accepted message and let strict timestamp policy mark degraded fallbacks.
- Follow the inherited Wave 1 source-record ID conventions already seeded by existing fixtures and contract work; implementation should wire those builders, not redesign them.

## ASCII Flow

```text
Binance USD-M websocket
  - <symbol>@markPrice@1s
  - <symbol>@forceOrder
          |
          v
services/venue-binance
  - subscription registry
  - reconnect/resubscribe loop
  - parse markPrice / forceOrder
  - runtime feed-health
          |
          +----> funding-rate candidate
          +----> mark-index candidate
          +----> liquidation-print candidate
          +----> feed-health candidate
                    |
                    v
services/normalizer + libs/go/ingestion
  - canonical symbol/provenance
  - exchangeTs / recvTs policy
  - degraded timestamp handling
                    |
                    v
canonical events for BTC-USD / ETH-USD
```
