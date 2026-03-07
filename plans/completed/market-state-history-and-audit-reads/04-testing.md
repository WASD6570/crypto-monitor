# Testing

## Test Matrix

- historical symbol read, healthy path: exact closed-window lookup returns the expected version-pinned state payload for `BTC-USD` and `ETH-USD`
- historical global read, healthy path: exact closed-window lookup returns the expected global ceiling payload and capped symbol summaries
- version pin mismatch: request fails closed with explicit unavailable or pin-mismatch status and no silent fallback
- replay correction path: late-event replay returns the corrected authoritative historical state plus superseded lineage
- audit provenance path: audit read exposes composite, bucket, regime, and replay provenance for the same lookup tuple as the history read

## Validation Commands

- `go test ./services/feature-engine/... -run 'TestMarketStateHistory(ReadByBucketKey|UnavailableOnPinMismatch)'`
- `go test ./services/regime-engine/... -run 'TestMarketStateHistory(GlobalLookup|VersionPinnedContext)'`
- `go test ./services/replay-engine/... -run 'TestMarketStateAuditProvenance(AuthoritativeOriginal|ReplayCorrected)'`
- `go test ./tests/integration/... -run 'TestMarketStateHistory(ClosedWindowLookup|VersionPinnedBucketContext|AuditProvenanceConsistency)'`
- `go test ./tests/replay/... -run 'TestMarketStateHistoryReplay(LateEventCorrection|ConfigVersionPinnedLookup|DeterministicAuditLineage)'`

## Execution Notes

- Use deterministic fixtures that already cover composite snapshots, bucket summaries, regime outputs, and replay manifests for the same symbols and closed windows.
- Run replay coverage against pinned manifests only; do not use live network dependencies.
- Record the exact fixture or manifest identifiers used for each passing command in `plans/completed/market-state-history-and-audit-reads/testing-report.md`.

## Verification Checklist

- historical reads resolve only closed windows and echo the exact lookup tuple in the response
- history responses reuse the current-state family sections and preserve `configVersion`, `algorithmVersion`, and schema metadata
- pin mismatch and artifact-gap cases fail closed with explicit machine-readable reason codes
- replay-corrected reads return the corrected authoritative state and identify the superseded lineage
- repeated replay runs with the same inputs produce identical state payloads and audit provenance ids

## Expected Report Artifact

- `plans/completed/market-state-history-and-audit-reads/testing-report.md`
