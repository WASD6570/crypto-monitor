# Implementation: Snapshot Refresh And Feed-Health Handoff

## Module Requirements

- Add explicit handling for periodic snapshot refresh, snapshot-stale degradation, and recovery-state feed-health emission without reopening the completed bootstrap alignment rules.
- Keep machine-readable health semantics aligned with the existing ingestion vocabulary from `plans/completed/market-ingestion-and-feed-health/`.
- Reuse current config defaults for refresh cadence, snapshot freshness, reconnect loops, and clock thresholds unless concrete implementation gaps force a narrow config addition.

## Target Repo Areas

- `services/venue-binance`
- `libs/go/ingestion`
- `services/normalizer`
- `configs/local/ingestion.v1.json`
- `configs/dev/ingestion.v1.json`
- `configs/prod/ingestion.v1.json`

## Key Decisions

- Keep refresh scheduling inside the venue-owned recovery path; `services/normalizer` should only see the resulting canonical depth and feed-health messages.
- Distinguish three related but different recovery concerns:
  - `gap-triggered resync` after explicit sequence loss
  - `refresh due` when the configured interval asks for a new snapshot even without a gap
  - `snapshot stale` when recovery or refresh has not restored recent snapshot state within the configured freshness window
- Prefer using existing feed-health reasons and precedence rules rather than inventing new severity vocabulary.
- Only update config files if implementation proves the current `snapshotRefreshPolicy`, `snapshotStaleAfterMs`, cooldown, or per-minute allowance fields cannot express the desired recovery behavior.

## Data And State Notes

- The recovery owner should track at minimum:
  - last accepted snapshot time
  - last successful recovery time
  - next refresh eligibility or due state
  - whether the current block is caused by cooldown, rate limit, reconnect posture, or stale snapshot state
- Feed-health emission should preserve:
  - `connectionState`
  - `messageFreshness`
  - `snapshotFreshness`
  - `sequenceGapDetected`
  - `clockState`
  - explicit degradation reasons in stable order

## Unit Test Expectations

- refresh-due logic requests a replacement snapshot only after the configured interval elapses
- snapshot-stale logic degrades to `STALE` when the last successful snapshot ages beyond the configured threshold even if messages are still arriving
- cooldown or rate-limit blocked refresh attempts remain machine-visible in the resulting health state
- stale precedence remains intact when stale freshness overlaps with degraded loop or connection reasons
- normalizer handoff preserves the chosen reasons and source identity for emitted feed-health events

## Contract / Fixture / Replay Notes

- If a canonical feed-health fixture needs richer reason combinations, update only the touched Binance fixtures and the shared manifest in the same slice.
- If the implementation needs a new source-record naming convention for recovery-originated feed-health messages, keep it explicit and deterministic so later replay work can reuse it.

## Summary

This module converts recovery decisions into stable operator-visible health output by attaching refresh cadence, snapshot freshness, and blocked-recovery posture to the existing feed-health boundary instead of hiding them in runtime-only state.
