# Binance Runtime Status Endpoint And Ops Handoff

## Ordered Implementation Plan

1. Add one bounded operator-facing runtime-status route that serializes the existing command-owned Binance runtime-health snapshot without changing `/healthz` or `/api/market-state/*`.
2. Wire the live `cmd/market-state-api` provider into that route through a narrow optional runtime-status reader seam and prove the live route works without adding new Binance polling or websocket ownership.
3. Update operator runbooks and `services/market-state-api/README.md` so runtime-status, degraded reasons, and `/healthz` process-health semantics stay explicit and machine-readable.
4. Add focused handler, command, and route smoke coverage, then record validation evidence in `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/testing-report.md` as part of the archived feature history.

## Requirements

- Scope is limited to exposing the already-landed internal runtime-health snapshot through one additive Go-served operator surface plus the matching ops handoff.
- Keep `/healthz` process health only; it must continue to return a simple process-health payload and must not encode runtime freshness.
- Keep `GET /api/market-state/global` and `GET /api/market-state/:symbol` backward-compatible; do not depend on current-state payload changes for runtime-health visibility.
- Keep the tracked symbol set fixed to `BTC-USD` and `ETH-USD`.
- Reuse the shared feed-health vocabulary exactly: `HEALTHY`, `DEGRADED`, `STALE` plus canonical reasons such as `connection-not-ready`, `message-stale`, `snapshot-stale`, `sequence-gap`, `reconnect-loop`, `resync-loop`, `rate-limit`, and `clock-degraded`.
- Keep machine-readable status primary; logs and prose are supporting evidence only.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Public surface choice

- Default to `GET /api/runtime-status` as the additive operator route because it stays clearly separate from `/healthz` and the current-state read APIs.
- Return one response rooted in the existing command snapshot with a stable top-level generation time and deterministic per-symbol ordering.
- Prefer a public response type in `services/market-state-api` that mirrors the settled internal snapshot semantics rather than leaking `cmd/market-state-api` types across the package boundary.

### Provider and handler boundary

- Keep `services/market-state-api.Provider` focused on current-state reads.
- Add a narrow optional runtime-status reader interface that `cmd/market-state-api` can satisfy through the existing `providerWithRuntime` seam.
- If a provider does not support runtime status, return an explicit unsupported response from the new route rather than widening all providers for a Binance-only concern.

### Response content posture

- Preserve deterministic symbol ordering: `BTC-USD` first, `ETH-USD` second.
- Keep readiness explicit alongside feed-health state so warm-up remains distinguishable from degraded-but-readable runtime behavior.
- Carry through the existing settled timestamps, reconnect counters, connection posture, and depth recovery posture from the internal snapshot seam without recomputing new policy inside the handler.
- Keep any future discoverability metadata on `/api/market-state/*` out of scope unless implementation proves a tiny additive seam is unavoidable.

### Ops handoff posture

- Update runbooks to explain when operators should check `/api/runtime-status` versus `/healthz`.
- Reuse the shared vocabulary from `docs/runbooks/ingestion-feed-health-ops.md` and `docs/runbooks/degraded-feed-investigation.md`; do not introduce aliases or endpoint-specific reason names.
- Treat the new endpoint as the primary bounded operator contract for warm-up, reconnect, stale, recovery, and rate-limit posture.

## ASCII Flow

```text
Binance Spot runtime owner in cmd/market-state-api
  - supervisor state
  - depth recovery state
  - read-model readiness
            |
            v
internal runtime-health snapshot
  - generatedAt
  - BTC-USD status
  - ETH-USD status
            |
            v
services/market-state-api
  GET /api/runtime-status
            |
            +--> operator automation / smoke checks
            +--> runbooks and README examples

existing routes remain unchanged
  - GET /healthz
  - GET /api/market-state/global
  - GET /api/market-state/:symbol
```

## Archive Intent

- This feature is complete and archived under `plans/completed/binance-runtime-status-endpoint-and-ops-handoff/`.
