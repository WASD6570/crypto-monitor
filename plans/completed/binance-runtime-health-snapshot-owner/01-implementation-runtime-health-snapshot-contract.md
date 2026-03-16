# Implementation Module 1: Runtime Health Snapshot Contract

## Scope

- Define the internal snapshot types, field semantics, and status-mapping rules for the command-owned Binance runtime-health view.
- Cover per-symbol readiness, feed-health state, degradation reasons, connection posture, depth posture, and stable ordering.
- Exclude public HTTP exposure, runbook rewrites, and any change to current-state payloads.

## Target Repo Areas

- `cmd/market-state-api`
- `services/venue-binance` only for narrow helper exports if the existing runtime surfaces cannot be consumed cleanly as-is

## Requirements

- The snapshot contract must be internal and implementation-ready for later API serialization, but this feature should not yet make it a public HTTP schema.
- Use the shared feed-health vocabulary exactly; do not rename states or reasons.
- Distinguish read-model readiness from feed degradation so operators can tell the difference between startup warm-up and a degraded but still readable runtime.
- Preserve deterministic symbol ordering and deterministic field presence rules across repeated identical runtime inputs.
- Prefer existing venue/runtime structs and typed enums over string duplication.

## Key Decisions

- Use one command-local snapshot root with a generation time and one entry per tracked symbol.
- Represent each symbol with explicit fields for `symbol`, readiness, feed-health state/reasons, connection state, depth status, and selected timestamps/counters needed later by the endpoint and runbooks.
- Derive snapshot state from the sustained runtime owner plus completed supervisor/depth-recovery surfaces rather than by re-evaluating policy from raw frames.
- Treat missing publishable observations during startup as a first-class readiness state instead of inventing placeholder market-state values.

## Unit Test Expectations

- Startup before any publishable observation produces stable not-ready snapshot entries for both symbols.
- Healthy runtime inputs produce `HEALTHY` with no lingering degraded reasons.
- Reconnect, stale, resync, and rate-limit paths preserve canonical reasons and explicit readiness posture.
- Repeated identical runtime inputs produce byte-for-byte or field-for-field equivalent snapshots in deterministic symbol order.

## Contract / Fixture / Replay Impacts

- No public schema change is expected in this feature.
- Replay-sensitive assumptions remain implicit in the runtime inputs; keep status mapping deterministic so the next endpoint feature can expose it without reopening semantics.
- Add focused test fixtures only if they improve deterministic coverage of warm-up and degradation transitions.

## Summary

This module settles the internal status contract so later handlers and runbooks can depend on one explicit operator-facing runtime-health model instead of reassembling health state from scattered runtime pieces.
