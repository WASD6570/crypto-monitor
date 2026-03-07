# Normalizer Feed Health Handoff

## Ordered Implementation Plan

1. Define the adapter-to-normalizer input boundary for shared ingestion messages and feed-health status.
2. Add the minimal `services/normalizer` entry points that accept venue outputs and emit canonical event/health outputs.
3. Validate timestamp, degradation, and source metadata preservation through the handoff.
4. Add deterministic service-level tests for the handoff layer.

## Problem Statement

Shared normalization logic exists in `libs/go/ingestion`, but the explicit service boundary in `services/normalizer` is still missing from the runnable ingestion path.

## Requirements

- Keep venue-specific parsing in venue services.
- Use `services/normalizer` as the canonical handoff point.
- Preserve both event outputs and feed-health outputs.
- Preserve `exchangeTs`, `recvTs`, venue, symbol, market type, and degraded markers.

## Out Of Scope

- Raw storage.
- Replay engine integration.
- UI consumers.

## Target Repo Areas

- `services/normalizer`
- `libs/go/ingestion`
- `tests/fixtures`

## ASCII Flow

```text
venue parser/runtime output
        |
        v
services/normalizer handoff
        |
        +--> canonical event stream
        +--> canonical feed-health stream
```
