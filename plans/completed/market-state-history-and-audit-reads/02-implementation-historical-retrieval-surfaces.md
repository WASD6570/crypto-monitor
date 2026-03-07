# Implementation Historical Retrieval Surfaces

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `services/regime-engine`, `services/replay-engine`, `libs/go`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Define the Go-owned query surfaces and lookup algorithm for historical market-state retrieval.
- Keep retrieval bounded to closed windows and version-pinned contexts backed by replay-safe artifacts.

## In Scope

- symbol historical current-state query surface for `BTC-USD` and `ETH-USD`
- global historical current-state query surface for the global ceiling and capped symbol summaries at a closed window
- exact-match lookup logic for bucket key or closed `asOf` plus explicit version context
- assembly rules that join stored composite, bucket, regime, and replay metadata into the history response family
- deterministic handling of unavailable artifacts and mismatched version pins

## Out Of Scope

- new storage retention policy, artifact backfills, or replay orchestration changes
- fuzzy time search, nearest-neighbor fallback, or implicit version selection
- UI adapter logic or consumer-side recovery when history is unavailable

## Service Surface Recommendations

- Keep retrieval service-owned in Go with the same trust boundary as the current-state reads.
- Recommended surfaces:
  - `GET /api/market-state/{symbol}/history/{bucketFamily}/{bucketKey}` with required version-pinning query parameters or equivalent RPC
  - `GET /api/market-state/global/history/{bucketFamily}/{bucketKey}` with the same version-pinning context or equivalent RPC
  - `GET /api/market-state/{symbol}/audit/{bucketFamily}/{bucketKey}` or equivalent RPC for lineage-only reads when state payload reuse is not required by the caller
- Require callers to provide either an exact closed bucket key or exact closed `asOf`; do not accept open-ended ranges in this slice.

## Assembly Rules

- Resolve historical reads from closed composite, bucket, and regime artifacts that share the requested version tuple.
- Prefer exact version-pin matching over best-effort assembly; when the tuple does not resolve, return an explicit unavailable response with machine-readable reason codes.
- Reuse the current-state response assembler for nested state sections so historical and current answers stay structurally aligned.
- Include bounded recent-context only when the underlying closed artifacts for that same resolved context exist; never mix windows from a different replay or config version.
- Preserve idempotency: repeated reads for the same lookup tuple must return byte-equivalent payloads apart from transport-level metadata such as request id.

## Failure And Gap Handling

- Missing artifacts: return unavailable status with the unresolved lookup tuple echoed back.
- Version mismatch: return explicit pin-mismatch status rather than silently selecting the newest version.
- Partial lineage: return the historical state only if the state tuple is authoritative; otherwise fail closed and mark the read unavailable.
- Replay supersession: return the corrected state and mark the response as superseded-from another lineage entry when a replay has replaced the earlier answer.

## Test Expectations

- integration tests for healthy historical symbol and global reads across `30s`, `2m`, and `5m` bucket families
- integration tests for pin-mismatch and missing-artifact negative cases
- replay tests proving repeated pinned lookups return the same payload and that late-event corrections resolve to the replay-corrected authoritative state

## Summary

This module defines how Go services resolve historical state from replay-safe closed artifacts. The core boundary is exact, version-pinned retrieval with explicit unavailable responses instead of best-effort fallback or consumer-side joins.
