# Binance Spot Depth Resync And Snapshot Health

## Ordered Implementation Plan

1. Add a bounded Binance Spot depth recovery owner in `services/venue-binance` that reacts to sequence gaps or bootstrap failures by entering explicit resync state, consulting existing cooldown and rate-limit primitives, and requesting the next eligible snapshot without re-owning websocket lifecycle behavior.
2. Wire snapshot refresh cadence, snapshot staleness, and repeated resync tracking through the existing Binance runtime and loop-state path so degraded depth health remains machine-readable and explicit for BTC/ETH Spot symbols.
3. Add deterministic recovery fixtures and integration coverage that prove gap-triggered resync, cooldown/rate-limit blocking, refresh-driven snapshot replacement, and stale snapshot degradation before direct live validation.
4. Run attached direct Binance validation for the implemented recovery path, write `plans/binance-spot-depth-resync-and-snapshot-health/testing-report.md`, then move the full directory to `plans/completed/` after implementation and validation finish.

## Requirements

- Scope is limited to the post-bootstrap depth recovery path for Binance Spot: gap-triggered resync, bootstrap retry posture, snapshot refresh cadence, snapshot staleness visibility, bounded resync-loop tracking, and machine-readable feed-health output.
- Inherit the completed startup semantics from `plans/completed/binance-spot-depth-bootstrap-and-buffering/`; this feature must treat its synchronized snapshot and bridged delta rules as settled behavior.
- Reuse existing runtime primitives in `services/venue-binance/runtime.go` for snapshot cooldown, per-minute snapshot recovery rate limiting, snapshot freshness, reconnect-loop evaluation, and resync-loop thresholds rather than inventing a second health model.
- Preserve the completed Spot lifecycle owner from `plans/completed/binance-spot-ws-runtime-supervisor/`; recovery must not create a second connection manager or duplicate subscribe/reconnect logic.
- Preserve canonical contract rules: asset-centric symbols remain `BTC-USD` and `ETH-USD`, while `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs` stay explicit on emitted order-book and feed-health outputs.
- Keep sequence gaps, stale snapshots, cooldown blocks, rate-limit blocks, resync loops, and connection-not-ready posture machine-visible instead of silently repairing depth state.
- Keep Go as the live runtime path; Python remains offline-only.

## Design Notes

### Repository state to preserve

- `plans/completed/binance-spot-depth-bootstrap-and-buffering/` already defines the startup owner, one-shot snapshot alignment, and bridging delta acceptance window.
- `services/venue-binance/runtime.go` already exposes the recovery primitives this feature should compose: `SnapshotRecoveryStatus(...)`, `SnapshotRecoveryRateLimitStatus(...)`, `EvaluateStaleness(...)`, `EvaluateResyncLoop(...)`, and `EvaluateLoopState(...)`.
- `tests/fixtures/events/binance/BTC-USD/edge-sequence-gap-usdt.fixture.v1.json` already encodes the canonical degraded feed-health outcome for a sequence-gap path and should be reused where practical.
- `plans/completed/market-ingestion-and-feed-health/00-overview.md` already fixes the shared feed-health expectations: connection state, message freshness, snapshot freshness, gap visibility, bounded resync logic, and clock degradation stay machine-readable.

### Recovery owner boundary

- Introduce one depth recovery owner beside the completed bootstrap owner in `services/venue-binance`.
- The owner should accept already-synchronized depth state, detect when synchronization is lost, decide whether an immediate snapshot retry is allowed, and surface an explicit blocked or resyncing state when cooldown or rate-limit rules prevent recovery.
- The owner should reset depth sequencing only when a new snapshot actually replaces state; it should never silently continue after a detected gap.

### Snapshot refresh and stale posture

- Treat snapshot refresh cadence separately from gap-triggered resync: periodic refresh exists to keep depth state trustworthy even when no explicit gap is observed.
- Use the existing `snapshotRefreshPolicy` and `snapshotStaleAfterMs` defaults as the control plane for refresh and stale degradation.
- A refresh attempt blocked by cooldown or rate limit should remain visible in feed health so operators can distinguish "healthy but pending refresh" from "degraded and cannot recover yet".

### Feed-health handoff

- Recovery state should end in canonical `feed-health` output through the existing `services/normalizer` path rather than venue-only logs.
- Degradation reasons should stay explicit and composable: `sequence-gap`, `snapshot-stale`, `resync-loop`, `reconnect-loop`, `connection-not-ready`, and clock degradation should remain distinguishable.
- Stale precedence should remain intact: if snapshot freshness or message freshness crosses the stale boundary, the final status should stay `STALE` even when other degraded reasons are also present.

### Live vs research boundary

- Recovery policy, snapshot scheduling, and feed-health emission all stay in Go under `services/venue-binance` plus `services/normalizer` and `libs/go/ingestion`.
- Offline research may inspect fixture traces later, but it must not become a dependency of the live recovery path.

## ASCII Flow

```text
synchronized depth state
  - snapshot + accepted deltas
          |
          v
services/venue-binance depth recovery owner
  - detect gap or refresh due
  - consult cooldown / rate limit
  - request replacement snapshot when eligible
  - mark blocked or resyncing when not eligible
          |
          +------------------------------+
          |                              |
          v                              v
replacement snapshot + aligned deltas    explicit recovery status
          |                              |
          v                              v
shared sequencer reset/restart      services/normalizer feed-health output
          |                              |
          +---------------+--------------+
                          v
                 canonical depth + health events
```
