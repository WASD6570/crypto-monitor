# Testing: Binance Live Raw Append And Feed-Health Provenance

## Validation Matrix

| Case | Flow | Expected Result |
|---|---|---|
| Spot websocket raw append | completed Spot trade or top-of-book output -> Binance seam helper -> shared raw builder | raw entry preserves websocket connection/session provenance, stable source identity, and expected partition routing |
| Spot depth degraded health | completed depth recovery output -> feed-health raw append | degraded depth output preserves stable depth source identity and explicit `degradedFeedRef` when recovery remains degraded |
| USD-M mixed-surface provenance | websocket-derived context plus REST-polled open-interest append to raw entries | raw append provenance stays distinct across websocket and REST surfaces while preserving canonical symbol identity |
| Duplicate-input stability | identical Binance inputs append twice | duplicate audit occurrence increments deterministically without partition or identity drift |
| Timestamp provenance retention | degraded timestamp fallback path appends to raw entry | raw entry preserves `bucketTimestampSource`, `timestampDegradationReason`, and late-event posture where applicable |
| Replay-readiness smoke | appended Binance raw entries are consumable by existing raw manifest/index helpers | raw entries validate and remain partition-resolvable for the later replay feature |

## Commands

### Shared Raw Append Boundary

- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestRawWriteBoundary|TestRawPartitionRouting|TestBuildRawAppendEntry' -v`

Expected coverage:
- raw append builder coverage for completed Binance families
- duplicate identity precedence and partition routing
- timestamp and degraded-feed provenance retention

### Binance Venue Seams

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'Test.*RawAppend|Test(Spot|USDM).*' -v`

Expected coverage:
- Binance seam helper behavior for Spot websocket, Spot depth recovery, USD-M websocket, and USD-M REST surfaces
- stable connection/session provenance and degraded-feed linkage

### Fixture-Backed Integration

- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinance.*RawAppend' -v`

Expected coverage:
- raw append proof for representative completed Binance families
- duplicate-input stability and mixed-surface provenance checks

### Replay-Readiness Guardrail

- `"/usr/local/go/bin/go" test ./tests/replay -run 'TestReplay.*Retention|TestReplay.*Determinism' -v`

Expected coverage:
- shared replay helpers still accept the touched raw append contract
- no drift is introduced to existing deterministic replay harness assumptions

## Inputs / Env

- Default deterministic suite requires no secrets.
- This feature should not require live Binance network access.
- If implementation adds any non-default build metadata or session reference fixtures, keep them deterministic and local.

## Verification Checklist

- each touched Binance surface emits a valid raw append entry with explicit market provenance
- degraded feed-health and timestamp fallback facts are visible in appended raw entries
- duplicate-input behavior is deterministic for representative Binance families
- raw partition lookup remains compatible with the existing replay manifest/index helpers
- any new fixture or integration expectations are captured in the active plan's testing report

## Testing Report Output

- While active: `plans/binance-live-raw-append-and-feed-health-provenance/testing-report.md`
- After implementation and validation complete: move the full directory to `plans/completed/binance-live-raw-append-and-feed-health-provenance/`
