# Bybit Adapter Foundation

## Ordered Implementation Plan

1. Add the first native Bybit parser boundaries for the MVP public streams.
2. Map native payloads into shared ingestion message types and feed shared normalization.
3. Add a Bybit runtime helper surface aligned with the shared health model.
4. Validate trade, book, and degraded runtime paths with deterministic fixtures.

## Problem Statement

The umbrella ingestion plan requires resilient multi-venue coverage, but only Binance has real parser and runtime surfaces today.

## Requirements

- Cover the MVP Bybit public spot/perp streams needed by the visibility initiative.
- Preserve source timestamps, source ordering metadata, symbol identity, and market type.
- Reuse `libs/go/ingestion` for normalization and health semantics.
- Keep Bybit-specific protocol handling inside `services/venue-bybit`.

## Out Of Scope

- Generic multi-venue runtime abstractions.
- Live websocket connectivity.

## Target Repo Areas

- `services/venue-bybit`
- `libs/go/ingestion`
- `tests/fixtures`

## ASCII Flow

```text
Bybit native payloads
        |
        v
services/venue-bybit parsers
        |
        v
shared ingestion messages
        |
        +--> normalization tests
        +--> runtime health helpers
```
