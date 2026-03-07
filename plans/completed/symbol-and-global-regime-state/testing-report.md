# Testing Report: Symbol And Global Regime State

## Outcome

- Passed deterministic 5m symbol regime coverage for `TRADEABLE`, `WATCH`, and `NO-OPERATE` classification with explicit reasons, trigger metrics, and recovery hysteresis.
- Passed global ceiling coverage for shared-failure downgrade, symbol-local preservation, and staged global recovery.
- Passed focused integration and replay checks for deterministic reason ordering, config-version provenance, and stable late-window transition behavior.

## Commands

1. `/usr/local/go/bin/go test ./libs/go/features/... -run 'TestRegimeClassification|TestRegimeDowngradePrecedence|TestRegimeRecoveryRequiresPersistence|TestRegimeThresholdEdgesAreDeterministic|TestGlobalCeilingRules|TestGlobalRecoveryRequiresPersistence|TestGlobalTransitionReasonsAreDeterministic|TestRegimeReasonsIncludeTriggerMetrics'`
   - Result: passed
2. `/usr/local/go/bin/go test ./services/regime-engine/... -run 'TestRegimeClassification|TestFragmentedMarketDowngrade|TestGlobalCeilingRules|TestGlobalCeilingDoesNotHideSymbolSpecificDifferences|TestWorldUSAMarketStateTransitionReasons'`
   - Result: passed
3. `/usr/local/go/bin/go test ./tests/integration/... -run 'TestWorldUSAMarketStateTransitions|TestWorldUSAGlobalCeilingTransitions|TestWorldUSASymbolSpecificDegradeWithoutGlobalStop'`
   - Result: passed
4. `/usr/local/go/bin/go test ./tests/replay/... -run 'TestWorldUSAReplayDeterminism|TestWorldUSARegimeReplayDeterminism|TestWorldUSARegimeConfigVersionPinning|TestWorldUSALateEventReplayCorrectionDoesNotChangeLiveTransitionRules'`
   - Result: passed

## Notes

- The runtime seam stays Go-only and consumes 5m `MarketQualityBucket` inputs directly instead of re-deriving bucket math.
- Global `appliedCeilingToSymbols` is emitted whenever the global state is not `TRADEABLE`, so downstream readers can see which configured symbols are under a market-wide ceiling.
- Assumption: `recvTs`-dominant 5m buckets and incomplete 5m windows are treated as immediate `NO-OPERATE` conditions to preserve the conservative trust gate described in the plan.
