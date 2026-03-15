# Epic Plans

This folder holds refined epic context that is ready to feed `feature-planning`, but is still too broad to hand directly to `feature-implementing`.

- Epics in this folder are already the output of `program-refining`.
- `feature-planning` consumes these refined epics to create bounded child feature plans under `plans/{feature_name}/`.
- Refined, implementation-ready child plans belong in `plans/{feature_name}/`.
- Finished plans belong in `plans/completed/{feature_name}/`.

Do not treat this folder listing as an actionable queue. Use `plans/STATE.md` to see which refined epics are still `ready_to_plan`, which are blocked, and which are only retained as historical context.
