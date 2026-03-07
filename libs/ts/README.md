# TypeScript Shared Contracts

- Put canonical contract types or runtime validators in `libs/ts` so `apps/web` and TS tooling reuse one schema interpretation path.
- TS consumers should gate on `schemaVersion` and fixture-backed validation before rendering payloads.
- Keep UI transforms downstream of validated canonical payloads rather than re-deriving contract meaning in components.
