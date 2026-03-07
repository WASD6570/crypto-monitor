# Child Plan Seeds: World USA Composites And Market State

## `world-usa-composite-snapshots` (completed)

- Outcome: Go services emit trusted WORLD and USA composite snapshots for `BTC-USD` and `ETH-USD`, including venue eligibility, quote-normalization mode, weighting, clamping, degraded reasons, and contributor provenance.
- Primary repo areas: `services/feature-engine`, `libs/go`, `configs/*`, `schemas/json/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Depends on: `plans/completed/canonical-contracts-and-fixtures/`, `plans/completed/market-ingestion-and-feed-health/`, `plans/completed/raw-event-log-boundary/` for replay-ready raw facts
- Validation shape: `go test ./libs/go/... -run 'Test(CompositeWeighting|StablecoinNormalization|CompositeClamping)'`, `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeConstruction|CompositeDegradedVenueHandling)'`, and a focused replay check such as `go test ./tests/replay/... -run 'TestWorldUSACompositeDeterminism'`
- Why it stands alone: it establishes the smallest service-trust boundary every later bucket, regime, dashboard, and alert consumer needs.
- Archive: `plans/completed/world-usa-composite-snapshots/`

## `market-quality-and-divergence-buckets` (completed)

- Outcome: the feature engine emits deterministic 30s, 2m, and 5m bucket families for WORLD/USA divergence, fragmentation, coverage, timestamp trust, and market-quality summaries without mixing in final regime decisions.
- Primary repo areas: `services/feature-engine`, `libs/go`, `configs/*`, `schemas/json/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Depends on: `plans/completed/world-usa-composite-snapshots/`, operating defaults in `docs/specs/crypto-market-copilot-program/03-operating-defaults.md`
- Validation shape: `go test ./services/feature-engine/... -run 'Test(BucketAssignment|DivergenceMetrics|LateEventHandling)'`, `go test ./tests/integration/... -run 'TestWorldUSABucketReplayWindow'`, and `go test ./tests/replay/... -run 'TestWorldUSABucketDeterminism'`
- Why it stands alone: bucket assignment and divergence/quality math are a separate deterministic layer that both regime logic and consumer read models can reuse.
- Archive: `plans/completed/market-quality-and-divergence-buckets/`

## `symbol-and-global-regime-state` (completed)

- Outcome: the regime engine publishes 5m symbol state and global ceiling state as `TRADEABLE`, `WATCH`, or `NO-OPERATE`, with conservative downgrade/recovery rules and explicit trigger reasons.
- Primary repo areas: `services/regime-engine`, `libs/go`, `configs/*`, `schemas/json/features`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Depends on: `plans/completed/market-quality-and-divergence-buckets/`, feed-health semantics from `plans/completed/market-ingestion-and-feed-health/`
- Validation shape: `go test ./services/regime-engine/... -run 'Test(RegimeClassification|FragmentedMarketDowngrade|GlobalCeilingRules)'`, `go test ./tests/integration/... -run 'TestWorldUSAMarketStateTransitions'`, and `go test ./tests/replay/... -run 'TestWorldUSAReplayDeterminism'`
- Why it stands alone: it isolates the user-facing trust gate and ceiling policy from lower-level feature math, which keeps later alert consumers and dashboards aligned on one authoritative state output.
- Archive: `plans/completed/symbol-and-global-regime-state/`

## `market-state-current-query-contracts` (completed)

- Outcome: services expose versioned current-state read models for dashboard and service consumers, including current WORLD/USA composites, divergence summaries, trust metadata, symbol regime, global ceiling, and a small recent-context window without client recomputation.
- Primary repo areas: `schemas/json/features`, `services/feature-engine`, `services/regime-engine`, `tests/fixtures`, `tests/integration`, `apps/web` as a consuming surface only
- Depends on: `plans/completed/world-usa-composite-snapshots/`, `plans/completed/market-quality-and-divergence-buckets/`, `plans/completed/symbol-and-global-regime-state/`, `plans/epics/visibility-dashboard-core/92-refinement-handoff.md`
- Validation shape: `go test ./services/feature-engine/... -run 'Test(MarketStateQueryResponse|CompositeSnapshotSchema)'`, `go test ./services/regime-engine/... -run 'Test(MarketRegimeSnapshotSchema|HistoricalStateVersionContext)'`, and a dashboard-consumer smoke such as `pnpm --dir apps/web test -- --runInBand`
- Why it stands alone: it unblocks the dashboard query-adapter child and future service consumers while preserving the rule that the UI formats market state but never computes it.
- Archive: `plans/completed/market-state-current-query-contracts/`

## `market-state-history-and-audit-reads` (completed)

- Outcome: replay-aware historical query surfaces can retrieve market-state artifacts by symbol, bucket, and version context, including replay-corrected state and audit provenance for dashboards and service-side investigations.
- Primary repo areas: `services/replay-engine`, `services/feature-engine`, `services/regime-engine`, `schemas/json/features`, `schemas/json/replay`, `tests/fixtures`, `tests/integration`, `tests/replay`
- Depends on: `plans/completed/market-state-current-query-contracts/`, `plans/completed/raw-storage-and-replay-foundation/`, and the completed replay child slices that preserve partition, ordering, and retention provenance
- Validation shape: `go test ./tests/replay/... -run 'TestWorldUSALateEventReplayCorrection'`, `go test ./tests/integration/... -run 'TestWorldUSAConfigVersionPinnedReplay'`, and replay-history contract checks for bucket/version lookup fidelity
- Why it stands alone: historical and replay-corrected reads are valuable for auditability, but they carry upstream replay/storage dependency risk and should not delay current-state consumer delivery.
- Archive: `plans/completed/market-state-history-and-audit-reads/`
