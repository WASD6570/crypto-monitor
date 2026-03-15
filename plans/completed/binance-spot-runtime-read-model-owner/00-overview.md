# Binance Spot Runtime Read Model Owner

## Ordered Implementation Plan

1. Add one sustained Spot runtime owner in `cmd/market-state-api` that composes the completed Binance websocket supervisor and depth recovery surfaces into process-owned per-symbol state for `BTC-USD` and `ETH-USD`.
2. Expose one explicit read-model snapshot seam that satisfies `marketstateapi.SpotCurrentStateReader` without reopening the stable `services/market-state-api` provider contract.
3. Preserve honest startup, reconnect, resync, and stale behavior by carrying machine-readable feed-health and depth recovery status through the read model instead of hiding temporary runtime gaps.
4. Add deterministic runtime and integration proof for startup, accepted frame progression, gap-triggered resync, reconnect carry-forward, and repeated-input stability.
5. Run the attached validation matrix, write `plans/binance-spot-runtime-read-model-owner/testing-report.md`, then move the full directory to `plans/completed/binance-spot-runtime-read-model-owner/` after implementation and validation finish.

## Requirements

- Scope is limited to the sustained Spot runtime owner and read-model seam that later cutover work will consume.
- Keep `services/market-state-api` on the existing `marketstateapi.SpotCurrentStateReader` contract and current-state response shape.
- Reuse the completed Spot websocket supervisor, bootstrap, and depth recovery behavior; do not create a second websocket lifecycle owner or a second depth policy implementation.
- Keep the supported symbol set fixed to `BTC-USD` and `ETH-USD` for this feature.
- Preserve machine-readable warm-up, degraded, stale, partial, and unavailable behavior while the runtime is connecting, bootstrapping, resyncing, refreshing snapshots, or disconnected.
- Keep replay-sensitive assumptions explicit enough that repeated accepted-input sequences can still prove deterministic read-model output ordering.
- Keep Go as the live runtime path; Python remains offline-only.
- Leave provider cutover, handler changes, browser smoke, and broader operator observability to the later `binance-market-state-live-reader-cutover` slice.

## Design Notes

### Parent context to preserve

- `plans/epics/binance-streaming-market-state-runtime-integration/` defines this slice as the runtime plumbing seam that replaces the temporary command-local polling reader.
- `plans/completed/binance-live-market-state-api-provider-cutover/` already moved the command to a live-backed provider, but its `cmd/market-state-api/live_provider.go` reader still polls `/api/v3/depth` on demand instead of owning a sustained runtime.
- `plans/completed/binance-spot-ws-runtime-supervisor/`, `plans/completed/binance-spot-depth-bootstrap-and-buffering/`, and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` already settled connection, bootstrap, and recovery semantics that this feature must compose rather than redesign.
- `services/market-state-api/live_spot_provider.go` already defines the consumer seam as a read-model snapshot of `SpotCurrentStateObservation` values with feed-health and depth status attached.

### Planned runtime boundary

- Keep runtime ownership in `cmd/market-state-api` because this process is the direct owner of provider lifecycle and shutdown.
- Introduce one command-owned runtime component that owns startup, shutdown, symbol state, and snapshot reads for the provider.
- Keep `services/venue-binance` responsible for venue-native runtime logic and parsing helpers; only add shared helpers there when the sustained owner cannot be built cleanly from existing exported seams.
- Treat the read model as the latest accepted current-state observations plus explicit per-symbol readiness and degradation status, not as a new canonical event store.

### Read-model posture

- The runtime owner should maintain deterministic symbol ordering and snapshot reads so `BTC-USD` and `ETH-USD` responses remain stable across repeated accepted inputs.
- A symbol may remain absent from the snapshot until the runtime has enough accepted data to publish a trustworthy observation.
- When an observation exists but depth recovery is degraded or stale, the read model should continue returning that observation with the current `FeedHealth` and `DepthStatus` attached.
- Warm-up and reconnect behavior must stay honest: no fallback to ad hoc REST polling once the sustained owner exists.

### Live vs research boundary

- All runtime orchestration, read-model state, websocket handling, and validation hooks stay in Go under `cmd/market-state-api`, `services/venue-binance`, and Go tests.
- No Python runtime, notebook artifact, or offline research dependency is introduced for live current-state serving.

## ASCII Flow

```text
market-state-api process
  |
  +--> sustained spot runtime owner
  |      - start supervisor
  |      - drive depth bootstrap/recovery
  |      - maintain BTC/ETH read model
  |      - expose shutdown + snapshot seam
  |
  +--> marketstateapi.NewLiveSpotProvider(...)
             |
             v
     current-state handlers
       - same routes
       - same response contract
       - honest partial/unavailable startup

accepted spot runtime inputs
  - trade/bookTicker/depth
  - reconnect and resync state
             |
             v
read-model observations
  - symbol
  - best bid/ask
  - exchangeTs / recvTs
  - feed health
  - depth status
```

## Archive Intent

- Keep this feature active under `plans/binance-spot-runtime-read-model-owner/` while implementation and validation are in progress.
- When complete, move the directory and `testing-report.md` to `plans/completed/binance-spot-runtime-read-model-owner/`.
