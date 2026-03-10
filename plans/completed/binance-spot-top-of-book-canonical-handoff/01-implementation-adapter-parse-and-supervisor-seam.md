# Implementation: Adapter Parse And Supervisor Seam

## Module Requirements

- Add one bounded Binance Spot parser entrypoint for native `bookTicker` payloads and one helper that consumes supervisor-accepted Spot frames.
- Keep the parser scoped to `bookTicker` only; no depth snapshot, delta, or gap-recovery behavior belongs here.
- Preserve native `sourceSymbol`, best bid price, best ask price, `recvTs`, and any usable native update identifier needed by later canonical/raw identity.
- Prefer a native exchange timestamp only when the payload actually provides one; do not derive exchange time from local receipt or supervisor state.

## Target Repo Areas

- `services/venue-binance`
- `plans/completed/binance-spot-ws-runtime-supervisor/`

## Key Decisions

- Mirror the completed trade feature seam with explicit parser functions such as `ParseTopOfBookEvent(...)` and `ParseTopOfBookFrame(...)` so the supervisor remains the only runtime owner.
- Reuse `ingestion.OrderBookMessage` with `Type=top-of-book` and carry the Binance update ID through `Sequence` when available.
- Reject malformed payloads that omit `sourceSymbol`, best bid, or best ask rather than silently emitting partial top-of-book state.

## Unit Test Expectations

- native Binance `bookTicker` payloads parse into the shared ingestion order-book message model with `Type=top-of-book`
- supervisor-fed accepted `bookTicker` frames can be consumed by the parser without creating a second lifecycle owner
- native exchange time, when present, is passed through; missing exchange time stays empty so canonical timestamp fallback remains explicit
- Binance update ID, when present, survives parser output for later canonical/raw identity checks

## Summary

This module locks the adapter-side seam so later validation can prove Binance Spot top-of-book handoff uses the completed supervisor boundary and shared book model instead of bypassing them.
