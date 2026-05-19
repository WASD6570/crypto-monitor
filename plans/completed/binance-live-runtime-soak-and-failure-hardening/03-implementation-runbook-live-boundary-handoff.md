# Implementation: Runbook And Live Boundary Handoff

## Requirements And Scope

- Document the required local deterministic runtime-soak checks.
- Document optional public Binance live validation and Compose proof without making either required in environments that cannot run them.
- Keep docs aligned with the existing operator vocabulary and route ownership.
- Do not introduce secrets, private endpoints, browser-side Binance logic, or Python live-runtime dependencies.

## Target Files

- `docs/runbooks/binance-runtime-soak-and-failure-check.md`
- `docs/runbooks/binance-compose-rollout.md`
- `docs/runbooks/ingestion-feed-health-ops.md`
- `docs/runbooks/degraded-feed-investigation.md`
- `services/market-state-api/README.md`
- `README.md` only if implementation adds or changes a top-level command

## Runbook Content

- Required local validation command sequence from `04-testing.md`.
- How to read `/api/runtime-status` after the additive `usdmStatus` lands.
- The distinction between warm-up, readable degradation, stale state, and process health.
- What operators should record for Spot depth recovery, USD-M websocket health, and USD-M open-interest health.
- Optional live-boundary command: `BINANCE_LIVE_VALIDATION=1 go test ./tests/integration -run TestIngestionBinanceSpotDepthLive`.
- Optional Compose command: `make compose-smoke` when Docker is available.
- Explicit note that optional live and Compose checks create no exchange mutations and require no credentials.

## Documentation Rules

- Reuse canonical feed-health names exactly: `HEALTHY`, `DEGRADED`, `STALE`.
- Reuse canonical reasons exactly: `connection-not-ready`, `message-stale`, `snapshot-stale`, `sequence-gap`, `reconnect-loop`, `resync-loop`, `rate-limit`, `clock-degraded`.
- Keep `/healthz` described as process health only.
- Keep `/api/market-state/*` described as consumer current-state reads.
- Keep `/api/runtime-status` described as the bounded operator status route for fixed `BTC-USD` and `ETH-USD`.

## Validation Notes

- If docs mention a command, the command must be present in `04-testing.md` or explicitly marked optional.
- If Docker remains unavailable in the implementation environment, record that in `testing-report.md` rather than weakening the required deterministic local matrix.
- If public network access is unavailable, record skipped live-boundary evidence and keep the local deterministic proof as the required acceptance gate.

## Summary For Next Agent

Add one focused runtime-soak runbook and update existing route/runbook docs only where needed to include `usdmStatus` and the exact required versus optional validation boundaries.
