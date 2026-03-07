# Testing

## Testing Goals

- prove the current-state contract family is versioned, complete, and stable for dashboard and service consumers
- prove symbol and global current-state query surfaces return service-owned outputs without client recomputation
- prove degraded, unavailable, and bounded recent-context cases are explicit in responses
- prove repeated replay of the same fixture window emits identical current-state payloads and version metadata

## Output Artifact

- Write the implementation-phase test report to `plans/completed/market-state-current-query-contracts/testing-report.md`.

## Test Matrix

### 1. Schema And Fixture Contract Validation

- Purpose: validate the current-state schema family and fixture-backed contract examples.
- Target repo areas: `schemas/json/features`, `tests/fixtures`, `tests/integration`
- Fixtures:
  - healthy current BTC symbol payload
  - degraded ETH payload with excluded contributors and timestamp fallback markers
  - unavailable composite case packaged inside a valid current-state response
- Validation commands:
  - `go test ./services/feature-engine/... -run 'Test(MarketStateCurrentSchema|MarketStateCurrentResponseShape|CompositeSnapshotSchemaCompatibility)'`
  - `go test ./services/regime-engine/... -run 'Test(MarketStateCurrentRegimeSection|MarketStateCurrentGlobalSchema)'`
- Verify:
  - all required sections are present with explicit version metadata
  - degraded and unavailable states use machine-readable reason codes
  - fixtures validate without adding history/audit-only fields

### 2. Symbol Current-State Query Assembly

- Purpose: verify the symbol current-state surface packages composite, bucket, and regime data in one response.
- Target repo areas: `services/feature-engine`, `services/regime-engine`, `tests/integration`
- Fixtures:
  - healthy aligned BTC market
  - fragmented ETH market with partial USA coverage loss
  - symbol response with recent-context gaps represented explicitly
- Validation commands:
  - `go test ./services/feature-engine/... -run 'Test(SymbolCurrentStateQuery|CurrentStateRecentContext|CurrentStateUnavailableSections)'`
  - `go test ./tests/integration/... -run 'TestMarketStateCurrentSymbolQuery|TestMarketStateCurrentRecentContextOrdering'`
- Verify:
  - one symbol response is sufficient to render current trust state
  - recent context uses only closed windows and preserves deterministic ordering
  - missing or unavailable upstream sections stay explicit and never silently disappear

### 3. Global Ceiling Query And Consumer Trust Mapping

- Purpose: verify the global current-state surface and capped symbol summaries stay aligned with regime outputs.
- Target repo areas: `services/regime-engine`, `tests/integration`, `apps/web` as a consuming seam only
- Fixtures:
  - both symbols healthy with `TRADEABLE` ceiling
  - one symbol degraded while the global ceiling remains `WATCH`
  - global `NO-OPERATE` case that caps both symbols
- Validation commands:
  - `go test ./services/regime-engine/... -run 'Test(GlobalCurrentStateQuery|GlobalCeilingAppliedToSymbolResponse|CurrentStateTransitionReasons)'`
  - `go test ./tests/integration/... -run 'TestMarketStateCurrentGlobalQuery|TestMarketStateCurrentConsumerContractSeam'`
- Verify:
  - symbol responses carry the effective global ceiling context
  - dashboard/service consumers can trust service-owned capped state without reapplying ceiling logic
  - transport contracts remain UI-neutral while still readable by dashboard adapters

### 4. Replay Determinism And Version Pinning

- Purpose: prove current-state payloads stay replay-safe and explainable across pinned inputs.
- Target repo areas: `tests/replay`, `tests/integration`
- Fixtures:
  - pinned BTC one-day window
  - pinned ETH one-day window with degraded feeds
  - config-version change fixture showing a deliberate version-context difference
- Validation commands:
  - `go test ./tests/replay/... -run 'TestMarketStateCurrentReplayDeterminism|TestMarketStateCurrentVersionPinning'`
  - `go test ./tests/integration/... -run 'TestMarketStateCurrentConfigVersionContext'`
- Verify:
  - repeated runs with identical fixtures emit identical current-state payloads
  - version metadata explains intentional output differences across config versions
  - current-state responses remain bounded and do not turn into history retrieval APIs during replay testing

## Required Negative Cases

- no healthy WORLD contributors but the response contract still returns explicit unavailable composite state
- no healthy USA contributors while global ceiling remains present
- mixed timestamp trust across recent-context buckets
- regime output available but one recent-context bucket missing
- config version changes between adjacent replayed windows
- consumer attempts to rely on absent history fields are rejected by schema shape

## Exit Criteria For Implementation

- targeted schema, integration, and replay commands pass
- symbol and global current-state responses expose enough provenance for dashboard and service consumers without extra joins
- bounded recent-context behavior is explicit and clearly separate from later history/audit reads
- implementation evidence is recorded in `plans/completed/market-state-current-query-contracts/testing-report.md`
