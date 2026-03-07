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
- which feature should be planned next

### Feature Plans

Location:

- `plans/{feature_name}/`

Use feature plans for implementation-ready work.

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

### 2. Use `feature-planning` Per Feature

Once an initiative exists, use `feature-planning` for each slice under that initiative.

The feature planner should read:

- the relevant initiative `00-overview.md`
- the relevant entry and wave info in initiative `03-handoff.md`
- only the relevant program docs sections when they materially affect the feature

Then it should write the feature plan under `plans/{feature_name}/`.

### 3. Use Parallel Planning In Waves

When dependencies allow, use OpenCode subagents to plan multiple features in parallel.

Safe pattern:

- Wave 1: prerequisites
- Wave 2: independent consumers of those prerequisites
- Wave 3: dependent follow-ons

Do not parallel-plan features that still depend on unresolved:

- contracts
- replay semantics
- market-state semantics
- shared query surfaces

## Standard `03-handoff.md` Format For Initiatives

Every initiative handoff should contain:

1. `Feature Queue`
2. `Planning Waves`
3. `Child Plan Seeds`
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
3. the parent initiative `initiatives/{initiative_name}/00-overview.md`
4. the relevant slice and wave from `initiatives/{initiative_name}/03-handoff.md`
5. the relevant program docs sections from `docs/specs/{program_name}/`, usually:
   - `02-product-success.md`
   - `03-operating-defaults.md`
   - `04-handoff.md` when sequencing matters

Do not load the whole program or all initiative docs unless the feature actually depends on them.

## OpenCode-Specific Guidance

- Use the built-in task tracker for execution state.
- Use `Read`, `Grep`, and `Glob` to load only the necessary context.
- Use subagents for parallel planning waves when dependencies are stable.
- Keep program docs, initiative docs, and feature plans separate so later agents do not guess which layer they are reading.
