# Implementation Module 2: Evaluator And Feature-Engine Integration

## Scope

- Implement the deterministic USD-M influence evaluator and integrate it into the Go-owned feature-engine path as an internal capability.
- Cover bounded policy decisions, degraded-context handling, and no-context fallback.
- Exclude application of the signal to current-state or regime outputs.

## Target Repo Areas

- `services/feature-engine`
- `libs/go/features`
- `services/venue-binance` only for narrow input-seam support
- focused unit tests in the same areas

## Requirements

- Keep the evaluator deterministic for the same accepted Spot plus USD-M input bundle.
- Default to auxiliary or degrade-cap posture only; no positive weighting or direct regime mutation in this child.
- Preserve Spot-only external behavior when USD-M context is absent.
- Keep the evaluator internal so the follow-on child can choose how to apply or surface the signal.
- Emit explicit reason codes and trigger metrics for later operator and replay auditability.

## Key Decisions

- Implement the evaluator in `services/feature-engine` because the policy belongs closest to market-state feature assembly rather than venue adapters.
- Keep any config or algorithm versioning explicit on the signal so follow-on application work can evolve independently from existing current-state response versions.
- Treat stale or degraded USD-M context as a bounded cap/degrade input, not as silent acceptance or outright positive influence.
- Add regression proof that existing current-state and regime responses remain unchanged until the follow-on child applies the signal.

## Unit Test Expectations

- Fresh, internally consistent USD-M inputs yield the expected auxiliary or degrade-cap signal.
- Missing context yields the explicit no-context signal and unchanged current external behavior.
- Degraded websocket or REST inputs yield degraded-context posture with stable reason ordering.
- Repeated identical inputs produce identical signal outputs across repeated runs.

## Contract / Fixture / Replay Impacts

- No public API contract changes are expected in this module.
- Unit tests should rely on deterministic in-memory inputs rather than live runtime ownership.
- If config/version constants are added, keep them explicit and easy for replay tests to assert.

## Summary

This module turns the policy contract into working Go evaluator behavior while intentionally stopping short of consumer-facing application, preserving the repo's plan to settle semantics before changing outputs.
