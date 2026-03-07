# Implementation Snapshot Output And Degradation Seams

## Module Requirements And Scope

- Target repo areas: `services/feature-engine`, `libs/go`, `schemas/json/features` if needed, `tests/fixtures`, `tests/integration`, `tests/replay`
- Define deterministic WORLD and USA snapshot outputs produced from the contributor pipeline.
- Preserve contributor provenance, degraded/unavailable semantics, and downstream seams without absorbing later bucket, regime, or query work.

## In Scope

- composite snapshot payload fields for WORLD and USA groups
- contributor provenance and exclusion-reason structure
- composite-level degraded, unavailable, and coverage semantics
- explicit version fields and downstream seam fields needed by later service-side consumers
- deterministic snapshot emission behavior for identical inputs and config

## Out Of Scope

- divergence bucket payloads or summaries
- regime output payloads
- current-state query response envelopes
- persistence schema, raw storage design, or dashboard presentation fields

## Recommended Repo Breakdown

- `services/feature-engine`: build and emit snapshot structs from finalized contributor decisions.
- `libs/go`: shared pure types/helpers for contributor summaries, degrade reason enums, and deterministic field assembly.
- `schemas/json/features`: update only if the live service-owned snapshot contract needs an explicit shared schema.
- `tests/fixtures`, `tests/integration`, `tests/replay`: fixture-backed snapshot assertions and deterministic replay coverage.

## Snapshot Boundary

The output of this feature is one authoritative snapshot per composite group and symbol for the active bucket boundary. That snapshot should be sufficient for later consumers to answer:

- which venues were configured, eligible, contributing, or excluded
- which quote-normalization mode was used
- what composite value and contribution weights were produced
- whether the snapshot is degraded or unavailable
- which config and algorithm versions governed the decision

## Required Snapshot Fields

- `symbol`
- `bucketTs`
- `compositeGroup` as `WORLD` or `USA`
- `priceBasis`
- `quoteNormalizationMode`
- `compositePrice` when available
- `contributors[]` with venue, market type, raw weight, final weight, penalties, include/exclude status, and reason codes
- `configuredContributorCount`
- `eligibleContributorCount`
- `contributingContributorCount`
- `coverageRatio`
- `healthScore` or equivalent bounded composite-trust field
- `degraded`
- `degradedReasons[]`
- `unavailable`
- `configVersion`
- `algorithmVersion`
- `schemaVersion`

## Degradation And Unavailable Policy

- `degraded` means a snapshot exists but trust is reduced because of contributor penalties, timestamp fallback, quote-proxy confidence loss, missing peers, or clamp concentration concerns.
- `unavailable` means the composite cannot be emitted as a trustworthy value because no contributors survived eligibility rules or the remaining data cannot satisfy minimum trust requirements.
- Do not collapse degraded and unavailable into one flag; later bucket and regime logic needs the distinction.
- Carry explicit reason codes so downstream consumers can aggregate trust loss without reverse-engineering contributor state.

## Downstream Seams To Preserve

- Later bucket work reads `compositePrice`, contributor coverage, concentration, timestamp trust, and degraded reason fields from this snapshot boundary.
- Later regime work reads bucketed outputs derived from these snapshots, not raw contributors directly.
- Later query-contract work may expose this snapshot, but that contract planning should not change contributor or degradation semantics here.
- The UI remains read-only and may label or sort snapshot fields, but must not recompute weighting, trust, or availability.

## Deterministic Emission Notes

- Keep contributor arrays in stable order.
- Keep reason codes explicit and enumerable rather than free-form.
- Emit unavailable snapshots consistently for the same no-contributor or below-minimum-trust scenarios.
- Preserve the timestamp source and config version so replay and future audit readers can explain why a snapshot differed across versions.

## Unit And Integration Test Expectations

- `go test ./services/feature-engine/... -run 'Test(WorldUSACompositeSnapshotShape|CompositeDegradedVenueHandling|CompositeUnavailableState)'`
- `go test ./tests/integration/... -run 'TestWorldUSACompositeSnapshotSeams'`
- `go test ./tests/replay/... -run 'TestWorldUSACompositeDeterminism'`

## Summary

This module locks the service-owned snapshot seam. It gives later features a deterministic composite artifact with provenance and trust metadata while keeping divergence buckets, regimes, and query envelopes clearly outside this slice.
