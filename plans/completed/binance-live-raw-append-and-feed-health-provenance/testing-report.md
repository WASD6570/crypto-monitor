# Testing Report: Binance Live Raw Append And Feed-Health Provenance

## Result

- Status: passed
- Date: 2026-03-09

## Commands

- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestRawWriteBoundary|TestRawPartitionRouting|TestBuildRawAppendEntry' -v`
- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'Test.*RawAppend|Test(Spot|USDM).*' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinance.*RawAppend' -v`
- `"/usr/local/go/bin/go" test ./tests/replay -run 'TestReplay.*Retention|TestReplay.*Determinism' -v`

## Coverage Notes

- Shared raw append duplicate identity now stays distinct across Binance USD-M funding and mark-index families that share exchange timestamps.
- Venue-owned raw append seams now provide stable websocket and REST provenance for Spot, Spot depth health, USD-M websocket, and USD-M open-interest paths.
- Integration coverage proves Spot websocket raw append context, degraded Spot depth feed-health linkage, USD-M mixed-surface provenance, and deterministic duplicate behavior for message-ID and sequence-based identities.
- Replay retention guardrails still accept the touched raw append contract without determinism drift.
