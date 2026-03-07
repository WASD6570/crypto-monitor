# Testing Plan: Symbol And Global Regime State

Expected output artifact: `plans/completed/symbol-and-global-regime-state/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Clean symbol classification | Feed healthy aligned 5m bucket summaries into symbol classification for `BTC-USD` and `ETH-USD` | Symbols emit `TRADEABLE` with stable reason ordering and version provenance | Unit plus integration assertions |
| Moderate degradation downgrade | Feed partial coverage loss, timestamp fallback pressure, or moderate fragmentation into one symbol | Symbol downgrades to `WATCH` with explicit reasons and trigger metrics | Service tests plus fixture output |
| Severe trust-loss hard stop | Feed severe fragmentation, unavailable composite side, or critical trust cap into a symbol window | Symbol emits `NO-OPERATE` on the first qualifying closed 5m window | Unit and integration assertions |
| Conservative recovery hysteresis | Run sequential windows that recover after `WATCH` or `NO-OPERATE` | State does not recover until the configured consecutive healthy windows are satisfied | Transition-sequence assertions |
| Global shared-failure ceiling | Run both symbols through a shared severe failure sequence | Global state becomes `NO-OPERATE` and caps both symbols accordingly | Integration and replay assertions |
| Symbol-local degradation without global ceiling | Run one symbol degraded and one symbol healthy | Global state remains less severe when shared conditions are absent; healthy symbol is not over-downgraded beyond the configured ceiling | Integration assertions |
| Replay determinism | Run the same pinned regime fixture window twice | Symbol and global outputs, reasons, transition metadata, and version fields match exactly | Replay digests and test equality |
| Config-version pinning | Re-run the same fixture with a different regime config version | Differences are intentional, bounded, and explained by changed config/version fields | Replay/config assertions |

## Required Commands

The implementing agent should provide these exact commands or repo-standard direct equivalents:

- `/usr/local/go/bin/go test ./libs/go/features/... -run 'TestRegimeClassification|TestRegimeDowngradePrecedence|TestRegimeRecoveryRequiresPersistence|TestRegimeThresholdEdgesAreDeterministic|TestGlobalCeilingRules|TestGlobalRecoveryRequiresPersistence'`
- `/usr/local/go/bin/go test ./services/regime-engine/... -run 'TestRegimeClassification|TestFragmentedMarketDowngrade|TestGlobalCeilingRules|TestGlobalCeilingDoesNotHideSymbolSpecificDifferences|TestWorldUSAMarketStateTransitionReasons'`
- `/usr/local/go/bin/go test ./tests/integration/... -run 'TestWorldUSAMarketStateTransitions|TestWorldUSAGlobalCeilingTransitions|TestWorldUSASymbolSpecificDegradeWithoutGlobalStop'`
- `/usr/local/go/bin/go test ./tests/replay/... -run 'TestWorldUSAReplayDeterminism|TestWorldUSARegimeReplayDeterminism|TestWorldUSARegimeConfigVersionPinning|TestWorldUSALateEventReplayCorrectionDoesNotChangeLiveTransitionRules'`

If package paths differ during implementation, replace them with equally explicit commands that still isolate symbol classification, global ceiling logic, hysteresis, and replay determinism.

## Verification Checklist

- Symbol classification consumes completed 5m bucket summaries without recomputing upstream bucket math.
- `TRADEABLE`, `WATCH`, and `NO-OPERATE` outputs include ordered reason codes, primary reason, trigger metrics, and version provenance.
- Downgrades happen immediately on configured severe trust-loss conditions.
- Recovery requires deterministic consecutive healthy windows and does not use wall-clock timers.
- Global state acts only as a ceiling and preserves symbol-specific differences when shared conditions do not justify a full market-wide downgrade.
- Repeated replay runs with the same inputs and config emit identical symbol/global states and reason ordering.
- Config-version changes are auditable through output metadata and explainable state differences.
- Python remains optional and offline-only.

## Negative Cases

- one symbol has no trustworthy 5m bucket because upstream completeness is too low
- one symbol oscillates around a threshold and would flap without hysteresis
- both symbols degrade for different local reasons that should not force an automatic global `NO-OPERATE`
- timestamp-trust loss is severe enough to cap confidence even when price alignment looks clean
- late-event replay correction reproduces the same pinned regime result for the same ordered replay scope
- one symbol recovers sooner than the global ceiling, and the ceiling must remain active until shared persistence is satisfied

## Handoff Notes

- Use deterministic local fixtures and pinned replay seeds only; no live venue calls or credentials.
- Keep tests focused on the conservative trust gate and shared ceiling behavior, not query-contract or storage concerns.
- The feature is done only when another agent can run the commands above and produce `plans/completed/symbol-and-global-regime-state/testing-report.md` with symbol/global transition evidence.
