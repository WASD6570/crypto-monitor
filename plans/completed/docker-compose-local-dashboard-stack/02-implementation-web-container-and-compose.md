# Implementation: Web Container And Compose

## Requirements And Scope

- Add a Docker image for `apps/web` that installs dependencies, builds the SPA, and serves it on a stable port.
- Add a repo-root Compose service for the current dashboard stack.
- Configure the container to enable the mock API seam by default for local Compose startup.
- Keep the stack intentionally small and transparent.

## Target Repo Areas

- `apps/web/Dockerfile`
- `docker-compose.yml`

## Implementation Notes

- Use the existing app-local `pnpm-lock.yaml` and `package.json` as the container dependency boundary.
- Serve the built app through `vite preview` so Compose exercises the production bundle rather than a dev-only server path.
- Expose a single host port and keep configuration readable from the compose file.

## Unit / Smoke Expectations

- `docker compose config` succeeds.
- `docker compose up --build -d` starts the web service.
- `curl http://127.0.0.1:4173/dashboard` returns HTML and `curl http://127.0.0.1:4173/api/market-state/global` returns JSON.

## Summary

This step creates one reproducible startup path for the current dashboard and its mock API without pretending that the rest of the live stack already exists.
