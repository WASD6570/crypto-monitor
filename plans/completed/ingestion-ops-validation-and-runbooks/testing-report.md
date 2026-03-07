# Ingestion Ops Validation And Runbooks Testing Report

## Commands

- `"/usr/local/go/bin/go" test ./services/venue-binance/... ./services/venue-bybit/... ./services/venue-coinbase/... ./services/venue-kraken/... ./services/normalizer/... ./libs/go/...`
- `"/usr/local/go/bin/go" test ./tests/integration -run Ingestion`
- `"/usr/local/go/bin/go" test ./tests/integration -run Ingestion -v`

## Results

All commands passed on 2026-03-07 in the local workspace.

## Smoke Matrix Evidence

| Case | Evidence |
|---|---|
| Adapter happy path | `TestIngestionAdapterHappyPath` |
| Gap path | `TestIngestionGapPathEmitsDeterministicResyncOutput` |
| Stale path | `TestIngestionStalePathPreservesFeedHealthOutput` |
| Retry safety | `TestIngestionRetrySafetyStaysBounded` |
| Runbook alignment | `TestIngestionRunbookAlignmentUsesSharedHealthVocabulary` |

## Runbook References

- Ops metric inventory and alert-condition matrix: `docs/runbooks/ingestion-feed-health-ops.md`
- Degraded-feed investigation steps: `docs/runbooks/degraded-feed-investigation.md`

## Verification Notes

- Validation stays local and deterministic; no live venue credentials or network access are required.
- Feed-health outputs now preserve degradation reasons through both order-book gap resync events and explicit feed-health normalization.
- Runbook language matches the shared feed states and degradation reasons emitted by the code.
