# Binance Live Replay Binance Family Determinism

## Ordered Implementation Plan

1. Extend `services/replay-engine` manifest resolution and validation coverage so replay accepts the settled Binance raw partition posture from the completed raw-append slice without guessing paths, collapsing stream families, or drifting from manifest-provided continuity facts.
2. Add deterministic replay-engine ordering and counter proof for the completed Binance family set so repeated runs, mixed sequence availability, duplicate inputs, degraded timestamps, and degraded feed-health evidence remain stable across inspect, rebuild, and compare modes.
3. Add fixture-backed replay proof in `tests/replay` and focused end-to-end audit coverage in `tests/integration` so accepted Binance raw entries flow through manifest build and replay execution with unchanged ordered IDs, duplicate counters, and degraded evidence.
4. Run the attached validation matrix, write `plans/binance-live-replay-binance-family-determinism/testing-report.md`, then move the full directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to replay acceptance, partition resolution, deterministic ordering, duplicate-input handling, and degraded evidence retention for already-settled Binance raw append entries.
- Do not reopen Binance parsing, canonical schema, raw append contract shape, connection/session provenance, or partition posture from `plans/completed/binance-live-raw-append-and-feed-health-provenance/`.
- Preserve the existing replay ordering policy unless a concrete Binance family exposes a deterministic gap that cannot be covered by targeted shared replay changes.
- Keep replay audit behavior manifest-driven: storage state, location, continuity checksum, and entry counts must come from the resolved manifest records rather than inferred tier paths or Binance-specific conventions.
- Preserve explicit degraded feed-health references and timestamp fallback facts so replay output remains auditable for Spot depth recovery and USD-M mixed-surface health states.
- Keep Go as the live and replay runtime path; Python remains offline-only.
- Leave market-state API cutover and any downstream feature-engine behavior to later slices.

## Design Notes

### Repository state to preserve

- `plans/completed/binance-live-raw-append-and-feed-health-provenance/` already fixed Binance stream-family partition posture, stream-key-scoped duplicate identity, and degraded-feed retention format.
- `services/replay-engine/runtime.go` already builds manifests from `RawPartitionLookupScope`, validates resolved partitions, sorts replay entries by persisted replay facts, and emits replay counters and digests.
- `services/replay-engine/manifest_lookup.go` and current tests already encode the rule that replay must trust manifest-provided locations rather than guessing storage paths.
- `tests/replay/replay_retention_safety_test.go` already covers generic retention and degraded timestamp evidence; this feature extends that posture to the completed Binance family set.

### Binance partition posture to preserve

- Spot `trades` and top-of-book remain on the shared date-symbol-venue partition because their raw append entries do not split by `streamFamily`.
- Spot `order-book`, `feed-health`, USD-M `funding-rate`, `mark-index`, `open-interest`, and `liquidation` remain dedicated stream-family partitions because the completed raw append boundary already settled that routing.
- Replay scope resolution must accept both shared and dedicated Binance partition families in the same symbol window without inventing a Binance-only manifest mode.

### Determinism expectations

- Repeated replay runs over identical Binance manifests and raw inputs must produce the same ordered canonical event IDs, output digest, and input counters.
- Duplicate Binance raw entries remain visible to replay counters and audit consumers; replay must not silently deduplicate them.
- Feed-health entries with degraded reasons and raw entries with `recvTs` bucket fallback remain ordered by persisted replay fields only, not recomputed venue semantics.
- USD-M websocket and REST-originated families must remain distinct in replay ordering and digest outcomes when their raw append identities or stream families differ.

### Live vs research boundary

- All replay acceptance, manifest handling, and deterministic proof stay in Go under `services/replay-engine`, `tests/replay`, and `tests/integration`.
- Offline analysis may consume replay artifacts later, but this feature must not make research code part of replay execution.

## ASCII Flow

```text
completed Binance raw append entries
  - spot trades / top-book / order-book / feed-health
  - usdm funding / mark-index / open-interest / liquidation / feed-health
            |
            v
services/replay-engine manifest build
  - resolve manifest-provided Binance partitions
  - preserve storage state / checksum / location
            |
            v
services/replay-engine execute
  - load raw entries
  - sort by persisted ordering policy
  - count duplicates / degraded timestamps / late events
            |
            v
tests/replay + tests/integration proof
  - repeated runs stay identical
  - mixed spot/usdm families stay distinct
  - degraded evidence survives unchanged
```
