# Planning Workflow

This runbook defines how planning and implementation artifacts relate to each other in this repository.

## Artifact Types

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
- which epic should be refined next

### Epic Plans

Location:

- `plans/epics/{epic_name}/`

Use epic plans for broad work that is still too large for direct implementation.

Expected files:

- existing broad context such as `00-overview.md` and any inherited implementation/testing docs
- `90-refinement-map.md`
- `91-child-plan-seeds.md`
- `92-refinement-handoff.md`

Epic plans answer:

- what is already done
- what is still missing
- which child feature plans should be created next
- which child slices are safe to refine or plan in parallel
- what still blocks direct implementation

### Feature Plans

Location:

- epic / too large: `plans/epics/{epic_name}/`
- active: `plans/{feature_name}/`
- completed archive: `plans/completed/{feature_name}/`

Use epic plans for broad work that still needs refinement. Use active feature plans for implementation-ready work. Move active plans to `plans/completed/{feature_name}/` after implementation and validation are finished.

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

- what still needs refinement
- which child plans should be created next
- which dependencies or decisions block direct implementation

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

### 2. Use `program-refining` Per Epic

Once an initiative exists and broad slices live under `plans/epics/`, use `program-refining` for each epic that still needs decomposition.

The refiner should read:

- the relevant initiative `00-overview.md`
- the relevant entry and wave info in initiative `03-handoff.md`
- the relevant epic under `plans/epics/{epic_name}/`
- only the relevant program docs sections when they materially affect the epic

Then it should write or update refinement artifacts under `plans/epics/{epic_name}/`.

### 3. Use `feature-planning` Per Child Feature

Once an epic has bounded child slices, use `feature-planning` for each child feature.

The feature planner should read:

- the relevant initiative `00-overview.md`
- the relevant entry and wave info in initiative `03-handoff.md`
- the relevant epic refinement docs under `plans/epics/{epic_name}/`
- only the relevant program docs sections when they materially affect the feature

Then it should read any relevant epic under `plans/epics/{epic_name}/` and write the active feature plan under `plans/{feature_name}/`.

Active feature plans remain in `plans/{feature_name}/` only until implementation and validation are complete. After that, move the full plan directory, including `testing-report.md`, to `plans/completed/{feature_name}/` and update handoff references to the archive path.

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

1. `Epic Queue`
2. `Planning Waves`
3. `Epic Seeds`
4. `Open Questions That Still Matter` if needed

The `Planning Waves` section should make it obvious:

- what to plan first
- what can be planned in parallel
- what must wait and why

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
Do not implement directly from `plans/epics/{epic_name}/`; refine epics into active plans first.
Read archived prerequisite plans under `plans/completed/` only when they directly constrain the active feature.

## OpenCode-Specific Guidance

- Use the built-in task tracker for execution state.
- Use `Read`, `Grep`, and `Glob` to load only the necessary context.
- Use subagents for parallel planning waves when dependencies are stable.
- Keep program docs, initiative docs, and feature plans separate so later agents do not guess which layer they are reading.
