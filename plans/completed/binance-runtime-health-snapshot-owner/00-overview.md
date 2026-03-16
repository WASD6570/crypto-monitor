# Binance Runtime Health Snapshot Owner

## Ordered Implementation Plan

1. Define one command-owned runtime-health snapshot contract that merges Spot supervisor health, depth recovery posture, and read-model readiness into stable per-symbol operator status for `BTC-USD` and `ETH-USD`.
2. Implement the snapshot owner and lifecycle wiring in `cmd/market-state-api` so warm-up, reconnect, stale, recovery, and rate-limit transitions become deterministically readable without changing `/healthz` or `/api/market-state/*`.
3. Add focused tests for startup, degradation, recovery, and repeated-input stability, then write `plans/binance-runtime-health-snapshot-owner/testing-report.md` before moving the directory to `plans/completed/binance-runtime-health-snapshot-owner/` after implementation and validation finish.

## Requirements

- Scope is limited to the internal runtime-health snapshot seam that the later `binance-runtime-status-endpoint-and-ops-handoff` feature will expose; do not add the public status endpoint in this slice.
- Keep `/healthz` process-only and keep the existing `GET /api/market-state/global` plus `GET /api/market-state/:symbol` contracts unchanged.
- Reuse the completed Spot supervisor, depth recovery, and read-model owner surfaces; do not create a second websocket lifecycle owner or duplicate feed-health policy.
- Keep the tracked symbol set fixed to `BTC-USD` and `ETH-USD`.
- Preserve the shared health vocabulary: `HEALTHY`, `DEGRADED`, `STALE` plus canonical reasons such as `connection-not-ready`, `message-stale`, `snapshot-stale`, `reconnect-loop`, `resync-loop`, `rate-limit`, and `clock-degraded`.
- Keep machine-readable runtime status primary; logs may support debugging but must not become the only source of operator truth.
- Keep repeated accepted runtime inputs deterministic so later endpoint exposure can rely on stable symbol ordering and stable field semantics.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Parent context to preserve

- `initiatives/crypto-market-copilot-binance-integration-completion/00-overview.md` defines this slice as the first Wave 2 step after the sustained Spot runtime cutover.
- `plans/epics/binance-runtime-health-and-operator-observability/00-overview.md` defines the parent epic and the additive operator-surface constraint.
- `plans/completed/binance-spot-runtime-read-model-owner/00-overview.md` and `plans/completed/binance-market-state-live-reader-cutover/00-overview.md` define the sustained command-owned runtime and the stable API contract this feature must preserve.
- `docs/runbooks/ingestion-feed-health-ops.md` and `docs/runbooks/degraded-feed-investigation.md` define the vocabulary and operator meaning this snapshot must reuse.

### Planned snapshot boundary

- Keep ownership in `cmd/market-state-api` because that process already owns runtime startup, shutdown, and provider lifecycle.
- Introduce one snapshot owner that can read the latest status for all tracked symbols without hitting Binance or depending on handler logic.
- Model per-symbol status from existing runtime inputs rather than from log scraping or ad hoc string synthesis.
- Capture both readiness and degradation explicitly so operators can distinguish `not yet publishable`, `healthy`, `degraded`, and `stale` conditions.

### Snapshot content expectations

- Keep symbol ordering deterministic: always emit `BTC-USD` then `ETH-USD`.
- For each symbol, carry the latest available feed-health state and reasons, connection state, depth status, read-model readiness, and the most relevant timestamps/counters already available from the completed runtime surfaces.
- Keep counter and timestamp semantics explicit: if a metric is unavailable at this layer, prefer omission or a neutral zero-value rule that is stable and testable rather than ambiguous inferred values.
- Keep the snapshot internal to the command for this feature; a later feature may serialize it for operator APIs.

### Live vs research boundary

- All snapshot assembly, runtime reads, lifecycle wiring, and validation stay in Go under `cmd/market-state-api`, `services/venue-binance`, and Go tests.
- No Python runtime, notebook artifact, or offline analysis dependency enters the live status path.

## ASCII Flow

```text
binance spot runtime pieces
  - websocket supervisor state
  - depth recovery state
  - read-model readiness
            |
            v
cmd/market-state-api
  runtime health snapshot owner
  - gather per-symbol status
  - preserve shared health vocabulary
  - expose deterministic snapshot reads
            |
            +--> later status endpoint feature
            |
            +--> operator tests / runbooks

existing routes remain unchanged
  - GET /healthz
  - GET /api/market-state/global
  - GET /api/market-state/:symbol
```

## Archive Intent

- Keep this feature active under `plans/binance-runtime-health-snapshot-owner/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/binance-runtime-health-snapshot-owner/`.
