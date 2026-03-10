# Refinement Handoff: Binance Live Contract Seams And Fixtures

## Next Recommended Child Feature

- `binance-live-identity-and-time-policy`

## Why This Is Next

- every later Binance live epic depends on canonical identity, source-record ID, and exchange-time rules being explicit first
- existing shared schema families already exist, so the highest-value missing work is not more family scaffolding but Binance-specific seam locking
- fixture expansion and validation work would otherwise risk encoding conflicting semantics

## Safe Parallel Planning

- none yet
- `binance-live-fixture-corpus-expansion` should wait for `binance-live-identity-and-time-policy`
- `binance-contract-validation-and-runbook-alignment` should wait for both earlier child features so it validates settled rules instead of drafting them

## Blocked Until

- `binance-live-fixture-corpus-expansion` is blocked on written source-record ID and timestamp rules for Spot `bookTicker`, Spot depth handling, USD-M `markPrice`, USD-M `forceOrder`, and REST `openInterest`
- `binance-contract-validation-and-runbook-alignment` is blocked on the final fixture matrix and policy decisions from the first two child features

## What Already Exists

- `plans/completed/canonical-contracts-and-fixtures/` already supplies the shared event-family and fixture-manifest foundation
- `plans/completed/market-ingestion-and-feed-health/` already supplies feed-health vocabulary and the rule that gaps and stale inputs degrade explicitly
- `tests/fixtures/events/binance/` already contains starter cases that the next child plan should treat as seed material rather than complete coverage

## Assumptions To Preserve

- canonical downstream symbols remain `BTC-USD` and `ETH-USD`
- provenance remains explicit through `sourceSymbol`, `quoteCurrency`, `venue`, and `marketType`
- `recvTs` stays mandatory for every accepted event or sensor sample
- the epic only locks contract seams and fixtures; runtime implementation belongs to later epics

## Recommended Follow-On After This Child Feature

1. `binance-live-fixture-corpus-expansion`
2. `binance-contract-validation-and-runbook-alignment`
