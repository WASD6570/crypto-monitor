# Slow Context Query Surface And Freshness Testing Report

## Result

- Status: passed
- Date: 2026-03-08

## Validation Commands

```bash
/usr/local/go/bin/go test ./services/slow-context/... -run 'TestSlowContextFreshnessClassification|TestSlowContextUnavailableState|TestSlowContextLatestRevisionSelection'
/usr/local/go/bin/go test ./services/feature-engine/... -run 'TestCurrentStateSucceedsWhenSlowContextFails|TestSlowContextResponseExplicitlyUnavailable'
/usr/local/go/bin/go test ./services/slow-context/...
/usr/local/go/bin/go test ./services/feature-engine/...
```

## Notes

- Added accepted-record storage and latest-query assembly to `services/slow-context` without reopening the completed source-acquisition boundary.
- Added deterministic freshness classification for CME and ETF slow context with explicit `fresh`, `delayed`, `stale`, and `unavailable` states plus threshold-basis metadata.
- Added explicit unavailable response blocks and correction-aware latest revision selection for slow-context reads.
- Added non-blocking current-state integration in `services/feature-engine` so slow-context lookup failures degrade only the slow-context block.

## Assumptions

- Freshness becomes `delayed` after the next expected publish window closes and becomes `stale` only after an additional source-family stale duration (36h for CME, 48h for ETF); this preserves a distinct delayed state before stale.
- `services/feature-engine` can depend on the public slow-context query seam directly for now because this slice is still service-owned and has no consumer-facing schema requirement yet.
- Slow-context availability remains separate from existing market-state availability and does not change regime/composite/bucket semantics.
