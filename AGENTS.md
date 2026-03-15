# AGENTS

## External File Loading

CRITICAL: When you encounter a file reference (for example `@rules/general.md`), use the Read tool to load it on a need-to-know basis. Those references are relevant to the specific task at hand.

Instructions:

- Do NOT preemptively load all references; use lazy loading based on actual need.
- When loaded, treat referenced content as mandatory instructions that override defaults.
- Follow references recursively when needed.

## Repository Intent

This repository is the `Crypto Market Copilot` monorepo.

Stack ownership is fixed:

- `apps/web` is the TypeScript + React + Vite SPA.
- `services/*` are Go services for live, realtime, and production-path responsibilities.
- `apps/research` and `libs/python` are for Python research, ML, offline analysis, and experiments.
- Python must never be required for the platform to run in live mode.

Repository structure principles:

- Keep the monorepo legible and boring.
- Prefer explicit service names and boundaries over clever abstractions.
- Do not create deep folder trees for hypothetical future complexity.
- Keep live operation and research clearly separated.
- Put shared contracts in one obvious home under `schemas/`.

## Repository Boundaries

- Do not implement business logic, concrete schemas, migrations, or data models from high-level setup briefs alone.
- Use `services/` for focused Go service boundaries, `libs/` for shared language-specific code, and `schemas/` for canonical shared contracts.
- Keep venue-specific behavior inside the relevant adapter or service boundary unless a shared contract or library is clearly justified.
- Keep research artifacts, notebooks, and experimental code out of the live runtime path.
- Favor minimal scaffolding over speculative infrastructure.

## MCP Access

- Agents have access to MCP servers/tools available in the current OpenCode session.
- Before MCP-dependent work, verify connectivity with `opencode mcp list`.
- If MCP is unavailable, continue with non-blocked work and report the blocker in the current task state and handoff.

## Contracts, Data, And Migration Policy

- Shared payload and contract definitions belong under `schemas/json/`.
- Future database-oriented definitions belong under `schemas/sql/`.
- The canonical homes reserved today are:
  - `schemas/json/events`
  - `schemas/json/features`
  - `schemas/json/alerts`
  - `schemas/json/outcomes`
  - `schemas/json/replay`
  - `schemas/json/simulation`
- Do not invent or fill in concrete schemas just because a high-level brief mentions them.
- If a task changes a shared contract, update the affected fixtures, docs, and consumer validation for the touched languages.
- If a change requires rollout sequencing, backfills, replay compatibility work, or migration planning, escalate to the planning flow: `program-planning`, `program-refining`, or `feature-planning` depending on scope.

## Skill Routing (Always Available)

Workflow reference for non-micro planning and implementation: `@docs/runbooks/planning-workflow.md`

The following local skills are available under `.agents/skills/` and should be loaded on demand for matching work:

- `frontend-design`
- `program-planning`
- `program-refining`
- `feature-planning`
- `feature-implementing`
- `feature-testing`
- `web-design-guidelines`
- `code-reviewer`
- `go-reviewer`
- `python-reviewer`
- `security-reviewer`
- `database-reviewer`

## Durable Planning State

- `plans/STATE.md` is the authoritative durable source of truth for repo planning and execution state.
- Before non-micro planning, refinement, implementation, or feature-testing work, read the relevant parts of `plans/STATE.md` first.
- When initiative, epic seed, active plan, testing, blocker, archive, next-step, or parallelization state changes, update `plans/STATE.md` in the same pass.
- Use initiative docs, epic docs, feature plans, and testing reports for scope, rationale, and evidence; use `plans/STATE.md` for the quick-look answer to what is active, what is next, what is blocked, and what can run in parallel.
- Keep the smallest relevant parent planning doc coherent with the state change, usually the relevant initiative `03-handoff.md` and, when needed, the epic handoff/refinement docs.
- Session execution tasks live only in the OpenCode native task tracker. Do not create durable repo todo/checklist files for session state.

Routing rules:

- Use `frontend-design` for new or updated UI work in `apps/web`, especially dashboards, charts, alert views, and data-heavy screens.
- Use `program-planning` when a large brief contains multiple initiatives, features, workstreams, or rollout concerns that should be decomposed before epic refinement.
- Use `program-refining` when work starts from initiative seeds or a broad initiative slice that still needs to be turned into refined epic context under `plans/epics/`.
- Use `feature-planning` when work starts from refined epic context in `plans/epics/` or when one bounded child feature is already identified and needs an implementation-ready plan under `plans/`.
- Use `feature-implementing` only after a plan exists in `plans/{feature_name}/`.
- Use `feature-testing` after implementation to run smoke, integration, replay, parity, or side-effect checks.
- Do not create initiative, epic, or feature plans whose only deliverable is smoke or integration validation; run that validation directly after implementation, preferably against the real external/API boundary when practical.
- Use `web-design-guidelines` when auditing UI, UX, or accessibility quality against the embedded local guideline set.
- Use `code-reviewer` as the default reviewer after every code-changing micro-implementation and after any implementing skill.
- Add `go-reviewer` when the patch touches Go code or Go-owned live/runtime behavior.
- Add `python-reviewer` when the patch touches Python research, offline analysis, or parity-sensitive Python code.
- Add `security-reviewer` when the patch touches auth, secrets, endpoints, user input, webhooks, external integrations, or other security-sensitive flows.
- Add `database-reviewer` when the patch touches SQL, migrations, persistence layers, query paths, schemas, backfills, or replay-sensitive data handling.
- Mandatory review dispatch: reviewers MUST run in a newly spawned agent with fresh context; never reuse the implementing agent's context and never resume a prior reviewer session for the same pass.
- Reviewer intent inputs MUST make the core intent set explicit: user request, current task contract or active plan step, changed diff or file list, and relevant repo context. Overview/testing artifacts, changed tests/fixtures, and validation commands/results are supporting evidence.
- If the request, plan, diff, and repo context still do not make intent clear, the reviewer agent MUST ask for clarification instead of guessing.

Execution order for non-micro feature work:

1. `program-planning`
2. `program-refining`
3. `feature-planning`
4. `feature-implementing`
5. fresh-context reviewer pass (`code-reviewer` plus any relevant specialist reviewers)
6. `feature-testing`

Approval boundary for that flow:

- `program-refining` and `feature-planning` may be chained in one planning pass when the user is asking to continue planning and has not asked to stop after refinement.
- Stop after `feature-planning` and keep the user in the loop before `feature-implementing` unless the user explicitly asked to continue into implementation.

Skill interaction rules:

- Keep micro-implementing as the default unless escalation is triggered.
- If escalated, follow the skill flow above and keep task state and handoff context updated.
- All planning, implementing, and testing skills must keep `plans/STATE.md` synchronized with the current repo state.
- Smoke-only or integration-only validation is never its own planning slice in this repo; keep it attached to the owning implementation slice or execute it directly via `feature-testing`.
- For initiative-scale briefs, run `program-planning` first, let it decide whether the work should become one initiative or many, write initiative artifacts under `initiatives/`, then use `program-refining` to materialize refined epic context under `plans/epics/` before creating active plans under `plans/`.
- When a refinement pass identifies an obvious next child feature and the user is still in planning mode, continue directly into `feature-planning` without asking for a separate approval step.
- When dependencies allow, use parallel subagents for `program-refining` waves first, then `feature-planning` waves for already-bounded child slices.
- Do not auto-chain from `feature-planning` into `feature-implementing`; pause for explicit user approval before implementation starts.
- When implementing a feature from `plans/{feature_name}/`, read the relevant parts of the parent initiative under `initiatives/` and the relevant program docs under `docs/specs/` before editing.
- For frontend-heavy feature work, combine the active feature skill with `frontend-design`.
- For UI audits, run `web-design-guidelines` using its embedded local rule spec.
- For replay-sensitive or cross-language work, ensure `feature-testing` covers replay determinism and parity where applicable.
- After `feature-implementing`, launch the required reviewer agent(s) with fresh context before considering the slice ready for `feature-testing` or handoff.
- If review findings are clear from the request, plan, diff, and repo context, address them before handoff rather than deferring them.
- Ask for clarification only when review findings expose genuinely missing or conflicting intent that cannot be resolved from the request, plan, diff, and repo context.
- If reviewer feedback results in code changes, rerun the touched validation commands and, when appropriate, rerun the same fresh-context reviewer pass on the updated diff.

---

## Default Execution Mode (Always On): Micro-Implementing

Unless explicitly instructed otherwise, all agents MUST operate in **micro-implementing mode**.

Micro-implementing means assuming the task is a small, scoped change and acting accordingly.

The default behavior is to:

- implement one small, concrete change,
- validate it,
- run the required fresh-context reviewer pass for the change,
- report,
- and stop.

Feature-level planning, orchestration, or multi-step execution is NOT assumed by default.

---

## What Micro-Implementing Means (Default)

By default, the agent works on:

- adding or adjusting a focused Go service change,
- tweaking an existing React/Vite screen or component,
- small refactors,
- bug fixes,
- shared contract or fixture updates,
- local script improvements,
- follow-ups to recently implemented features.

If the task grows beyond a single small change, the agent MUST stop and request escalation.

---

## Default Micro Task Contract

Before editing, the agent must identify a **single micro task**.

If no task exists, create one implicitly.

A micro task must define:

- Outcome: what changes
- Files: files to touch
- Command: validation command
- Accepts: observable done criteria

If more than one logical change is required, the agent MUST split the work and stop after the first task unless instructed to continue.

If a task touches live ingestion, alerts, risk, replay, simulation, backfills, or shared contracts:

- Command MUST include at least one of: targeted integration test, direct API/system validation, parity check, deterministic fixture run, or focused end-to-end check.
- Accepts MUST include idempotency, determinism, or backward-compatibility verification as applicable.

---

## Mandatory Execution Loop

For any non-trivial task:

1. Identify the smallest deliverable that changes system behavior.
2. Express it as a concrete task.
3. Implement immediately.
4. Run a validation command.
5. Launch the required fresh-context reviewer pass with intent artifacts.
6. Record state and stop unless instructed otherwise.

No step is considered complete without either:

- code changes, or
- a runnable validation command.

## Market System Non-Negotiables (Always On)

### Live vs Research Boundary

- Go owns the live production path.
- Python research may inform live logic, but Python cannot be a runtime dependency for live operation.
- Research outputs must move into the live path through reviewed code, contracts, fixtures, or generated artifacts.

### Trust Boundaries

- Do not trust client-computed market state, alerts, outcomes, risk decisions, or derived analytics.
- The service side is the source of truth for canonical events, normalization, features, alerts, outcomes, and risk state.
- Venue adapters ingest raw external payloads; downstream consumers should rely on canonicalized forms.

### Idempotency, Ordering, And Determinism

- Ingestion, replay, backfill, and alert-triggering side effects must be safe to retry.
- Deduplicate external events using stable source identifiers, sequence values, or canonical event IDs when available.
- Preserve a clear distinction between event time and processing time.
- Replay and simulation should produce deterministic results for the same inputs and configuration unless explicitly documented otherwise.

### Contracts And Compatibility

- Shared payload definitions live under `schemas/json/...`.
- Cross-language behavior that must match belongs in `tests/parity` with deterministic inputs in `tests/fixtures`.
- Contract changes require validation in the affected consumers and touched languages.

### Security Minimums

- Verify external webhook or signed payload inputs before mutation.
- Protect public ingestion and control-plane endpoints with appropriate rate limiting or equivalent safeguards.
- Never commit secrets, exchange credentials, or production tokens.
- Operator or admin actions require server-side authorization, never client-only gating.

---

## Task Tracking

All ongoing work must be represented as an explicit task checklist using the OpenCode native task tracker for the current session.

- Keep session task state accurate while you work.
- Record durable cross-session project state in `plans/STATE.md`, not in repo todo files.
- Use `plans/epics/{epic_name}/` for durable refined epic context that is ready to feed `feature-planning`, and `plans/{feature_name}/` for active implementation-ready plans, not execution checklists.
- When pausing, surface blockers, assumptions, the next recommended step, and any needed `plans/STATE.md` updates in the handoff message.

---

## Task Format (Required)

Each task must be executable without interpretation.

- [ ] <task title: concrete outcome>
              - Outcome: what changes for the user or system
              - Files: paths or globs to modify
              - Command: exact command to validate (test, lint, build, API check)
              - Accepts: observable done criteria

Tasks that cannot be validated must be broken down further.

---

## Anti-Stall Rules

To prevent planning loops and analysis paralysis:

- Planning is capped at **5 bullets or 10 lines**, whichever comes first.
- Begin editing files in the same turn a task is identified.
- Ask questions only if blocked from making a safe minimal change.
- If ambiguity is low-risk, choose the smallest backward-compatible assumption and proceed.
- Log assumptions in the current task state or handoff.

Explanations that do not unblock execution are disallowed.

---

## Patch Discipline

- Prefer minimal, reviewable diffs.
- One task means one logical patch.
- No refactors unless required for the current task.
- Defer unrelated improvements.

---

## Validation Discipline

- Every task must specify a validation command.
- Prefer targeted tests.
- If no tests exist, add the smallest direct validation check that exercises the real boundary when practical.
- Validation must be runnable by another agent without context.

---

## Mandatory Fresh-Context Review

- After every micro-implementation that changes code, tests, contracts, queries, migrations, or runtime behavior, spawn a new review agent with fresh context.
- After every implementing skill execution, including `feature-implementing`, spawn a new review agent with fresh context before closing the work or handing off to the next execution phase.
- Use a fresh OpenCode subagent/task session for each review pass; do not reuse the implementing agent and do not continue an earlier reviewer task.
- Always run `code-reviewer`; add `go-reviewer`, `python-reviewer`, `security-reviewer`, and `database-reviewer` when the touched surface matches their domain.
- Reviewer artifact bundle must be explicit and minimal. The core intent set is: user request, plan or task contract, changed file list or diff summary, and relevant repo context such as `00-overview.md` or equivalent design notes. Testing guidance, touched tests/fixtures, and validation commands/results are supporting evidence.
- Intent artifacts are part of implementation hygiene. If they are missing, create or update the smallest appropriate artifact before dispatching review. Examples: focused tests, testing notes, active plan overview, task contract, or direct validation evidence.
- If a reviewer cannot determine intent from the request, plan, diff, and repo context, the reviewer must ask for clarification rather than infer requirements.
- A review pass is not complete until findings are resolved or a blocking clarification request is surfaced in the task state and handoff.
- Implementers must treat clear reviewer findings as required follow-up work, not optional suggestions.
- Only leave a reviewer finding unresolved when the reviewer explicitly asked for clarification or the repo intent remains materially ambiguous after reading the relevant context.

---

## Hand-off Rules

When pausing or handing off work:

- Task state reflects reality; checked tasks are truly done.
- Next task is clearly identified.
- Validation command for the next task is explicit.
- Review status is explicit, including which fresh-context reviewer pass ran, what artifacts it used, and any unresolved findings or clarification requests.
- Any blockers or assumptions are written down in the handoff.

---

## Auto-Escalation Rules

The agent MUST stop and request escalation into the planning flow if any of the following occur:

- More than one app, service, or shared library is involved.
- A contract rollout, migration, replay, or backfill sequence is required.
- Multiple sequential steps are necessary.
- The task expands into workflow or orchestration design.
- More than one micro task is clearly required.

Absent explicit escalation, micro-implementing remains active.

Escalation target defaults:

- use `program-planning` when initiative or program boundaries are still unclear
- use `program-refining` when the work exists only as a broad initiative seed and has not yet been materialized as refined epic context under `plans/epics/`
- use `feature-planning` when one bounded child feature is known but lacks an active implementation-ready plan

---

## Hard Stop Conditions

If any of the following occur, stop and log:

- A task cannot be validated.
- A requirement blocks safe implementation.
- A decision requires user input.

Do not speculate. Do not continue blindly.

---

## Definition of Done

Work is done only when:

- code is implemented,
- validation command passes,
- required fresh-context review has run with adequate intent artifacts,
- no blocking reviewer finding or unanswered reviewer clarification request remains,
- task is checked off,
- durable project state is updated in `plans/STATE.md` and any relevant parent planning doc.

## Frontend Constraints (Vite SPA)

- `apps/web` is a Vite SPA. Do not introduce Next.js, SSR, or server-component assumptions.
- Prefer React + TypeScript patterns that keep domain logic testable and portable.
- Optimize for readable market, chart, alert, and table-heavy interfaces: dense enough to be useful, clear enough to scan quickly.
- Components must be mobile-first:
  - tap targets >= 44px
  - forms and filter panels stay single-column on mobile unless a different pattern is clearly better
  - critical status and alert information stays visible on smaller screens
- Avoid heavy animations or decorative motion that obscures changing market data.

## Performance Budgets (Always On)

- Minimize bundle growth and dependency additions.
- Route-split or lazy-load heavy views and charting code when practical.
- Avoid blocking work on first paint; defer non-essential analytics and debug tooling.
- Prefer platform APIs and small utilities over large framework additions.
- For expensive browser-side calculations, use memoization or workers only when justified.
- Keep Python entirely optional for local SPA and live-service runtime paths.
