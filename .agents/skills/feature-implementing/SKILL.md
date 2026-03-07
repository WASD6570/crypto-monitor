---
name: feature-implementing
description: Execute a feature plan produced by feature-planning (execution-first with feature-driving task loop)
compatibility: opencode
---

## What I do

- Execute a feature plan produced by `feature-planning`
- Work strictly in plan order, resuming from last recorded progress
- Convert each plan step into an explicit built-in task checklist (feature-driving)
- Implement tasks immediately with a tight edit → test → log loop
- Persist context between agents through accurate task state and clear handoff summaries

Repo-specific operating assumptions:

- `apps/web` is the React + Vite SPA
- `services/*` are Go live/realtime services
- `apps/research` and `libs/python` are offline Python surfaces
- Python must not become a live runtime dependency
- `schemas/` and `tests/fixtures` are shared coordination points when contracts change

## When to use me

Use this after `feature-planning` has produced an active plan in `plans/{feature_name}/`.

## How I work

### 0) Locate plan + load overview (required)
1. Locate the active plan for the feature name in `plans/{feature_name}/`.
2. If the feature only exists in `plans/epics/{feature_name}/`, stop and refine it with `program-refining`, then `feature-planning`, before implementing.
3. If the feature only exists in `plans/completed/{feature_name}/`, treat it as already completed historical context and do not restart implementation unless the user explicitly asks.
4. Always load `00-overview.md` before any step to understand flow, design, and requirements.
5. If the active plan came from an epic, load only the relevant refinement docs from `plans/epics/{epic_name}/`.
6. Load only the relevant initiative docs under `initiatives/{initiative_name}/` when the feature belongs to a larger initiative:
    - at minimum, read the matching initiative `00-overview.md`
    - read the relevant slice entry from the initiative `03-handoff.md`
7. Load only the relevant program docs under `docs/specs/{program_name}/` when they materially affect the feature:
   - success metrics
   - operating defaults
   - handoff or ordering constraints

### 0.1) Context loading discipline (required)
- Read only the relevant parts of the program and initiative context for the active feature.
- Do not reread the entire program or all initiative files if the feature only depends on one section.
- Treat program docs as product and operating constraints, initiative docs as decomposition and dependency context, and feature plans as execution instructions.

### 1) Determine starting point (required)
- Resume from the last completed step you can verify from repository state and current handoff context.
- Otherwise start at the first step file (`01-*.md`).

### 1.1) Task relevance check (required)
- Use the built-in task tracker as the execution queue for the active session.
- `plans/epics/{feature_name}/` is broad context only and must be refined before execution.
- `plans/{feature_name}/` files are durable active plan artifacts, not required execution checklists.
- `plans/completed/{feature_name}/` is read-only history for prerequisite context and prior testing evidence.

### 2) Feature-driving: create / maintain task queue (required)
For the current step file:
- Create or update the built-in task checklist as concrete, executable tasks derived from the step.
- The task tracker is the execution source of truth for the step.

Each task must include:
- Outcome: what changes for the system/user
- Files: paths or globs to modify
- Command: exact validation command
- Accepts: observable done criteria

Task format:

- [ ] <task title>
  - Outcome: ...
  - Files: ...
  - Command: ...
  - Accepts: ...

### 3) Execution bias (anti-stall rules)
- Start implementing immediately after reading the step file and task checklist.
- Planning cap: max 5 bullets or 10 lines of reasoning before editing files.
- Questions policy: ask questions only if blocked from making a safe minimal change.
- Default assumptions: if ambiguity is low-risk, choose the smallest backward-compatible assumption and proceed.
- Record assumptions in the task state or handoff.

### 4) Implement in strict step order (required)
- Execute steps in strict order; never skip or reorder.
- Keep changes focused to the current step.
- Never complete tasks for a future step while on the current step.
- Defer unrelated fixes.
- Preserve the live/research boundary. Do not make `services/*` depend on Python code or notebooks.

### 5) Task execution loop (required)
Repeat until all tasks for the step are complete:
1. Pick the next unchecked task in the built-in task tracker
2. Implement only what that task requires (small, reviewable edits)
3. Run the task’s Command
4. If it fails:
   - fix within scope
   - re-run the command
5. Mark the task complete only when acceptance is met

### 6) Validation
- Validate according to `04-testing.md` when you reach it.
- Prefer targeted tests over full suites unless explicitly required.
- Use stack-appropriate commands for the touched area:
  - Go: `go test`, service-specific smoke checks, replay fixtures
  - TypeScript: `pnpm`, `npm`, Vite, lint, typecheck, UI smoke checks
  - Python: `pytest`, job/script smoke checks, offline fixture validation
- If contracts or fixtures change, validate the affected consumers or document the exact remaining gap.

### 6.1) End-of-step hygiene (required)
- Run formatting, linting, and type checks only for the touched area when applicable.
- Update docs, fixtures, or shared contracts if the current step depends on them.
- Keep validation focused and reproducible.

### 7) End-of-step logging (required)
At the end of each step, capture in the handoff summary:
- Completed step(s)
- Current state of work
- Key files changed
- Next step to run
- Next unchecked task in the built-in task tracker
- Any blockers or decisions needed
- Any assumptions made
- Test handoff block (when implementation step is complete):
  - endpoint, CLI, replay, or job sequence for smoke testing
  - required env vars/credentials
  - expected side effects, artifacts, persisted state, and parity/replay checks to verify

## Required conventions

- Start from the beginning unless repository state or current handoff context says otherwise.
- Always use a clear handoff summary when pausing.
- Keep the built-in task checklist accurate; it is the execution queue.
- Keep changes aligned to the monorepo boundaries; shared code goes in `libs/*`, not arbitrary duplication.
- Preserve deterministic fixtures when touching replay, simulation, or parity-sensitive logic.

## Recommended step loop

1. Read `00-overview.md`
2. Read current step file
3. Create / update the built-in task checklist
4. Implement next unchecked task
5. Run validation command
6. Update the current task state and handoff notes
7. Continue or hand off

## Extra guidance for productive execution

- Prefer small, reviewable edits per task.
- Code > prose.
- If blocked, log the blocker precisely and stop.
- Done means: implemented, validated, logged.
