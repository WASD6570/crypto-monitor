# Kraken L2 Adapter Foundation

## Ordered Implementation Plan

1. Add explicit Kraken parser boundaries for MVP trade and L2/book payloads.
2. Implement strong L2 integrity handling with deterministic gap/resync behavior.
3. Add a Kraken runtime health surface aligned with the shared health model.
4. Validate happy-path and degraded L2 transitions with fixtures.

## Problem Statement

The umbrella plan calls out Kraken as the venue with the strongest order-book integrity requirements, so it needs its own bounded feature plan instead of being folded into a generic venue bucket.

## Requirements

- Keep Kraken-specific sequencing and integrity logic inside `services/venue-kraken`.
- Treat L2 gaps as explicit degraded/resync conditions.
- Preserve source ordering metadata and timestamps for replay.
- Reuse the shared feed-health vocabulary.

## Out Of Scope

- Full live websocket orchestration.
- Generic multi-venue L2 engines.

## Target Repo Areas

- `services/venue-kraken`
- `libs/go/ingestion`
- `tests/fixtures`

## ASCII Flow

```text
Kraken L2/trade payloads
        |
        v
services/venue-kraken parsers
        |
        v
integrity checks + gap detection
        |
        v
shared ingestion messages + health output
```
