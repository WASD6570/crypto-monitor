# Coinbase Adapter Foundation

## Ordered Implementation Plan

1. Add the first native Coinbase parser boundaries for MVP public market data.
2. Map Coinbase payloads into shared ingestion message types and feed shared normalization.
3. Add Coinbase runtime health helpers aligned with the common feed-health vocabulary.
4. Validate normalization and degraded runtime cases with deterministic tests.

## Problem Statement

Coinbase is still only planned at the umbrella level. It needs a bounded feature plan that can be implemented without reopening the whole ingestion umbrella.

## Requirements

- Cover MVP public Coinbase spot feeds needed by the initiative.
- Preserve exchange timestamps, receive timestamps, source identity, and relevant ordering metadata.
- Reuse shared normalization and feed-health semantics.

## Out Of Scope

- Authenticated channels.
- Live websocket integration.

## Target Repo Areas

- `services/venue-coinbase`
- `libs/go/ingestion`
- `tests/fixtures`

## ASCII Flow

```text
Coinbase native payloads
        |
        v
services/venue-coinbase parsers
        |
        v
shared ingestion messages
        |
        +--> normalization tests
        +--> runtime health helpers
```
