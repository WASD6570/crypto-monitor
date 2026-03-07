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
