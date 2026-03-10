# Child Plan Seeds: Binance Live Contract Seams And Fixtures

## `binance-live-identity-and-time-policy`

- Outcome: the repo gains one explicit Binance live policy for canonical symbol mapping, market-type separation, source-record ID patterns, and stream-specific exchange-time selection across the chosen Spot and USD-M inputs.
- Primary repo areas: `schemas/json/events`, `libs/go/contracts`, `docs/runbooks`, `tests/integration`
- Depends on: `plans/completed/canonical-contracts-and-fixtures/`, `plans/completed/market-ingestion-and-feed-health/`, and the parent epic context in `plans/epics/binance-live-contract-seams-and-fixtures/`
- Validation shape: targeted integration and contract checks proving the chosen identity and timestamp rules are stable across duplicate, degraded, and mixed Spot/USD-M examples
- Why it stands alone: later fixture, runtime, replay, and API-cutover plans all depend on one written semantics package before code-level adapter work starts

## `binance-live-fixture-corpus-expansion`

- Outcome: the Binance fixture corpus covers the selected Spot and USD-M streams with paired raw/canonical examples for happy path, timestamp degradation, sequence-gap, duplicate or replay-sensitive, and mixed quote-context cases.
- Primary repo areas: `tests/fixtures/events/binance`, `tests/fixtures/manifest.v1.json`, `tests/integration`
- Depends on: `binance-live-identity-and-time-policy` and existing fixture conventions from `plans/completed/canonical-contracts-and-fixtures/`
- Validation shape: fixture manifest validation plus targeted fixture-backed tests proving expected canonical outputs exist for the chosen stream matrix
- Why it stands alone: fixtures are the shared truth set for later adapter, replay, and consumer tests and should be finished before runtime implementation broadens

## `binance-contract-validation-and-runbook-alignment`

- Outcome: schema expectations, consumer checks, and operator runbook notes explicitly reflect the final Binance live identity, timestamp, and fixture rules so later planners do not reinterpret them.
- Primary repo areas: `docs/runbooks`, `tests/integration`, optional `libs/go/contracts`, and the parent epic docs
- Depends on: `binance-live-identity-and-time-policy`, `binance-live-fixture-corpus-expansion`
- Validation shape: targeted validation commands that prove fixture indexing, schema compatibility, and runbook terminology stay aligned with the chosen contract seam
- Why it stands alone: it closes the epic by turning the chosen semantics into durable guidance and proof instead of leaving them as implicit planner knowledge
