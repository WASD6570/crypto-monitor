# Implementation Module 3: API Compatibility And Deterministic Proof

## Scope

- Add focused proof that the new USD-M output application is deterministic, backward-compatible, and understandable from the API.
- Cover route-level API behavior, replay determinism, and the focused integration cases for `BTC-USD` and `ETH-USD`.
- Exclude broader rollout hardening and long-run soak work.

## Target Repo Areas

- `services/market-state-api`
- `tests/integration`
- `tests/replay`
- touched schemas/tests under `libs/go/features`

## Requirements

- Prove auxiliary/no-context/degraded-context inputs keep the prior spot-derived output stable.
- Prove `DEGRADE_CAP` changes output only in the bounded planned way and that the response exposes enough provenance to explain that change.
- Prove repeated pinned Spot plus USD-M inputs produce byte-stable or value-stable symbol/global current-state responses.
- Keep the route contract backward-compatible aside from the planned additive provenance metadata.

## Key Decisions

- Reuse the archived first-child fixtures and add the smallest follow-on cases that exercise actual consumer-facing changes.
- Cover both symbol and global responses because global state may inherit capped symbol posture.
- Prefer focused route/integration tests over broad end-to-end harness expansion.
- Record the exact commands and expected evidence in the testing file so `feature-implementing` and `feature-testing` can hand off cleanly.

## Unit Test Expectations

- API handler/provider tests assert the additive provenance field shape and bounded state changes.
- Integration tests cover one unchanged case and one degrade-cap case for both tracked symbols.
- Replay tests prove stable repeated outputs for the same pinned current-state and USD-M inputs.
- Regression tests confirm `/healthz` and unrelated runtime-status behavior remain unchanged.

## Contract / Fixture / Replay Impacts

- Update any touched JSON schemas and schema-loading tests together.
- Extend replay fixtures only as far as needed to pin the new application behavior.
- Keep fixture ordering, timestamps, and version metadata explicit so repeated runs remain deterministic.

## Summary

This module turns the plan into evidence: the output change is conservative, auditable, and stable across repeated current-state and replay runs.
