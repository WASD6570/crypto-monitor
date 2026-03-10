# Binance Live Raw Append And Feed-Health Provenance

## Ordered Implementation Plan

1. Extend the shared raw append boundary in `libs/go/ingestion` so every completed Binance Spot and USD-M canonical family can produce raw entries with stable stream-family routing, duplicate identity precedence, degraded-feed references, and timestamp provenance without inventing a Binance-only storage format.
2. Add Binance-owned raw append seam helpers in `services/venue-binance` that assemble the right connection/session/degraded-feed context for Spot websocket, Spot depth recovery, and USD-M mixed-surface outputs before handing entries to the shared raw writer.
3. Add deterministic fixture-backed and integration proof that accepted Binance trade, top-of-book, order-book, funding, mark-index, open-interest, liquidation, and feed-health outputs append into stable raw entries with unchanged partition keys and duplicate audit facts.
4. Run the attached validation matrix, write `plans/binance-live-raw-append-and-feed-health-provenance/testing-report.md`, then move the full directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to raw append identity, partitioning, connection/session provenance, degraded-feed retention, and stable source-ID posture for already-completed Binance live families.
- Do not reopen parsing, canonical schema, symbol normalization, timestamp policy, or feed-health vocabulary decisions from the completed Binance Spot and USD-M features.
- Preserve asset-centric canonical symbols (`BTC-USD`, `ETH-USD`) while keeping `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs` explicit in appended raw entries.
- Retain degraded feed-health evidence and timestamp fallback facts as part of the raw audit boundary so later replay work can reconstruct why an output was degraded.
- Reuse the shared raw append contract in `libs/go/ingestion/raw_event_log.go`; add only the minimum shared changes needed for concrete Binance family gaps.
- Keep Go as the live runtime path; Python remains offline-only.
- Leave replay-engine ordering and cross-tier determinism to the next child feature once raw append semantics are settled.

## Design Notes

### Repository state to preserve

- `libs/go/ingestion/raw_event_log.go` already defines the shared raw append entry schema, duplicate audit rules, partition routing, and builder functions for canonical families.
- `plans/completed/binance-spot-trade-canonical-handoff/`, `plans/completed/binance-spot-top-of-book-canonical-handoff/`, `plans/completed/binance-spot-depth-bootstrap-and-buffering/`, and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` already settle the Binance Spot event families this slice must append without reinterpretation.
- `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/`, `plans/completed/binance-usdm-open-interest-rest-polling/`, and `plans/completed/binance-usdm-context-sensor-fixtures-and-integration/` already settle the USD-M families and mixed-surface feed-health semantics this slice must retain.
- The replay epic depends on this slice fixing final stream-family partition posture, duplicate identity precedence, and degraded-feed retention format.

### Shared vs venue-owned responsibility

- Shared raw entry structure, partition routing, and duplicate audit precedence belong in `libs/go/ingestion`.
- Binance-specific connection/session reference assembly and the mapping from completed venue outputs into raw append builders belong in `services/venue-binance`.
- Integration proof belongs in `tests/integration` and should use completed Binance fixtures where practical instead of inventing parallel fixture vocabularies.

### Provenance expectations

- Spot websocket families should preserve a stable websocket-owned `connectionRef` and `sessionRef` that match the settled Spot supervisor/runtime seams.
- Spot depth feed-health should preserve the dedicated recovery-owned source record identity and degraded-feed reference when recovery leaves depth unsynchronized.
- USD-M websocket and REST-originated families should stay distinct in raw append provenance just as they already stay distinct in canonical feed-health output.
- Duplicate identity should continue to prefer venue message ID, then sequence-plus-stream key, then canonical event ID fallback unless a concrete Binance family proves that order insufficient.

### Live vs research boundary

- Raw append helpers, provenance mapping, and deterministic integration all stay in Go under `libs/go/ingestion`, `services/venue-binance`, and `tests/`.
- Offline analysis may later consume the raw corpus, but this feature must not make research code a dependency of raw append behavior.

## ASCII Flow

```text
completed Binance canonical outputs
  - spot trade / top-book / depth / depth-health
  - usdm funding / mark-index / open-interest / liquidation / feed-health
            |
            v
services/venue-binance raw append seams
  - attach connectionRef / sessionRef / degradedFeedRef
  - preserve source identity per live surface
            |
            v
libs/go/ingestion raw append builders
  - canonical payload -> raw entry
  - duplicate identity
  - partition routing
  - timestamp provenance
            |
            v
raw writer / integration proof
  - stable partitions
  - retained degraded evidence
  - replay-ready raw append entries
```
