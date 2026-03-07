# Testing Report

- Feature: `market-state-history-and-audit-reads`
- Scope: version-pinned historical market-state wrappers, global history lookup, and replay-corrected audit provenance.

## Commands

- `/usr/local/go/bin/go test ./services/feature-engine/... -run 'TestMarketStateHistory(ReadByBucketKey|UnavailableOnPinMismatch)'`
- `/usr/local/go/bin/go test ./services/regime-engine/... -run 'TestMarketStateHistory(GlobalLookup|VersionPinnedContext)'`
- `/usr/local/go/bin/go test ./services/replay-engine/... -run 'TestMarketStateAuditProvenance(AuthoritativeOriginal|ReplayCorrected)'`
- `/usr/local/go/bin/go test ./tests/integration/... -run 'TestMarketStateHistory(ClosedWindowLookup|VersionPinnedBucketContext|AuditProvenanceConsistency)'`
- `/usr/local/go/bin/go test ./tests/replay/... -run 'TestMarketStateHistoryReplay(LateEventCorrection|ConfigVersionPinnedLookup|DeterministicAuditLineage)'`

## Results

- All targeted Go service, integration, and replay commands passed in this session.

## Notes

- Historical symbol/global reads stay aligned to the completed current-state response family by wrapping the existing state payload and validating exact lookup tuples at the service boundary.
- Replay audit provenance is bounded to authoritative lineage plus one optional superseded lineage, which keeps correction semantics machine-readable without reopening raw-event inspection flows.
