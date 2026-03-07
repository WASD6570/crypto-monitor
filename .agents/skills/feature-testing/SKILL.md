---
name: feature-testing
description: Run smoke, replay, parity, and side-effect testing for implemented features
compatibility: opencode
---

## What I do

- Execute real-flow smoke tests for a feature that has already been implemented
- Test endpoint, job, replay, or UI behavior end-to-end for the critical path
- Verify side effects in the relevant system of record (database, files, fixtures, logs, or generated artifacts)
- Record reproducible results and clear pass/fail outcomes

## When to use me

Use this after implementation to validate that a feature works in real conditions, especially for multi-step service flows, replay-sensitive paths, cross-language algorithms, or critical UI journeys.

## Preconditions

1. Feature implementation is already present in the branch.
2. The relevant target is reachable or runnable (API, web app, CLI job, replay runner, batch job, or script).
3. Test inputs are available (fixture payloads, IDs, signatures, tokens, symbols, date ranges, config files, etc.).
4. If helper assets exist (for example `plans/{feature_name}/test-helpers/`), validate them against the specific flow before running tests.

If required credentials are missing, stop and report exactly what is needed.

## How I work

### 0) Discover and validate run-specific test inputs

Before execution, determine what this exact test run needs:

1. Infer required inputs from the feature flow itself by checking handlers, services, routes, jobs, fixtures, contracts, and docs.
2. If helper assets exist (for example `plans/{feature_name}/test-helpers/` or feature-local helper files), parse and validate them against that inferred checklist.
3. For env-like files, confirm required keys are present and non-empty; do not assume fixed key names across features.
4. If required tooling is missing (for example signature generators, request scripts, payload builders), search for existing scripts first; if none exist, create minimal local tooling for the test and document how it is used.
5. If any required input cannot be inferred or generated safely, stop and ask for an explicit missing-items checklist (keys/files/passphrases/scripts/IDs).
6. Record in the testing report which helper assets and generated tooling were used.

### 1) Load feature context first

1. Read `plans/{feature_name}/00-overview.md`.
2. Extract the concrete user journeys and endpoint sequence to test.
3. Use the current handoff context as the primary source for smoke setup when it is available.

Context relevance rules:

- Do not read unrelated `plans/{other_feature}` files.
- Use built-in task context for current-session execution details rather than repo task files.

### 2) Build a minimal smoke matrix

Create a short test matrix with:

- Happy-path flow (start to finish)
- Authentication/authorization checks
- Critical validation failures (bad signature, wrong role, missing required fields, invalid step transitions, malformed payloads)
- Idempotency/retry check for key flow entry points
- Replay/determinism checks when the feature touches normalization, alerts, outcomes, simulation, or backfills
- Go/Python parity checks when a shared algorithm or fixture is expected to match across runtimes

Prefer a small, high-signal matrix over broad coverage.

### 3) Execute endpoint tests

- Run the relevant sequence in real order: requests, jobs, replay runs, or UI actions.
- Record command or request intent, status or exit code, and key outputs.
- Reuse existing conventions/scripts when available.
- Stop immediately on blocker failures that invalidate the rest of the flow.

### 4) Verify side effects and artifacts (required)

Validate the actual system of record for the feature.

Validate at least:

- Created or updated records exist where expected
- Status or state transitions match observed results
- Output artifacts, fixtures, or logs are present and correct when applicable
- No duplicate side effects were created in idempotent paths
- Replay and parity results match expectations when relevant

### 5) Write test report artifacts

Write a concise report to:

`plans/{feature_name}/testing-report.md`

Include:

- Tested flows
- Command/request results (pass/fail)
- Side-effect verification results (queries, files, fixtures, or other evidence)
- Failures/blockers
- Recommended fixes or next checks

Use this template:

```md
# Testing Report: <feature_name>

## Environment
- Target:
- Date/time:
- Commit/branch:

## Smoke Matrix
| Case | Endpoint/Flow | Expected | Actual | Verdict |
|---|---|---|---|---|

## Execution Evidence
### <case-name>
- Command/Request: <method/path or command + key input>
- Expected:
- Actual:
- Verdict: PASS|FAIL

## Side-Effect Verification
### <case-name>
- Evidence: <SQL, file diff, fixture output, log path, screenshot, etc.>
- Expected state:
- Actual state:
- Verdict: PASS|FAIL

## Blockers / Risks
- ...

## Next Actions
1. ...
```

## Safety rules (mandatory)

- Prefer read verification first.
- If setup writes are needed, scope them to explicit test records, fixtures, or temp outputs and clean them up.
- Never run broad destructive operations.
- Never run schema-changing commands in this skill.
- If a verification step could affect production-like data broadly, stop and ask.

## Completion criteria

Testing is complete only when:

- Smoke matrix executed
- Execution results recorded
- Side effects verified
- `testing-report.md` written with clear pass/fail status
