# Binance Market State Live Reader Cutover

## Ordered Implementation Plan

1. Replace the temporary command-local polling seam with the sustained Spot runtime read model so `cmd/market-state-api` serves current-state responses from one process-owned source of truth.
2. Preserve the existing `services/market-state-api` provider contract and route shapes while proving startup, partial, degraded, and unavailable behavior still remain machine-readable through the real API path.
3. Refresh local compose/browser validation and operator-facing docs so the dashboard smoke checks the real same-origin `/api/market-state/*` path without value-pinning live Binance output.
4. Run the attached validation matrix, write `plans/binance-market-state-live-reader-cutover/testing-report.md`, then move the full directory to `plans/completed/binance-market-state-live-reader-cutover/` after implementation and validation finish.

## Requirements

- Scope is limited to consumer-facing cutover work that uses the sustained Spot runtime owner from `binance-spot-runtime-read-model-owner`; do not reopen websocket, depth recovery, or venue-runtime design.
- Keep `services/market-state-api` on the existing `marketstateapi.Provider` and `marketstateapi.SpotCurrentStateReader` contracts plus the existing routes: `GET /healthz`, `GET /api/market-state/global`, and `GET /api/market-state/:symbol`.
- Keep `BTC-USD` and `ETH-USD` as the only supported symbols for this slice.
- Preserve honest startup posture: before publishable observations exist, symbol/global responses may remain partial or unavailable rather than fabricating values.
- Preserve machine-readable degraded behavior across reconnect, stale, resync, and partial-read paths; `/healthz` stays process health only.
- Keep `apps/web` read-only and same-origin; do not move market-state derivation into the browser or add direct Binance access there.
- Keep validation contract-oriented rather than value-pinned: browser smoke should prove symbols, navigation, route wiring, and availability/degradation handling instead of exact live labels.
- Keep Go as the live runtime path; Python remains offline-only.
- Carry forward the current prerequisite risk: the runtime-owner slice still has a failing full regression in its active testing report, so this feature should not be considered implementation-ready to merge until that baseline is green.

## Design Notes

### Parent context to preserve

- `initiatives/crypto-market-copilot-binance-integration-completion/00-overview.md` defines this slice as the final Wave 1 consumer-facing step before operator observability and USD-M semantics.
- `plans/epics/binance-streaming-market-state-runtime-integration/91-child-plan-seeds.md` defines this feature as the API-contract-preserving cutover that follows the sustained runtime owner.
- `plans/binance-spot-runtime-read-model-owner/00-overview.md` defines the read-model seam this cutover must consume rather than redesign.
- `plans/completed/binance-live-market-state-api-provider-cutover/00-overview.md` is the earlier local-first cutover that should be treated as superseded temporary context, not reopened scope.

### Planned cutover boundary

- Keep provider ownership in `cmd/market-state-api`; this process owns runtime startup, shutdown, and the live reader wiring.
- Use the command-owned runtime owner as the single reader beneath `marketstateapi.NewLiveSpotProvider(...)`.
- Remove any remaining per-request polling or temporary bootstrap assumptions from the command/provider path once the runtime owner is the source.
- Treat the dashboard-visible change as proof that the same-origin API path can eventually render live shell content after runtime warm-up, not as a frontend redesign project.

### API and browser posture

- `services/market-state-api` stays the browser-facing contract boundary; any change here must be additive and backward-compatible.
- `/healthz` remains a process/readiness signal only; warm-up and degradation continue to show up in JSON payloads and the dashboard fallback/shell states.
- Browser smoke should accept that live runtime output changes over time; assert presence of both tracked symbols, section navigation, same-origin loading, and readable state transitions instead of exact tradeability text.
- If the dashboard still legitimately starts in the unavailable state, the smoke should allow that initial state but verify it can recover into the shell once current state becomes readable.

### Docs and operator handoff posture

- Update the local docs that still imply a temporary cutover or deterministic assumptions.
- Keep broader runtime health/status surfaces for `binance-runtime-health-and-operator-observability`; this feature only documents the live-reader cutover and the expected warm-up/degraded behavior users see today.

### Live vs research boundary

- All runtime ownership, API serving, integration validation, and compose/browser proof stay in Go plus the existing SPA consumer.
- No Python runtime, notebook, or offline artifact becomes a dependency of the live current-state path.

## ASCII Flow

```text
browser (/dashboard)
  |
  +--> same-origin /api/market-state/*
          |
          v
   cmd/market-state-api
     - start sustained Spot runtime owner
     - build live provider from read-model snapshot seam
     - serve stable handlers
          |
          v
   services/market-state-api
     - unchanged routes/contracts
     - global + symbol assembly
          |
          v
   dashboard states
     - initial unavailable or partial while warming
     - live shell once observations publish
     - machine-readable degradation on reconnect/stale/resync
```

## Archive Intent

- Keep this feature active under `plans/binance-market-state-live-reader-cutover/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/binance-market-state-live-reader-cutover/`.
