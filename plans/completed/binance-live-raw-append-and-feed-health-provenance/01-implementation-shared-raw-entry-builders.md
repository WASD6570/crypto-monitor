# Implementation: Shared Raw Entry Builders

## Module Requirements

- Ensure the shared raw append boundary in `libs/go/ingestion` can represent every completed Binance live family without losing provenance needed by later replay.
- Preserve current raw entry schema and duplicate-audit precedence unless a concrete Binance family exposes a narrow gap.
- Keep stream-family routing stable and explicit for Spot trades, top-of-book, order-book, feed-health, funding, mark-index, open-interest, and liquidation.

## Target Repo Areas

- `libs/go/ingestion`
- `libs/go/ingestion/raw_event_log.go`
- `libs/go/ingestion/raw_event_log_test.go`

## Key Decisions

- Extend shared builder coverage rather than creating Binance-only raw entry structs or encoders.
- Treat degraded feed-health and timestamp fallback as first-class raw provenance that later replay logic should not have to infer.
- Preserve partition routing as a shared rule: only stream families that already require dedicated partitions should continue to split by family.
- Keep duplicate identity precedence deterministic across Binance families, especially depth/order-book and mixed-surface USD-M health outputs.

## Data And Algorithm Notes

- Review whether the existing builder set already covers all completed Binance families; if any family is missing a builder or helper, add it in `libs/go/ingestion`.
- Confirm the builder inputs preserve:
  - canonical payload and schema version
  - stable `streamKey` and `streamFamily`
  - `sourceInstrumentID`
  - `connectionRef`, `sessionRef`, and optional `degradedFeedRef`
  - timestamp provenance (`exchangeTs`, `recvTs`, bucket source, fallback reason)
  - duplicate identity precedence (`venueMessageID` -> `sequence` + `streamKey` -> canonical ID)
- Keep any contract change narrowly scoped and backwards-compatible; if a shared raw field must change, update touched tests and consumer assumptions in the same slice.

## Unit Test Expectations

- raw append builders cover all completed Binance families without omitting market provenance
- depth and top-of-book families keep the intended stream-family routing and sequence-backed duplicate identity
- feed-health raw entries retain degraded-feed references and stable source-record identities
- USD-M REST and websocket-originated families remain distinguishable in raw append provenance
- duplicate inputs preserve identical partition keys and deterministic duplicate-audit increments

## Contract / Fixture / Replay Notes

- Avoid schema churn unless a concrete replay blocker is found.
- If a shared raw append field changes, update only the touched replay-sensitive tests and note the dependency in the later replay feature handoff.

## Summary

This module settles the shared raw append contract for completed Binance live families so later replay work can trust stable identity, partitioning, and degraded provenance without venue-specific storage logic.
