---
name: program-planning
description: Decompose a large brief into program docs, one or more bounded initiatives, and dependency-aware epic-refinement waves
compatibility: opencode
---

## What I do

- Ingest a long, dense, high-level brief that may contain multiple initiatives, features, workstreams, or rollout tracks
- Separate the brief into a program-level view plus one or more bounded initiatives when needed
- Decompose each initiative into feature slices, platform tracks, and cross-cutting concerns
- Identify dependencies, contract touchpoints, replay/backfill risks, rollout sequencing, and safe parallelization opportunities
- Write durable program docs under `docs/specs/` and initiative plans under `initiatives/`, then hand off refined-epic materialization to `program-refining`
- Keep `plans/STATE.md` synchronized so later agents can see which initiative is active, what should be refined next, and what is safe in parallel
- When appropriate, hand off to `program-refining` in dependency-aware parallel waves using OpenCode's native subagent workflow

## When to use me

Use this before `program-refining` and `feature-planning` when the input is too large or mixed for a single feature plan, for example:

- a long product or architecture brief covering multiple capabilities
- a roadmap or initiative description that spans several services or apps
- a dense planning memo that mixes contracts, rollout concerns, research, frontend, and live service work
- any brief where one feature plan would become bloated, ambiguous, or hard to hand off

Do not use this for a single bounded feature. Use `feature-planning` directly in that case.

Do not use this when initiative and epic boundaries already exist and you only need to turn one initiative seed into refined epic context and then plan child features. Use `program-refining` then `feature-planning`.

Use this when you need to answer either of these:

- "Is this one initiative or several initiatives?"
- "Which initiative seeds should be refined first, and which refined epics can be planned in parallel without drifting on contracts or semantics?"

## Repo assumptions

- `apps/web` is the React + Vite SPA
- `services/*` are Go live and realtime services
- `apps/research` and `libs/python` are offline Python research surfaces
- `schemas/` is the home for shared contracts
- Python must not become a live runtime dependency
- `plans/STATE.md` is the repo-level source of truth for durable planning and execution state

## How I work

### 0) Read durable state first

Before decomposing a new brief, read `plans/STATE.md` to understand the current initiative stack, what is already active or archived, and where a new program or initiative should fit.

### 1) Read the brief and identify planning boundaries

First, infer structure from the brief itself before asking questions.

Classify the content into:

- user-facing features
- service or platform capabilities
- shared contract changes
- data/replay/backfill concerns
- research-only work
- rollout and operational needs

Look for natural seams such as:

- separate user journeys
- distinct service boundaries
- contract/version changes
- independent rollout units
- different validation strategies

### 2) Decide whether the brief is one initiative or many

Before decomposing to feature slices, decide whether the brief should become:

- one initiative under `initiatives/{initiative_name}/`, or
- multiple initiatives plus a parent program doc

Create multiple initiatives when any of these are true:

- there are clearly different user loops or product phases
- one group of work must land before another can even be trusted
- there are different success metrics, consumers, or operating modes
- one initiative is primarily visibility/platform work and another is alerting/execution/review work
- the brief would otherwise create an initiative plan that is too broad to hand off cleanly

If multiple initiatives are created, give each one an explicit, boring name and a clear dependency relationship.

### 3) Produce bounded slices

Break the initiative into small named slices that can each become a separate refined epic for later feature planning.

Each slice should have:

- a clear outcome
- an obvious primary repo area
- limited cross-surface coupling
- a validation path
- a reason it should be planned independently

Do not over-split into trivial fragments. Do not keep unrelated work bundled together.
Do not create initiative or epic slices whose only deliverable is smoke or integration validation; keep that work as direct validation attached to the owning implementation slice.

### 4) Identify cross-cutting tracks explicitly

Some work should not be hidden inside a single feature plan. Call out shared tracks such as:

- shared contracts in `schemas/json/...`
- fixtures and parity work in `tests/fixtures` and `tests/parity`
- replay or backfill sequencing
- shared libraries in `libs/go`, `libs/ts`, or `libs/python`
- rollout coordination or migration constraints

Mark whether each cross-cutting item is:

- prerequisite
- parallelizable
- rollout-sensitive
- research-only

### 5) Build an execution roadmap

Produce an order that later agents can follow.

For each slice, state:

- depends on
- unlocks
- risk level
- recommended next skill (`program-refining`, then `feature-planning`, then later `feature-implementing` and `feature-testing`)

Default to the smallest safe ordering. Prefer reducing dependency chains.

### 6) Write program and initiative artifacts

If the brief becomes a multi-initiative program, create parent program docs under:

`docs/specs/{program_name}/`

Write these files:

- `00-overview.md`
  - first user
  - core product loop
  - program structure
  - scope vs out of scope
  - system map
- `01-initiative-map.md`
  - each initiative, why it exists, and what depends on it
- `02-product-success.md`
  - success metrics, baselines, and gates
- `03-operating-defaults.md`
  - safe defaults, time rules, storage defaults, delivery defaults, and governance defaults
- `04-handoff.md`
  - initiative order and next recommended planning wave
- `05-open-questions.md` if needed

For each actual implementation initiative, write artifacts under `initiatives/`.

Create a top-level initiative folder:

`initiatives/{initiative_name}/`

Write these files:

- `00-overview.md`
  - initiative summary
  - scope vs out of scope
  - high-level system map
  - key constraints and assumptions
- `01-feature-map.md`
  - the decomposed feature slices
  - each slice name, goal, primary repo area, and why it stands alone
- `02-dependencies.md`
  - dependency graph
  - contract/replay/backfill/parity risks
  - suggested implementation order
- `03-handoff.md`
  - explicit queue of which refined epic should go through `feature-planning` next when refined epic context already exists
  - an `Execution State` section that marks which initiative seeds are next, which refined epics are parallel-safe, and which items are blocked or archived
  - exact epic folder names to create or refresh under `plans/epics/{epic_name}/`
  - which slices are safe to plan in parallel and which are not
  - a standard `Planning Waves` section with `Wave 1`, `Wave 2`, `Wave 3`, ... and the rationale for each wave
  - any open questions that still block epic-level refinement

If useful, also create:

- `04-rollout-notes.md`
- `05-open-questions.md`

### 7) Prepare epic refinement inputs

For each slice, create a compact refinement seed inside `03-handoff.md` that includes:

- epic name
- problem statement
- in scope
- out of scope
- target repo areas
- contract/fixture/parity/replay implications
- likely validation shape

These seeds should be good enough for a later agent to run `program-refining` and materialize refined epic context without rereading the entire original brief.

### 7.1) Standard handoff format for initiative docs

Every `initiatives/{initiative_name}/03-handoff.md` should contain these sections in this order:

1. `Refined Epic Queue`
2. `Execution State`
3. `Planning Waves`
4. `Initiative Seeds` or `Refined Epics`
5. `Open Questions That Still Matter` if needed

`Refined Epic Queue` should list only actionable refined epics that are currently `ready_to_plan`.
Blocked refined epics, initiative seeds that still need `program-refining`, and historical references belong in `Execution State`, not in the actionable queue.

Use `Initiative Seeds` when the handoff is still describing pre-refinement seed detail.
Use `Refined Epics` only when real `plans/epics/...` artifacts already exist and the section is describing those materialized refined epics.

The `Planning Waves` section should:

- group slices into dependency-safe planning waves
- state which slices can be planned in parallel
- explain why a later wave is blocked on an earlier one
- identify the first recommended refinement/planning wave explicitly

The `Execution State` section should:

- mark which initiative seeds are `ready_to_refine`
- mark which refined epics are `ready_to_plan`, `blocked`, or `archived`
- name the next recommended seed or epic
- identify what is safe to plan in parallel
- stay coherent with `plans/STATE.md`

Example shape:

```md
## Execution State

- Historical reference only: `plans/epics/canonical-contracts-and-fixtures/`
- Next recommended seed or epic: `raw-storage-and-replay-foundation`
- Parallel-safe now: `market-ingestion-and-feed-health`

| Item | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `raw-storage-and-replay-foundation` | `ready_to_refine` | contracts complete | `market-ingestion-and-feed-health` | Run `program-refining` and materialize `plans/epics/raw-storage-and-replay-foundation/` | Default next seed |
| `market-ingestion-and-feed-health` | `ready_to_refine` | contracts complete | `raw-storage-and-replay-foundation` | Refine in parallel when capacity allows | Parallel-safe seed |

## Planning Waves

### Wave 1
- `canonical-contracts-and-fixtures`
- Why now: all later slices depend on shared contract decisions.

### Wave 2
- `market-ingestion-and-feed-health`
- `raw-storage-and-replay-foundation`
- Why parallel: both consume wave-1 outputs and do not redefine the same contract vocabulary.

### Wave 3
- `world-usa-composites-and-market-state`
- Why later: depends on ingestion and replay semantics being stable.
```

### 7.2) Sync durable state (required)

After writing or updating program and initiative artifacts, update `plans/STATE.md` in the same pass.

At minimum:

- mark the initiative as `seeded`, `in_progress`, `blocked`, or `archived`
- record the next recommended initiative-seed or epic transition
- record which refined epics are safe to plan in parallel
- note blockers or open questions that materially change sequencing

Use `plans/STATE.md` for quick-look status. Use program and initiative docs for decomposition detail and rationale.

### 8) Use OpenCode-native orchestration when continuing planning

This skill is designed for OpenCode and should lean on its native tools and workflow.

- Use the OpenCode native task tracker for current-session execution state.
- Use `plans/STATE.md` as the durable current-state index and update it whenever initiative or epic-seed state changes.
- Use `Read`, `Grep`, and `Glob` to gather only the necessary repo context.
- Use OpenCode subagents when planning can safely fan out.

If the user wants decomposition only:

- stop after writing the program and initiative artifacts plus `03-handoff.md`

If the user wants decomposition plus epic refinement:

- launch multiple `program-refining` subagents in parallel only for initiative seeds whose dependencies are already stable
- do not parallelize slices that still depend on unresolved contract, replay, or state semantics
- prefer wave-based planning:
  - wave 1: prerequisite epics
  - wave 2: independent epics that consume those prerequisites
  - wave 3: dependent follow-ons
- mirror the exact wave names and ordering from `initiatives/{initiative_name}/03-handoff.md`

Each subagent should:

- read only the relevant initiative docs under `initiatives/`, relevant epics under `plans/epics/`, active prerequisite feature plans under `plans/`, and direct historical prerequisite plans under `plans/completed/` when needed
- write only to its own `plans/epics/{epic_name}/`
- return a concise summary of files created and key assumptions

## Decomposition rules

- Preserve live vs research boundaries.
- Do not invent concrete schemas, migrations, or business logic from a high-level brief.
- Separate rollout-sensitive infrastructure from user-facing slices when that reduces risk.
- Keep names explicit and boring.
- Do not keep umbrella program docs inside `plans/`; reserve `initiatives/` for real implementation initiatives and pre-refinement seed queues, `plans/epics/` for refined epic context that is ready for `feature-planning`, `plans/` for active implementation-ready feature plans, and `plans/completed/` for archived finished features.
- Prefer one plan per independently reviewable capability.
- If a slice cannot be validated on its own, it is probably not a good slice yet.

## Output quality bar

A good decomposition should let another agent answer all of these quickly:

- Is this one initiative or several?
- What are the distinct features or workstreams?
- Which ones require shared contract work?
- What must happen first?
- Which slices are safe to plan in parallel?
- Which refined epic should go into `feature-planning` next, and which initiative seed still needs `program-refining` first?

## Handoff to program-refining

After this skill finishes:

1. If needed, read `docs/specs/{program_name}/04-handoff.md` to identify initiative order.
2. Pick the next highest-priority initiative seed from `Execution State` when no refined epic is ready yet, or the next refined epic from `Refined Epic Queue` when it is ready.
3. Run `program-refining` for initiative seeds and `feature-planning` for refined epics.
4. When dependencies allow, run multiple `program-refining` agents in parallel for the next safe wave.
5. Write or update refined epic context under `plans/epics/{epic_name}/`.
6. Repeat until the initiative backlog is fully decomposed into refined epic artifacts that can feed `feature-planning`.

## Stop conditions

Stop and ask questions only if one of these blocks safe decomposition:

- the brief does not describe any meaningful boundaries
- the initiative mixes unrelated efforts with no stated goal
- naming or ownership is so ambiguous that slice boundaries would be misleading
- rollout or contract risk cannot even be framed from the available context
