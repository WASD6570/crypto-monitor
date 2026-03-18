# AGENTS

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

## Skill Routing For This Repo

Workflow reference for non-micro planning and implementation: `@docs/runbooks/planning-workflow.md`

Use the available planning, implementation, testing, review, and frontend skills when they match the work.

Project-specific routing rules:

- Use `frontend-design` for new or updated UI work in `apps/web`, especially dashboards, charts, alert views, and data-heavy screens.
- Use `program-planning` when a large brief contains multiple initiatives, features, workstreams, or rollout concerns that should be decomposed before epic refinement.
- Use `program-refining` when work starts from initiative seeds or a broad initiative slice that still needs to be turned into refined epic context under `plans/epics/`.
- Use `feature-planning` when work starts from refined epic context in `plans/epics/` or when one bounded child feature is already identified and needs an implementation-ready plan under `plans/`.
- Use `feature-implementing` only after a plan exists in `plans/{feature_name}/`.
- Use `feature-testing` after implementation to run smoke, integration, replay, parity, or side-effect checks.
- Do not create initiative, epic, or feature plans whose only deliverable is smoke or integration validation; run that validation directly after implementation, preferably against the real external or API boundary when practical.
- Use `web-design-guidelines` when auditing UI, UX, or accessibility quality; apply it together with this repo's frontend constraints.

Project planning flow:

1. `program-planning`
2. `program-refining`
3. `feature-planning`
4. `feature-implementing`
5. fresh-context reviewer pass (`code-reviewer` plus any relevant specialist reviewers)
6. `feature-testing`

Planning-flow specifics:

- `program-refining` and `feature-planning` may be chained in one planning pass when the user is asking to continue planning and has not asked to stop after refinement.
- Stop after `feature-planning` and keep the user in the loop before `feature-implementing` unless the user explicitly asked to continue into implementation.
- For initiative-scale briefs, run `program-planning` first, let it decide whether the work should become one initiative or many, write initiative artifacts under `initiatives/`, then use `program-refining` to materialize refined epic context under `plans/epics/` before creating active plans under `plans/`.
- When a refinement pass identifies an obvious next child feature and the user is still in planning mode, continue directly into `feature-planning` without asking for a separate approval step.
- When dependencies allow, use parallel subagents for `program-refining` waves first, then `feature-planning` waves for already-bounded child slices.
- Do not auto-chain from `feature-planning` into `feature-implementing`; pause for explicit user approval before implementation starts.
- When implementing a feature from `plans/{feature_name}/`, read the relevant parts of the parent initiative under `initiatives/` and the relevant program docs under `docs/specs/` before editing.
- For frontend-heavy feature work, combine the active feature skill with `frontend-design`.
- For UI audits, run `web-design-guidelines` and apply it together with this repo's frontend constraints.
- For replay-sensitive or cross-language work, ensure `feature-testing` covers replay determinism and parity where applicable.

## Durable Planning State

- `plans/STATE.md` is the authoritative durable source of truth for repo planning and execution state.
- Before non-micro planning, refinement, implementation, or feature-testing work, read the relevant parts of `plans/STATE.md` first.
- When initiative, epic seed, active plan, testing, blocker, archive, next-step, or parallelization state changes, update `plans/STATE.md` in the same pass.
- Use initiative docs, epic docs, feature plans, and testing reports for scope, rationale, and evidence; use `plans/STATE.md` for the quick-look answer to what is active, what is next, what is blocked, and what can run in parallel.
- Keep the smallest relevant parent planning doc coherent with the state change, usually the relevant initiative `03-handoff.md` and, when needed, the epic handoff or refinement docs.

## Market System Non-Negotiables

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
