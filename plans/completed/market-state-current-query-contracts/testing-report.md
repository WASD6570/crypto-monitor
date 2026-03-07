# Testing Report

- Feature: `market-state-current-query-contracts`
- Date: 2026-03-07

## Commands

- Passed: `/usr/local/go/bin/go test ./services/feature-engine/... -run 'Test(MarketStateCurrentSchema|MarketStateCurrentResponseShape|CompositeSnapshotSchemaCompatibility|SymbolCurrentStateQuery|CurrentStateRecentContext|CurrentStateUnavailableSections)'`
- Passed: `/usr/local/go/bin/go test ./services/regime-engine/... -run 'Test(MarketStateCurrentRegimeSection|MarketStateCurrentGlobalSchema|GlobalCurrentStateQuery|GlobalCeilingAppliedToSymbolResponse|CurrentStateTransitionReasons)'`
- Passed: `/usr/local/go/bin/go test ./tests/integration/... -run 'TestMarketStateCurrentSymbolQuery|TestMarketStateCurrentRecentContextOrdering|TestMarketStateCurrentGlobalQuery|TestMarketStateCurrentConsumerContractSeam|TestMarketStateCurrentConfigVersionContext'`
- Passed: `/usr/local/go/bin/go test ./tests/replay/... -run 'TestMarketStateCurrentReplayDeterminism|TestMarketStateCurrentVersionPinning'`

## Evidence

- Added versioned current-state schema files for symbol, global, response, and recent-context payload families.
- Added Go read-model builders for symbol/global current-state responses with bounded recent-context assembly and reserved history/audit seams.
- Added service query methods in `services/feature-engine` and `services/regime-engine` so consumers read one authoritative Go-owned contract.
- Verified degraded, partial, unavailable, consumer-seam, and version-pinning behavior with focused service, integration, and replay tests.

## Assumption

- Top-level current-state `configVersion` and `algorithmVersion` use symbol-regime version metadata when present, while provenance retains bucket and composite version context so the response stays backward-compatible without inventing a new live config source.
