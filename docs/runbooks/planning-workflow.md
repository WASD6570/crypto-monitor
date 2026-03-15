# Planning Workflow

This runbook defines how planning and implementation artifacts relate to each other in this repository.

## Artifact Types

### Durable State Index

Location:

- `plans/STATE.md`

`plans/STATE.md` is the authoritative quick-look source of truth for current project state.

Use it to answer:

- what is active right now
- what should happen next
- what is blocked
- what can be planned or implemented in parallel
- what was recently archived

Expected sections:

- `Current Snapshot`
- `Next Recommended`
- `Ready In Parallel`
- `Blocked`
- `Initiative State`
- `Active Feature Plans`
- `Recently Archived`

Every planning, implementing, and testing pass should update `plans/STATE.md` whenever durable project state changes. Session execution checklists belong only in the OpenCode native task tracker.

### Program Docs

Location:

- `docs/specs/{program_name}/`

Use program docs when work is large enough to require multiple initiatives or when product-level defaults and success metrics must be shared.

Expected files:

- `00-overview.md`
- `01-initiative-map.md`
- `02-product-success.md`
- `03-operating-defaults.md`
- `04-handoff.md`
- `05-open-questions.md` if needed

Program docs answer:

- who the user is
- what success means
- what defaults and constraints apply across the whole program
- which initiative should happen first

### Initiative Docs

Location:

- `initiatives/{initiative_name}/`

Use initiative docs when a body of work is too large for a single feature plan but still represents one coherent implementation track.

Expected files:

- `00-overview.md`
- `01-feature-map.md`
- `02-dependencies.md`
- `03-handoff.md`
- optional rollout or open-question docs

Initiative docs answer:

- which feature slices belong to the initiative
- what depends on what
- which slices are safe to plan in parallel
- which refined epic should be planned next, and which initiative seeds still need refinement first

They should also contain a literal `## Execution State` section that stays coherent with `plans/STATE.md` for the initiative's own epic/seed queue.

Do not create initiative-only tracks whose sole output is smoke or integration validation. Validation-only work should stay attached to the owning implementation slice or be executed directly after implementation.

### Epic Plans

Location:

- `plans/epics/{epic_name}/`

Use epic plans for work that has already been refined enough to feed `feature-planning`, but is still broader than one active implementation-ready feature plan.

Expected files:

- existing broad context such as `00-overview.md` and any inherited implementation/testing docs
- `90-refinement-map.md`
- `91-child-plan-seeds.md`
- `92-refinement-handoff.md`

Epic plans answer:

- what child feature plans should be created next
- which child slices are safe to plan in parallel
- what still blocks feature planning or later implementation
- what refined assumptions must carry into feature planning

Do not refine smoke-only or integration-only child features. If an epic needs combined validation after implementation, record it as a direct post-implementation check rather than a child plan.

### Feature Plans

Location:

- epic / too large: `plans/epics/{epic_name}/`
- active: `plans/{feature_name}/`
- completed archive: `plans/completed/{feature_name}/`

Use epic plans for refined epic context that is ready to feed `feature-planning`. Use active feature plans for implementation-ready work. Move active plans to `plans/completed/{feature_name}/` after implementation and validation are finished.

Expected files:

- `00-overview.md`
- `01-implementation-*.md`
- `02-implementation-*.md`
- `03-implementation-*.md`
- `04-testing.md`

Feature plans answer:

- what to implement
- in what order
- how to validate it
- what constraints matter for this specific feature

Epic plans answer:

- which child plans should be created next
- which dependencies or decisions block feature planning or later implementation
- what refined boundaries and assumptions must carry into child plans

## Recommended Workflow

### 1. Start With `program-planning` When The Brief Is Too Big

Use `program-planning` when the input is large enough that you must first decide:

- one initiative or many
- feature boundaries
- planning waves
- dependency order

`program-planning` should:

- create program docs under `docs/specs/` when needed
- create initiative docs under `initiatives/`
- create a standard `Planning Waves` section in each initiative `03-handoff.md`

### 2. Use `program-refining` From Initiative Seeds

Once an initiative exists and it has broad seeds or slices that are not yet materialized as refined epic context, use `program-refining` to create or update `plans/epics/{epic_name}/`.

The refiner should read:

- the relevant initiative `00-overview.md`
- the relevant entry and wave info in initiative `03-handoff.md`
- the relevant epic under `plans/epics/{epic_name}/` only when it already exists and needs refinement-state maintenance
- only the relevant program docs sections when they materially affect the epic

Then it should write or update refined epic artifacts under `plans/epics/{epic_name}/`.

It should also update:

- `plans/STATE.md`
- the relevant initiative `03-handoff.md` execution-state summary

### 3. Use `feature-planning` From Refined Epics

Once refined epic context exists under `plans/epics/{epic_name}/`, use `feature-planning` for each child feature.

The feature planner should read:

- the relevant initiative `00-overview.md`
- the relevant entry and wave info in initiative `03-handoff.md`
- the relevant epic refinement docs under `plans/epics/{epic_name}/`
- only the relevant program docs sections when they materially affect the feature

Then it should read any relevant epic under `plans/epics/{epic_name}/` and write the active feature plan under `plans/{feature_name}/`.

It should also update:

- `plans/STATE.md`
- the smallest relevant parent planning doc when the next/parallel recommendations change

If the remaining work is only smoke or integration validation, do not write a feature plan. Run the validation directly and record the result in the current handoff or testing report.

Active feature plans remain in `plans/{feature_name}/` only until implementation and validation are complete. After a passing `feature-testing` run, move the full plan directory, including `testing-report.md`, to `plans/completed/{feature_name}/` and update handoff references to the archive path.

That archive move must be reflected in `plans/STATE.md` in the same pass.

`program-refining` and `feature-planning` may be chained in one planning pass when the user is continuing through planning and has not asked to stop after refinement.

### 3.5 Stop Before `feature-implementing`

Once the active feature plan exists under `plans/{feature_name}/`, pause and keep the user in the loop before implementation starts.

Do not auto-chain into `feature-implementing` unless the user explicitly asked to continue into implementation.

After `feature-implementing`, run the required fresh-context reviewer pass before treating the feature as ready for `feature-testing` or closure.

If reviewer findings are clear from the request, plan, diff, and repo context, address them before handoff. Ask for clarification only when the review exposes genuinely missing or conflicting intent that cannot be resolved from the request, plan, diff, and repo context.

### 4. Use Parallel Refinement And Planning In Waves

When dependencies allow, use OpenCode subagents to refine multiple epics in parallel and then plan multiple bounded child features in parallel.

Safe pattern:

- Wave 1: prerequisite epics
- Wave 2: bounded child features from stabilized epics
- Wave 3: dependent follow-ons

Do not parallel-plan features that still depend on unresolved:

- contracts
- replay semantics
- market-state semantics
- shared query surfaces

## Standard `03-handoff.md` Format For Initiatives

Every initiative handoff should contain:

1. `Refined Epic Queue`
2. `Execution State`
3. `Planning Waves`
4. `Initiative Seeds` or `Refined Epics`
5. `Open Questions That Still Matter` if needed

`Refined Epic Queue` should list only actionable refined epics that are currently `ready_to_plan`.
Blocked refined epics, initiative seeds that still need `program-refining`, and historical references belong in `Execution State`, not in the actionable queue.

Use `Initiative Seeds` when the handoff is still describing pre-refinement seed detail.
Use `Refined Epics` only when real `plans/epics/...` artifacts already exist and the section is describing those materialized refined epics.

The `Planning Waves` section should make it obvious:

- what to plan first
- what can be planned in parallel
- what must wait and why

The `Execution State` section should make it obvious:

- which initiative seeds are `ready_to_refine`
- which refined epics are `ready_to_plan`, `blocked`, or `archived`
- what depends on what
- what is next
- what is parallel-safe

Example shape:

```md
## Execution State

- Initiative status: `in_progress`
- Next recommended seed or epic: `foo-runtime-health`
- Parallel-safe now: `foo-semantic-follow-on`

| Item | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `foo-runtime-health` | `ready_to_refine` | Wave 1 complete | `foo-semantic-follow-on` | Run `program-refining` and materialize `plans/epics/foo-runtime-health/` | Default next seed |
| `foo-semantic-follow-on` | `ready_to_refine` | Wave 1 complete | `foo-runtime-health` | Refine in parallel when capacity allows | Parallel-safe seed |
```

If a prior epic is complete but its refined epic files are still useful as historical context, keep it out of the actionable refined-epic queue and call it out separately as historical reference rather than listing it as an active/refined epic state row.

## What Implementers Should Read

Yes: implementers should read the relevant spec and initiative context, but only the relevant parts.

Before implementation, load in this order:

1. `plans/{feature_name}/00-overview.md`
2. the current step file in `plans/{feature_name}/`
3. if relevant, the parent epic refinement handoff in `plans/epics/{epic_name}/92-refinement-handoff.md`
4. the parent initiative `initiatives/{initiative_name}/00-overview.md`
5. the relevant slice and wave from `initiatives/{initiative_name}/03-handoff.md`
6. the relevant program docs sections from `docs/specs/{program_name}/`, usually:
    - `02-product-success.md`
    - `03-operating-defaults.md`
    - `04-handoff.md` when sequencing matters

Do not load the whole program or all initiative docs unless the feature actually depends on them.
Do not implement directly from `plans/epics/{epic_name}/`; convert refined epic context into active plans first.
Read archived prerequisite plans under `plans/completed/` only when they directly constrain the active feature.

## OpenCode-Specific Guidance

- Use the OpenCode native task tracker for execution state.
- Use `plans/STATE.md` as the durable planning/execution status source of truth.
- Use `Read`, `Grep`, and `Glob` to load only the necessary context.
- Use subagents for parallel planning waves when dependencies are stable.
- Keep program docs, initiative docs, and feature plans separate so later agents do not guess which layer they are reading.
