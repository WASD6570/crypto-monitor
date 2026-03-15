# Implementation: Compose Browser Smoke And Docs

## Requirements And Scope

- Prove the local compose stack loads the dashboard through the same-origin Go API path after the live-reader cutover.
- Update browser smoke so it no longer depends on deterministic live labels.
- Refresh the small set of docs that describe local startup, warm-up, and degraded behavior for the live current-state path.
- Keep scope limited to local compose/browser verification and docs; broader observability/runbook redesign stays out of scope.

## Target Repo Areas

- `apps/web/tests/e2e/*.ts`
- `docker-compose.yml` only if startup wiring needs a bounded adjustment
- `README.md`
- `services/market-state-api/README.md`

## Implementation Notes

- Update `apps/web/tests/e2e/dashboard-compose-api-smoke.spec.ts` so it proves same-origin data loading and symbol switching without asserting exact live regime text.
- Prefer assertions that the dashboard renders both symbol cards, focused navigation works, the shell sections appear when readable data exists, and honest fallback/degraded UI remains possible during warm-up.
- If compose startup timing is flaky, fix it with the smallest bounded readiness or test-wait adjustment instead of hard-coding deterministic market output.
- Refresh docs to describe that local startup may briefly show `Current State Unavailable` before the runtime publishes observations, and that `/healthz` does not encode market-data freshness.

## Unit Test Expectations

- Browser smoke passes against the compose stack in desktop and mobile projects.
- Docs accurately describe the same-origin live path, supported symbols, warm-up behavior, and retry expectations.

## Contract / Rollout Notes

- The dashboard remains a pure consumer of `/api/market-state/*`.
- No non-local environment defaults should be introduced here; those belong to the later environment-hardening epic.

## Summary

This module closes the local user-visible loop for the cutover: the compose stack and dashboard smoke prove the API path works without falling back to deterministic frontend assumptions, and the docs explain the expected warm-up experience.
