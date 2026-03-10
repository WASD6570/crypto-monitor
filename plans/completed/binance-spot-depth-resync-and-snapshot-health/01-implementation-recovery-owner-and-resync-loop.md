# Implementation: Recovery Owner And Resync Loop

## Module Requirements

- Add a recovery-focused owner in `services/venue-binance` that starts from the completed bootstrap slice's synchronized state and handles sequence-gap or bootstrap-failure transitions explicitly.
- Reuse the completed Spot websocket supervisor as the only lifecycle owner for transport; this module only decides when depth state is synchronized, resyncing, blocked by cooldown/rate limit, or stale.
- Use existing runtime primitives for cooldown, retry allowance, reconnect-loop, and resync-loop evaluation.
- Ensure state transitions remain deterministic for BTC and ETH Spot symbols and never silently reaccept deltas after a gap.

## Target Repo Areas

- `services/venue-binance`
- `services/venue-binance/runtime.go`
- `services/venue-binance/runtime_test.go`
- `tests/integration`

## Key Decisions

- Add a bounded recovery state machine beside the completed bootstrap owner instead of folding all recovery rules into `Runtime` or `services/normalizer`.
- Model explicit recovery states such as `synchronized`, `resyncing`, `cooldown-blocked`, `rate-limit-blocked`, and `bootstrap-failed` so the later feed-health handoff does not need to infer hidden venue state.
- Record each actual snapshot recovery attempt through existing loop-state counters and history so `ResyncLoop` and per-minute allowance calculations remain authoritative.
- Treat a detected sequence gap as an immediate loss of synchronized depth state: accepted deltas stop, sequence-gap stays visible, and the recovery owner decides whether a snapshot retry can start now or must wait.
- Reset the shared sequencer only when a replacement snapshot is actually accepted; blocked retry states should not fabricate a clean depth baseline.

## Algorithm Notes

- The recovery owner should accept these inputs at minimum:
  - current synchronized source symbol and last accepted sequence
  - detected sequence gap or startup bootstrap failure
  - current loop state / counters
  - snapshot fetcher used by the completed bootstrap path
  - current time for cooldown and rate-limit evaluation
- Candidate transition flow:
  1. mark unsynchronized and preserve the triggering reason
  2. consult cooldown via `SnapshotRecoveryStatus(...)`
  3. consult rate limit via `SnapshotRecoveryRateLimitStatus(...)`
  4. if blocked, expose the blocking posture and retry timing without mutating accepted depth state
  5. if allowed, record a snapshot recovery attempt, enter resyncing, request a snapshot, and reuse bootstrap alignment semantics to re-establish synchronized depth state
  6. increment or reset resync counters based on success/failure and expose loop detection through runtime health

## Unit Test Expectations

- gap-triggered recovery enters explicit resync posture and stops treating depth as synchronized
- cooldown-blocked recovery reports remaining cooldown and does not fetch a replacement snapshot early
- rate-limit-blocked recovery reports retry timing and does not exceed configured per-minute allowance
- successful replacement snapshot plus aligned delta path resets the recovery state to synchronized and clears the sequence-gap flag
- repeated failed recovery attempts trip the configured resync-loop threshold visibly

## Contract / Fixture / Replay Notes

- No new canonical event family should be introduced; recovery state should continue to map to existing order-book and feed-health outputs.
- If the recovery owner introduces a narrow new status enum or internal state shape, keep it venue-local in `services/venue-binance` unless multiple venues demonstrably need the same contract.
- Preserve explicit triggering reasons so later replay/raw work can audit why depth state left synchronized mode.

## Summary

This module turns the completed bootstrap slice into a bounded recovery path by making resync entry, cooldown/rate-limit blocking, and loop-sensitive replacement snapshots explicit under the existing Binance Spot runtime owner.
