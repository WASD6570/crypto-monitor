---
name: program-refining
description: Refine broad epic plans into dependency-aware child feature queues before feature-planning
compatibility: opencode
---

## What I do

- Enforce the planning flow `initiatives -> epics -> plans`
- Read one epic under `plans/epics/{epic_name}/` as broad source context, or materialize that epic from an initiative handoff seed when the epic folder does not exist yet
- Compare the epic against initiative docs, active plans in `plans/`, and completed plans in `plans/completed/`
- Decide what is already done, what is still missing, and what is still too broad
- Split the epic into bounded child features that are safe to hand to `feature-planning`
- Create or update the epic under `plans/epics/{epic_name}/` and write durable refinement artifacts back into that folder so later agents can continue without rereading the whole epic
- Stop before implementation and before writing active plan files unless the user explicitly asks to continue into `feature-planning`

## When to use me

Use this after `program-planning` when an initiative has identified one or more epic slices.

Use this to move from initiative context into durable epic context under `plans/epics/`, and then refine those epics into bounded child features for active plans under `plans/`.

The intended sequence is:

- `program-planning` -> initiative docs under `initiatives/`
- `program-refining` -> epic docs under `plans/epics/`
- `feature-planning` -> active implementation-ready plans under `plans/`
- `feature-implementing` -> implementation of one active feature plan

Use this when you need to answer questions like:

- which epic folders should be created from the initiative handoff
- which parts of this epic are already covered by completed work
- what child feature plans should exist under `plans/`
- what is still missing or blocked
- what should be refined next and what can be refined in parallel

Do not use this for a single bounded feature that is already ready for `feature-planning`.

## Repo assumptions

- `initiatives/{initiative_name}/` is the parent layer that defines the slice queue, planning waves, and epic seeds
- `plans/epics/{epic_name}/` holds broad unfinished plan context
- `plans/{feature_name}/` holds active implementation-ready feature plans
- `plans/completed/{feature_name}/` holds completed read-only history
- `initiatives/` and `docs/specs/` remain the parent decomposition and product-constraint layers

## How I work

### 1) Load only the relevant planning context

For the active epic:

1. Read only the relevant initiative docs under `initiatives/{initiative_name}/`, especially the active slice entry and epic seed in `03-handoff.md`.
2. If `plans/epics/{epic_name}/00-overview.md` already exists, read it.
3. If the epic folder does not exist yet, derive the initial epic context from the initiative handoff seed and create the epic scaffold first.
4. Read only the relevant program docs under `docs/specs/{program_name}/` when they materially constrain the epic.
5. Read direct prerequisite active plans under `plans/` only when they affect slice boundaries.
6. Read direct prerequisite completed plans under `plans/completed/` only when they affect what is already done or what constraints must carry forward.

Do not read unrelated epics, unrelated active plans, or unrelated completed plans.

### 2) Determine refinement status

Decide which of these is true:

- the epic is still broad and needs decomposition
- part of the epic is already covered by completed work
- the epic has one obvious next child feature to plan
- the epic should be split into multiple child features or refinement waves

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

### 4) Write refinement artifacts inside the epic folder

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

### 5) Respect plan-state boundaries

- Do not treat the epic itself as implementation-ready.
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
- broad epic context created or updated here: `plans/epics/{epic_name}/`
- active child plans created later by `feature-planning`: `plans/{feature_name}/`
- completed history: `plans/completed/{feature_name}/`

## Handoff to feature-planning

After this skill finishes:

1. Read `plans/epics/{epic_name}/92-refinement-handoff.md`.
2. Pick the next bounded child feature from that handoff.
3. Run `feature-planning` for that child feature.
4. Create the active implementation-ready plan under `plans/{feature_name}/`.
5. Hand off to `feature-implementing` only after the active plan exists.

## Stop conditions

Stop and ask questions only if:

- the epic has no meaningful decomposition seams
- ownership is too ambiguous to name child features safely
- the repo state contradicts the epic strongly enough that refinement would be misleading
- a blocking product or rollout decision cannot be inferred from initiative/program context
