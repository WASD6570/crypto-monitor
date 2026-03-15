# Refinement Map

## Already Done

- `plans/completed/binance-spot-ws-runtime-supervisor/` settled the Spot websocket lifecycle owner, bounded reconnect posture, and machine-readable supervisor state
- `plans/completed/binance-spot-depth-bootstrap-and-buffering/` settled buffered delta plus snapshot alignment for first trustworthy depth sync
- `plans/completed/binance-spot-depth-resync-and-snapshot-health/` settled explicit recovery, refresh, stale, cooldown, and rate-limit depth semantics
- `plans/completed/binance-live-current-state-query-assembly/` settled the Spot-driven current-state assembly seam and preserved the `services/market-state-api` response contract
- `plans/completed/binance-live-market-state-api-provider-cutover/` cut the command entrypoint over to a live-backed provider, but only through a bounded command-local snapshot polling reader

## Remaining Work

- replace the command-local polling reader in `cmd/market-state-api/live_provider.go` with a sustained runtime owner that continuously consumes the completed Spot supervisor plus depth recovery surfaces
- introduce one explicit process-owned read model that stores the latest accepted Spot observations for `BTC-USD` and `ETH-USD` and can satisfy `marketstateapi.SpotCurrentStateReader`
- preserve warm-up honesty and machine-readable degraded states while the sustained runtime is connecting, bootstrapping, resyncing, stale, or temporarily disconnected
- prove the new runtime path through targeted Go tests plus focused integration checks that hit the existing API boundary

## Overlap And Non-Goals

- do not reopen websocket subscription, ping/pong, rollover, or reconnect semantics already fixed by the completed supervisor work
- do not redesign depth bootstrap or recovery policy; compose the completed owners and carry their state through the new runtime loop
- do not change `/api/market-state/global`, `/api/market-state/:symbol`, or `/healthz` contracts in this epic
- do not bundle operator-facing status expansion, USD-M semantics, environment rollout defaults, or long-run soak validation; those belong to later epics in the initiative queue

## Refinement Waves

### Wave 1

- `binance-spot-runtime-read-model-owner`
- Why first: the repo still lacks the sustained runtime loop that turns completed venue-runtime pieces into a process-owned state source

### Wave 2

- `binance-market-state-live-reader-cutover`
- Why later: provider cutover should only happen after the read model exposes stable warm-up and degradation behavior to the command path

### Direct Post-Implementation Checks

- same-origin API smoke against the command-backed `/api/market-state/*` path
- repeated deterministic fixture/integration proof for identical accepted Spot input sequences
- these stay attached to the owning implementation slices rather than becoming standalone child plans
