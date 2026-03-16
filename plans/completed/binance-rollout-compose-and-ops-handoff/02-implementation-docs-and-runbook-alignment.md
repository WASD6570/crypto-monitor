# Implementation: Docs And Runbook Alignment

## Module Requirements And Scope

- Update repo-facing startup docs to describe one prod-like startup posture everywhere.
- Keep the service README authoritative for route semantics, warm-up expectations, and override guardrails.
- Add or update operator runbook coverage so startup verification, `/api/runtime-status`, `/healthz`, and degraded-feed investigation form one coherent handoff.
- Remove stale wording that implies local-only startup behavior or future environment-specific defaults are part of the current supported path.

## Target Repo Areas

- `README.md`
- `services/market-state-api/README.md`
- `docs/runbooks/`

## Key Decisions

- Add one rollout-oriented runbook under `docs/runbooks/` for compose startup and verification instead of overloading the existing degradation runbooks with startup steps.
- Cross-link `docs/runbooks/ingestion-feed-health-ops.md` and `docs/runbooks/degraded-feed-investigation.md` from the new rollout runbook rather than duplicating their vocabulary or troubleshooting tables.
- Document override variables as deliberate testing or operator tools that must preserve paired Spot/USD-M overrides when the Spot URLs change.

## Implementation Notes

- Keep root README guidance short and point detailed operational steps to the runbook.
- Preserve the explicit warm-up caveat that current-state payloads may be unavailable before the sustained runtime publishes readable observations.
- Keep `/api/runtime-status` positioned as the primary bounded runtime-health route and `/healthz` as process health only.
- Keep `GET /api/market-state/global` and `GET /api/market-state/:symbol` described as consumer read paths, not operator health gates.

## Unit Test And Proof Expectations

- Updated docs contain no instructions that depend on choosing different runtime behavior for `local`, `dev`, or `prod`.
- All documented routes and environment variables match the current Go service behavior.
- The rollout runbook gives an operator enough detail to distinguish warm-up from degradation and to hand off to the existing health runbooks when the runtime is not healthy.

## Summary

- This module turns the settled runtime and config posture into a readable operator handoff without reopening API or runtime-shape decisions.
