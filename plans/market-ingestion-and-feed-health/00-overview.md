# Market Ingestion And Feed Health

## Ordered Implementation Plan

1. Define venue adapter boundaries, stream inventory, and health-state outputs
2. Implement order book bootstrap, delta sequencing, and resync behavior per venue
3. Implement canonical normalization handoff with timestamp and degradation semantics
4. Add operational controls for reconnects, rate limits, freshness, and observability
5. Run targeted ingestion, gap-recovery, and staleness validation

## Problem Statement

The first user cannot trust a market-state dashboard unless venue ingestion is resilient, order book state is reconstructable, and degraded feeds are surfaced explicitly instead of silently poisoning downstream features.

This feature turns Binance, Bybit, Coinbase, and Kraken market feeds into dependable canonical event streams with health signals that later visibility and alerting features can consume.

## Requirements

- Consume the shared contract vocabulary from `plans/canonical-contracts-and-fixtures/`.
- Support BTC and ETH only for the MVP.
- Cover these live feed groups:
  - Binance spot market data
  - Binance USD-M perp sensors relevant to the visibility initiative
  - Bybit spot market data
  - Bybit perp sensors relevant to the visibility initiative
  - Coinbase spot market data
  - Kraken spot market data with strong L2 integrity handling
- Store both `exchangeTs` and `recvTs` on emitted canonical events.
- Detect sequence gaps, stale streams, repeated reconnect loops, and snapshot drift.
- Expose feed health in a way downstream services can use for market-state degradation.
- Respect the operating defaults for timestamp precedence, lateness handling, and degraded markers.
- Rate-limit WS and REST recovery paths so implementation does not invite venue bans.
- Keep the live runtime in Go; Python may only support offline analysis or parity later.

## Out Of Scope

- WORLD vs USA composite weighting logic
- 5m tradeability regime logic
- alert setup logic and delivery
- raw append-only persistence implementation details beyond ingestion handoff requirements
- dashboard rendering beyond payloads needed for later views

## Design Notes

### Service Boundaries

- Keep venue-specific websocket and REST behavior inside the relevant adapter service.
- Use `services/normalizer` as the canonical handoff boundary rather than duplicating shared normalization rules inside every downstream consumer.
- If Kraken needs a dedicated service folder that is not scaffolded yet, the implementation plan should add `services/venue-kraken/` explicitly rather than hiding Kraken logic inside another service.

### Feed Health Model

- Health must be machine-readable and not just log text.
- At minimum, define health dimensions for:
  - connection state
  - message freshness
  - resync count
  - gap detection
  - snapshot freshness
  - local clock health if timestamp trust is affected
- Degradation should flow forward into canonical payloads or side-channel health outputs that later regime logic can consume.

### Order Book Integrity

- Snapshot + delta reconstruction rules must be explicit per venue.
- Sequence gaps should force a degraded state and resync path, not best-effort guessing.
- Resync logic should be bounded and observable so persistent failure becomes visible as `NO-OPERATE` input later.

### Timestamp Rules

- `exchangeTs` remains the canonical event-time source when plausible.
- `recvTs` remains mandatory for freshness, latency, and auditability.
- If time trust degrades, normalization must preserve both the fallback behavior and the degraded reason.

### Live vs Research Boundary

- All live ingestion and normalization logic stays in Go services and shared Go helpers.
- Offline Python may later replay or inspect fixtures, but it must not sit in the ingestion runtime path.

## Target Repo Areas

- `services/venue-binance`
- `services/venue-bybit`
- `services/venue-coinbase`
- `services/venue-kraken` if added during implementation
- `services/normalizer`
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/integration`

## ASCII Flow

```text
exchange ws/rest streams
        |
        v
venue adapter service
  - connect / heartbeat
  - snapshot bootstrap
  - delta sequencing
  - resync on gap
  - feed health outputs
        |
        v
normalizer
  - canonical symbol + venue mapping
  - exchangeTs / recvTs semantics
  - degraded markers
        |
        v
canonical event stream + feed-health stream
        |
        +----> later raw storage / replay
        +----> later market-state computation
        +----> later dashboard health panels
```
