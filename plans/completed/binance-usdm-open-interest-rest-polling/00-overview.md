# Binance USD-M Open Interest Rest Polling

## Ordered Implementation Plan

1. Add a Binance USD-M REST open-interest parser and poll scheduler that keeps cadence, per-symbol freshness, and rate-limit posture explicit.
2. Normalize parsed samples into canonical `open-interest-snapshot` events while preserving symbol, provenance, exchange time, `recvTs`, and degraded timestamp behavior.
3. Add Binance fixtures and targeted integration coverage for happy-path polling, missing exchange time fallback, stale polls, and rate-limit degradation.
4. Record validation evidence in `plans/binance-usdm-open-interest-rest-polling/testing-report.md`, then move the full plan directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to Binance USD-M REST polling for `/fapi/v1/openInterest` in the Go live path.
- Emit canonical `open-interest-snapshot` outputs plus adapter-scoped `feed-health` outputs that make poll freshness and rate-limit degradation machine-visible.
- Keep downstream symbols asset-centric as `BTC-USD` and `ETH-USD` while preserving `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType` on every canonical output.
- Prefer Binance's REST `time` field as `exchangeTs`; if it is missing or invalid, preserve `recvTs` and let canonical timestamp policy degrade explicitly.
- Keep websocket and REST freshness semantics explicit and separate; do not fold open-interest polling into the websocket runtime.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Polling Boundary

- Keep request-shape, cadence, poll state, and feed-health behavior inside `services/venue-binance`.
- Keep canonical event building in `libs/go/ingestion` plus `services/normalizer`.
- Use per-symbol poll state so REST freshness remains visible even though Binance serves one symbol per request.

### Polling Defaults

- Use config-backed local defaults for open-interest poll cadence and per-minute rate limiting.
- Treat a successful poll sample as the freshness-bearing event for the symbol.
- Surface rate-limit pressure through machine-readable feed-health reasons instead of logs only.

## ASCII Flow

```text
Binance REST /fapi/v1/openInterest?symbol=<symbol>
          |
          v
services/venue-binance
  - request shape
  - poll cadence + rate limit checks
  - per-symbol freshness state
  - parse openInterest payload
          |
          +----> open-interest candidate
          +----> feed-health candidate
                    |
                    v
services/normalizer + libs/go/ingestion
  - canonical symbol/provenance
  - exchangeTs / recvTs policy
  - degraded timestamp fallback
                    |
                    v
canonical open-interest + feed-health events
```
