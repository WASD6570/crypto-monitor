# Implementation: Fixtures And Live Validation

## Module Requirements

- Add deterministic fixtures and integration coverage that prove the implemented recovery path before direct live Binance validation is run.
- Keep validation attached to the owning recovery implementation rather than creating a standalone smoke-only feature.
- Cover both the happy recovery path and failure-adjacent states that should remain explicit to downstream consumers.

## Target Repo Areas

- `tests/fixtures/events/binance`
- `tests/fixtures/manifest.v1.json`
- `tests/integration`
- `services/venue-binance`

## Key Decisions

- Reuse the completed bootstrap fixtures for startup where possible, then add only the minimum new fixture set required for recovery-specific behavior.
- Add at least one deterministic fixture for each recovery posture that materially changes downstream behavior:
  - sequence-gap followed by successful resync
  - sequence-gap with cooldown-blocked retry
  - sequence-gap with per-minute rate-limit block
  - snapshot-stale degradation while websocket messages still arrive
- Keep direct live validation attached to this implementation slice and gate any externally calling test with an explicit environment opt-in so default local runs stay deterministic.

## Integration Expectations

- integration proves the completed Spot supervisor can feed depth frames into the recovery owner without introducing a second transport boundary
- resync success path emits replacement canonical depth output only after the new snapshot plus alignment succeeds
- blocked recovery path emits feed-health degradation without pretending synchronization was restored
- snapshot-stale path remains machine-visible even when message freshness is still fresh

## Direct Live Validation Expectations

- Attach one public Binance validation path after implementation, for example:
  - REST check against `/api/v3/depth`
  - optional live WS + REST recovery probe behind an env gate such as `BINANCE_LIVE_VALIDATION=1`
- The live check should confirm at minimum:
  - snapshot response shape still matches the parser assumptions
  - replacement snapshots can still be matched to a known source symbol
  - degraded recovery posture stays visible if live resync cannot complete within policy bounds

## Unit Test Expectations

- fixture-backed integration proves successful resync clears the prior sequence-gap posture
- fixture-backed integration proves cooldown and rate-limit blocked recovery remain unsynchronized
- parity validation continues to pass after any new Binance recovery fixtures join the manifest

## Summary

This module closes the recovery proof gap with deterministic fixtures and one attached live validation step so the completed feature can be archived with both bounded tests and real Binance shape confirmation.
