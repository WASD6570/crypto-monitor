# Implementation: Compose Smoke Proof

## Module Requirements And Scope

- Add one repeatable, repo-owned compose smoke proof for the rollout handoff.
- Keep the proof narrow: verify startup contract, route reachability, warm-up interpretation, and operator entry points.
- Avoid turning this slice into a long-run soak or failure-injection suite; deeper runtime hardening remains in `binance-long-run-runtime-hardening`.

## Target Repo Areas

- `Makefile`
- `scripts/dev/*` if a small helper script is needed
- `docs/runbooks/` for invocation and expected output guidance

## Key Decisions

- Prefer a single copy-paste-safe command such as `make compose-smoke` or an equivalently small helper invoked from the runbook.
- The smoke proof should run `docker compose config`, launch the stack, verify `/api/runtime-status` and `/api/market-state/global` through `http://127.0.0.1:4173/api/...`, verify `http://127.0.0.1:8080/healthz` from inside the `market-state-api` container, and tear the stack down even on failure.
- Accept `readiness=NOT_READY` during initial warm-up; success is route availability and correct contract posture, not immediate fully-ready market state.

## Implementation Notes

- Keep the helper dependency-light; prefer POSIX shell and Docker CLI over new language-specific tooling.
- Make failure messages operator-readable so the runbook can point directly at the helper output.
- Keep cleanup explicit so repeated runs do not leave stale containers or networks behind.

## Unit Test And Proof Expectations

- The helper or command sequence is runnable from the repo root without additional repo context.
- A passing run proves compose rendering, stack startup, same-origin API reachability, and process-health separation.
- A failing run leaves clear evidence for whether the problem is compose rendering, container startup, `/healthz`, same-origin proxying, or runtime-status/current-state route availability.

## Summary

- This module gives the rollout handoff one concrete verification path that future operators and feature-testing can rerun without reconstructing the steps from prose.
