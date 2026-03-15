---
name: feature-planning
description: Plan features to prevent context rot
compatibility: opencode
---

## What I do

- Gather requirements and confirm scope/constraints for one bounded child feature
- Ask clarifying questions only when the repo and request do not provide enough context for a safe plan
- Produce a structured plan with implementation order
- Write active implementation-ready plan artifacts to `plans/{feature_name}/` with overview + implementation steps + testing plan
- Treat `plans/epics/{epic_name}/` as refined source context that is ready to feed this skill
- Treat `plans/completed/{feature_name}/` as read-only archive for already-finished features
- Keep `plans/STATE.md` and the smallest relevant parent planning doc synchronized with the new plan state
- Plan with the expectation that once implementation and validation finish, `feature-testing` archives the active plan from `plans/{feature_name}/` to `plans/completed/{feature_name}/`
- Stop after the plan is written and hand back to the user before `feature-implementing` unless the user explicitly asked to continue into implementation

## When to use me

Use this for a single bounded child feature that needs a clear, durable active plan across agents.

Do not use this for smoke-only or integration-only validation work. If the work is just proving an already-implemented flow, run the validation directly or use `feature-testing` instead of writing `plans/{feature_name}/` artifacts.

If the work still exists only as an initiative seed and has not been materialized under `plans/epics/{epic_name}/`, stop and run `program-refining` first.

If the input is a large initiative brief that clearly contains multiple features or workstreams, use `program-planning` first and then return to this skill for each bounded slice.

This repository is a multi-language monorepo:

- `apps/web` is the React + Vite SPA
- `services/*` are Go live/realtime services
- `apps/research` and `libs/python` are offline Python research surfaces
- `schemas/` holds shared contract definitions

Plans should preserve the live/research boundary and avoid overdesign.

## How I work

1. **Read `plans/STATE.md` first** to confirm the feature is actually the next or a parallel-safe item and to reconcile any stale active/archive state before planning.
2. **Confirm feature name** and target area of the codebase.
3. **Load the relevant parent context first** when the feature comes from a larger initiative:
   - read the relevant initiative docs under `initiatives/{initiative_name}/`
   - read the relevant program docs under `docs/specs/{program_name}/` when product defaults, metrics, or sequencing materially affect the feature
   - read the relevant refined epic docs under `plans/epics/{epic_name}/` when this child feature comes from broader epic context
   - read direct prerequisite history under `plans/completed/{feature_name}/` only when prior finished slices materially constrain the new plan
4. **Infer as much as possible from the repo first**, then ask only blocking questions:
   - What is the exact problem to solve?
   - What is in scope vs out of scope?
   - What constraints must be respected (time, tech, legacy, compliance)?
5. **Check project-specific constraints** as needed:
   - Does the change belong in `apps/web`, `services/*`, `apps/research`, `libs/*`, `schemas/`, or `tests/`?
   - Does the change affect live/realtime behavior, offline research only, or both?
   - Will shared contracts, fixtures, replay payloads, or parity expectations change?
   - Is Python staying out of the live runtime path?
6. **Ask context-sensitive questions** only when still needed, based on the feature:
   - Dependencies or integrations?
   - Data model changes?
   - API changes or contract impacts?
   - Migration/backfill requirements?
   - Replay, determinism, or parity requirements?
   - Rollout, monitoring, or metrics?
   - Risks and edge cases?
7. **Draft the plan** and write it to `plans/{feature_name}/` using this structure:
    - `00-overview.md` (overview + design + requirements; always load before any step)
    - `01-implementation-<module>.md`, `02-implementation-<module>.md`, ...
    - `0n-testing.md` or `04-testing.md` (required)
8. **Ensure each implementation file** is a functional/testable module or part
    with its own requirements and unit-test plan.
9. **Prepare testing handoff inputs** in the testing file:
    - endpoint, CLI, or job sequence to execute
    - required env vars and credentials
    - expected side effects, artifacts, or state transitions to verify
    - replay, determinism, and parity checks when relevant
10. **Record archive intent explicitly**:
    - active plans live under `plans/{feature_name}/` only while work is still being implemented or validated
    - after a passing `feature-testing` run, move the full plan directory and `testing-report.md` to `plans/completed/{feature_name}/`
    - downstream handoff docs should reference the completed archive path, not the former active path
11. **Sync durable planning state**:
    - update `plans/STATE.md` to add or refresh the feature-plan entry as `ready_to_implement` (or `blocked` if planning uncovered a real blocker)
    - update the relevant initiative `03-handoff.md` and, when needed, the epic handoff/refinement docs so the next/parallel recommendations stay coherent
    - use `plans/STATE.md` as the quick-look truth for what is now planned, next, and parallel-safe
12. **Pause at the implementation boundary**:
    - present the completed plan and recommended next step
    - do not auto-run `feature-implementing` unless the user explicitly asked for implementation too

## Context Discipline

- Do not read unrelated feature plans under `plans/`, unrelated epics under `plans/epics/`, or unrelated completed plans under `plans/completed/`.
- Do not use this skill to turn initiative seeds into epic context; that belongs to `program-refining`.
- Read only the relevant initiative docs under `initiatives/` and relevant program docs under `docs/specs/` for the active feature.
- Only inspect files required for the current feature being planned.
- If examples are needed, prefer local conventions in source code/docs rather than opening other feature plan folders.
- Treat the OpenCode native task tracker as the session-only execution checklist; do not require repo task files for planning.
- Treat `plans/STATE.md` as the durable planning-status source of truth.
- Goal: avoid context bloat and keep planning focused on the requested feature.

## Plan file expectations

### `00-overview.md`

- Start with the ordered implementation plan (module list).
- Put requirements after the plan outline, then design details.
- Always include an ASCII flow of data, inputs, outputs, and any other helpful
  system or domain flow to make the feature understandable at a glance.
- Call out the live-path vs research-path boundary whenever the feature could blur it.

### Implementation files (`0n-implementation-*.md`)

- Each file represents a functional/testable module or project part.
- Begin with module-specific requirements and scope.
- Name the target repo areas explicitly, such as `apps/web`, `services/normalizer`, `libs/go`, or `schemas/json/events`.
- Include key decisions, data structures, and algorithm notes.
- Add code snippets only when needed for precision.
- Include unit test expectations for that module.
- Include contract, fixture, replay, or parity impacts when relevant.
- Close each file with **Summary** of what was implemented and critical details
  for the next agent to continue without context loss.

### Testing file (`0n-testing.md` or `04-testing.md`) (required)

- Define a minimal, high-signal validation matrix for the feature's key journeys.
- Include auth/authz checks, critical negative cases, and idempotency checks when relevant.
- Specify concrete endpoint, CLI, replay, or job sequence and required inputs.
- Specify verification checklist for side effects, artifacts, persisted state, and logs as needed.
- Add replay determinism and Go/Python parity checks when the feature depends on shared algorithms or fixtures.
- Include expected output artifact path: `plans/{feature_name}/testing-report.md` while the feature is active; once the feature is implemented and validated, move the entire directory to `plans/completed/{feature_name}/` so the report archives with the rest of the plan.

Never create a feature plan whose only purpose is smoke coverage, integration proof, or test orchestration. That work belongs in direct validation after the owning implementation lands.

## Output location

Write implementation-ready plan files to:

```
{repo_root}/plans/{feature_name}/
```

Use the ordering above and number sequentially for each module.

Remember: this is the active implementation path only. Finished feature plans belong under:

```
{repo_root}/plans/completed/{feature_name}/
```

If the work is still too broad for direct implementation, keep or move that refined epic context under:

```
{repo_root}/plans/epics/{epic_name}/
```

Do not hand an epic directly to `feature-implementing`.
Do not materialize initiative seeds here when `program-refining` should be used first.
This skill may be chained directly after `program-refining`, but it should still stop before `feature-implementing` unless the user explicitly asked to continue.
