# Testing Plan: Raw Storage And Replay Foundation

Expected output artifact: `plans/raw-storage-and-replay-foundation/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Raw persistence happy path | Ingest deterministic canonical fixtures into raw storage | Append-only records persist with symbol, venue, `exchangeTs`, `recvTs`, provenance, and degradation metadata | Integration test output + persisted manifest summary |
| Hot-to-cold retention continuity | Move scoped raw partitions from hot to cold tier and replay them | Replay resolves partitions without changing counts, ordering, or event identity | Retention test output + manifest/checksum comparison |
| Replay determinism | Replay the same symbol/day scope twice with the same config snapshot | Output digests, counts, ordering, late-event markers, and manifests are identical | Replay smoke output + checksum/diff evidence |
| Late-event correction path | Inject events beyond live watermarks and run inspect/compare replay | Late events persist, are marked late, and produce correction candidates without silent live mutation | Replay compare output |
| Backfill resume | Fail a scoped rebuild mid-run, then resume from checkpoint | Resume continues from durable checkpoint without duplicate materialization or audit gaps | Backfill integration output |
| Side-effect safety | Run inspect, rebuild, and apply-mode tests against instrumented sinks | Inspect/rebuild emit no alerts or webhooks; apply requires explicit gate and idempotency protection | Sink assertion output |
| Missing snapshot negative | Remove required config or contract snapshot reference and run replay | Replay fails clearly before rebuilding outputs | Error-path test output |
| Overlapping backfill risk | Submit overlapping apply-mode backfill requests for the same scope | Requests are serialized, rejected, or otherwise resolved deterministically with audit evidence | Control-plane integration output |

## Required Commands

The implementing agent should provide these exact commands or direct equivalents:

- `go test ./services/normalizer/...`
- `go test ./libs/go/...`
- `go test ./tests/integration -run RawStorage`
- `go test ./tests/integration -run Replay`
- `go test ./tests/integration -run Backfill`
- `go test ./tests/replay/...`
- `make replay-smoke SYMBOL=BTC-USD DAY=2026-01-15`
- `make retention-smoke RAW=1`
- optional offline parity command if added: `pytest tests/parity -k replay`

If package names or make targets differ during implementation, replace them with equally explicit repo-standard commands.

## Verification Checklist

- Raw storage is append-only and preserves canonical event identity plus ingest provenance.
- Partitioning supports one-day symbol replay without guessing hot vs cold locations.
- Retention transitions preserve replayability and audit metadata.
- Replay loads preserved config and contract snapshots rather than mutable current defaults.
- Repeated replay runs with identical inputs and config are byte-for-byte or digest-identical at the manifest/output level.
- Late and timestamp-degraded events remain visible and auditable.
- Backfill checkpoints support idempotent resume behavior.
- Replay and backfill default to no external side effects.
- Python remains optional and offline-only.

## Negative Cases

- Missing `recvTs` or missing timestamp-selection metadata should fail raw-persistence validation.
- Contract/version mismatch between stored events and replay runtime should fail before processing.
- Partition manifest gaps should fail replay rather than silently skip history.
- Duplicate event identities across reconnect windows should be surfaced and handled deterministically.
- Apply-mode replay without explicit gate or idempotency protection should be rejected.
- Resume request with a changed config snapshot should fail and require a new run.

## Replay-Risk Cases

- Out-of-order events with identical timestamps but different sequence availability
- Timestamp fallback from implausible `exchangeTs` to degraded `recvTs`
- Late events that would change already-materialized 30s, 2m, or 5m buckets
- Cold-storage replay for a historical day where some partitions require staged restore
- Rebuild of a scope that previously emitted operator-visible alerts
- Mixed healthy and degraded feed-health windows in the same replay scope

## Handoff Notes

- No exchange credentials should be required for these tests; use deterministic fixtures, local harnesses, or controlled persisted samples.
- The feature is done only when raw history, replay inputs, and correction flows are auditable enough that later state and alert features never have to guess how historical outputs were produced.
