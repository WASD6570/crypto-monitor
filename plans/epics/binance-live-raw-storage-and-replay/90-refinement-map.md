# Refinement Map: Binance Live Raw Storage And Replay

## Current Status

This Wave 5 epic is newly materialized from the initiative handoff and still needs decomposition before `feature-planning`.

## What Is Already Covered

- `plans/completed/binance-spot-trade-canonical-handoff/`, `plans/completed/binance-spot-top-of-book-canonical-handoff/`, `plans/completed/binance-spot-depth-bootstrap-and-buffering/`, and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` already settle the Binance Spot canonical event families that raw append must preserve instead of redefining.
- `plans/completed/binance-usdm-mark-funding-index-and-liquidation-runtime/`, `plans/completed/binance-usdm-open-interest-rest-polling/`, and `plans/completed/binance-usdm-context-sensor-fixtures-and-integration/` already settle the USD-M canonical event families and mixed-surface feed-health posture this epic must retain.
- `libs/go/ingestion/raw_event_log.go` and `libs/go/ingestion/raw_event_log_test.go` already provide raw append entry builders, duplicate audit facts, partition routing, and manifest lookup primitives for shared event families.
- `services/replay-engine/runtime.go`, `services/replay-engine/manifest_lookup.go`, and `tests/replay/*` already provide the baseline replay manifest, partition loading, ordering, and determinism harness.
- `tests/fixtures/manifest.v1.json` and the completed Binance fixture corpus already provide stable accepted inputs for later replay-sensitive validation.

## What Remains

- wire the completed Binance live families into the shared raw append boundary with stable connection/session/degraded-feed provenance and deterministic partition decisions
- confirm that Binance Spot and USD-M feed-health outputs survive raw append with explicit degraded reasons and source identities intact
- add replay-engine acceptance for Binance raw partitions and live event families without output drift across repeated runs
- prove duplicate-input and idempotency behavior for Binance live families using replay-sensitive tests rather than logs alone

## What This Epic Must Not Absorb

- new venue parsing or canonical schema work that belongs to already-completed Binance features
- market-state API cutover logic from the later cutover epic
- speculative storage backend or archive infrastructure redesign beyond the shared raw append/replay contracts already present
- smoke-only validation slices detached from the owning implementation work

## Refinement Waves

### Wave 5A

- `binance-live-raw-append-and-feed-health-provenance`
- Why first: replay is only trustworthy after the accepted Binance live families are appended with stable partition keys, duplicate identity, and degraded-feed provenance.

### Wave 5B

- `binance-live-replay-binance-family-determinism`
- Why next: replay acceptance and deterministic/idempotent behavior depend on the raw append identity and partition decisions from Wave 5A being settled.

## Notes For Future Planning

- preserve the existing shared raw append builders and extend them only where concrete Binance provenance or stream-family gaps remain
- keep duplicate-input and idempotency checks attached to the replay implementation slice instead of creating a validation-only follow-on
- use the completed Binance Spot and USD-M fixtures as the seed corpus for replay-sensitive tests wherever possible
- plan the replay slice only after the raw append child feature fixes the final partition and source-ID posture
