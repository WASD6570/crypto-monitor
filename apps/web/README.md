# Web App

Home for the TypeScript + React + Vite single-page application.

Planned structure:

- `public/`: static assets
- `src/app`: app shell and composition
- `src/pages`: top-level screens/routes
- `src/components`: reusable UI pieces
- `src/features`: domain-focused frontend feature folders
- `src/api`: API client layer
- `src/state`: app state management
- `src/hooks`: shared React hooks
- `src/utils`: general utilities

Current implemented slice:

- fixture-backed `dashboard-shell-and-summary-strip` route at `/dashboard`
- Vite + React + TypeScript app scaffold
- unit tests and Playwright smoke spec for the dashboard shell

Live query adapters, detailed market panels, and replay-aware UI work remain planned follow-on slices.

## Docker Compose

- From the repo root, run `docker compose up --build`.
- Visit `http://127.0.0.1:4173/dashboard`.
- The web container serves the built SPA behind Nginx and proxies `/api` to the Go `market-state-api` service.
- For local frontend development outside Compose, run the Go API separately and let Vite proxy `/api` via `VITE_API_PROXY_TARGET` (defaults to `http://127.0.0.1:8080`).
- For the isolated dev-overlay workflow, run `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build`; the `web` service switches to Vite with HMR but keeps the same-origin `/api` boundary by proxying to `http://market-state-api:8080` inside Compose.
- The dev overlay does not add runtime mock routes or fixture-backed API behavior; it still reads live Go-owned market-state responses.
