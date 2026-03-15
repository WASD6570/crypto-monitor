---
name: program-refining
description: Turn initiative seeds into refined epic context that is ready for feature-planning
compatibility: opencode
---

## What I do

- Enforce the planning flow `initiatives -> epics -> plans`
- Read initiative handoff seeds and materialize or refresh refined epic context under `plans/epics/{epic_name}/`
- Compare the seed and epic context against initiative docs, active plans in `plans/`, and completed plans in `plans/completed/`
- Decide what is already done, what is still missing, and which child features should be handed to `feature-planning`
- Create or update the epic under `plans/epics/{epic_name}/` and write durable refined context back into that folder so later agents can continue without rereading the whole initiative
- Keep `plans/STATE.md` and the relevant initiative handoff in sync so later agents can see what is ready next without rereading the whole planning stack
- Stop before implementation
- When the user is still asking for planning work and has not asked to stop after refinement, it is acceptable to continue directly into `feature-planning` for the next recommended child feature without a separate approval step

## When to use me

Use this after `program-planning` when an initiative has identified one or more epic seeds that still need to become refined epic context.

Use this to move from initiative seed context into durable refined epic context under `plans/epics/`, so `feature-planning` can create active plans under `plans/`.

The intended sequence is:

- `program-planning` -> initiative docs under `initiatives/`
- `program-refining` -> epic docs under `plans/epics/`
- `feature-planning` -> active implementation-ready plans under `plans/`
- `feature-implementing` -> implementation of one active feature plan

Use this when you need to answer questions like:

- which epic folders should be created from the initiative handoff
- which parts of this seed/epic are already covered by completed work
- what child feature plans should exist under `plans/`
- what is still missing or blocked
- what initiative seeds should be refined next and what refined epics can later be planned in parallel

Do not use this for a single bounded feature that is already ready for `feature-planning`.

## Repo assumptions

- `initiatives/{initiative_name}/` is the parent layer that defines the slice queue, planning waves, and epic seeds
- `plans/epics/{epic_name}/` holds refined epic context that is ready to feed `feature-planning`
- `plans/{feature_name}/` holds active implementation-ready feature plans
- `plans/completed/{feature_name}/` holds completed read-only history
- `plans/STATE.md` is the repo-level source of truth for current planning and execution status
- `initiatives/` and `docs/specs/` remain the parent decomposition and product-constraint layers

## How I work

### 0) Load durable state first

1. Read `plans/STATE.md` before refinement so you know the currently active initiative, what is already archived, what is blocked, and what is recommended next.
2. If repo state and `plans/STATE.md` disagree, resolve the mismatch in the same pass instead of silently trusting one source.

### 1) Load only the relevant planning context

For the active initiative seed / epic:

1. Read only the relevant initiative docs under `initiatives/{initiative_name}/`, especially the active slice entry and epic seed in `03-handoff.md`.
2. If `plans/epics/{epic_name}/00-overview.md` already exists, read it as refined epic context.
3. If the epic folder does not exist yet, derive the initial refined epic context from the initiative handoff seed and create the epic scaffold first.
4. Read only the relevant program docs under `docs/specs/{program_name}/` when they materially constrain the epic.
5. Read direct prerequisite active plans under `plans/` only when they affect slice boundaries.
6. Read direct prerequisite completed plans under `plans/completed/` only when they affect what is already done or what constraints must carry forward.

Do not read unrelated epics, unrelated active plans, or unrelated completed plans.

### 2) Determine refinement status

Decide which of these is true:

- the initiative seed still needs to be materialized as refined epic context
- part of the seed/epic is already covered by completed work
- the refined epic has one obvious next child feature to plan
- the refined epic should present multiple child features or planning waves

Call out overlap explicitly so later agents do not duplicate work.

### 3) Produce bounded child feature seeds

For each proposed child feature, define:

- feature name
- concrete outcome
- primary repo area
- dependencies
- validation shape
- why it stands alone

Prefer the smallest independently reviewable slices.
Do not bundle unrelated work just because it shares one epic.
Do not emit smoke-only or integration-only child features; attach those validations to the owning implementation slice or note them as direct post-implementation checks.

### 4) Write refined epic artifacts inside the epic folder

If the epic does not exist yet, create `plans/epics/{epic_name}/00-overview.md` first from the initiative seed.

Then write or update these files under `plans/epics/{epic_name}/`:

- `00-overview.md`
  - epic summary derived from the initiative seed
  - in scope vs out of scope
  - target repo areas
  - validation shape and major constraints

- `90-refinement-map.md`
  - what parts are already done
  - what remains
  - dependency-safe refinement waves
- `91-child-plan-seeds.md`
  - one compact planning seed per proposed child feature
- `92-refinement-handoff.md`
  - next recommended child feature for `feature-planning`
  - which child seeds are safe to plan in parallel
  - blockers, assumptions, and prerequisite references

These files should be concise and should let `feature-planning` start without rereading the whole epic.

### 4.1) Sync durable state (required)

After updating the epic files, update the smallest durable state surfaces that changed:

- `plans/STATE.md`
  - set initiative-seed status where relevant and mark the epic itself as `ready_to_plan` or `blocked`
  - record what is `ready_to_plan`, what remains `blocked`, what is next, and what can be planned in parallel
  - keep `recently_archived` and `active_feature_plans` accurate when refinement changes queue interpretation
- the relevant initiative `03-handoff.md`
  - keep `Epic Queue`, `Planning Waves`, and the literal `## Execution State` section coherent with the refined epic state
  - make it obvious which epic is next and which epics are safe in parallel

Use `plans/STATE.md` as the authoritative quick-look status layer. Use initiative and epic docs for rationale and decomposition detail.

### 5) Respect plan-state boundaries

- Do not treat the epic itself as implementation-ready; it is refined enough for `feature-planning`, not `feature-implementing`.
- Do not write active feature plans to `plans/{feature_name}/` unless the user explicitly asks you to continue into `feature-planning`.
- Do not move finished work out of `plans/completed/`.
- Do not skip the epic layer; initiative slices must become explicit epics before child feature plans are written.

## Refinement quality bar

A good refinement pass lets another agent answer all of these quickly:

- what is already complete
- what is still missing
- which active feature plan should be created next
- which child plans can be planned in parallel
- which dependencies still block refinement or implementation

## Output locations

- initiative seeds consumed here: `initiatives/{initiative_name}/`
- refined epic context created or updated here: `plans/epics/{epic_name}/`
- active child plans created later by `feature-planning`: `plans/{feature_name}/`
- completed history: `plans/completed/{feature_name}/`

## Handoff to feature-planning

After this skill finishes:

1. Read `plans/epics/{epic_name}/92-refinement-handoff.md`.
2. Pick the next bounded child feature from that handoff.
3. Run `feature-planning` for that child feature.
4. Create the active implementation-ready plan under `plans/{feature_name}/`.
5. Stop after the active plan exists unless the user explicitly asks to continue into `feature-implementing`.

If the current planning workflow was invoked as a continuation request, it is acceptable to chain this skill directly into `feature-planning` for the recommended child feature. Do not auto-chain into `feature-implementing`.

## Stop conditions

Stop and ask questions only if:

- the epic has no meaningful decomposition seams
- ownership is too ambiguous to name child features safely
- the repo state contradicts the epic strongly enough that refinement would be misleading
- a blocking product or rollout decision cannot be inferred from initiative/program context
