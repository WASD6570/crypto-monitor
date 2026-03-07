# Go Shared Contracts

- Put canonical contract decoding and version guards in `libs/go` so live services share one interpretation path.
- New Go consumers should be fixture-backed and reject unsupported `schemaVersion` values before business logic runs.
- Keep Go as the live-path consumer of canonical market state; do not route live validation through Python.
