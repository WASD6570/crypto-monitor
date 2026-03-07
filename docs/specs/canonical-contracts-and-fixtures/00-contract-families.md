# Canonical Contract Families

## Family Inventory

| Family | Directory | Purpose | Planned starter schemas |
| --- | --- | --- | --- |
| Events | `schemas/json/events` | Canonical normalized venue payloads and feed-health source events used by ingestion, replay, and downstream derivations. | `market-trade.v1.schema.json`, `order-book-top.v1.schema.json`, `feed-health.v1.schema.json`, `funding-rate.v1.schema.json`, `open-interest-snapshot.v1.schema.json`, `mark-index.v1.schema.json`, `liquidation-print.v1.schema.json` |
| Features | `schemas/json/features` | Derived per-venue and composite state used by visibility, regime gating, and later alert logic. | `venue-bar-30s.v1.schema.json`, `venue-bar-2m.v1.schema.json`, `venue-bar-5m.v1.schema.json`, `composite-market-state.v1.schema.json`, `feed-health-state.v1.schema.json` |
| Alerts | `schemas/json/alerts` | Operator-facing alert payloads, routing metadata, and delivery status records. | `trade-alert.v1.schema.json`, `alert-delivery.v1.schema.json` |
| Outcomes | `schemas/json/outcomes` | Evaluation payloads that explain whether a delivered alert later worked and why. | `alert-outcome.v1.schema.json`, `alert-review.v1.schema.json` |
| Replay | `schemas/json/replay` | Deterministic replay seeds, replay windows, and replay result metadata. | `replay-seed.v1.schema.json`, `replay-window.v1.schema.json`, `replay-run-result.v1.schema.json` |
| Simulation | `schemas/json/simulation` | Offline simulated execution requests and results used to estimate edge after costs. | `simulated-execution-request.v1.schema.json`, `simulated-execution-result.v1.schema.json` |

## Naming And Versioning Standard

- Every concrete schema file uses `<schema-name>.v{major}.schema.json`.
- Every family publishes a versioned manifest at `schemas/json/<family>/family.v1.json`.
- Additive changes stay within the active major version.
- Breaking changes require a new schema filename major version and an explicit consumer rollout.
- Payloads must carry `schemaVersion` so fixtures, replays, and downstream consumers do not guess the contract shape.

## Canonical Identity Rules

### Symbols

- `symbol` is the normalized tradable identity and is currently limited to `BTC-USD` and `ETH-USD`.
- `sourceSymbol` preserves the venue-native symbol such as `BTCUSDT`, `XBT/USD`, or `ETH-USD`.
- `quoteCurrency` preserves source quote context and is currently limited to `USD`, `USDT`, and `USDC`.

### Venues And Market Context

- `venue` identifies a concrete source venue: `BINANCE`, `BYBIT`, `COINBASE`, or `KRAKEN`.
- `marketType` identifies the market surface and is currently `spot` or `perpetual`.
- `compositeId` is optional and reserved for derived payloads such as `WORLD_SPOT_COMPOSITE`, `WORLD_PERP_COMPOSITE`, `USA_SPOT_COMPOSITE`, and `USA_PERP_COMPOSITE`.
- Composite payloads may not overload `venue` with composite identifiers; use `compositeId` instead.

## Timestamp Semantics

- `exchangeTs` is the primary event-time field when present and plausible.
- `recvTs` is always required for auditability, feed staleness, and transport latency analysis.
- `timestampStatus` records whether event-time handling is `normal` or `degraded`.
- `bucketSource` is reserved for derived payloads and records whether bucket assignment used `exchangeTs` or `recvTs`.
- Replay payloads must retain the timestamp choice and source record references needed to reproduce ordering later.

## Reserved Field Glossary

| Field | Meaning |
| --- | --- |
| `schemaVersion` | Concrete schema version for the payload. |
| `symbol` | Canonical tradable identity such as `BTC-USD`. |
| `sourceSymbol` | Venue-native symbol string preserved for provenance. |
| `quoteCurrency` | Venue-native quote context used by the source market. |
| `venue` | Concrete venue identifier for source payloads. |
| `marketType` | Source market surface such as `spot` or `perpetual`. |
| `compositeId` | Optional aggregate identifier for WORLD or USA derived payloads. |
| `exchangeTs` | Primary event-time timestamp when plausible. |
| `recvTs` | Processing-time timestamp captured on receipt. |
| `timestampStatus` | Indicates whether timestamp handling stayed normal or degraded. |
| `bucketSource` | States whether downstream bucket assignment used `exchangeTs` or `recvTs`. |
| `configVersion` | Versioned configuration tag used by derived outputs and later alert logic. |
| `regimeTags` | Ordered list of market-state labels carried by derived outputs. |
| `feedHealthState` | Explicit feed-health label such as `HEALTHY`, `DEGRADED`, or `STALE`. |
| `replayRef` | Replay provenance object containing deterministic source references. |
| `sourceRecordId` | Stable source-side identifier or sequence used for deduplication and replay ordering. |

## Notes For Follow-Up Work

- The family manifests under `schemas/json/*/family.v1.json` are the machine-readable source of truth for this contract inventory.
- Starter concrete schemas now exist for the `events` and `replay` families so fixture and replay metadata can point at real versioned files.
- Remaining families stay reserved until later modules define concrete payload shapes.
