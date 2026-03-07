# Implementation: Evaluation Storage And Consumer Contracts

## Module Requirements And Scope

- Define immutable storage and versioned contracts for alert outcome records.
- Make the outcome record queryable by review surfaces, replay tooling, and baseline comparison without client recomputation.
- Preserve a clean append-only correction model for replay and config evolution.

## Target Repo Areas

- `services/outcome-engine`
- `schemas/json/outcomes`
- `schemas/json/alerts`
- `schemas/json/replay`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Storage Principles

- Outcome records are service-owned and append-only.
- One alert may accumulate multiple outcome artifacts over time due to normal horizon closure progression and later replay correction.
- Consumers should read the latest non-superseded artifact by default, while retaining access to historical versions for audit.
- Storage retention should follow operating defaults for alerts and outcomes: hot for at least 180 days and retained for at least 2 years.

## Record Shape

Recommended top-level fields:

- `outcomeRecordId`
- `alertId`
- `alertVersion`
- `symbol`
- `setupFamily`
- `baselineId` when applicable
- `direction`
- `evaluationSource`
- `status` as `PENDING`, `PARTIAL`, or `COMPLETE`
- `openedAt`
- `latestClosedAt`
- `horizons[]`
- `configVersion`
- `algorithmVersion`
- `costModelVersion`
- `schemaVersion`
- `replayProvenance`
- `supersedesOutcomeRecordId` when a replay correction replaces a prior record
- `createdAt`

Each `horizons[]` entry should include:

- `horizon`
- `result`
- decisive timestamps and duration fields
- target/invalidation/timeout details
- MAE and MFE metrics
- regime attribution block
- net-viability block
- trusted-data coverage flags

## Consumer Contracts

### Review UI And API Consumers

- Must be able to fetch the latest outcome record for an `alertId`.
- Must be able to slice by symbol, setup family, baseline, regime, and horizon.
- Must not recompute the decisive result client-side.

### Replay Consumers

- Must be able to resolve which config, schema, and cost model produced the stored record.
- Must write replay manifests or provenance references that explain whether the artifact is original live evaluation or replay-corrected.

### Baseline And Tuning Consumers

- Must be able to compare production and baseline records on the same key dimensions:
  - symbol
  - setup family or baseline ID
  - horizon
  - open regime
  - config version
  - cost-model version

## Correction And Supersession Model

- Live evaluation may first write `PENDING`, then `PARTIAL`, then `COMPLETE` artifacts as horizons close.
- Replay corrections should append a new artifact, mark the prior artifact superseded, and include a replay reason or manifest reference.
- Never silently rewrite an already reviewed outcome in place.

## Contract Boundaries

- `schemas/json/outcomes` should define canonical outcome payloads.
- `schemas/json/alerts` should carry the minimum evaluation inputs the outcome service depends on.
- `schemas/json/replay` should define provenance references or replay-manifest linkage rather than embedding large replay payloads directly into outcome records.

## Query And Review Expectations

- Recent alert drill-down should target the operating-default budget of under 5 seconds.
- 24h review queries should prefer precomputed outcome slices over recomputing raw market paths.
- Baseline and production review queries should be directly comparable without special-case transforms.

## Negative Cases To Preserve In Contract Design

- Missing threshold inputs in the alert payload should fail contract validation early.
- A consumer should not be able to confuse a replay correction with the original live artifact.
- A record with unknown net viability should remain queryable and countable without being coerced into positive or negative buckets.
- Regime attribution fields should tolerate missing downstream enrichment only by marking explicit unknown values, not by omitting required keys.

## Summary

This module turns outcome evaluation into a durable product artifact. It preserves auditability, baseline comparability, and replay-safe corrections while keeping all decisive logic on the service side.
